package tdlib

type ManagerHandlers struct {
	onRawIncomingEvent func(eventBytes []byte)
}

func NewManagerHandlers() *ManagerHandlers {
	return &ManagerHandlers{}
}

func (h *ManagerHandlers) SetRawIncomingEventHandler(fn func(eventBytes []byte)) *ManagerHandlers {
	h.onRawIncomingEvent = fn

	return h
}
