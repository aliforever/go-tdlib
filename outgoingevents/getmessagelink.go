package outgoingevents

type GetMessageLink struct {
	ChatID          int64  `json:"chat_id"`
	MessageID       int64  `json:"message_id"`
	MediaTimestamp  *int64 `json:"media_timestamp,omitempty"`
	ForAlbum        *bool  `json:"for_album,omitempty"`
	InMessageThread *bool  `json:"in_message_thread,omitempty"`
}

func (s GetMessageLink) Type() string {
	return "getMessageLink"
}
