package entities

import "encoding/json"

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

type InlineKeyboardButtonTypeCallback struct {
	Type string `json:"@type"`
	Data string `json:"data"`
}

type InlineKeyboardButtonType struct {
	Type string `json:"@type"`

	Callback *InlineKeyboardButtonTypeCallback `json:"callback"`
}

// UnmarshalJSON Overrides UnmarshalJSON for InlineKeyboardButton
func (m *InlineKeyboardButtonType) UnmarshalJSON(b []byte) error {
	type baseMarkup struct {
		Type string          `json:"@type"`
		Data json.RawMessage `json:"type"`
	}

	var t baseMarkup

	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}

	rp := InlineKeyboardButtonType{Type: t.Type}

	switch t.Type {
	case "inlineKeyboardButtonTypeCallback":
		err = json.Unmarshal(t.Data, &rp.Callback)
	}

	if err != nil {
		return err
	}

	*m = rp

	return nil
}

type InlineKeyboardButton struct {
	Type string `json:"@type"`
	Text string `json:"text"`

	Data *InlineKeyboardButtonTypeCallback `json:"type"`
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
