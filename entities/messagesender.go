package entities

import "encoding/json"

type MessageSenderType string

const (
	MessageSenderTypeChat MessageSenderType = "messageSenderChat"
	MessageSenderTypeUser MessageSenderType = "messageSenderUser"
)

type MessageSenderChat struct {
	ChatId int64 `json:"chat_id"`
}

type MessageSenderUser struct {
	UserId int64 `json:"user_id"`
}

type MessageSender struct {
	Type MessageSenderType `json:"@type"`

	MessageSenderChat *MessageSenderChat `json:"messageSenderChat,omitempty"`
	MessageSenderUser *MessageSenderUser `json:"messageSenderUser,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (messageSender *MessageSender) UnmarshalJSON(data []byte) error {
	var err error

	type plain MessageSender

	var plainMessageSender plain

	if err = json.Unmarshal(data, &plainMessageSender); err != nil {
		return err
	}

	if plainMessageSender.MessageSenderChat != nil {
		messageSender.MessageSenderChat = plainMessageSender.MessageSenderChat
		messageSender.Type = MessageSenderTypeChat
	} else if plainMessageSender.MessageSenderUser != nil {
		messageSender.MessageSenderUser = plainMessageSender.MessageSenderUser
		messageSender.Type = MessageSenderTypeUser
	}

	return nil
}
