package entities

type ChatEvents struct {
	Type   string      `json:"@type"`
	Events []ChatEvent `json:"events"`
}

type ChatEvent struct {
	ID       int64            `json:"id"`
	Date     int64            `json:"date"`
	MemberID *MessageSender   `json:"member_id"`
	Action   *ChatEventAction `json:"action"`
}

type ChatEventAction struct {
	Type string `json:"@type"`
}
