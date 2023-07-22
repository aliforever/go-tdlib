package entities

type CallbackQueryPayloadData struct {
	Data []byte `json:"data"`
}

// NewCallbackQueryPayloadData creates a new CallbackQueryPayloadData.
func NewCallbackQueryPayloadData(data string) *CallbackQueryPayload {
	return &CallbackQueryPayload{
		Type: "callbackQueryPayloadData",
		CallbackQueryPayloadData: &CallbackQueryPayloadData{
			Data: []byte(data),
		},
	}
}

type CallbackQueryPayload struct {
	Type string `json:"@type"`

	*CallbackQueryPayloadData
}
