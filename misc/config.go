package misc

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

const GlobalConfigPath = "/etc/bb"
const DefaultConfigName = "config"
const DefaultConfigType = "toml"

var GlobalConfigFullPath = filepath.Join(GlobalConfigPath, DefaultConfigName+"."+DefaultConfigType)

func setupViper(searchPaths []string) {
	viper.SetConfigName(DefaultConfigName)
	viper.SetConfigType(DefaultConfigType)

	for _, path := range searchPaths {
		viper.AddConfigPath(path)
	}
	viper.AddConfigPath(GlobalConfigPath)

	viper.SetDefault("log.path", "/var/log/bb.log")
	viper.SetDefault("log.console", true)
	viper.SetDefault("log.pam_debug_dump", false)

	viper.SetDefault("push.tg0.type", "telegram")
	viper.SetDefault("push.tg0.token", "")
	viper.SetDefault("push.tg0.to_username", "")
	viper.SetDefault("push.tg0.to_chatid", 0)
	viper.SetDefault("push.tg0.timeout", time.Duration(5)*time.Second)

	viper.SetDefault("push.lark0.type", "feishu")
	viper.SetDefault("push.lark0.app_id", "")
	viper.SetDefault("push.lark0.app_secret", "")
	viper.SetDefault("push.lark0.template_id", "AAAA:1.0.0")
	viper.SetDefault("push.lark0.dst", "open_id:ou_7d8a6e6df7621556ce0d21922b676706ccs")
	viper.SetDefault("push.lark0.timeout", time.Duration(5)*time.Second)

	viper.SetDefault("ppp.ssh_user", "ppp")
	viper.SetDefault("ppp.ssh_keys", []string{"/etc/ssh/ssh_host_ed25519_key"})
}

func LoadConfig(extraSearchPaths []string) error {
	setupViper(extraSearchPaths)

	if _, err := os.Stat("/etc/bb"); os.IsNotExist(err) {
		os.Mkdir("/etc/bb", 0755)
	}
	viper.SafeWriteConfig()
	err := viper.ReadInConfig()
	return err
}
