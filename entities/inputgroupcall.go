package entities

type InputGroupCall interface {
	Type() string
}

type InputGroupCallLink struct {
	Link string `json:"link"`
}

func (i *InputGroupCallLink) Type() string {
	return "inputGroupCallLink"
}

type InputGroupCallMessage struct {
	ChatID    int64 `json:"chat_id"`
	MessageID int64 `json:"message_id"`
}

func (i *InputGroupCallMessage) Type() string {
	return "inputGroupCallMessage"
}

type GroupCallJoinParameters struct {
	AudioSourceID    int64  `json:"audio_source_id"`
	Payload          string `json:"payload"`
	IsMuted          bool   `json:"is_muted"`
	IsMyVideoEnabled bool   `json:"is_my_video_enabled"`
}
