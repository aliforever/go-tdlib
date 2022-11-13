package outgoingevents

import "github.com/aliforever/go-tdlib/entities"

type SetAuthenticationPhoneNumber struct {
	PhoneNumber string                                      `json:"phone_number"`
	Settings    *entities.PhoneNumberAuthenticationSettings `json:"settings"`
}

func (s SetAuthenticationPhoneNumber) Type() string {
	return "setAuthenticationPhoneNumber"
}
