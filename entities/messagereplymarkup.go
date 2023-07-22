package entities

import "encoding/json"

type MessageReplyMarkup struct {
	Type string `json:"@type"`

	ForceReply     *ReplyMarkupForceReply     `json:"force_reply"`
	InlineKeyboard *ReplyMarkupInlineKeyboard `json:"inline_keyboard"`
	RemoveKeyboard *ReplyMarkupRemoveKeyboard `json:"remove_keyboard"`
	ShowKeyboard   *ReplyMarkupShowKeyboard   `json:"show_keyboard"`
}

// UnmarshalJSON Overrides UnmarshalJSON for MessageReplyMarkup
func (m *MessageReplyMarkup) UnmarshalJSON(b []byte) error {
	type baseMarkup struct {
		Type string `json:"@type"`
	}

	var t baseMarkup

	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}

	rp := MessageReplyMarkup{Type: t.Type}

	switch t.Type {
	case "replyMarkupForceReply":
		err = json.Unmarshal(b, &rp.ForceReply)
	case "replyMarkupInlineKeyboard":
		err = json.Unmarshal(b, &rp.InlineKeyboard)
	case "replyMarkupRemoveKeyboard":
		err = json.Unmarshal(b, &rp.RemoveKeyboard)
	case "replyMarkupShowKeyboard":
		err = json.Unmarshal(b, &rp.ShowKeyboard)
	}

	if err != nil {
		return err
	}

	*m = rp

	return nil
}
