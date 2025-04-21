package entities

type MiniThumbnail struct {
	Type   string `json:"@type"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
	Data   string `json:"data"`
}

type PhotoSize struct {
	Type          string `json:"@type"`
	PhotoSizeType string `json:"type"`
	Width         int64  `json:"width"`
	Height        int64  `json:"height"`
	Photo         *File  `json:"photo"`
}

type Photo struct {
	Type          string         `json:"@type"`
	HasStickers   bool           `json:"has_stickers"`
	MiniThumbnail *MiniThumbnail `json:"mini_thumbnail"`
	Sizes         []PhotoSize    `json:"sizes"`
}
