package jsonschema

import (
	"bytes"
	"github.com/mnavarrocarter/jsonapi"
	"github.com/xeipuuv/gojsonschema"
	"io"
	"net/http"
)

func WithSchema(schema io.Reader) jsonapi.OptsFn {
	return func(h *jsonapi.JsonHandler) {
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

func NewJsonSchemaValidator(l gojsonschema.JSONLoader) *jsonSchemaValidator {
	return &jsonSchemaValidator{loader: l}
}

// jsonSchemaValidator validates a request body using json testdata
// It uses the "github.com/xeipuuv/gojsonschema" library to validate
type jsonSchemaValidator struct {
	loader gojsonschema.JSONLoader
}

func (v *jsonSchemaValidator) Validate(req *http.Request) ([]*jsonapi.ErrorItem, error) {
	defer func(c io.Closer) {
		_ = c.Close()
	}(req.Body)

	b, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, jsonapi.ErrUnexpected
	}

	loader := gojsonschema.NewBytesLoader(b)

	buff := bytes.NewBuffer(b)

	// We change the body to the buffer
	req.Body = io.NopCloser(buff)

	result, err := gojsonschema.Validate(v.loader, loader)
	if err == io.EOF {
		return nil, jsonapi.ErrEmptyBody
	}

	if err != nil {
		return nil, jsonapi.ErrUnexpected
	}

	if result.Valid() {
		return nil, nil
	}

	errors := make([]*jsonapi.ErrorItem, 0, len(result.Errors()))

	for _, res := range result.Errors() {
		errors = append(errors, &jsonapi.ErrorItem{
			Field: res.Field(),
			Value: res.Value(),
			Msg:   res.Description(),
		})
	}

	return errors, nil
}
