package outgoingevents

type SetTdlibParameters struct {
	UseTestDc              bool        `json:"use_test_dc"`
	DatabaseDirectory      string      `json:"database_directory"`
	FilesDirectory         string      `json:"files_directory"`
	DatabaseEncryptionKey  interface{} `json:"database_encryption_key"`
	UseFileDatabase        bool        `json:"use_file_database"`
	UseChatInfoDatabase    bool        `json:"use_chat_info_database"`
	UseMessageDatabase     bool        `json:"use_message_database"`
	UseSecretChats         bool        `json:"use_secret_chats"`
	ApiId                  interface{} `json:"api_id"`
	ApiHash                interface{} `json:"api_hash"`
	SystemLanguageCode     string      `json:"system_language_code"`
	DeviceModel            string      `json:"device_model"`
	SystemVersion          string      `json:"system_version"`
	ApplicationVersion     string      `json:"application_version"`
	EnableStorageOptimizer bool        `json:"enable_storage_optimizer"`
	IgnoreFileNames        bool        `json:"ignore_file_names"`
}

func (SetTdlibParameters) Type() string {
	return "setTdlibParameters"
}
