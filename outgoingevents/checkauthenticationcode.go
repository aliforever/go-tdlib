package outgoingevents

type CheckAuthenticationCode struct {
	Code string `json:"code"`
}

func (s CheckAuthenticationCode) Type() string {
	return "checkAuthenticationCode"
}
