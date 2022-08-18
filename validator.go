package jsonapi

import (
	"errors"
	"net/http"
)

var ErrValidation = errors.New("validation error")

type RequestValidator interface {
	// Validate validates the request
	//
	// If the body of the request is read, then the validator needs to restore
	// the request body so future handlers still have access to it.
	//
	// Validate MUST return ErrValidation error when an error running the validation occurs.
	//
	// If there are actual validation errors, then a slice of ErrorItem should be returned
	Validate(req *http.Request) ([]*ErrorItem, error)
}

type ErrorItem struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
	Msg   string      `json:"msg"`
}
