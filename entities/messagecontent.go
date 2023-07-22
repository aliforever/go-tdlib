package entities

import "encoding/json"

type MessageText struct {
	Text    *FormattedText `json:"text"`
	WebPage interface{}    `json:"webPage"`
}

type MessageContent struct {
	Type string `json:"@type"`

	MessageText *MessageText `json:"messageText,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (messageContent *MessageContent) UnmarshalJSON(data []byte) error {
	type tmp MessageContent

	var tmpMessageContent tmp

	err := json.Unmarshal(data, &tmpMessageContent)
	if err != nil {
		return err
	}

	*messageContent = MessageContent(tmpMessageContent)

	switch messageContent.Type {
	case "messageText":
		var messageText MessageText
		err = json.Unmarshal(data, &messageText)
		if err != nil {
			return err
		}
		messageContent.MessageText = &messageText
	}

	return nil
}
