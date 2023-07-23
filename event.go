package tdlib

import (
	"encoding/json"
	"fmt"
)

type event interface {
	Handle(message json.RawMessage) error
}

type Empty struct{}

type Event[T any] struct {
	handler func(*T)
	filter  func(*T) bool
}

func NewEventHandler[T any](handler func(data *T)) Event[T] {
	return Event[T]{handler: handler}
}

func NewEventHandlerWithFilter[T any](handler func(data *T), filter func(*T) bool) Event[T] {
	return Event[T]{handler: handler, filter: filter}
}

func (e Event[T]) Handle(data json.RawMessage) error {
	var t *T

	if _, ok := any(t).(*Empty); !ok {
		err := json.Unmarshal(data, &t)
		if err != nil {
			return fmt.Errorf("error unmarshaling event data in custom event: %s : %s : %T", string(data), err, t)
		}
	}

	if e.filter != nil && !e.filter(t) {
		return nil
	}

	e.handler(t)

	return nil
}
