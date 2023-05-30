package incomingevents

import "github.com/aliforever/go-tdlib/entities"

type GetMe struct {
	Event

	*entities.User `json:"user"`
}
