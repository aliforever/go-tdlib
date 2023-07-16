package entities

type InputFileID struct {
	ID int64 `json:"id"`
}

func (s InputFileID) Type() string {
	return "inputFileId"
}
