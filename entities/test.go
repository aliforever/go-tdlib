package entities

type Text struct {
	Type     string        `json:"@type"`
	Text     string        `json:"text"`
	Entities []interface{} `json:"entities"`
}
