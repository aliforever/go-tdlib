package outgoingevents

type GetUser struct {
	UserID int64 `json:"user_id"`
}

func (g GetUser) Type() string {
	return "getUser"
}
