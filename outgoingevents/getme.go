package outgoingevents

type GetMe struct {
}

func (s GetMe) Type() string {
	return "getMe"
}
