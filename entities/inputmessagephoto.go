package entities

import (
	"encoding/json"
	"fmt"
)

type InputMessagePhoto struct {
	Photo     InputFile       `json:"photo"`
	Thumbnail *InputThumbnail `json:"thumbnail"`
	Width     int64           `json:"width"`
	Height    int64           `json:"height"`
	Caption   *FormattedText  `json:"caption"`
}

func (s InputMessagePhoto) Type() string {
	return "inputMessagePhoto"
}

func (s InputMessagePhoto) MarshalJSON() ([]byte, error) {
	type basicPhoto struct {
		Type      string          `json:"@type"`
		Thumbnail *InputThumbnail `json:"thumbnail"`
		Width     int64           `json:"width"`
		Height    int64           `json:"height"`
		Caption   *FormattedText  `json:"caption"`
	}

	b := basicPhoto{
		Type:      s.Type(),
		Thumbnail: s.Thumbnail,
		Width:     s.Width,
		Height:    s.Height,
		Caption:   s.Caption,
	}

	switch t := s.Photo.(type) {
	case *InputFileID:
		type InputPhotoFileID struct {
			Type string `json:"@type"`
			*InputFileID
		}

		return json.Marshal(&struct {
			*basicPhoto
			Photo *InputPhotoFileID `json:"photo"`
		}{
			basicPhoto: &b,
			Photo: &InputPhotoFileID{
				Type:        t.Type(),
				InputFileID: t,
			},
		})
	case *InputFileRemote:
		type InputPhotoFileRemote struct {
			Type string `json:"@type"`
			*InputFileRemote
		}

		return json.Marshal(&struct {
			*basicPhoto
			Photo *InputPhotoFileRemote `json:"photo"`
		}{
			basicPhoto: &b,
			Photo: &InputPhotoFileRemote{
				Type:            t.Type(),
				InputFileRemote: t,
			},
		})

	case *InputFileLocal:
		type InputPhotoFileLocal struct {
			Type string `json:"@type"`
			*InputFileLocal
		}

		return json.Marshal(&struct {
			*basicPhoto
			Photo *InputPhotoFileLocal `json:"photo"`
		}{
			basicPhoto: &b,
			Photo: &InputPhotoFileLocal{
				Type:           t.Type(),
				InputFileLocal: t,
			},
		})
	default:
		return nil, fmt.Errorf("invalid type")
	}
}
