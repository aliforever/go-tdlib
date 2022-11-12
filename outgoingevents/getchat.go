package outgoingevents

type GetChat struct {
	ChatID int64 `json:"chat_id"`
}

func (s GetChat) Type() string {
	return "getChat"
}
