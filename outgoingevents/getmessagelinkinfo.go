package outgoingevents

type GetMessageLinkInfo struct {
	URL string `json:"url"`
}

func (s GetMessageLinkInfo) Type() string {
	return "getMessageLinkInfo"
}
