package entities

type CallbackQueryPayloadData struct {
	Data []byte `json:"data"`
}

// NewCallbackQueryPayloadData creates a new CallbackQueryPayloadData.
func NewCallbackQueryPayloadData(data []byte) *CallbackQueryPayload {
	return &CallbackQueryPayload{
		Type: "callbackQueryPayloadData",
		CallbackQueryPayloadData: &CallbackQueryPayloadData{
			Data: data,
		},
	}
}

type CallbackQueryPayload struct {
	Type string `json:"@type"`

	*CallbackQueryPayloadData
}
