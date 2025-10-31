package incomingevents

type UpdateDeleteMessages struct {
	ChatID      int64   `json:"chat_id"`
	MessageIDs  []int64 `json:"message_ids"`
	IsPermanent bool    `json:"is_permanent"`
	FromCache   bool    `json:"from_cache"`
}
