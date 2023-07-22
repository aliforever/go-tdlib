package entities

type FormattedText struct {
	Text     string       `json:"text"`
	Entities []TextEntity `json:"entities"`
}
