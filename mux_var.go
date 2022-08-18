package jsonapi

import (
	"fmt"
	"net/http"
	"reflect"
)

var VarFunc = func(r *http.Request) map[string]string {
	return map[string]string{}
}

func WithVar(key string, pos int) OptsFn {
	return func(h *JsonHandler) {
		h.ArgumentResolver = &varInjector{
			next: h.ArgumentResolver,
			key:  key,
			pos:  pos,
		}
	}
}

// A VarInjector is designed to check the route vars and inject them in a function.
//
// It has to be pre-configured by using the position of the argument and the key
// that should be injected in that position.
type varInjector struct {
	next ArgumentResolver
	key  string
	pos  int
}

func (vi *varInjector) Resolve(req *http.Request, t reflect.Type, pos int) (reflect.Value, error) {
	if vi.pos != pos {
		return vi.next.Resolve(req, t, pos)
	}

	vars := VarFunc(req)

	val, ok := vars[vi.key]
	if !ok {
		return reflect.Value{}, fmt.Errorf("%w: key '%s' does not exist in mux vars", ErrArgumentResolution, vi.key)
	}

	if !reflect.TypeOf(val).AssignableTo(t) {
		return reflect.Value{}, fmt.Errorf("%w: argument cannot be assigned to declared type", ErrArgumentResolution)
	}

	return reflect.ValueOf(val), nil
}
