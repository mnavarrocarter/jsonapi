package jsonapi

import (
	"fmt"
	"reflect"
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// reflectFunc creates a *reflectedFn
func reflectFunc(fn any) *reflectedFn {
	t := reflect.TypeOf(fn)
	if t.Kind() != reflect.Func {
		panic("function expected")
	}

	v := reflect.ValueOf(fn)

	in := make([]reflect.Type, 0, t.NumIn())
	for i := 0; i < t.NumIn(); i++ {
		in = append(in, t.In(i))
	}

	rFn := reflectedFn{
		fn: v,
		in: in,
	}

	switch t.NumOut() {
	case 0:
		rFn.outFn = func(in []reflect.Value) (interface{}, error) {
			return nil, nil
		}
	case 1:
		if t.Out(0).Implements(errorType) {
			rFn.outFn = func(out []reflect.Value) (interface{}, error) {
				err, _ := out[0].Interface().(error)

				return nil, err
			}
		} else {
			rFn.outFn = func(out []reflect.Value) (interface{}, error) {
				return out[0].Interface(), nil
			}
		}
	case 2:
		if !t.Out(1).Implements(errorType) {
			panic(fmt.Sprintf("function %s second return value must be an error", t.Name()))
		}

		rFn.outFn = func(out []reflect.Value) (interface{}, error) {
			err, _ := out[1].Interface().(error)

			return out[0].Interface(), err
		}
	default:
		panic(fmt.Sprintf("function %s cannot return more than two values", t.Name()))
	}

	return &rFn
}

type reflectedFn struct {
	fn    reflect.Value
	in    []reflect.Type
	outFn func(out []reflect.Value) (interface{}, error)
}

func (h *reflectedFn) call(in []reflect.Value) (interface{}, error) {
	out := h.fn.Call(in)

	return h.outFn(out)
}
