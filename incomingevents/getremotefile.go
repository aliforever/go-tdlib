package incomingevents

import "github.com/aliforever/go-tdlib/entities"

type GetRemoteFile struct {
	Event

	*entities.File `json:"file"`
}
