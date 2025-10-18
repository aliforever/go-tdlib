package outgoingevents

type AddChatMember struct {
	ChatID       int64 `json:"chat_id"`
	UserID       int64 `json:"user_id"`
	ForwardLimit int64 `json:"forward_limit"`
}

func (s AddChatMember) Type() string {
	return "addChatMember"
}
