package outgoingevents

type DownloadFile struct {
	FileID      int32 `json:"file_id"`
	Priority    int32 `json:"priority"`
	Offset      int32 `json:"offset"`
	Limit       int32 `json:"limit"`
	Synchronous bool  `json:"synchronous"`
}

func (s DownloadFile) Type() string {
	return "downloadFile"
}
