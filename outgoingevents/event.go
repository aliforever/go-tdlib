package outgoingevents

import (
	"encoding/json"
)

type EventInterface interface {
	Type() string
}

func NewEventJSON(requestID string, data EventInterface) (string, error) {
	m := map[string]interface{}{}

	js, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(js, &m)
	if err != nil {
		return "", err
	}

	m["@type"] = data.Type()
	m["@extra"] = requestID

	js, err = json.Marshal(m)
	if err != nil {
		return "", err
	}

	return string(js), nil
}
