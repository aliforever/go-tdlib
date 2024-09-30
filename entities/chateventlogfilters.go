package entities

type ChatEventLogFilters struct {
	// True, if message edits need to be returned.
	MessageEdits bool `json:"message_edits"`
	// True, if message deletions need to be returned.
	MessageDeletions bool `json:"message_deletions"`
	// True, if pin/unpin events need to be returned.
	MessagePins bool `json:"message_pins"`
	// True, if members joining events need to be returned.
	MemberJoins bool `json:"member_joins"`
	// True, if members leaving events need to be returned.
	MemberLeaves bool `json:"member_leaves"`
	// True, if invited member events need to be returned.
	MemberInvites bool `json:"member_invites"`
	// True, if member promotion/demotion events need to be returned.
	MemberPromotions bool `json:"member_promotions"`
	// True, if member restricted/unrestricted/banned/unbanned events need to be returned.
	MemberRestrictions bool `json:"member_restrictions"`
	// True, if changes in chat information need to be returned.
	InfoChanges bool `json:"info_changes"`
	// True, if changes in chat settings need to be returned.
	SettingChanges bool `json:"setting_changes"`
	// True, if changes to invite links need to be returned.
	InviteLinkChanges bool `json:"invite_link_changes"`
	// True, if video chat actions need to be returned.
	VideoChatChanges bool `json:"video_chat_changes"`
	// True, if forum-related actions need to be returned.
	ForumChanges bool `json:"forum_changes"`
	// True, if subscription extensions need to be returned.
	SubscriptionExtensions bool `json:"subscription_extensions"`
}
