package entities

type MessageSenders struct {
	TotalCount int64           `json:"total_count"`
	Senders    []MessageSender `json:"senders"`
}
