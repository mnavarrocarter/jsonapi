package jsonapi

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	StatusCode int          `json:"status"`
	Details    string       `json:"details"`
	Errors     []*ErrorItem `json:"errors,omitempty"`
}

var SendResponse = sendResponse

type ResponseSenderFunc = func(w http.ResponseWriter, req *http.Request, v interface{})

func sendResponse(w http.ResponseWriter, _ *http.Request, v interface{}) {
	if v == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	status := http.StatusOK

	switch t := v.(type) {
	case error:
		status = http.StatusInternalServerError
		if c, ok := v.(Coder); ok {
			status = c.Code()
		}

		v = &errorResponse{
			StatusCode: status,
			Details:    t.Error(),
		}
	case []*ErrorItem:
		status = http.StatusBadRequest
		v = &errorResponse{
			StatusCode: status,
			Details:    "Validation errors",
			Errors:     t,
		}
	default:
		// Override status code if we can
		if c, ok := v.(Coder); ok {
			status = c.Code()
		}
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
