package incomingevents

type GetChatResponse struct {
	Event

	Id   int64 `json:"id"`
	Type struct {
		Type         string `json:"@type"`
		SupergroupId int    `json:"supergroup_id"`
		IsChannel    bool   `json:"is_channel"`
	} `json:"type"`
	Title       string `json:"title"`
	Permissions struct {
		Type                  string `json:"@type"`
		CanSendMessages       bool   `json:"can_send_messages"`
		CanSendMediaMessages  bool   `json:"can_send_media_messages"`
		CanSendPolls          bool   `json:"can_send_polls"`
		CanSendOtherMessages  bool   `json:"can_send_other_messages"`
		CanAddWebPagePreviews bool   `json:"can_add_web_page_previews"`
		CanChangeInfo         bool   `json:"can_change_info"`
		CanInviteUsers        bool   `json:"can_invite_users"`
		CanPinMessages        bool   `json:"can_pin_messages"`
		CanManageTopics       bool   `json:"can_manage_topics"`
	} `json:"permissions"`
	LastMessage struct {
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
		Content           struct {
			Type  string `json:"@type"`
			Video struct {
				Type              string `json:"@type"`
				Duration          int    `json:"duration"`
				Width             int    `json:"width"`
				Height            int    `json:"height"`
				FileName          string `json:"file_name"`
				MimeType          string `json:"mime_type"`
				HasStickers       bool   `json:"has_stickers"`
				SupportsStreaming bool   `json:"supports_streaming"`
				Minithumbnail     struct {
					Type   string `json:"@type"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
					Data   string `json:"data"`
				} `json:"minithumbnail"`
				Thumbnail struct {
					Type   string `json:"@type"`
					Format struct {
						Type string `json:"@type"`
					} `json:"format"`
					Width  int `json:"width"`
					Height int `json:"height"`
					File   struct {
						Type         string `json:"@type"`
						Id           int    `json:"id"`
						Size         int    `json:"size"`
						ExpectedSize int    `json:"expected_size"`
						Local        struct {
							Type                   string `json:"@type"`
							Path                   string `json:"path"`
							CanBeDownloaded        bool   `json:"can_be_downloaded"`
							CanBeDeleted           bool   `json:"can_be_deleted"`
							IsDownloadingActive    bool   `json:"is_downloading_active"`
							IsDownloadingCompleted bool   `json:"is_downloading_completed"`
							DownloadOffset         int    `json:"download_offset"`
							DownloadedPrefixSize   int    `json:"downloaded_prefix_size"`
							DownloadedSize         int    `json:"downloaded_size"`
						} `json:"local"`
						Remote struct {
							Type                 string `json:"@type"`
							Id                   string `json:"id"`
							UniqueId             string `json:"unique_id"`
							IsUploadingActive    bool   `json:"is_uploading_active"`
							IsUploadingCompleted bool   `json:"is_uploading_completed"`
							UploadedSize         int    `json:"uploaded_size"`
						} `json:"remote"`
					} `json:"file"`
				} `json:"thumbnail"`
				Video struct {
					Type         string `json:"@type"`
					Id           int    `json:"id"`
					Size         int    `json:"size"`
					ExpectedSize int    `json:"expected_size"`
					Local        struct {
						Type                   string `json:"@type"`
						Path                   string `json:"path"`
						CanBeDownloaded        bool   `json:"can_be_downloaded"`
						CanBeDeleted           bool   `json:"can_be_deleted"`
						IsDownloadingActive    bool   `json:"is_downloading_active"`
						IsDownloadingCompleted bool   `json:"is_downloading_completed"`
						DownloadOffset         int    `json:"download_offset"`
						DownloadedPrefixSize   int    `json:"downloaded_prefix_size"`
						DownloadedSize         int    `json:"downloaded_size"`
					} `json:"local"`
					Remote struct {
						Type                 string `json:"@type"`
						Id                   string `json:"id"`
						UniqueId             string `json:"unique_id"`
						IsUploadingActive    bool   `json:"is_uploading_active"`
						IsUploadingCompleted bool   `json:"is_uploading_completed"`
						UploadedSize         int    `json:"uploaded_size"`
					} `json:"remote"`
				} `json:"video"`
			} `json:"video"`
			Caption struct {
				Type     string        `json:"@type"`
				Text     string        `json:"text"`
				Entities []interface{} `json:"entities"`
			} `json:"caption"`
			IsSecret bool `json:"is_secret"`
		} `json:"content"`
	} `json:"last_message"`
	Positions                  []interface{} `json:"positions"`
	HasProtectedContent        bool          `json:"has_protected_content"`
	IsMarkedAsUnread           bool          `json:"is_marked_as_unread"`
	IsBlocked                  bool          `json:"is_blocked"`
	HasScheduledMessages       bool          `json:"has_scheduled_messages"`
	CanBeDeletedOnlyForSelf    bool          `json:"can_be_deleted_only_for_self"`
	CanBeDeletedForAllUsers    bool          `json:"can_be_deleted_for_all_users"`
	CanBeReported              bool          `json:"can_be_reported"`
	DefaultDisableNotification bool          `json:"default_disable_notification"`
	UnreadCount                int           `json:"unread_count"`
	LastReadInboxMessageId     int           `json:"last_read_inbox_message_id"`
	LastReadOutboxMessageId    int           `json:"last_read_outbox_message_id"`
	UnreadMentionCount         int           `json:"unread_mention_count"`
	UnreadReactionCount        int           `json:"unread_reaction_count"`
	ActionBar                  struct {
		Type         string `json:"@type"`
		CanUnarchive bool   `json:"can_unarchive"`
	} `json:"action_bar"`
	NotificationSettings struct {
		Type                                        string `json:"@type"`
		UseDefaultMuteFor                           bool   `json:"use_default_mute_for"`
		MuteFor                                     int    `json:"mute_for"`
		UseDefaultSound                             bool   `json:"use_default_sound"`
		SoundId                                     string `json:"sound_id"`
		UseDefaultShowPreview                       bool   `json:"use_default_show_preview"`
		ShowPreview                                 bool   `json:"show_preview"`
		UseDefaultDisablePinnedMessageNotifications bool   `json:"use_default_disable_pinned_message_notifications"`
		DisablePinnedMessageNotifications           bool   `json:"disable_pinned_message_notifications"`
		UseDefaultDisableMentionNotifications       bool   `json:"use_default_disable_mention_notifications"`
		DisableMentionNotifications                 bool   `json:"disable_mention_notifications"`
	} `json:"notification_settings"`
	AvailableReactions struct {
		Type string `json:"@type"`
	} `json:"available_reactions"`
	MessageTtl int    `json:"message_ttl"`
	ThemeName  string `json:"theme_name"`
	VideoChat  struct {
		Type            string `json:"@type"`
		GroupCallId     int    `json:"group_call_id"`
		HasParticipants bool   `json:"has_participants"`
	} `json:"video_chat"`
	ReplyMarkupMessageId int    `json:"reply_markup_message_id"`
	ClientData           string `json:"client_data"`
	Extra                string `json:"@extra"`
}
