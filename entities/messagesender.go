package entities

type MessageSenderType string

const (
	MessageSenderTypeChat MessageSenderType = "messageSenderChat"
	MessageSenderTypeUser MessageSenderType = "messageSenderUser"
)

type MessageSender struct {
	Type MessageSenderType `json:"@type"`

	ChatId *int64 `json:"chat_id,omitempty"`
	UserId *int64 `json:"user_id,omitempty"`
}

func MessageSenderUser(userId int64) MessageSender {
	return MessageSender{
		Type:   MessageSenderTypeUser,
		UserId: &userId,
	}
}

func MessageSenderChat(chatId int64) MessageSender {
	return MessageSender{
		Type:   MessageSenderTypeChat,
		ChatId: &chatId,
	}
}
