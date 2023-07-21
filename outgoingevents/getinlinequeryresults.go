package outgoingevents

type GetInlineQueryResults struct {
	BotUserID    int64       `json:"bot_user_id"`
	ChatID       int64       `json:"chat_id"`
	UserLocation interface{} `json:"user_location"`
	Query        string      `json:"query"`
	Offset       string      `json:"offset"`
}

func (s GetInlineQueryResults) Type() string {
	return "getInlineQueryResults"
}
