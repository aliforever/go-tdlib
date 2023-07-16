package entities

type InputMessageText struct {
	Text                  *FormattedText `json:"text"`
	DisableWebPagePreview bool           `json:"disable_web_page_preview"`
	ClearDraft            bool           `json:"clear_draft"`
}

func (s InputMessageText) Type() string {
	return "inputMessageText"
}
