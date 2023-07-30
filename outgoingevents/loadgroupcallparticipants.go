package outgoingevents

type LoadGroupCallParticipants struct {
	GroupCallID int64 `json:"group_call_id"`
	Limit       int64 `json:"limit"`
}

func (LoadGroupCallParticipants) Type() string {
	return "loadGroupCallParticipants"
}
