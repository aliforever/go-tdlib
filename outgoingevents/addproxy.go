package outgoingevents

import (
	"encoding/json"
	"fmt"

	"github.com/aliforever/go-tdlib/entities"
)

type AddProxy struct {
	Server string             `json:"server"`
	Port   int64              `json:"port"`
	Enable bool               `json:"enable"`
	Proxy  entities.ProxyType `json:"type"`
}

func (p AddProxy) Type() string {
	return "addProxy"
}

func (p AddProxy) MarshalJSON() ([]byte, error) {
	if p.Proxy == nil {
		return nil, fmt.Errorf("proxy not passed")
	}

	m := map[string]interface{}{
		"server": p.Server,
		"port":   p.Port,
		"enable": p.Enable,
	}

	proxyTypeJs, _ := json.Marshal(p.Proxy)

	proxyTypeMap := map[string]interface{}{}

	err := json.Unmarshal(proxyTypeJs, &proxyTypeMap)
	if err != nil {
		return nil, err
	}

	proxyTypeMap["@type"] = p.Proxy.Type()

	m["type"] = proxyTypeMap

	return json.Marshal(m)
}
