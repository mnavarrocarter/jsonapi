package jsonapi

import (
	"errors"
	"net/http"
	"reflect"
)

var ErrUnexpected = errors.New("unexpected error")
var ErrEmptyBody = errors.New("request body cannot be empty")
var ErrArgumentUnsupported = errors.New("argument resolution unsupported")
var ErrArgumentResolution = errors.New("argument resolution error")

type OptsFn func(h *JsonHandler)

// Wrap makes a JsonHandler using the Defaults
//
// See JsonHandler for documentation on how this handler works.
//
// Also, see Defaults to study the default implementations of the different components.
func Wrap(fn any, opts ...OptsFn) *JsonHandler {
	h := &JsonHandler{
		fn:               reflectFunc(fn),
		RequestValidator: Defaults,
		ArgumentResolver: Defaults,
		ErrorCaster:      Defaults,
		ErrorLogger:      Defaults,
		ResponseSender:   Defaults,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// An ErrorCaster casts any type to an error
//
// You can compose this interface to control how you want your errors to be serialized
type ErrorCaster interface {
	CastError(v interface{}) error
}

// A ResponseSender writes v to w
type ResponseSender interface {
	SendResponse(w http.ResponseWriter, v interface{})
}

type RequestValidator interface {
	// Validate validates the request
	//
	// If the body of the request is read, then the validator needs to restore
	// the request body so future handlers still have access to it.
	//
	// Validate MUST return ErrUnexpected when there was an unexpected issue while
	// running the validation.
	//
	// It MUST return ErrEmptyBody when the body is expected to contain something
	//
	// If there are actual validation errors, then ValidationErrors should be returned
	Validate(req *http.Request) ([]*ErrorItem, error)
}

// ApiError represents a standard error from the handler
type ApiError struct {
	StatusCode int                    `json:"status"`
	Kind       string                 `json:"kind"`
	Details    string                 `json:"details"`
	Errors     []*ErrorItem           `json:"errors,omitempty"`
	Meta       map[string]interface{} `json:"meta,omitempty"`
	sourceErr  error
}

func (ae *ApiError) Error() string {
	return ae.Details
}

func (ae *ApiError) Unwrap() error {
	return ae.sourceErr
}

func (ae *ApiError) Status() int {
	return ae.StatusCode
}

type ErrorItem struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
	Msg   string      `json:"msg"`
}

// An ErrorLogger logs an error in a particular request
type ErrorLogger interface {
	LogError(req *http.Request, err error)
}

// An ArgumentResolver resolves argument types from a function in the context of an HTTP Request
//
// For every argument in a function, it receives the current http.Request, the reflect.Type of
// the argument and the position on the argument on the function signature.
//
// Go does not have a notion of argument names, so the position is crucial to implement custom
// resolving logic.
type ArgumentResolver interface {
	// Resolve resolves an argument using the request information and the argument's type and position.
	// It must return a valid reflect.Value
	//
	// ErrArgumentUnsupported is returned when the type could not be resolved
	// ErrArgumentResolution is returned when something unexpected happens while resolving
	// ErrEmptyBody is returned when the body EOFs
	Resolve(req *http.Request, t reflect.Type, pos int) (reflect.Value, error)
}
