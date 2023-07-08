package entities

const (
	ContentTypeText      = "messageText"
	ContentTypeAnimation = "messageAnimation"
	ContentTypePhoto     = "messagePhoto"
	ContentTypeDocument  = "messageDocument"
	ContentTypeVideo     = "messageVideo"
	ContentTypeVideoNote = "messageVideoNote"
	ContentTypeVoiceNote = "messageVoiceNote"
	ContentTypeAudio     = "messageAudio"
)

type Content struct {
	Type      string     `json:"@type"`
	Animation *Animation `json:"animation"`
	Document  *Document  `json:"document"`
	Video     *Video     `json:"video"`
	Audio     *Audio     `json:"audio"`
	VoiceNote *VoiceNote `json:"voice_note"`
	Text      *Text      `json:"text"`
	Caption   struct {
		Type     string `json:"@type"`
		Text     string `json:"text"`
		Entities []struct {
			Type   string `json:"@type"`
			Offset int    `json:"offset"`
			Length int    `json:"length"`
			Type1  struct {
				Type string `json:"@type"`
			} `json:"type"`
		} `json:"entities"`
	} `json:"caption"`
}
