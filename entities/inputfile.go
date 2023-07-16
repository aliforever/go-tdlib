package entities

type InputFile interface {
	Type() string
}

func NewInputFileID(fileID int64) InputFile {
	return &InputFileID{fileID}
}

func NewInputFileRemote(fileID string) InputFile {
	return &InputFileRemote{fileID}
}

func NewInputFileLocal(path string) InputFile {
	return &InputFileLocal{path}
}
