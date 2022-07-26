package jsonapi_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mnavarrocarter/jsonapi"
)

func TestHandler(t *testing.T) {
	tt := []struct {
		name             string
		req              *http.Request
		handler          interface{}
		schema           io.Reader
		expectedResponse []byte
		customResolver   jsonapi.ArgumentResolver
		expectedStatus   int
		skipPanic        bool
	}{
		{
			name: "no arguments json handler",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
			}(),
			handler:          func() {},
			expectedResponse: []byte(""),
			expectedStatus:   http.StatusNoContent,
		},
		{
			name: "only context json handler",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
			}(),
			handler: func(ctx context.Context) {

			},
			expectedResponse: []byte(""),
			expectedStatus:   http.StatusNoContent,
		},
		{
			name: "custom context typed with context interface json handler",
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
				return req.WithContext(&customContext{req.Context()})
			}(),
			handler: func(ctx context.Context) {

			},
			expectedResponse: []byte(""),
			expectedStatus:   http.StatusNoContent,
		},
		{
			name: "custom context typed with concrete context json handler",
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
				return req.WithContext(&customContext{req.Context()})
			}(),
			handler: func(ctx *customContext) {

			},
			expectedResponse: []byte(""),
			expectedStatus:   http.StatusNoContent,
		},
		{
			name: "custom context typed with concrete context as value",
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
				return req.WithContext(customContext{req.Context()})
			}(),
			handler: func(ctx customContext) {

			},
			expectedResponse: []byte(""),
			expectedStatus:   http.StatusNoContent,
		},
		{
			name: "custom non assignable context",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
			}(),
			handler: func(ctx *customContext) {
				panic("should not reach here")
			},
			expectedResponse: []byte(`{"status":500,"details":"Error while trying to resolve handler arguments"}` + "\n"),
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name: "empty body",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
			}(),
			handler: func(cmd *testCmd) *testResp {
				return &testResp{Msg: "success"}
			},
			expectedResponse: []byte(`{"status":400,"details":"Request body cannot be empty"}` + "\n"),
			expectedStatus:   http.StatusBadRequest,
		},
		{
			name: "valid json struct pointer with response and no testdata",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", mustOpen(t, "valid.json"))
			}(),
			handler: func(cmd *testCmd) *testResp {
				return &testResp{Msg: "success"}
			},
			expectedResponse: []byte(`{"msg":"success"}` + "\n"),
			expectedStatus:   http.StatusOK,
		},
		{
			name: "valid json struct value with response and no testdata",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", mustOpen(t, "valid.json"))
			}(),
			handler: func(cmd testCmd) testResp {
				return testResp{Msg: "success"}
			},
			expectedResponse: []byte(`{"msg":"success"}` + "\n"),
			expectedStatus:   http.StatusOK,
		},
		{
			name: "error response",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", mustOpen(t, "valid.json"))
			}(),
			handler: func() error {
				return errors.New("there was an error")
			},
			expectedResponse: []byte(`{"status":500,"details":"there was an error"}` + "\n"),
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name: "nil error response",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", mustOpen(t, "valid.json"))
			}(),
			handler: func() error {
				return nil
			},
			expectedResponse: []byte(""),
			expectedStatus:   http.StatusNoContent,
		},
		{
			name: "double nil error response",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", mustOpen(t, "valid.json"))
			}(),
			handler: func() (map[string]interface{}, error) {
				return map[string]interface{}{
					"msg": "hello",
				}, nil
			},
			expectedResponse: []byte(`{"msg":"hello"}` + "\n"),
			expectedStatus:   http.StatusOK,
		},
		{
			name: "panic response",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", mustOpen(t, "valid.json"))
			}(),
			handler: func() error {
				panic("something really bad has happened")
			},
			expectedResponse: []byte(`{"status":500,"details":"An unexpected error has occurred"}` + "\n"),
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name: "custom error response",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", mustOpen(t, "valid.json"))
			}(),
			handler: func() error {
				return customErr(400, errors.New("server error"), "customer error")
			},
			expectedResponse: []byte(`{"status":400,"details":"customer error"}` + "\n"),
			expectedStatus:   http.StatusBadRequest,
		},
		{
			name: "struct and error response",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", mustOpen(t, "valid.json"))
			}(),
			handler: func() (*testResp, error) {
				return nil, errors.New("there was an error")
			},
			expectedResponse: []byte(`{"status":500,"details":"there was an error"}` + "\n"),
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name: "valid ctx and cmd with response and with schema",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", mustOpen(t, "valid.json"))
			}(),
			schema: mustOpen(t, "schema.json"),
			handler: func(ctx context.Context, cmd *testCmd) (*testResp, error) {
				return &testResp{Msg: "success"}, nil
			},
			expectedResponse: []byte(`{"msg":"success"}` + "\n"),
			expectedStatus:   http.StatusOK,
		},
		{
			name: "invalid body with schema",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", mustOpen(t, "invalid.json"))
			}(),
			schema:           mustOpen(t, "schema.json"),
			handler:          func() {},
			expectedResponse: []byte(`{"status":400,"details":"Validation errors","errors":[{"field":"id","value":"3f2476fd2b-270f-4baa-81c9-01e91fc87fd3","msg":"Does not match format 'uuid'"},{"field":"name","value":"","msg":"String length must be greater than or equal to 1"}]}` + "\n"),
			expectedStatus:   http.StatusBadRequest,
		},
		{
			name: "unresolvable element",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", mustOpen(t, "valid.json"))
			}(),
			handler: func(ctx context.Context, cmd *testCmd, params map[string]string) *testResp {
				panic("should not reach here")
			},
			expectedResponse: []byte(`{"status":500,"details":"Error while trying to resolve handler arguments"}` + "\n"),
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name: "custom resolver",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://example.com", mustOpen(t, "valid.json"))
			}(),
			handler: func(ctx context.Context, cmd *testCmd, params map[string]string) *testResp {
				return &testResp{Msg: "success"}
			},
			customResolver:   &wrappedResolver{next: jsonapi.Defaults},
			expectedResponse: []byte(`{"msg":"success"}` + "\n"),
			expectedStatus:   http.StatusOK,
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler := jsonapi.Wrap(test.handler, jsonapi.WithSchema(test.schema))

			if test.skipPanic {
				handler.SkipPanic = test.skipPanic
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("test should have panicked")
					}
				}()
			}

			if test.customResolver != nil {
				handler.ArgumentResolver = test.customResolver
			}

			handler.ServeHTTP(rec, test.req)
			resp := rec.Result()

			defer func(c io.Closer) {
				_ = c.Close()
			}(resp.Body)

			if test.expectedStatus != resp.StatusCode {
				t.Errorf("expected status %d does not match received %d", test.expectedStatus, resp.StatusCode)
			}

			b, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal("could not read response body")
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

func customErr(status int, prev error, msg string) *appError {
	return &appError{
		status: status,
		prev:   prev,
		msg:    msg,
	}
}

type appError struct {
	status int
	prev   error
	msg    string
}

func (e *appError) Code() int {
	return e.status
}

func (e *appError) Unwrap() error {
	return e.prev
}

func (e *appError) Error() string {
	return e.msg
}
