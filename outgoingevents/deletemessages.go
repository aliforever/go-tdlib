package outgoingevents

type DeleteMessages struct {
	ChatID     int64   `json:"chat_id"`
	MessageIDs []int64 `json:"message_ids"`
	Revoke     bool    `json:"revoke"`
}

func (s DeleteMessages) Type() string {
	return "deleteMessages"
}
