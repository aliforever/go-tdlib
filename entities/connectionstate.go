package entities

type ConnectionStateType string

const (
	ConnectionStateTypeConnectionStateReady    ConnectionStateType = "connectionStateReady"
	ConnectionStateTypeConnectionStateUpdating ConnectionStateType = "connectionStateUpdating"
)

type ConnectionState struct {
	Type ConnectionStateType `json:"@type"`
}
