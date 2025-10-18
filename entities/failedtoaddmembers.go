package entities

type FailedToAddMembers struct {
	FailedToAddMembers []struct {
		UserID                        int64 `json:"user_id"`
		PremiumWouldAllowInvite       bool  `json:"premium_would_allow_invite"`
		PremiumRequiredToSendMessages bool  `json:"premium_required_to_send_messages"`
	} `json:"failed_to_add_members"`
}
