package incomingevents

import "github.com/aliforever/go-tdlib/entities"

type GetChatHistoryResponse struct {
	Event

	TotalCount int64              `json:"total_count"`
	Messages   []entities.Message `json:"messages"`
}
