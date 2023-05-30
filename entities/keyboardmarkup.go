package entities

type ReplyMarkup interface {
	Type() string
}

type ReplyMarkupForceReply struct {
	IsPersonal            bool   `json:"is_personal"`
	InputFieldPlaceholder string `json:"input_field_placeholder"`
}

func (s ReplyMarkupForceReply) Type() string {
	return "replyMarkupForceReply"
}

type InlineKeyboardButton struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type ReplyMarkupInlineKeyboard struct {
	Rows [][]*InlineKeyboardButton `json:"rows"`
}

func (s ReplyMarkupInlineKeyboard) Type() string {
	return "replyMarkupInlineKeyboard"
}

type ReplyMarkupRemoveKeyboard struct {
	IsPersonal bool `json:"is_personal"`
}

func (s ReplyMarkupRemoveKeyboard) Type() string {
	return "replyMarkupRemoveKeyboard"
}

type KeyboardButtonType struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type ReplyMarkupShowKeyboard struct {
	IsPersonal            bool                    `json:"is_personal"`
	InputFieldPlaceholder string                  `json:"input_field_placeholder"`
	Rows                  [][]*KeyboardButtonType `json:"rows"`
	ResizeKeyboard        bool                    `json:"resize_keyboard"`
	OneTime               bool                    `json:"one_time"`
}

func (s ReplyMarkupShowKeyboard) Type() string {
	return "replyMarkupShowKeyboard"
}
