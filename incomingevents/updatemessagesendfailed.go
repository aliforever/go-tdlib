package incomingevents

import "github.com/aliforever/go-tdlib/entities"

type UpdateMessageSendFailed struct {
	Message      *entities.Message `json:"message"`
	OldMessageID int64             `json:"old_message_id"`
	Error        *entities.Error   `json:"error"`
}
