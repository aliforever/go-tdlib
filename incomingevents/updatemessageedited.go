package incomingevents

import "github.com/aliforever/go-tdlib/entities"

type UpdateMessageEdited struct {
	ChatID      int64                 `json:"chat_id"`
	MessageID   int64                 `json:"message_id"`
	EditDate    int64                 `json:"edit_date"`
	ReplyMarkup *entities.ReplyMarkup `json:"reply_markup"`
}
