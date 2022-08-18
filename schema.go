package jsonapi

import (
	"bytes"
	"github.com/xeipuuv/gojsonschema"
	"io"
	"net/http"
)

func WithSchema(schema io.Reader) OptsFn {
	return func(h *JsonHandler) {
		if schema == nil {
			return
		}

		b, err := io.ReadAll(schema)
		if err != nil {
			panic(err)
		}

		h.RequestValidator = &jsonSchemaValidator{
			loader: gojsonschema.NewBytesLoader(b),
		}
	}
}

// jsonSchemaValidator validates a request body using json _testdata
// It uses the "github.com/xeipuuv/gojsonschema" library to validate
type jsonSchemaValidator struct {
	loader gojsonschema.JSONLoader
}

func (v *jsonSchemaValidator) Validate(req *http.Request) ([]*ErrorItem, error) {
	defer func(c io.Closer) {
		_ = c.Close()
	}(req.Body)

	b, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, ErrUnexpected
	}

	loader := gojsonschema.NewBytesLoader(b)

	buff := bytes.NewBuffer(b)

	// We change the body to the buffer
	req.Body = io.NopCloser(buff)

	result, err := gojsonschema.Validate(v.loader, loader)
	if err == io.EOF {
		return nil, ErrEmptyBody
	}

	if err != nil {
		return nil, ErrUnexpected
	}

	if result.Valid() {
		return nil, nil
	}

	errors := make([]*ErrorItem, 0, len(result.Errors()))

	for _, res := range result.Errors() {
		errors = append(errors, &ErrorItem{
			Field: res.Field(),
			Value: res.Value(),
			Msg:   res.Description(),
		})
	}

	return errors, nil
}
