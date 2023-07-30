package entities

type RecentSpeaker struct {
	ParticipantID *MessageSender `json:"participant_id"`
	IsSpeaking    bool           `json:"is_speaking"`
}
