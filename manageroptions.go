package tdlib

type ManagerOptions struct {
	DisableLogging    bool
	LogVerbosityLevel int
	LogPath           string
}

func NewManagerOptions() *ManagerOptions {
	return &ManagerOptions{}
}

func (cfg *ManagerOptions) SetLogVerbosityLevel(level int) *ManagerOptions {
	cfg.LogVerbosityLevel = level

	return cfg
}

func (cfg *ManagerOptions) SetLogPath(path string) *ManagerOptions {
	cfg.LogPath = path

	return cfg
}

func (cfg *ManagerOptions) SetDisableLogging() *ManagerOptions {
	cfg.DisableLogging = true

	return cfg
}
