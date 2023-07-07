package incomingevents

import "github.com/aliforever/go-tdlib/entities"

type UpdateFile struct {
	Type string         `json:"@type"`
	File *entities.File `json:"file"`
}
