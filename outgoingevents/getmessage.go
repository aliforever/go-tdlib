package outgoingevents

type GetMessage struct {
	ChatID    int64 `json:"chat_id"`
	MessageID int64 `json:"message_id"`
}

func (s GetMessage) Type() string {
	return "getMessage"
}
