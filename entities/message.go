package entities

type Message struct {
	Type     string `json:"@type"`
	Id       int    `json:"id"`
	SenderId struct {
		Type   string `json:"@type"`
		UserId int64  `json:"user_id"`
	} `json:"sender_id"`
	ChatId                    int64 `json:"chat_id"`
	IsOutgoing                bool  `json:"is_outgoing"`
	IsPinned                  bool  `json:"is_pinned"`
	CanBeEdited               bool  `json:"can_be_edited"`
	CanBeForwarded            bool  `json:"can_be_forwarded"`
	CanBeSaved                bool  `json:"can_be_saved"`
	CanBeDeletedOnlyForSelf   bool  `json:"can_be_deleted_only_for_self"`
	CanBeDeletedForAllUsers   bool  `json:"can_be_deleted_for_all_users"`
	CanGetAddedReactions      bool  `json:"can_get_added_reactions"`
	CanGetStatistics          bool  `json:"can_get_statistics"`
	CanGetMessageThread       bool  `json:"can_get_message_thread"`
	CanGetViewers             bool  `json:"can_get_viewers"`
	CanGetMediaTimestampLinks bool  `json:"can_get_media_timestamp_links"`
	CanReportReactions        bool  `json:"can_report_reactions"`
	HasTimestampedMedia       bool  `json:"has_timestamped_media"`
	IsChannelPost             bool  `json:"is_channel_post"`
	IsTopicMessage            bool  `json:"is_topic_message"`
	ContainsUnreadMention     bool  `json:"contains_unread_mention"`
	Date                      int   `json:"date"`
	EditDate                  int   `json:"edit_date"`
	InteractionInfo           struct {
		Type         string `json:"@type"`
		ViewCount    int    `json:"view_count"`
		ForwardCount int    `json:"forward_count"`
		ReplyInfo    struct {
			Type                    string        `json:"@type"`
			ReplyCount              int           `json:"reply_count"`
			RecentReplierIds        []interface{} `json:"recent_replier_ids"`
			LastReadInboxMessageId  int           `json:"last_read_inbox_message_id"`
			LastReadOutboxMessageId int           `json:"last_read_outbox_message_id"`
			LastMessageId           int           `json:"last_message_id"`
		} `json:"reply_info"`
		Reactions []interface{} `json:"reactions"`
	} `json:"interaction_info"`
	UnreadReactions   []interface{} `json:"unread_reactions"`
	ReplyInChatId     int           `json:"reply_in_chat_id"`
	ReplyToMessageId  int           `json:"reply_to_message_id"`
	MessageThreadId   int           `json:"message_thread_id"`
	Ttl               int           `json:"ttl"`
	TtlExpiresIn      float64       `json:"ttl_expires_in"`
	ViaBotUserId      int           `json:"via_bot_user_id"`
	AuthorSignature   string        `json:"author_signature"`
	MediaAlbumId      string        `json:"media_album_id"`
	RestrictionReason string        `json:"restriction_reason"`
	Content           Content       `json:"content"`
}
