package jsonapi

import (
	"net/http"
)

var NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	resp := &ApiError{
		StatusCode: http.StatusNotFound,
		Kind:       "Not Found",
		Details:    "Handler not found for request",
		Meta: map[string]interface{}{
			"method": req.Method,
			"path":   req.URL.Path,
		},
	}

	Defaults.SendResponse(w, resp)
})

var MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

	resp := &ApiError{
		StatusCode: http.StatusMethodNotAllowed,
		Kind:       "Method Not Allowed",
		Details:    "Method not allowed for request",
		Meta: map[string]interface{}{
			"method": req.Method,
			"path":   req.URL.Path,
		},
	}

	Defaults.SendResponse(w, resp)
})
