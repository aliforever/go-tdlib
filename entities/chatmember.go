package entities

type ChatMember struct {
	MemberID      *MessageSender `json:"member_id"`
	InviterUserID int64          `json:"inviter_user_id"`
	JoinChatDate  int64          `json:"join_chat_date"`
	Status        any            `json:"status"`
}
