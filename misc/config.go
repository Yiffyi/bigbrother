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

	viper.SetDefault("push.channel", "telegram")
	viper.SetDefault("push.telegram.token", "")
	viper.SetDefault("push.telegram.to_username", "")
	viper.SetDefault("push.telegram.to_chatid", 0)
	viper.SetDefault("push.telegram.timeout", time.Duration(5)*time.Second)

	viper.SetDefault("honeypot.enabled", false)
	viper.SetDefault("honeypot.server_version", "SSH-2.0-OpenSSH_8.4p1 Debian-5+deb11u3")
	viper.SetDefault("honeypot.server_host_keys", []string{"id_rsa"})
	viper.SetDefault("honeypot.listen_addrs", []string{"0.0.0.0:2022"})
	viper.SetDefault("honeypot.allow_any_creds", false)

	viper.SetDefault("installer.honeypot_path", "/usr/local/bin/honeypot")
	viper.SetDefault("installer.honeypot_service_unit", "/etc/systemd/system/bb-honeypot.service")
	viper.SetDefault("installer.pam_bb_path", "/usr/local/lib/pam_bb.so")
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
