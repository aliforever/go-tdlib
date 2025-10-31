package entities

type InternalLinkType interface {
	Type() string
}
type InternalLinkTypeMessage struct {
	URL string `json:"url"`
}

func (i InternalLinkTypeMessage) Type() string {
	return "internalLinkTypeMessage"
}
