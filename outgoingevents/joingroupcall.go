package outgoingevents

import "github.com/aliforever/go-tdlib/entities"

type JoinGroupCall struct {
	GroupCallID   int64                   `json:"group_call_id"`
	ParticipantID *entities.MessageSender `json:"participant_id"`
	AudioSourceID int64                   `json:"audio_source_id"`
	Payload       string                  `json:"payload"`
	IsMuted       bool                    `json:"is_muted"`
	IsMyVideo     bool                    `json:"is_my_video_enabled"`
	InviteHash    string                  `json:"invite_hash"`
}

func (JoinGroupCall) Type() string {
	return "joinGroupCall"
}
