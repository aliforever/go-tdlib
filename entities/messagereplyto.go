package entities

import "encoding/json"

type MessageReplyTo struct {
	Type string `json:"@type"`

	Message *MessageReplyToMessage `json:"force_reply"`
	Story   *MessageReplyToStory   `json:"inline_keyboard"`
}

type MessageReplyToMessage struct {
	ChatID         int64 `json:"chat_id"`
	MessageID      int64 `json:"message_id"`
	Quote          any   `json:"quote"`
	Origin         any   `json:"origin"`
	OriginSendDate int64 `json:"origin_send_date"`
	Content        any   `json:"content"`
}

type MessageReplyToStory struct {
	StorySenderChatID int64 `json:"story_sender_chat_id"`
	StoryID           int64 `json:"story_id"`
}

// UnmarshalJSON Overrides UnmarshalJSON for MessageReplyMarkup
func (m *MessageReplyTo) UnmarshalJSON(b []byte) error {
	type baseReplyTo struct {
		Type string `json:"@type"`
	}

	var t baseReplyTo

	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}

	rp := MessageReplyTo{Type: t.Type}

	switch t.Type {
	case "messageReplyToMessage":
		err = json.Unmarshal(b, &rp.Message)
	case "messageReplyToStory":
		err = json.Unmarshal(b, &rp.Story)
	}

	if err != nil {
		return err
	}

	*m = rp

	return nil
}
