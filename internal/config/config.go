package config

import (
	"github.com/spf13/viper"
)

var vip *viper.Viper

func LoadConfig() {
	vip = viper.New()
	vip.SetConfigName("config")
	vip.SetConfigType("yaml")
	vip.AddConfigPath("configs")
	vip.AddConfigPath(".")

	vip.SetDefault(KeyUISidebarWidth, 26)
	vip.SetDefault(KeyUIShowKeyHints, true)
	vip.SetDefault(KeyUIScrollSpeed, 3)
	vip.SetDefault(KeyUIMouseEnabled, true)

	// config file is optional — defaults cover everything
	_ = vip.ReadInConfig()
}

func GetUIConfig() UIConfig {
	return UIConfig{
		SidebarWidth: vip.GetInt(KeyUISidebarWidth),
		ShowKeyHints: vip.GetBool(KeyUIShowKeyHints),
		ScrollSpeed:  vip.GetInt(KeyUIScrollSpeed),
		MouseEnabled: vip.GetBool(KeyUIMouseEnabled),
	}
}
