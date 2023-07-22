package entities

type TextEntityType string

const (
	TextEntityTypeBold TextEntityType = "textEntityTypeBold"
)

type TextEntity struct {
	Type   TextEntityType `json:"@type"`
	Offset int64          `json:"offset"`
	Length int64          `json:"length"`
}
