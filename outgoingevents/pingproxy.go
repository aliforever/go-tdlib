package outgoingevents

type PingProxy struct {
	ProxyID int64 `json:"proxy_id"`
}

func (p PingProxy) Type() string {
	return "pingProxy"
}
