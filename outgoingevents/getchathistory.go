package outgoingevents

type GetChatHistory struct {
	ChatID        int64  `json:"chat_id"`
	FromMessageID int64  `json:"from_message_id"`
	Offset        int64  `json:"offset"`
	Limit         uint64 `json:"limit"`
	OnlyLocal     bool   `json:"only_local"`
}

func (s GetChatHistory) Type() string {
	return "getChatHistory"
}
