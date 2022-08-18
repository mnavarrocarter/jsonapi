package jsonapi

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorHandlerFunc = func(w http.ResponseWriter, req *http.Request, err error)

var ErrEmptyBody = errors.New("request body is empty")
var ErrArgumentResolution = errors.New("argument resolution error")
var ErrArgumentUnsupported = errors.New("argument resolution unsupported")

var HandleError ErrorHandlerFunc = handleError

func handleError(w http.ResponseWriter, req *http.Request, err error) {
	SendResponse(w, req, err)
}

type apiError struct {
	code int
	msg  string
	prev error
}

type Wrapper interface {
	Unwrap() error
}

type Coder interface {
	Code() int
}

func (e *apiError) Unwrap() error {
	return e.prev
}

func (e *apiError) Error() string {
	return e.msg
}

func (e *apiError) Code() int {
	return e.code
}

func panicToError(v any) (err error) {
	switch t := v.(type) {
	case string:
		err = errors.New(t)
	case error:
		err = t
	default:
		err = errors.New(fmt.Sprintf("%v", v))
	}

	return &apiError{
		code: http.StatusInternalServerError,
		msg:  "An unexpected error has occurred",
		prev: err,
	}
}
