package entities

type FileLocal struct {
	Type                   string `json:"@type"`
	Path                   string `json:"path"`
	CanBeDownloaded        bool   `json:"can_be_downloaded"`
	CanBeDeleted           bool   `json:"can_be_deleted"`
	IsDownloadingActive    bool   `json:"is_downloading_active"`
	IsDownloadingCompleted bool   `json:"is_downloading_completed"`
	DownloadOffset         int    `json:"download_offset"`
	DownloadedPrefixSize   int    `json:"downloaded_prefix_size"`
	DownloadedSize         int64  `json:"downloaded_size"`
}
