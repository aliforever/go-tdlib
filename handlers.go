package tdlib

import "github.com/aliforever/go-tdlib/incomingevents"

type Handlers struct {
	RawIncomingEvent        func(event []byte)
	IncomingEvent           func(event incomingevents.Event)
	OnUpdateConnectionState func(newState string)
}

func NewHandlers() *Handlers {
	return &Handlers{}
}

func (h *Handlers) SetRawIncomingEventHandler(fn func(eventBytes []byte)) *Handlers {
	h.RawIncomingEvent = fn
	return h
}

func (h *Handlers) SetIncomingEventHandler(fn func(event incomingevents.Event)) *Handlers {
	h.IncomingEvent = fn
	return h
}

func (h *Handlers) SetOnUpdateConnectionStateEventHandler(fn func(newState string)) *Handlers {
	h.OnUpdateConnectionState = fn
	return h
}
