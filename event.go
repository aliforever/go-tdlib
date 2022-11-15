package tdlib

import "encoding/json"

type event interface {
	Handle(message json.RawMessage) error
}

type Empty struct{}

type Event[T any] struct {
	handler func(*T)
}

func (e Event[T]) Handle(data json.RawMessage) error {
	var t *T

	if _, ok := any(t).(*Empty); !ok {
		err := json.Unmarshal(data, &t)
		if err != nil {
			return err
		}
	}

	e.handler(t)
	return nil
}
