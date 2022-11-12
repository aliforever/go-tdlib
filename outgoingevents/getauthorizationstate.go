package outgoingevents

type GetAuthorizationState struct{}

func (GetAuthorizationState) Type() string {
	return "getAuthorizationState"
}
