package jsonapi_test

import (
	"context"
	"embed"
	"fmt"
	"github.com/mnavarrocarter/jsonapi"
	"io/fs"
	"net/http"
	"reflect"
	"testing"
)

//go:embed testdata
var testdata embed.FS

func mustOpen(t *testing.T, name string) fs.File {
	t.Helper()
	f, err := testdata.Open(fmt.Sprintf("testdata/%s", name))
	if err != nil {
		panic(err)
	}

	return f
}

type testCmd struct {
	Id      string  `json:"id,omitempty"`
	Name    string  `json:"name,omitempty"`
	Months  int     `json:"months,omitempty"`
	Rate    float64 `json:"rate,omitempty"`
	Deposit bool    `json:"deposit,omitempty"`
}

type testResp struct {
	Msg string `json:"msg"`
}

var stringMapType = reflect.TypeOf((*map[string]string)(nil)).Elem()

type wrappedResolver struct {
	next jsonapi.ArgumentResolver
}

func (w *wrappedResolver) Resolve(req *http.Request, t reflect.Type, pos int) (reflect.Value, error) {
	if t == stringMapType {
		return reflect.ValueOf(map[string]string{"hello": "world"}), nil
	}

	return w.next.Resolve(req, t, pos)
}

type customContext struct {
	context.Context
}
