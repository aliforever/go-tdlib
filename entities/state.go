package entities

type StateType string

const (
	StateTypeAwaitingTdlibParameters StateType = "authorizationStateWaitTdlibParameters"
	StateTypeAwaitingPhoneNumber     StateType = "authorizationStateWaitPhoneNumber"
	StateTypeAwaitingCode            StateType = "authorizationStateWaitCode"
	StateTypeReady                   StateType = "authorizationStateReady"
)

type State struct {
	Type StateType `json:"@type"`
}
