package incomingevents

import (
	"encoding/json"
	"github.com/aliforever/go-tdlib/entities"
)

type Event struct {
	Type      string          `json:"@type"`
	Code      int64           `json:"code"`
	Message   json.RawMessage `json:"message"`
	RequestID string          `json:"@extra"`

	UpdateFile *File           `json:"file"`
	State      *entities.State `json:"state"`

	*GetChatsResponse
	*GetChatResponse
	*File

	Raw []byte `json:"-"`
}

func IncomingEventFromBytes(b []byte) (Event, error) {
	var event Event
	err := json.Unmarshal(b, &event)
	if err != nil {
		return Event{}, err
	}

	event.Raw = b

	return event, err
}
