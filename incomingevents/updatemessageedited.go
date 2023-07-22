package incomingevents

import (
	"encoding/json"
	"github.com/aliforever/go-tdlib/entities"
)

type UpdateMessageEdited struct {
	ChatID      int64                `json:"chat_id"`
	MessageID   int64                `json:"message_id"`
	EditDate    int64                `json:"edit_date"`
	ReplyMarkup entities.ReplyMarkup `json:"reply_markup"`
}

// UnmarshalJSON implements json.Unmarshaler
func (updateMessageEdited *UpdateMessageEdited) UnmarshalJSON(b []byte) error {
	type baseKeyboard struct {
		Type string `json:"@type"`
	}

	type tmpStruct struct {
		ChatID      int64           `json:"chat_id"`
		MessageID   int64           `json:"message_id"`
		EditDate    int64           `json:"edit_date"`
		ReplyMarkup json.RawMessage `json:"reply_markup"`
	}

	var tmp tmpStruct

	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	var base baseKeyboard

	err = json.Unmarshal(tmp.ReplyMarkup, &base)
	if err != nil {
		return err
	}

	switch base.Type {
	case "replyMarkupForceReply":
		var replyMarkupForceReply entities.ReplyMarkupForceReply
		err = json.Unmarshal(tmp.ReplyMarkup, &replyMarkupForceReply)
		if err != nil {
			return err
		}
		updateMessageEdited.ReplyMarkup = &replyMarkupForceReply
	case "replyMarkupInlineKeyboard":
		var replyMarkupInlineKeyboard entities.ReplyMarkupInlineKeyboard
		err = json.Unmarshal(tmp.ReplyMarkup, &replyMarkupInlineKeyboard)
		if err != nil {
			return err
		}
		updateMessageEdited.ReplyMarkup = &replyMarkupInlineKeyboard
	case "replyMarkupRemoveKeyboard":
		var replyMarkupRemoveKeyboard entities.ReplyMarkupRemoveKeyboard
		err = json.Unmarshal(tmp.ReplyMarkup, &replyMarkupRemoveKeyboard)
		if err != nil {
			return err
		}
		updateMessageEdited.ReplyMarkup = &replyMarkupRemoveKeyboard
	case "replyMarkupShowKeyboard":
		var replyMarkupShowKeyboard entities.ReplyMarkupShowKeyboard
		err = json.Unmarshal(tmp.ReplyMarkup, &replyMarkupShowKeyboard)
		if err != nil {
			return err
		}
		updateMessageEdited.ReplyMarkup = &replyMarkupShowKeyboard
	}

	updateMessageEdited.ChatID = tmp.ChatID
	updateMessageEdited.MessageID = tmp.MessageID
	updateMessageEdited.EditDate = tmp.EditDate

	return nil
}
