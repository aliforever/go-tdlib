package outgoingevents

type GetProxyLink struct {
	ProxyID int64 `json:"proxy_id"`
}

func (p GetProxyLink) Type() string {
	return "getProxyLink"
}
