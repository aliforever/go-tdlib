package entities

type InputFileLocal struct {
	Path string `json:"path"`
}

func (s InputFileLocal) Type() string {
	return "inputFileLocal"
}
