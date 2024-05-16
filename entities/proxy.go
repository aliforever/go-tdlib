package entities

import (
	"encoding/json"
	"fmt"
)

type ProxyType interface {
	Type() string
}

type Proxy struct {
	ID           int64  `json:"proxy_id"`
	Server       string `json:"server"`
	Port         int64  `json:"port"`
	LastUsedDate int64  `json:"last_used_date"`
	IsEnabled    bool   `json:"is_enabled"`

	Http    *ProxyHttp    `json:"http"`
	Mtproto *ProxyMtproto `json:"mtproto"`
	Socks5  *ProxySocks5  `json:"socks5"`
}

func (s *Proxy) UnmarshalJSON(data []byte) error {
	// Intermediate struct to capture common fields and raw type field
	t := struct {
		ID           int64           `json:"id"`
		Server       string          `json:"server"`
		Port         int64           `json:"port"`
		LastUsedDate int64           `json:"last_used_date"`
		IsEnabled    bool            `json:"is_enabled"`
		Type         json.RawMessage `json:"type"`
	}{}

	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	// Intermediate struct to capture the @type field within the type field
	tType := struct {
		Type string `json:"@type"`
	}{}

	if err := json.Unmarshal(t.Type, &tType); err != nil {
		return err
	}

	// Initialize the Proxy struct with common fields
	proxy := Proxy{
		ID:           t.ID,
		Server:       t.Server,
		Port:         t.Port,
		LastUsedDate: t.LastUsedDate,
		IsEnabled:    t.IsEnabled,
	}

	var err error

	switch tType.Type {
	case "proxyTypeHttp":
		err = json.Unmarshal(t.Type, &proxy.Http)
	case "proxyTypeMtproto":
		err = json.Unmarshal(t.Type, &proxy.Mtproto)
	case "proxyTypeSocks5":
		err = json.Unmarshal(t.Type, &proxy.Socks5)
	default:
		err = fmt.Errorf("unknown proxy type: %s", tType.Type)
	}

	if err != nil {
		return err
	}

	*s = proxy

	return nil
}

type ProxyHttp struct {
	Username string `json:"username"`
	Password string `json:"password"`
	HttpOnly bool   `json:"http_only"`
}

func (s ProxyHttp) Type() string {
	return "proxyTypeHttp"
}

type ProxyMtproto struct {
	Secret string `json:"secret"`
}

func (s ProxyMtproto) Type() string {
	return "proxyTypeMtproto"
}

type ProxySocks5 struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s ProxySocks5) Type() string {
	return "proxyTypeSocks5"
}
