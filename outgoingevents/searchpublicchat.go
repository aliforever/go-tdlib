package outgoingevents

type SearchPublicChat struct {
	Username string `json:"username"`
}

func (s SearchPublicChat) Type() string {
	return "searchPublicChat"
}
