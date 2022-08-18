package jsonapi

import (
	"net/http"
	"reflect"
)

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
