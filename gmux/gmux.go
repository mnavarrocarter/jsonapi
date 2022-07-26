package gmux

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mnavarrocarter/jsonapi"
	"net/http"
	"reflect"
)

func WithVar(key string, pos int) jsonapi.OptsFn {
	return func(h *jsonapi.JsonHandler) {
		h.ArgumentResolver = NewVarInjector(h.ArgumentResolver, key, pos)
	}
}

func NewVarInjector(next jsonapi.ArgumentResolver, key string, pos int) *VarInjector {
	return &VarInjector{
		next: next,
		key:  key,
		pos:  pos,
	}
}

// A VarInjector is designed to check the route vars and inject them in a function.
//
// It has to be pre-configured by using the position of the argument and the key
// that should be injected in that position.
type VarInjector struct {
	next jsonapi.ArgumentResolver
	key  string
	pos  int
}

func (ai *VarInjector) Resolve(req *http.Request, t reflect.Type, pos int) (reflect.Value, error) {
	if ai.pos != pos {
		return ai.next.Resolve(req, t, pos)
	}

	vars := mux.Vars(req)

	val, ok := vars[ai.key]
	if !ok {
		return reflect.Value{}, fmt.Errorf("%w: key '%s' does not exist in mux vars", jsonapi.ErrUnexpected, ai.key)
	}

	if !reflect.TypeOf(val).AssignableTo(t) {
		return reflect.Value{}, fmt.Errorf("%w: argument cannot be assigned to declared type", jsonapi.ErrUnexpected)
	}

	return reflect.ValueOf(val), nil
}
