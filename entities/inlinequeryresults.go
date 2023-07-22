package entities

import "encoding/json"

/*
Inherited by inlineQueryResultAnimation, inlineQueryResultArticle, inlineQueryResultAudio, inlineQueryResultContact, inlineQueryResultDocument, inlineQueryResultGame, inlineQueryResultLocation, inlineQueryResultPhoto, inlineQueryResultSticker, inlineQueryResultVenue, inlineQueryResultVideo, and inlineQueryResultVoiceNote.
*/

type InlineQueryResultAnimation struct {
	ID        string    `json:"id"`
	Animation Animation `json:"animation"`
	Title     string    `json:"title"`
}

type InlineQueryResultArticle struct {
	ID          string      `json:"id"`
	URL         string      `json:"url"`
	HideURL     bool        `json:"hide_url"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Thumbnail   interface{} `json:"thumbnail"`
}

type InlineQueryResultAudio struct {
	ID    string      `json:"id"`
	Audio interface{} `json:"audio"`
}

type InlineQueryResultDocument struct {
	ID          string      `json:"id"`
	Document    interface{} `json:"document"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
}

type InlineQueryResultPhoto struct {
	ID          string      `json:"id"`
	Photo       interface{} `json:"photo"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
}

type InlineQueryResult struct {
	Type string `json:"@type"`

	*InlineQueryResultAnimation
	*InlineQueryResultArticle
	*InlineQueryResultAudio
	// *InlineQueryResultContact
	*InlineQueryResultDocument
	// *InlineQueryResultGame
	// *InlineQueryResultLocation
	*InlineQueryResultPhoto
}

// UnmarshalJSON implements json.Unmarshaler.
func (i *InlineQueryResult) UnmarshalJSON(data []byte) error {
	type baseInlineQueryResult struct {
		Type string `json:"@type"`
	}

	var b baseInlineQueryResult

	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}

	switch b.Type {
	case "inlineQueryResultAnimation":
		var inlineQueryResultAnimation InlineQueryResultAnimation

		if err := json.Unmarshal(data, &inlineQueryResultAnimation); err != nil {
			return err
		}

		i.InlineQueryResultAnimation = &inlineQueryResultAnimation
	case "inlineQueryResultArticle":
		var inlineQueryResultArticle InlineQueryResultArticle

		if err := json.Unmarshal(data, &inlineQueryResultArticle); err != nil {
			return err
		}

		i.InlineQueryResultArticle = &inlineQueryResultArticle
	case "inlineQueryResultAudio":
		var inlineQueryResultAudio InlineQueryResultAudio

		if err := json.Unmarshal(data, &inlineQueryResultAudio); err != nil {
			return err
		}

		i.InlineQueryResultAudio = &inlineQueryResultAudio
	case "inlineQueryResultDocument":
		var inlineQueryResultDocument InlineQueryResultDocument

		if err := json.Unmarshal(data, &inlineQueryResultDocument); err != nil {
			return err
		}

		i.InlineQueryResultDocument = &inlineQueryResultDocument
	case "inlineQueryResultPhoto":
		var inlineQueryResultPhoto InlineQueryResultPhoto

		if err := json.Unmarshal(data, &inlineQueryResultPhoto); err != nil {
			return err
		}

		i.InlineQueryResultPhoto = &inlineQueryResultPhoto
	}

	return nil
}

type InlineQueryResults struct {
	InlineQueryID     string              `json:"inline_query_id"`
	NextOffset        string              `json:"next_offset"`
	Results           []InlineQueryResult `json:"results"`
	SwitchPMText      string              `json:"switch_pm_text"`
	SwitchPMParameter string              `json:"switch_pm_parameter"`
}
