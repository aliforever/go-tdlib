package entities

type InputFileRemote struct {
	ID string `json:"id"`
}

func (s InputFileRemote) Type() string {
	return "inputFileRemote"
}
