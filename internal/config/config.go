package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var vip *viper.Viper

func LoadConfig() {
	vip = viper.New()
	vip.SetConfigName("config")
	vip.SetConfigType("yaml")
	vip.AddConfigPath("configs")
	vip.AddConfigPath(".")

	if err := vip.ReadInConfig(); err != nil {
		log.Fatalf("error reading config file: %v", err)
	}

	vip.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	vip.AutomaticEnv()
}

func GetGeneralConfig() GeneralConfig {
	return GeneralConfig{
		DefaultUser: vip.GetString(KeyGeneralDefaultUser),
		Theme:       vip.GetString(KeyGeneralTheme),
		LogLevel:    vip.GetString(KeyGeneralLogLevel),
	}
}

func GetSSHConfig() SSHConfig {
	return SSHConfig{
		ConnectTimeout: vip.GetInt(KeySSHConnectTimeout),
		KeepaliveEvery: vip.GetInt(KeySSHKeepaliveEvery),
		KnownHostsFile: vip.GetString(KeySSHKnownHostsFile),
		StrictHostKey:  vip.GetBool(KeySSHStrictHostKey),
	}
}

func GetUIConfig() UIConfig {
	return UIConfig{
		SidebarWidth: vip.GetInt(KeyUISidebarWidth),
		ShowKeyHints: vip.GetBool(KeyUIShowKeyHints),
		ScrollSpeed:  vip.GetInt(KeyUIScrollSpeed),
		MouseEnabled: vip.GetBool(KeyUIMouseEnabled),
	}
}

func GetPathsConfig() PathsConfig {
	dataDir := expandHome(vip.GetString(KeyPathsDataDir))
	dbFile := vip.GetString(KeyPathsDBFile)
	logFile := vip.GetString(KeyPathsLogFile)

	return PathsConfig{
		DataDir: dataDir,
		DBFile:  dbFile,
		LogFile: logFile,
		DBPath:  filepath.Join(dataDir, dbFile),
		LogPath: filepath.Join(dataDir, logFile),
	}
}

func GetDBPath() string {
	return GetPathsConfig().DBPath
}

func IsMouseEnabled() bool {
	return vip.GetBool(KeyUIMouseEnabled)
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
