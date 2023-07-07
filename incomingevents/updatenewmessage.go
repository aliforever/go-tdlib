package incomingevents

import "github.com/aliforever/go-tdlib/entities"

type UpdateNewMessage struct {
	Type    string            `json:"@type"`
	Message *entities.Message `json:"message"`
}
