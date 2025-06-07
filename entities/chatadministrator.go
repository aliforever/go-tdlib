package entities

type ChatAdministrator struct {
	UserId      int64  `json:"user_id"`
	CustomTitle string `json:"custom_title"`
	IsOwner     bool   `json:"is_owner"`
}

type ChatAdministrators struct {
	Administrators []ChatAdministrator `json:"administrators"`
}
