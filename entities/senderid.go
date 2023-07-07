package entities

type SenderID struct {
	Type   MessageSenderType `json:"@type"`
	UserID int64             `json:"user_id"`
	ChatID int64             `json:"chat_id"`
}
