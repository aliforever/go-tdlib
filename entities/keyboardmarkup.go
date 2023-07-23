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

type InlineKeyboardButton struct {
	Type string `json:"@type"`
	Text string `json:"text"`

	Callback *InlineKeyboardButtonTypeCallback `json:"callback"`

	Unknown interface{} `json:"unknown"`
}

// UnmarshalJSON Overrides UnmarshalJSON for InlineKeyboardButton
func (m *InlineKeyboardButton) UnmarshalJSON(b []byte) error {
	type baseData struct {
		Type string `json:"@type"`
	}

	type baseMarkup struct {
		Type string          `json:"@type"`
		Data json.RawMessage `json:"type"`
	}

	var t baseMarkup

	var bData baseData

	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}

	m.Type = t.Type

	if t.Data == nil || len(t.Data) == 0 {
		return nil
	}

	err = json.Unmarshal(t.Data, &bData)
	if err != nil {
		return err
	}

	rp := InlineKeyboardButton{Type: t.Type}

	switch bData.Type {
	case "inlineKeyboardButtonTypeCallback":
		err = json.Unmarshal(t.Data, &rp.Callback)
	default:
		rp.Unknown = t.Data
	}

	if err != nil {
		return err
	}

	*m = rp

	return nil
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
