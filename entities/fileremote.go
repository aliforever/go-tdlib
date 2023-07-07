package entities

type FileRemote struct {
	Type                 string `json:"@type"`
	Id                   string `json:"id"`
	UniqueId             string `json:"unique_id"`
	IsUploadingActive    bool   `json:"is_uploading_active"`
	IsUploadingCompleted bool   `json:"is_uploading_completed"`
	UploadedSize         int    `json:"uploaded_size"`
}
