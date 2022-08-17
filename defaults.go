package jsonapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
)

var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
var nilValue = reflect.ValueOf(nil)

// Defaults implements all the interfaces of this package in a default way.
//
// You can use this to compose custom behaviour on top.
var Defaults = &defaults{}

type defaults struct {
	LogDomainErrors bool
}

func (d *defaults) Resolve(req *http.Request, t reflect.Type, pos int) (reflect.Value, error) {
	if t.Implements(contextType) {
		val := reflect.ValueOf(req.Context())

		if !val.Type().AssignableTo(t) {
			return val, fmt.Errorf("%w: context of type %v is not assignable to argument in pos #%d (%v)", ErrArgumentResolution, val.Type(), pos, t)
		}

		return val, nil
	}

	if !isStructWithJson(t) {
		return nilValue, fmt.Errorf("%w: argument #%d (%v)", ErrArgumentUnsupported, pos, t)
	}

	ptr := false

	if t.Kind() == reflect.Ptr {
		ptr = true
		t = t.Elem()
	}

	v := reflect.New(t).Interface()

	err := json.NewDecoder(req.Body).Decode(&v)
	if err == io.EOF {
		return nilValue, ErrEmptyBody
	}

	if err != nil {
		return nilValue, fmt.Errorf("%w: %s", ErrArgumentResolution, err.Error())
	}

	if ptr {
		return reflect.Indirect(reflect.ValueOf(&v).Elem()).Elem(), nil
	}

	return reflect.Indirect(reflect.ValueOf(v).Elem()), nil
}

func (d *defaults) Validate(_ *http.Request) ([]*ErrorItem, error) {
	return nil, nil
}

func (d *defaults) CastError(v interface{}) error {
	apiErr := ApiError{
		StatusCode: http.StatusInternalServerError,
		Kind:       "Unknown",
		Details:    "Request failed with unknown error",
	}

	if rep, ok := v.(interface {
		Status() int
	}); ok {
		apiErr.StatusCode = rep.Status()
	}

	switch t := v.(type) {
	case string:
		apiErr.sourceErr = errors.New(t)
	case error:
		apiErr.sourceErr = t

		// If we are dealing with a wrapped error, we assume that the wrapped error is the source
		if wrapper, ok := t.(interface {
			Unwrap() error
		}); ok {
			apiErr.Kind = "Domain Error"
			apiErr.Details = t.Error()
			apiErr.sourceErr = wrapper.Unwrap()
		}

		if errors.Is(t, ErrUnexpected) {
			apiErr.StatusCode = http.StatusInternalServerError
			apiErr.Kind = "Handler Error"
			apiErr.Details = "Unexpected error while handling the request"
		}

		if errors.Is(t, ErrEmptyBody) {
			apiErr.StatusCode = http.StatusBadRequest
			apiErr.Kind = "Invalid Request"
			apiErr.Details = "Request body cannot be empty"
			apiErr.sourceErr = nil
		}

		if errors.Is(t, ErrArgumentUnsupported) || errors.Is(t, ErrArgumentResolution) {
			apiErr.StatusCode = http.StatusInternalServerError
			apiErr.Kind = "Handler Error"
			apiErr.Details = "Could not resolve handler arguments"
		}
	case []*ErrorItem:
		apiErr.StatusCode = http.StatusBadRequest
		apiErr.Kind = "Invalid Request"
		apiErr.Details = "Request body validation has failed"
		apiErr.Errors = t
	case *ApiError:
		return t
	default:
		// Noop
	}

	return &apiErr
}

// LogError logs only wrapped errors
func (d *defaults) LogError(_ *http.Request, err error) {
	if wrapped, ok := err.(interface {
		Unwrap() error
	}); ok {
		log.Println(wrapped.Unwrap().Error())
	}
}

func (d *defaults) SendResponse(w http.ResponseWriter, v interface{}) {
	if v == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	status := http.StatusOK

	if rep, ok := v.(interface {
		Status() int
	}); ok {
		status = rep.Status()
	}

	w.Header().Add("Content-Type", "application/json")

	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		panic(err)
	}
}

func isStructWithJson(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if _, ok := f.Tag.Lookup("json"); ok {
			return true
		}
	}

	return false
}
