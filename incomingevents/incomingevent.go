package incomingevents

import (
	"encoding/json"
	"fmt"
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
		return Event{}, fmt.Errorf("error unmarshaling event data from bytes: %s : %s", string(b), err)
	}

	event.Raw = b

	return event, err
}

func GenericFromBytes(b []byte) (Generic, error) {
	var event Generic
	err := json.Unmarshal(b, &event)
	if err != nil {
		return Generic{}, fmt.Errorf("error unmarshaling event data generic from bytes: %s : %s", string(b), err)
	}

	return event, err
}

func ErrorFromBytes(b []byte) (ErrorEvent, error) {
	var event ErrorEvent

	err := json.Unmarshal(b, &event)
	if err != nil {
		return ErrorEvent{}, fmt.Errorf("error unmarshaling error event data: %s : %s", string(b), err)
	}

	return event, err
}
