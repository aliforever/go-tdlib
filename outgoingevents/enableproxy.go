package outgoingevents

type EnableProxy struct {
	ProxyID int64 `json:"proxy_id"`
}

func (p EnableProxy) Type() string {
	return "enableProxy"
}
