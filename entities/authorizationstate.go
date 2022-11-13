package entities

type AuthorizationStateType string

const (
	AuthorizationStateTypeAwaitingTdlibParameters AuthorizationStateType = "authorizationStateWaitTdlibParameters"
	AuthorizationStateTypeAwaitingPhoneNumber     AuthorizationStateType = "authorizationStateWaitPhoneNumber"
	AuthorizationStateTypeAwaitingCode            AuthorizationStateType = "authorizationStateWaitCode"
	AuthorizationStateTypeReady                   AuthorizationStateType = "authorizationStateReady"
)

type AuthorizationState struct {
	Type AuthorizationStateType `json:"@type"`
}
