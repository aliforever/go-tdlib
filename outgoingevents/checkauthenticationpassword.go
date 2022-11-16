package outgoingevents

type CheckAuthenticationPassword struct {
	Password string `json:"password"`
}

func (s CheckAuthenticationPassword) Type() string {
	return "checkAuthenticationPassword"
}
