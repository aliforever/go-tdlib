package entities

type User struct {
	ID                int32       `json:"id"`
	FirstName         string      `json:"first_name"`
	LastName          string      `json:"last_name"`
	Username          string      `json:"username"`
	PhoneNumber       string      `json:"phone_number"`
	Status            interface{} `json:"status"`
	ProfilePhoto      interface{} `json:"profile_photo"`
	IsContact         bool        `json:"is_contact"`
	IsMutualContact   bool        `json:"is_mutual_contact"`
	IsVerified        bool        `json:"is_verified"`
	IsSupport         bool        `json:"is_support"`
	RestrictionReason string      `json:"restriction_reason"`
	IsScam            bool        `json:"is_scam"`
	IsFake            bool        `json:"is_fake"`
	HaveAccess        bool        `json:"have_access"`
	Type              interface{} `json:"type"`
	LanguageCode      string      `json:"language_code"`
}
