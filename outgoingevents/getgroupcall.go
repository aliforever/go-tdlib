package outgoingevents

type GetGroupCall struct {
	GroupCallID int64 `json:"group_call_id"`
}

func (GetGroupCall) Type() string {
	return "getGroupCall"
}
