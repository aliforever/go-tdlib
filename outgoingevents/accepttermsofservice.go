package outgoingevents

type AcceptTermsOfService struct {
	TermsOfServiceID string `json:"terms_of_service_id"`
}

func (s AcceptTermsOfService) Type() string {
	return "acceptTermsOfService"
}
