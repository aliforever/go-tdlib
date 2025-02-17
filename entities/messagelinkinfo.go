package entities

type MessageLinkInfo struct {
	IsPublic        bool     `json:"is_public"`
	ChatID          int64    `json:"chat_id"`
	MessageThreadID int64    `json:"message_thread_id"`
	Message         *Message `json:"message"`
	MediaTimestamp  int64    `json:"media_timestamp"`
	ForAlbum        bool     `json:"for_album"`
}
