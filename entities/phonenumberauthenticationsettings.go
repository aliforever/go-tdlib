package entities

type PhoneNumberAuthenticationSettings struct {
	AllowFlashCall       bool     `json:"allow_flash_call"`
	AllowMissedCall      bool     `json:"allow_missed_call"`
	IsCurrentPhoneNumber bool     `json:"is_current_phone_number"`
	AllowSmsRetrieverAPI bool     `json:"allow_sms_retriever_api"`
	AuthenticationTokens []string `json:"authentication_tokens"`
}
