package jsonapi_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mnavarrocarter/jsonapi"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	jsonapi.VarFunc = mux.Vars
}

func Test_WithVar(t *testing.T) {
	handler := jsonapi.Wrap(func(_ context.Context, id string) map[string]string {
		return map[string]string{
			"msg": fmt.Sprintf("user id is %s", id),
		}
	}, jsonapi.WithVar("id", 1))

	router := mux.NewRouter()
	router.NotFoundHandler = jsonapi.NotFoundHandler
	router.MethodNotAllowedHandler = jsonapi.MethodNotAllowedHandler
	router.Methods("GET").Path("/user/{id}").Handler(handler)

	tt := []struct {
		caseName         string
		req              *http.Request
		expectedResponse []byte
		expectedStatus   int
	}{
		{
			caseName:         "resolves a mux var",
			req:              httptest.NewRequest("GET", "http://localhost:8000/user/1234", http.NoBody),
			expectedResponse: []byte(`{"msg":"user id is 1234"}` + "\n"),
			expectedStatus:   http.StatusOK,
		},
		{
			caseName:         "not found",
			req:              httptest.NewRequest("GET", "http://localhost:8000/user", http.NoBody),
			expectedResponse: []byte(`{"status":404,"details":"No handler found for GET /user"}` + "\n"),
			expectedStatus:   http.StatusNotFound,
		},
		{
			caseName:         "method not allowed",
			req:              httptest.NewRequest("POST", "http://localhost:8000/user/1234", http.NoBody),
			expectedResponse: []byte(`{"status":405,"details":"Method not allowed for POST /user/1234"}` + "\n"),
			expectedStatus:   http.StatusMethodNotAllowed,
		},
	}

	for _, test := range tt {
		t.Run(test.caseName, func(t *testing.T) {
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, test.req)

			res := rec.Result()

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal("could not read response body")
			}

			defer func(c io.Closer) {
				_ = c.Close()
			}(res.Body)

			if res.StatusCode != test.expectedStatus {
				t.Errorf("expected status %d does not match received %d", test.expectedStatus, res.StatusCode)
			}

			if !bytes.Equal(test.expectedResponse, b) {
				t.Errorf(
					"response body does not match\nexpected: %s\nreceived: %s\n",
					string(test.expectedResponse),
					string(b),
				)
			}
		})
	}
}
