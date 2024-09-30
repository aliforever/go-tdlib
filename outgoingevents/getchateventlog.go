package outgoingevents

import "github.com/aliforever/go-tdlib/entities"

type GetChatEventLog struct {
	ChatID    int64                         `json:"chat_id"`
	Query     string                        `json:"query"`
	FromEvent int64                         `json:"from_event"`
	Limit     int64                         `json:"limit"`
	Filters   *entities.ChatEventLogFilters `json:"filters"`
	UserIDs   []int64                       `json:"user_ids"`
}

func (s GetChatEventLog) Type() string {
	return "getChatEventLog"
}
