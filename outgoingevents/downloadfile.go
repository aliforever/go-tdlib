package outgoingevents

type DownloadFile struct {
	FileID      int64 `json:"file_id"`
	Priority    int64 `json:"priority"`
	Offset      int64 `json:"offset"`
	Limit       int64 `json:"limit"`
	Synchronous bool  `json:"synchronous"`
}

func (s DownloadFile) Type() string {
	return "downloadFile"
}
