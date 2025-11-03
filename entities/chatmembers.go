package entities

type ChatMembers struct {
	TotalCount int64        `json:"total_count"`
	Members    []ChatMember `json:"members"`
}
