package config

type Config struct {
	General GeneralConfig
	SSH     SSHConfig
	UI      UIConfig
	Paths   PathsConfig
}

type GeneralConfig struct {
	DefaultUser string
	Theme       string
	LogLevel    string
}

type SSHConfig struct {
	ConnectTimeout int
	KeepaliveEvery int
	KnownHostsFile string
	StrictHostKey  bool
}

type UIConfig struct {
	SidebarWidth int
	ShowKeyHints bool
	ScrollSpeed  int
	MouseEnabled bool
}

type PathsConfig struct {
	DataDir string
	DBFile  string
	LogFile string
	DBPath  string
	LogPath string
}
