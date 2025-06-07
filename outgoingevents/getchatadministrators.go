package outgoingevents

type GetChatAdministrators struct {
	ChatID int64 `json:"chat_id"`
}

func (s GetChatAdministrators) Type() string {
	return "getChatAdministrators"
}
