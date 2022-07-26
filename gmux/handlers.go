package gmux

import (
	"github.com/mnavarrocarter/jsonapi"
	"net/http"
)

var NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	resp := &jsonapi.ApiError{
		Status:  http.StatusNotFound,
		Kind:    "Not Found",
		Details: "Handler not found for request",
		Meta: map[string]interface{}{
			"method": req.Method,
			"path":   req.URL.Path,
		},
	}

	jsonapi.Defaults.SendResponse(w, resp)
})

var MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

	resp := &jsonapi.ApiError{
		Status:  http.StatusMethodNotAllowed,
		Kind:    "Method Not Allowed",
		Details: "Method not allowed for request",
		Meta: map[string]interface{}{
			"method": req.Method,
			"path":   req.URL.Path,
		},
	}

	jsonapi.Defaults.SendResponse(w, resp)
})
