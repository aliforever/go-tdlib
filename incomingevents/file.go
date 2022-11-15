package incomingevents

type DownloadFileResponse struct {
	Id           int32 `json:"id"`
	Size         int   `json:"size"`
	ExpectedSize int   `json:"expected_size"`
	Local        struct {
		Type                   string `json:"@type"`
		Path                   string `json:"path"`
		CanBeDownloaded        bool   `json:"can_be_downloaded"`
		CanBeDeleted           bool   `json:"can_be_deleted"`
		IsDownloadingActive    bool   `json:"is_downloading_active"`
		IsDownloadingCompleted bool   `json:"is_downloading_completed"`
		DownloadOffset         int    `json:"download_offset"`
		DownloadedPrefixSize   int    `json:"downloaded_prefix_size"`
		DownloadedSize         int64  `json:"downloaded_size"`
	} `json:"local"`
	Remote struct {
		Type                 string `json:"@type"`
		Id                   string `json:"id"`
		UniqueId             string `json:"unique_id"`
		IsUploadingActive    bool   `json:"is_uploading_active"`
		IsUploadingCompleted bool   `json:"is_uploading_completed"`
		UploadedSize         int    `json:"uploaded_size"`
	} `json:"remote"`
}
