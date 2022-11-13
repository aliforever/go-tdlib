package entities

type StateType string

const (
	StateTypeAuthorizationStateAwaitingTdlibParameters StateType = "authorizationStateWaitTdlibParameters"
	StateTypeAuthorizationStateAwaitingPhoneNumber     StateType = "authorizationStateWaitPhoneNumber"
	StateTypeAuthorizationStateAwaitingCode            StateType = "authorizationStateWaitCode"
	StateTypeAuthorizationStateReady                   StateType = "authorizationStateReady"

	StateTypeConnectionStateReady    StateType = "connectionStateReady"
	StateTypeConnectionStateUpdating StateType = "connectionStateUpdating"
)

type State struct {
	Type StateType `json:"@type"`
}
