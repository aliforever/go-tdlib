package entities

type User struct {
	Id        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Usernames struct {
		Type              string        `json:"@type"`
		ActiveUsernames   []string      `json:"active_usernames"`
		DisabledUsernames []interface{} `json:"disabled_usernames"`
		EditableUsername  string        `json:"editable_username"`
	} `json:"usernames"`
	PhoneNumber string `json:"phone_number"`
	Status      struct {
		Type    string `json:"@type"`
		Expires int    `json:"expires"`
	} `json:"status"`
	ProfilePhoto struct {
		Type  string `json:"@type"`
		Id    string `json:"id"`
		Small struct {
			Type         string `json:"@type"`
			Id           int    `json:"id"`
			Size         int    `json:"size"`
			ExpectedSize int    `json:"expected_size"`
			Local        struct {
				Type                   string `json:"@type"`
				Path                   string `json:"path"`
				CanBeDownloaded        bool   `json:"can_be_downloaded"`
				CanBeDeleted           bool   `json:"can_be_deleted"`
				IsDownloadingActive    bool   `json:"is_downloading_active"`
				IsDownloadingCompleted bool   `json:"is_downloading_completed"`
				DownloadOffset         int    `json:"download_offset"`
				DownloadedPrefixSize   int    `json:"downloaded_prefix_size"`
				DownloadedSize         int    `json:"downloaded_size"`
			} `json:"local"`
			Remote struct {
				Type                 string `json:"@type"`
				Id                   string `json:"id"`
				UniqueId             string `json:"unique_id"`
				IsUploadingActive    bool   `json:"is_uploading_active"`
				IsUploadingCompleted bool   `json:"is_uploading_completed"`
				UploadedSize         int    `json:"uploaded_size"`
			} `json:"remote"`
		} `json:"small"`
		Big struct {
			Type         string `json:"@type"`
			Id           int    `json:"id"`
			Size         int    `json:"size"`
			ExpectedSize int    `json:"expected_size"`
			Local        struct {
				Type                   string `json:"@type"`
				Path                   string `json:"path"`
				CanBeDownloaded        bool   `json:"can_be_downloaded"`
				CanBeDeleted           bool   `json:"can_be_deleted"`
				IsDownloadingActive    bool   `json:"is_downloading_active"`
				IsDownloadingCompleted bool   `json:"is_downloading_completed"`
				DownloadOffset         int    `json:"download_offset"`
				DownloadedPrefixSize   int    `json:"downloaded_prefix_size"`
				DownloadedSize         int    `json:"downloaded_size"`
			} `json:"local"`
			Remote struct {
				Type                 string `json:"@type"`
				Id                   string `json:"id"`
				UniqueId             string `json:"unique_id"`
				IsUploadingActive    bool   `json:"is_uploading_active"`
				IsUploadingCompleted bool   `json:"is_uploading_completed"`
				UploadedSize         int    `json:"uploaded_size"`
			} `json:"remote"`
		} `json:"big"`
		Minithumbnail struct {
			Type   string `json:"@type"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
			Data   string `json:"data"`
		} `json:"minithumbnail"`
		HasAnimation bool `json:"has_animation"`
		IsPersonal   bool `json:"is_personal"`
	} `json:"profile_photo"`
	IsContact         bool   `json:"is_contact"`
	IsMutualContact   bool   `json:"is_mutual_contact"`
	IsVerified        bool   `json:"is_verified"`
	IsPremium         bool   `json:"is_premium"`
	IsSupport         bool   `json:"is_support"`
	RestrictionReason string `json:"restriction_reason"`
	IsScam            bool   `json:"is_scam"`
	IsFake            bool   `json:"is_fake"`
	HaveAccess        bool   `json:"have_access"`
	Type1             struct {
		Type string `json:"@type"`
	} `json:"type"`
	LanguageCode          string `json:"language_code"`
	AddedToAttachmentMenu bool   `json:"added_to_attachment_menu"`
}
