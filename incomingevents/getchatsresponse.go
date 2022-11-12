package incomingevents

type GetChatsResponse struct {
	ChatIDs    []int64 `json:"chat_ids,omitempty"`
	TotalCount int32   `json:"total_count"`
}
