package outgoingevents

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aliforever/go-tdlib/entities"
)

func TestAddProxy_MarshalJSON(t *testing.T) {
	ap := AddProxy{
		Server: "test",
		Port:   2020,
		Enable: true,
		Proxy: &entities.ProxyHttp{
			Username: "a",
			Password: "b",
			HttpOnly: false,
		},
	}

	j, _ := json.Marshal(ap)

	fmt.Println(string(j))
}
