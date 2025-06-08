package outgoingevents

import (
	"encoding/json"

	"github.com/aliforever/go-tdlib/entities"
)

type SetChatMemberStatus struct {
	ChatID   int64                     `json:"chat_id"`
	MemberID int64                     `json:"member_id"`
	Status   entities.ChatMemberStatus `json:"status"`
}

func (s SetChatMemberStatus) Type() string {
	return "setChatMemberStatus"
}

// MarshalJSON implements json.Marshaler
func (s SetChatMemberStatus) MarshalJSON() ([]byte, error) {
	type chatMemberStatus struct {
		Type string `json:"@type"`
		entities.ChatMemberStatus
	}

	m := map[string]interface{}{
		"chat_id":   s.ChatID,
		"member_id": entities.MessageSenderUser(s.MemberID),
		"status": chatMemberStatus{
			Type:             s.Status.Type(),
			ChatMemberStatus: s.Status,
		},
	}

	return json.Marshal(m)
}
