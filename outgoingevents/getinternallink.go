package outgoingevents

import "github.com/aliforever/go-tdlib/entities"

type GetInternalLink struct {
	LinkType entities.InternalLinkType `json:"_type"`
	IsHTTP   bool                      `json:"is_http"`
}

func (s GetInternalLink) Type() string {
	return "getInternalLink"
}
