package entities

type InputThumbnail struct {
	Thumbnail InputFile `json:"thumbnail"`
	Width     int64     `json:"width"`
	Height    int64     `json:"height"`
}
