package entities

type SupergroupMembersFilterAdministrators struct {
}

func (s *SupergroupMembersFilterAdministrators) Type() string {
	return "supergroupMembersFilterAdministrators"
}

type SupergroupMembersFilterBanned struct {
	Query string `json:"query"`
}

func (s *SupergroupMembersFilterBanned) Type() string {
	return "supergroupMembersFilterBanned"
}

type SupergroupMembersFilterBots struct {
}

func (s *SupergroupMembersFilterBots) Type() string {
	return "supergroupMembersFilterBots"
}

type SupergroupMembersFilterContacts struct {
	Query string `json:"query"`
}

func (s *SupergroupMembersFilterContacts) Type() string {
	return "supergroupMembersFilterContacts"
}

type SupergroupMembersFilterMention struct {
	Query           string `json:"query"`
	MessageThreadID int64  `json:"message_thread_id"`
}

func (s *SupergroupMembersFilterMention) Type() string {
	return "supergroupMembersFilterMention"
}

type SupergroupMembersFilterRecent struct {
}

func (s *SupergroupMembersFilterRecent) Type() string {
	return "supergroupMembersFilterRecent"
}

type SupergroupMembersFilterRestricted struct {
	Query string `json:"query"`
}

func (s *SupergroupMembersFilterRestricted) Type() string {
	return "supergroupMembersFilterRestricted"
}

type SupergroupMembersFilterSearch struct {
	Query string `json:"query"`
}

func (s *SupergroupMembersFilterSearch) Type() string {
	return "supergroupMembersFilterSearch"
}
