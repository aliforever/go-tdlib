package entities

type ChatMemberStatus interface {
	Type() string
}

type ChatMemberStatusAdministrator struct {
	CustomTitle string                   `json:"custom_title"`
	CanBeEdited bool                     `json:"can_be_edited"`
	Rights      *ChatAdministratorRights `json:"rights"`
}

func (ChatMemberStatusAdministrator) Type() string {
	return "chatMemberStatusAdministrator"
}

type ChatAdministratorRights struct {
	CanManageChat       bool `json:"can_manage_chat"`
	CanChangeInfo       bool `json:"can_change_info"`
	CanPostMessages     bool `json:"can_post_messages"`
	CanEditMessages     bool `json:"can_edit_messages"`
	CanDeleteMessages   bool `json:"can_delete_messages"`
	CanInviteUsers      bool `json:"can_invite_users"`
	CanRestrictMembers  bool `json:"can_restrict_members"`
	CanPinMessages      bool `json:"can_pin_messages"`
	CanManageTopics     bool `json:"can_manage_topics"`
	CanPromoteMembers   bool `json:"can_promote_members"`
	CanManageVideoChats bool `json:"can_manage_video_chats"`
	CanPostStories      bool `json:"can_post_stories"`
	CanEditStories      bool `json:"can_edit_stories"`
	CanDeleteStories    bool `json:"can_delete_stories"`
	IsAnonymous         bool `json:"is_anonymous"`
}

type ChatMemberStatusBanned struct {
	BannedUntilDate int64 `json:"banned_until_date"`
}

func (ChatMemberStatusBanned) Type() string {
	return "chatMemberStatusBanned"
}
