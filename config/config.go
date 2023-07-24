package config

type Config struct {
	UseTestDC             bool        `json:"use_test_dc"`
	DatabaseDirectory     string      `json:"database_directory"`
	FilesDirectory        string      `json:"files_directory"`
	DatabaseEncryptionKey interface{} `json:"database_encryption_key"`
	UseFileDatabase       bool        `json:"use_file_database"`
	UseChatInfoDatabase   bool        `json:"use_chat_info_database"`
	UseMessageDatabase    bool        `json:"use_message_database"`
	UseSecretChats        bool        `json:"use_secret_chats"`
	SystemLanguageCode    string      `json:"system_language_code"`
	DeviceModel           string      `json:"device_model"`
	SystemVersion         string      `json:"system_version"`
	ApplicationVersion    string      `json:"application_version"`
	StorageOptimizer      bool        `json:"enable_storage_optimizer"`
	FilenamesIgnored      bool        `json:"ignore_file_names"`
}

func New() *Config {
	return &Config{}
}

func (c *Config) EnableTestDC() *Config {
	c.UseTestDC = true

	return c
}

func (c *Config) SetDatabaseDirectory(dir string) *Config {
	c.DatabaseDirectory = dir

	return c
}

func (c *Config) SetFilesDirectory(dir string) *Config {
	c.FilesDirectory = dir

	return c
}

func (c *Config) SetDatabaseEncryptionKey(key interface{}) *Config {
	c.DatabaseEncryptionKey = key

	return c
}

func (c *Config) EnableFileDatabase() *Config {
	c.UseFileDatabase = true

	return c
}

func (c *Config) EnableChatInfoDatabase() *Config {
	c.UseChatInfoDatabase = true

	return c
}

func (c *Config) EnableMessageDatabase() *Config {
	c.UseChatInfoDatabase = true

	return c
}

func (c *Config) EnableSecretChats() *Config {
	c.UseSecretChats = true

	return c
}

func (c *Config) SetSystemLanguageCode(code string) *Config {
	c.SystemLanguageCode = code

	return c
}

func (c *Config) SetDeviceModel(model string) *Config {
	c.DeviceModel = model

	return c
}

func (c *Config) SetSystemVersion(version string) *Config {
	c.SystemVersion = version

	return c
}

func (c *Config) SetApplicationVersion(version string) *Config {
	c.ApplicationVersion = version

	return c
}

func (c *Config) EnableStorageOptimizer() *Config {
	c.StorageOptimizer = true

	return c
}

func (c *Config) IgnoreFileNames() *Config {
	c.FilenamesIgnored = true

	return c
}
