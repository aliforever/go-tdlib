package outgoingevents

type ForwardMessages struct {
	ChatID          int64   `json:"chat_id"`
	MessageThreadID int64   `json:"message_thread_id"`
	FromChatID      int64   `json:"from_chat_id"`
	MessageIDs      []int64 `json:"message_ids"`
	Options         any     `json:"options"`
	SendCopy        bool    `json:"send_copy"`
	RemoveCaption   bool    `json:"remove_caption"`
}

func (s ForwardMessages) Type() string {
	return "forwardMessages"
}
