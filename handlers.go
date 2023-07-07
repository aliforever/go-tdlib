package tdlib

import (
	"github.com/aliforever/go-tdlib/entities"
	"github.com/aliforever/go-tdlib/incomingevents"
	"sync"
)

type Handlers struct {
	rawIncomingEvent           func(event []byte)
	incomingEvent              func(event incomingevents.Event)
	onUpdateConnectionState    func(newState entities.ConnectionStateType)
	onUpdateAuthorizationState func(newState entities.AuthorizationStateType)
	onError                    func(err incomingevents.ErrorEvent)

	eventTypeHandlerLocker sync.Mutex
	eventTypeHandlers      map[string]event

	onNewMessageHandlers []*newMessageHandler
}

func NewHandlers() *Handlers {
	return &Handlers{eventTypeHandlers: map[string]event{}}
}

func (h *Handlers) SetErrorHandler(fn func(err incomingevents.ErrorEvent)) *Handlers {
	h.onError = fn
	return h
}

func (h *Handlers) SetRawIncomingEventHandler(fn func(eventBytes []byte)) *Handlers {
	h.rawIncomingEvent = fn
	return h
}

func (h *Handlers) SetIncomingEventHandler(fn func(event incomingevents.Event)) *Handlers {
	h.incomingEvent = fn
	return h
}

func (h *Handlers) SetOnUpdateConnectionStateEventHandler(fn func(newState entities.ConnectionStateType)) *Handlers {
	h.onUpdateConnectionState = fn
	return h
}

func (h *Handlers) SetOnUpdateAuthorizationStateEventHandler(fn func(newState entities.AuthorizationStateType)) *Handlers {
	h.onUpdateAuthorizationState = fn
	return h
}

func (h *Handlers) AddOnNewMessageHandler(
	fn func(data *incomingevents.UpdateNewMessage),
	filters *OnNewMessageFilters,
) *Handlers {
	h.eventTypeHandlerLocker.Lock()
	defer h.eventTypeHandlerLocker.Unlock()

	h.onNewMessageHandlers = append(h.onNewMessageHandlers, newOnNewMessageHandler(fn, filters))

	return h
}

func (h *Handlers) SetOnFile(fn func(data *incomingevents.UpdateFile)) *Handlers {
	h.eventTypeHandlerLocker.Lock()
	defer h.eventTypeHandlerLocker.Unlock()

	h.eventTypeHandlers["updateFile"] = NewEventHandler[incomingevents.UpdateFile](fn)

	return h
}

func (h *Handlers) AddEventTypeHandler(eventType string, e event) *Handlers {
	h.eventTypeHandlerLocker.Lock()
	defer h.eventTypeHandlerLocker.Unlock()

	h.eventTypeHandlers[eventType] = e

	return h
}
