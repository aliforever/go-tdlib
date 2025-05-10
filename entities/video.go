package entities

type AlternativeVideo struct {
	Type    string `json:"@type"`
	ID      int64  `json:"id"`
	Width   int64  `json:"width"`
	Height  int64  `json:"height"`
	Codec   string `json:"codec"`
	HlsFile *File  `json:"hls_file"`
	Video   *File  `json:"video"`
}

type Video struct {
	Duration          int64  `json:"duration"`
	Width             int64  `json:"width"`
	Height            int64  `json:"height"`
	FileName          string `json:"file_name"`
	MimeType          string `json:"mime_type"`
	HasStickers       bool   `json:"has_stickers"`
	SupportsStreaming bool   `json:"supports_streaming"`
	Minithumbnail     struct {
		Type   string `json:"@type"`
		Width  int64  `json:"width"`
		Height int64  `json:"height"`
		Data   string `json:"data"`
	} `json:"minithumbnail"`
	Thumbnail struct {
		Type   string `json:"@type"`
		Format struct {
			Type string `json:"@type"`
		} `json:"format"`
		Width  int64 `json:"width"`
		Height int64 `json:"height"`
		File   *File `json:"file"`
	} `json:"thumbnail"`
	Video *File `json:"video"`
}
