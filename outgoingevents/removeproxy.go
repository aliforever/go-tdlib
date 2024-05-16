package outgoingevents

type RemoveProxy struct {
	ProxyID int64 `json:"proxy_id"`
}

func (p RemoveProxy) Type() string {
	return "removeProxy"
}
