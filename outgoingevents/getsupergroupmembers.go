package outgoingevents

import "encoding/json"

type SupergroupMembersFilter interface {
	Type() string
}

type GetSuperGroupMembers struct {
	SupergroupID int64                   `json:"supergroup_id"`
	Filter       SupergroupMembersFilter `json:"filter"`
	Offset       int64                   `json:"offset"`
	Limit        int64                   `json:"limit"`
}

func (s GetSuperGroupMembers) Type() string {
	return "getSupergroupMembers"
}

func (s GetSuperGroupMembers) MarshalJSON() ([]byte, error) {
	j, _ := json.Marshal(s.Filter)

	var filterData map[string]interface{}

	err := json.Unmarshal(j, &filterData)
	if err != nil {
		return nil, err
	}

	filterData["@type"] = s.Filter.Type()

	return json.Marshal(map[string]interface{}{
		"supergroup_id": s.SupergroupID,
		"filter":        filterData,
		"offset":        s.Offset,
		"limit":         s.Limit,
	})
}
