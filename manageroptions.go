package tdlib

type ManagerOptions struct {
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
