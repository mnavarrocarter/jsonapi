package jsonapi

import (
	"errors"
	"io"
	"net/http"
	"reflect"
)

// A JsonHandler wraps a function in
// JsonHandler implements http.Handler. This handler is capable of resolving
// arguments at request time and injecting them into the function. See ArgumentResolver.
//
// For instance, the default maker will inject the context included on the request, and also
// any json body into a struct.
type JsonHandler struct {
	fn               *reflectedFn     // The wrapper over the reflected function
	RequestValidator RequestValidator // The validator for the request
	ArgumentResolver ArgumentResolver // The argument resolver to be used
	SkipPanic        bool             // Whether to skip panics or not
}

func (h *JsonHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func(c io.Closer) {
		_ = c.Close()
	}(req.Body)

	// Validate the request
	if h.RequestValidator != nil {
		items, err := h.RequestValidator.Validate(req)
		if errors.Is(err, ErrEmptyBody) {
			HandleError(w, req, &apiError{
				code: http.StatusBadRequest,
				msg:  "Request body cannot be empty",
			})
			return
		}

		if err != nil {
			HandleError(w, req, &apiError{
				code: http.StatusInternalServerError,
				msg:  "There was an error while validating the request",
				prev: err,
			})
			return
		}

		if len(items) != 0 {
			SendResponse(w, req, items)
			return
		}
	}

	// arguments is a slice of reflect.Value to pass to h.fn.Call
	args := make([]reflect.Value, 0, len(h.fn.in))

	for i, t := range h.fn.in {
		v, err := h.ArgumentResolver.Resolve(req, t, i)
		if errors.Is(err, ErrEmptyBody) {
			HandleError(w, req, &apiError{
				code: http.StatusBadRequest,
				msg:  "Request body cannot be empty",
				prev: err,
			})
			return
		}

		if err != nil {
			HandleError(w, req, &apiError{
				code: http.StatusInternalServerError,
				msg:  "Error while trying to resolve handler arguments",
				prev: err,
			})
			return
		}

		args = append(args, v)
	}

	defer func() {
		if r := recover(); r != nil && h.SkipPanic == false {
			err := panicToError(r)
			HandleError(w, req, err)
		}
	}()

	out, err := h.fn.call(args)

	if err != nil {
		HandleError(w, req, err)
		return
	}

	SendResponse(w, req, out)
	return
}
