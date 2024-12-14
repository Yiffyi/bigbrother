package misc

import (
	"time"

	"github.com/spf13/viper"
)

func setupViper() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("/etc/bb/")

	viper.SetDefault("log.path", "/var/log/bb.log")
	viper.SetDefault("log.console", true)
	viper.SetDefault("log.pam_debug_dump", false)
	viper.SetDefault("push.channel", "telegram")
	viper.SetDefault("push.telegram.token", "")
	viper.SetDefault("push.telegram.to_username", "")
	viper.SetDefault("push.telegram.to_chatid", 0)
	viper.SetDefault("push.telegram.timeout", time.Duration(5)*time.Second)
}

func LoadConfig() error {
	setupViper()
	viper.SafeWriteConfig()
	err := viper.ReadInConfig()
	return err
}
