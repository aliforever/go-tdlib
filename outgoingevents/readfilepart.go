package outgoingevents

type ReadFilePart struct {
	FileID int64 `json:"file_id"`
	Offset int64 `json:"offset"`
	Count  int64 `json:"count,omitempty"`
}

func (s ReadFilePart) Type() string {
	return "readFilePart"
}
