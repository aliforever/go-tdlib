package outgoingevents

type OpenChat struct {
	ChatID int64 `json:"chat_id"`
}

func (OpenChat) Type() string {
	return "openChat"
}
