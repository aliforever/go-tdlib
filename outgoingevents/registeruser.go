package outgoingevents

type RegisterUser struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (s RegisterUser) Type() string {
	return "registerUser"
}
