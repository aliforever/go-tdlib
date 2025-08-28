package outgoingevents

import "github.com/aliforever/go-tdlib/entities"

type JoinVideoChat struct {
	GroupCallID    int64                   `json:"group_call_id"`
	ParticipantID  *entities.MessageSender `json:"participant_id"`
	JoinParameters entities.JoinParameters `json:"join_parameters"`
	InviteHash     string                  `json:"invite_hash"`
}

func (JoinVideoChat) Type() string {
	return "joinVideoChat"
}
