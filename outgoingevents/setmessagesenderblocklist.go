package outgoingevents

import "github.com/aliforever/go-tdlib/entities"

type SetMessageSenderBlockList struct {
	SenderId  entities.SenderID   `json:"sender_id"`
	BlockList *entities.BlockList `json:"block_list"`
}

func (s SetMessageSenderBlockList) Type() string {
	return "setMessageSenderBlockList"
}
