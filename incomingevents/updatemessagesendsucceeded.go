package incomingevents

import "github.com/aliforever/go-tdlib/entities"

type UpdateMessageSendSucceeded struct {
	Message      *entities.Message `json:"message"`
	OldMessageID int64             `json:"old_message_id"`
}
