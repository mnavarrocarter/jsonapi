`jsonapi`
=========

A toolkit for building consistent json apis with ease.

Features:
- Custom request validation (we provide an adapter for `github.com/xeipuuv/gojsonschema`)
- Automatic parameter injection (we provide an adapter for `github.com/gorilla/mux`)

Check usage for examples.

## Install

```bash
go get -u github.com/mnavarrocarter/jsonapi
```

## Usage

You can wrap simple functions into a `JsonHandler` using `jsonapi.Wrap()`. This `JsonHandler` implements
`http.Handler` so you can use it as you would use any handler.

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/mnavarrocarter/jsonapi"
	"net/http"
)

type GreetCmd struct {
	Name        string `json:"name"`
	ShouldPanic bool   `json:"shouldPanic"`
}

func Greet(_ context.Context, cmd *GreetCmd) (map[string]string, error) {
	if cmd.ShouldPanic {
		panic("something unexpected has happened")
	}

	if cmd.Name == "" {
		return nil, errors.New("you must provide a name")
	}

	return map[string]string{
		"message": fmt.Sprintf("Hello %s", cmd.Name),
	}, nil
}

func main() {
	handler := jsonapi.Wrap(Greet)

	err := http.ListenAndServe(":8000", handler)
	if err != nil {
		panic(err)
	}
}
```

Then you can invoke the handler and see the magic in action!

```text
POST http://localhost:8000
Content-Type: application/json

{
  "name": "Matias"
}

---

HTTP/1.1 200 OK
Content-Type: application/json
Date: Mon, 25 Jul 2022 12:07:48 GMT
Content-Length: 27

{
  "message": "Hello Matias"
}
```

The json handler takes care of serializing the json and put it into the right struct, and then passing it into your
function in the correct order. It can also inject any type that implements `context.Context`.

The handler is smart enough to check that it needs a body and if none is present will report back to the user:

```text
GET http://localhost:8000

---

HTTP/1.1 400 Bad Request
Content-Type: application/json
Date: Mon, 25 Jul 2022 14:10:21 GMT
Content-Length: 81

{
  "status": 400,
  "kind": "Invalid Request",
  "details": "Request body cannot be empty"
}
```

### Validation

You can instruct the handler to validate payloads by passing a json schema. This gives you valid structs in your
handlers. The error reporting of the schema validation is consistent.

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/mnavarrocarter/jsonapi"
	"github.com/mnavarrocarter/jsonapi/jsonschema"
	"net/http"
	"strings"
)

type GreetCmd struct {
	Name        string `json:"name"`
	ShouldPanic bool   `json:"should_panic"`
}

func Greet(_ context.Context, cmd *GreetCmd) (map[string]string, error) {
	if cmd.ShouldPanic {
		panic("something unexpected has happened")
	}

	if cmd.Name == "" {
		return nil, errors.New("you must provide a name")
	}

	return map[string]string{
		"message": fmt.Sprintf("Hello %s", cmd.Name),
	}, nil
}

var schema = `{
    "type": "object",
    "properties": {
        "name": {
            "type": "string",
            "minLength": 2
        },
        "should_panic": {
            "type": "boolean"
        }
    }
}`

func main() {
	handler := jsonapi.Wrap(Greet, jsonschema.WithSchema(strings.NewReader(schema)))

	err := http.ListenAndServe(":8000", handler)
	if err != nil {
		panic(err)
	}
}
```

You can make your own validation logic by implementing `jsonapi.RequestValidator`.

### Error Handling

The `JsonHandler` has smart and consistent error handling baked in.

For instance, `error` types returned from inside the function are returned to the user in a message.

These are assumed to be domain errors that contain useful information to the user of your api.

```text
POST http://localhost:8000
Content-Type: application/json

{
  "name": ""
}

---

HTTP/1.1 400 Bad Request
Content-Type: application/json
Date: Mon, 25 Jul 2022 12:21:49 GMT
Content-Length: 73

{
  "status": 400,
  "kind": "Domain Error",
  "details": "you must provide a name"
}
```

Panics, on the other hand, represent unexpected server error conditions and are reported with a fixed error
message and a `500` status code.

```text
POST http://localhost:8000
Content-Type: application/json

{
  "name": "",
  "should_panic": true
}

---

HTTP/1.1 500 Internal Server Error
Content-Type: application/json
Date: Mon, 25 Jul 2022 12:25:33 GMT
Content-Length: 78

{
  "status": 500,
  "kind": "Unknown",
  "details": "Request failed with unknown error"
}
```

When a panic occurs from inside a handler, it is logged using the default `ErrorLogger` so you can access
the information about the error.

If you wish to handle panics yourself and let them bubble, you can set `SkipPanic = true` in the `JsonHandler`.