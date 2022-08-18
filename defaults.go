package jsonapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
