package jsonapi

type OptsFn func(h *JsonHandler)

// Wrap makes a JsonHandler using the Defaults
//
// See JsonHandler for documentation on how this handler works.
//
// Also, see Defaults to study the default implementations of the different components.
func Wrap(fn any, opts ...OptsFn) *JsonHandler {
	h := &JsonHandler{
		fn:               reflectFunc(fn),
		RequestValidator: Defaults,
		ArgumentResolver: Defaults,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}
