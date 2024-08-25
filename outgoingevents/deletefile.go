package outgoingevents

type DeleteFile struct {
	FileID int64 `json:"file_id"`
}

func (s DeleteFile) Type() string {
	return "deleteFile"
}
