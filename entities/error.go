package entities

type Error struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}
