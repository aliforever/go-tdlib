package outgoingevents

import "github.com/aliforever/go-tdlib/entities"

type GetChats struct {
	ChatList *entities.ChatList `json:"chat_list"`
	Limit    int32              `json:"limit"`
}

func (s GetChats) Type() string {
	return "getChats"
}
