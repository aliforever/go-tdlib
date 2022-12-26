package incomingevents

import "github.com/aliforever/go-tdlib/entities"

type GetRemoteFile struct {
	Event

	File *entities.File `json:"file"`
}
