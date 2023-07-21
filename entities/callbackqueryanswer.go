package entities

type CallbackQueryAnswer struct {
	Text      string `json:"text"`
	ShowAlert bool   `json:"show_alert"`
	URL       string `json:"url"`
}
