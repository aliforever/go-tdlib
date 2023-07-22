package incomingevents

import (
	"github.com/aliforever/go-tdlib/entities"
)

type UpdateMessageContent struct {
	ChatID     int64                    `json:"chat_id"`
	MessageID  int64                    `json:"message_id"`
	NewContent *entities.MessageContent `json:"new_content"`
}
