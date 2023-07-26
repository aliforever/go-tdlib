package outgoingevents

type CloseChat struct {
	ChatID int64 `json:"chat_id"`
}

func (CloseChat) Type() string {
	return "closeChat"
}
