package entities

type Video struct {
	Type              string `json:"@type"`
	Duration          int    `json:"duration"`
	Width             int    `json:"width"`
	Height            int    `json:"height"`
	FileName          string `json:"file_name"`
	MimeType          string `json:"mime_type"`
	HasStickers       bool   `json:"has_stickers"`
	SupportsStreaming bool   `json:"supports_streaming"`
	Minithumbnail     struct {
		Type   string `json:"@type"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
		Data   string `json:"data"`
	} `json:"minithumbnail"`
	Thumbnail struct {
		Type   string `json:"@type"`
		Format struct {
			Type string `json:"@type"`
		} `json:"format"`
		Width  int `json:"width"`
		Height int `json:"height"`
		File   struct {
			Type         string `json:"@type"`
			Id           int    `json:"id"`
			Size         int    `json:"size"`
			ExpectedSize int    `json:"expected_size"`
			Local        struct {
				Type                   string `json:"@type"`
				Path                   string `json:"path"`
				CanBeDownloaded        bool   `json:"can_be_downloaded"`
				CanBeDeleted           bool   `json:"can_be_deleted"`
				IsDownloadingActive    bool   `json:"is_downloading_active"`
				IsDownloadingCompleted bool   `json:"is_downloading_completed"`
				DownloadOffset         int    `json:"download_offset"`
				DownloadedPrefixSize   int    `json:"downloaded_prefix_size"`
				DownloadedSize         int    `json:"downloaded_size"`
			} `json:"local"`
			Remote struct {
				Type                 string `json:"@type"`
				Id                   string `json:"id"`
				UniqueId             string `json:"unique_id"`
				IsUploadingActive    bool   `json:"is_uploading_active"`
				IsUploadingCompleted bool   `json:"is_uploading_completed"`
				UploadedSize         int    `json:"uploaded_size"`
			} `json:"remote"`
		} `json:"file"`
	} `json:"thumbnail"`
	Video struct {
		Type         string `json:"@type"`
		Id           int    `json:"id"`
		Size         int    `json:"size"`
		ExpectedSize int    `json:"expected_size"`
		Local        struct {
			Type                   string `json:"@type"`
			Path                   string `json:"path"`
			CanBeDownloaded        bool   `json:"can_be_downloaded"`
			CanBeDeleted           bool   `json:"can_be_deleted"`
			IsDownloadingActive    bool   `json:"is_downloading_active"`
			IsDownloadingCompleted bool   `json:"is_downloading_completed"`
			DownloadOffset         int    `json:"download_offset"`
			DownloadedPrefixSize   int    `json:"downloaded_prefix_size"`
			DownloadedSize         int    `json:"downloaded_size"`
		} `json:"local"`
		Remote struct {
			Type                 string `json:"@type"`
			Id                   string `json:"id"`
			UniqueId             string `json:"unique_id"`
			IsUploadingActive    bool   `json:"is_uploading_active"`
			IsUploadingCompleted bool   `json:"is_uploading_completed"`
			UploadedSize         int    `json:"uploaded_size"`
		} `json:"remote"`
	} `json:"video"`
}
