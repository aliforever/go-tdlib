package entities

import "encoding/json"

type MessageReplyMarkup struct {
	Type string `json:"@type"`

	*ReplyMarkupForceReply
	*ReplyMarkupInlineKeyboard
	*ReplyMarkupRemoveKeyboard
	*ReplyMarkupShowKeyboard
}

// UnmarshalJSON Overrides UnmarshalJSON for MessageReplyMarkup
func (m *MessageReplyMarkup) UnmarshalJSON(b []byte) error {
	type temp MessageReplyMarkup

	var t temp

	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}

	*m = MessageReplyMarkup(t)

	switch t.Type {
	case "replyMarkupForceReply":
		err = json.Unmarshal(b, &m.ReplyMarkupForceReply)
	case "replyMarkupInlineKeyboard":
		err = json.Unmarshal(b, &m.ReplyMarkupInlineKeyboard)
	case "replyMarkupRemoveKeyboard":
		err = json.Unmarshal(b, &m.ReplyMarkupRemoveKeyboard)
	case "replyMarkupShowKeyboard":
		err = json.Unmarshal(b, &m.ReplyMarkupShowKeyboard)
	}

	return err
}
