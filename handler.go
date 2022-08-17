package jsonapi

import (
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
	ErrorLogger      ErrorLogger      // The error logger to be used
	ErrorCaster      ErrorCaster      // The error caster to be used
	ResponseSender   ResponseSender   // Sends the response
	SkipPanic        bool             // Whether to skip panics or not
}

func (h *JsonHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func(c io.Closer) {
		_ = c.Close()
	}(req.Body)

	// Validate the request
	if h.RequestValidator != nil {
		items, err := h.RequestValidator.Validate(req)
		if err != nil {
			h.sendError(w, req, err)
			return
		}

		if len(items) != 0 {
			h.sendError(w, req, items)
			return
		}
	}

	// arguments is a slice of reflect.Value to pass to h.fn.Call
	args := make([]reflect.Value, 0, len(h.fn.in))

	for i, t := range h.fn.in {
		v, err := h.ArgumentResolver.Resolve(req, t, i)
		if err != nil {
			h.sendError(w, req, err)
			return
		}
		args = append(args, v)
	}

	defer func() {
		if r := recover(); r != nil && h.SkipPanic == false {
			h.sendError(w, req, r)
		}
	}()

	out, err := h.fn.call(args)

	// If the error is not nil, then we have to return it
	// This is a domain error, so it should use the error status code and the original message.
	if err != nil {
		h.sendError(w, req, err)
		return
	}

	h.ResponseSender.SendResponse(w, out)
	return
}

func (h *JsonHandler) sendError(w http.ResponseWriter, r *http.Request, v interface{}) {
	err := h.ErrorCaster.CastError(v)

	if err != nil {
		h.ErrorLogger.LogError(r, err)
	}

	h.ResponseSender.SendResponse(w, err)
}
