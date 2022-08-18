package jsonapi

import (
	"fmt"
	"net/http"
)

var NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	HandleError(w, req, &apiError{
		code: http.StatusNotFound,
		msg:  fmt.Sprintf("No handler found for %s %s", req.Method, req.URL.Path),
	})
})

var MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	HandleError(w, req, &apiError{
		code: http.StatusMethodNotAllowed,
		msg:  fmt.Sprintf("Method not allowed for %s %s", req.Method, req.URL.Path),
	})
})
