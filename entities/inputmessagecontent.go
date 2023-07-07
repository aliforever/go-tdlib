package entities

type InputMessageContent interface {
	Type() string
}

type InputMessageText struct {
	Text                  *FormattedText `json:"text"`
	DisableWebPagePreview bool           `json:"disable_web_page_preview"`
	ClearDraft            bool           `json:"clear_draft"`
}

// NewInputMessageFormattedText creates a new InputMessageText
func NewInputMessageFormattedText(
	text string,
	disableWebPagePreview bool,
	clearDraft bool,
	entities []*TextEntity,
) *InputMessageText {
	return &InputMessageText{
		Text:                  &FormattedText{Text: text, Entities: entities},
		DisableWebPagePreview: disableWebPagePreview,
		ClearDraft:            clearDraft,
	}
}

func (s InputMessageText) Type() string {
	return "inputMessageText"
}
