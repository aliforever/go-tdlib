package entities

type LocalFile struct {
	Path                   string `json:"path"`
	CanBeDownloaded        bool   `json:"can_be_downloaded"`
	CanBeDeleted           bool   `json:"can_be_deleted"`
	IsDownloadingActive    bool   `json:"is_downloading_active"`
	IsDownloadingCompleted bool   `json:"is_downloading_completed"`
	DownloadOffset         int64  `json:"download_offset"`
	DownloadedPrefixSize   int64  `json:"downloaded_prefix_size"`
	DownloadedSize         int64  `json:"downloaded_size"`
}

type RemoteFile struct {
	ID                string `json:"id"`
	UniqueID          string `json:"unique_id"`
	IsUploadingActive bool   `json:"is_uploading_active"`
	UploadedSize      int64  `json:"uploaded_size"`
}

type File struct {
	ID           int64       `json:"id"`
	Size         int64       `json:"size"`
	ExpectedSize int64       `json:"expected_size"`
	Local        *LocalFile  `json:"local"`
	Remote       *RemoteFile `json:"remote"`
}
