package outgoingevents

type CheckAuthenticationBotToken struct {
	Token string `json:"token"`
}

func (s CheckAuthenticationBotToken) Type() string {
	return "checkAuthenticationBotToken"
}
