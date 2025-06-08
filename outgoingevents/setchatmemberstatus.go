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
	// Create a map with the base fields
	statusMap := map[string]interface{}{
		"@type": s.Status.Type(),
	}

	// Marshal the concrete status to get its fields
	statusBytes, err := json.Marshal(s.Status)
	if err != nil {
		return nil, err
	}

	// Unmarshal into the map to merge fields
	var statusFields map[string]interface{}
	if err := json.Unmarshal(statusBytes, &statusFields); err != nil {
		return nil, err
	}

	// Merge the fields
	for k, v := range statusFields {
		statusMap[k] = v
	}

	m := map[string]interface{}{
		"chat_id":   s.ChatID,
		"member_id": entities.MessageSenderUser(s.MemberID),
		"status":    statusMap,
	}

	return json.Marshal(m)
}
