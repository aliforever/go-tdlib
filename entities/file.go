package entities

type File struct {
	Type         string      `json:"@type"`
	ID           int64       `json:"id"`
	Size         int64       `json:"size"`
	ExpectedSize int64       `json:"expected_size"`
	Local        *FileLocal  `json:"local"`
	Remote       *FileRemote `json:"remote"`
}
