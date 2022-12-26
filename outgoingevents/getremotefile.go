package outgoingevents

import "github.com/aliforever/go-tdlib/entities"

type GetRemoteFile struct {
	RemoteFileID string             `json:"remote_file_id"`
	FileType     *entities.FileType `json:"file_type,omitempty"`
}

func (s GetRemoteFile) Type() string {
	return "getRemoteFile"
}
