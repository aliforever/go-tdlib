package incomingevents

type GetChatsResponse struct {
	Event

	ChatIDs    []int64 `json:"chat_ids,omitempty"`
	TotalCount int32   `json:"total_count"`
}
