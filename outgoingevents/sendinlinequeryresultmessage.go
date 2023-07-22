package outgoingevents

type SendInlineQueryResultMessage struct {
	ChatID           int64               `json:"chat_id"`
	MessageThreadID  int64               `json:"message_thread_id"`
	ReplyToMessageID int64               `json:"reply_to_message_id"`
	Options          *SendMessageOptions `json:"options,omitempty"`
	QueryID          string              `json:"query_id"`
	ResultID         string              `json:"result_id"`
	HideViaBot       bool                `json:"hide_via_bot"`
}

func (s SendInlineQueryResultMessage) Type() string {
	return "sendInlineQueryResultMessage"
}
