package entities

import (
	"encoding/json"
	"fmt"
)

type InputMessageVideo struct {
	Video             InputFile       `json:"video"`
	Thumbnail         *InputThumbnail `json:"thumbnail"`
	Duration          int64           `json:"duration"`
	Width             int64           `json:"width"`
	Height            int64           `json:"height"`
	SupportsStreaming bool            `json:"supports_streaming"`
	Caption           *FormattedText  `json:"caption"`
	Ttl               int64           `json:"ttl"`
}

func (s InputMessageVideo) Type() string {
	return "inputMessageVideo"
}

func (s InputMessageVideo) MarshalJSON() ([]byte, error) {
	type basicVideo struct {
		Type              string          `json:"@type"`
		Thumbnail         *InputThumbnail `json:"thumbnail"`
		Duration          int64           `json:"duration"`
		Width             int64           `json:"width"`
		Height            int64           `json:"height"`
		SupportsStreaming bool            `json:"supports_streaming"`
		Caption           *FormattedText  `json:"caption"`
		Ttl               int64           `json:"ttl"`
	}

	b := basicVideo{
		Type:              s.Type(),
		Thumbnail:         s.Thumbnail,
		Duration:          s.Duration,
		Width:             s.Width,
		Height:            s.Height,
		SupportsStreaming: s.SupportsStreaming,
		Caption:           s.Caption,
		Ttl:               s.Ttl,
	}

	switch t := s.Video.(type) {
	case *InputFileID:
		type InputVideoFileID struct {
			Type string `json:"@type"`
			*InputFileID
		}

		return json.Marshal(&struct {
			*basicVideo
			Video *InputVideoFileID `json:"video"`
		}{
			basicVideo: &b,
			Video: &InputVideoFileID{
				Type:        t.Type(),
				InputFileID: t,
			},
		})
	case *InputFileRemote:
		type InputVideoFileRemote struct {
			Type string `json:"@type"`
			*InputFileRemote
		}

		return json.Marshal(&struct {
			*basicVideo
			Video *InputVideoFileRemote `json:"video"`
		}{
			basicVideo: &b,
			Video: &InputVideoFileRemote{
				Type:            t.Type(),
				InputFileRemote: t,
			},
		})

	case *InputFileLocal:
		type InputVideoFileLocal struct {
			Type string `json:"@type"`
			*InputFileLocal
		}

		return json.Marshal(&struct {
			*basicVideo
			Video *InputVideoFileLocal `json:"video"`
		}{
			basicVideo: &b,
			Video: &InputVideoFileLocal{
				Type:           t.Type(),
				InputFileLocal: t,
			},
		})
	default:
		return nil, fmt.Errorf("invalid type")
	}
}
