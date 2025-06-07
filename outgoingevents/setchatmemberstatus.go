package outgoingevents

import "github.com/aliforever/go-tdlib/entities"

type SetChatMemberStatus struct {
	ChatID   int64                     `json:"chat_id"`
	MemberID int64                     `json:"user_id"`
	Status   entities.ChatMemberStatus `json:"status"`
}

func (s SetChatMemberStatus) Type() string {
	return "setChatMemberStatus"
}
