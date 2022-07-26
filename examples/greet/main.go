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
