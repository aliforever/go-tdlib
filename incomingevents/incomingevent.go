package incomingevents

import (
	"encoding/json"
	"github.com/aliforever/go-tdlib/entities"
)

type Generic struct {
	Type      string `json:"@type"`
	RequestID string `json:"@extra,omitempty"`
}

type ErrorEvent struct {
	Generic

	Code    int64           `json:"code"`
	Message json.RawMessage `json:"message"`
}

type Event struct {
	Generic

	Code    int64           `json:"code,omitempty"`
	Message json.RawMessage `json:"message,omitempty"`

	UpdateFile         *DownloadFileResponse        `json:"file,omitempty"`
	State              *entities.ConnectionState    `json:"state,omitempty"`
	AuthorizationState *entities.AuthorizationState `json:"authorization_state,omitempty"`

	Raw []byte `json:"-"`
}

func FromBytes(b []byte) (Event, error) {
	var event Event
	err := json.Unmarshal(b, &event)
	if err != nil {
		return Event{}, err
	}

	event.Raw = b

	return event, err
}

func GenericFromBytes(b []byte) (Generic, error) {
	var event Generic
	err := json.Unmarshal(b, &event)
	if err != nil {
		return Generic{}, err
	}

	return event, err
}
