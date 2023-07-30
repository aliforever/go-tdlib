package outgoingevents

type GetVideoChatAvailableParticipants struct {
	ChatID int64 `json:"chat_id"`
}

func (GetVideoChatAvailableParticipants) Type() string {
	return "getVideoChatAvailableParticipants"
}
