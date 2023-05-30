package incomingevents

type SendMessage struct {
	Event

	AuthorSignature           string  `json:"author_signature"`
	AutoDeleteIn              float64 `json:"auto_delete_in"`
	CanBeDeletedForAllUsers   bool    `json:"can_be_deleted_for_all_users"`
	CanBeDeletedOnlyForSelf   bool    `json:"can_be_deleted_only_for_self"`
	CanBeEdited               bool    `json:"can_be_edited"`
	CanBeForwarded            bool    `json:"can_be_forwarded"`
	CanBeSaved                bool    `json:"can_be_saved"`
	CanGetAddedReactions      bool    `json:"can_get_added_reactions"`
	CanGetMediaTimestampLinks bool    `json:"can_get_media_timestamp_links"`
	CanGetMessageThread       bool    `json:"can_get_message_thread"`
	CanGetStatistics          bool    `json:"can_get_statistics"`
	CanGetViewers             bool    `json:"can_get_viewers"`
	CanReportReactions        bool    `json:"can_report_reactions"`
	ChatId                    int64   `json:"chat_id"`
	ContainsUnreadMention     bool    `json:"contains_unread_mention"`
	Content                   struct {
		Type string `json:"@type"`
		Text struct {
			Type     string        `json:"@type"`
			Entities []interface{} `json:"entities"`
			Text     string        `json:"text"`
		} `json:"text"`
	} `json:"content"`
	Date                int64   `json:"date"`
	EditDate            int64   `json:"edit_date"`
	HasTimestampedMedia bool    `json:"has_timestamped_media"`
	Id                  int64   `json:"id"`
	IsChannelPost       bool    `json:"is_channel_post"`
	IsOutgoing          bool    `json:"is_outgoing"`
	IsPinned            bool    `json:"is_pinned"`
	IsTopicMessage      bool    `json:"is_topic_message"`
	MediaAlbumId        string  `json:"media_album_id"`
	MessageThreadId     int64   `json:"message_thread_id"`
	ReplyInChatId       int64   `json:"reply_in_chat_id"`
	ReplyToMessageId    int64   `json:"reply_to_message_id"`
	RestrictionReason   string  `json:"restriction_reason"`
	SelfDestructIn      float64 `json:"self_destruct_in"`
	SelfDestructTime    int64   `json:"self_destruct_time"`
	SenderId            struct {
		Type   string `json:"@type"`
		UserId int64  `json:"user_id"`
	} `json:"sender_id"`
	SendingState struct {
		Type      string `json:"@type"`
		SendingId int64  `json:"sending_id"`
	} `json:"sending_state"`
	UnreadReactions []interface{} `json:"unread_reactions"`
	ViaBotUserId    int64         `json:"via_bot_user_id"`
}
