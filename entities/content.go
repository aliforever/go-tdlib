package entities

type Content struct {
	Type     string    `json:"@type"`
	Document *Document `json:"document"`
	Video    *Video    `json:"video"`
	Text     *Text     `json:"text"`
	Caption  struct {
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
