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
	viper.SetDefault("log.pamDebugDump", false)

	viper.SetDefault("push.channel", "telegram")
	viper.SetDefault("push.telegram.token", "")
	viper.SetDefault("push.telegram.toUsername", "")
	viper.SetDefault("push.telegram.toChatID", 0)
	viper.SetDefault("push.telegram.timeout", time.Duration(5)*time.Second)

	viper.SetDefault("honeypot.enabled", false)
	viper.SetDefault("honeypot.serverVersion", "SSH-2.0-OpenSSH_8.4p1 Debian-5+deb11u3")
	viper.SetDefault("honeypot.serverHostKeys", []string{"id_rsa"})
	viper.SetDefault("honeypot.listenAddrs", []string{"0.0.0.0:2022"})
	viper.SetDefault("honeypot.allowAnyCreds", false)

	viper.SetDefault("installer.honeypotPath", "/usr/local/bin/honeypot")
	viper.SetDefault("installer.honeypotServiceUnit", "/etc/systemd/system/bb-honeypot.service")
	viper.SetDefault("installer.pamPath", "/usr/local/lib/pam_bb.so")
}

func LoadConfig() error {
	setupViper()
	viper.SafeWriteConfig()
	err := viper.ReadInConfig()
	return err
}
