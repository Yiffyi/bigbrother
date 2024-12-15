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

	viper.SetDefault("honeypot.enabled", false)
	viper.SetDefault("honeypot.server_version", "SSH-2.0-OpenSSH_8.4p1 Debian-5+deb11u3")
	viper.SetDefault("honeypot.server_host_keys", []string{"id_rsa"})
	viper.SetDefault("honeypot.listen_addrs", []string{"0.0.0.0:2022"})
	viper.SetDefault("honeypot.allow_any_creds", false)

	viper.SetDefault("installer.honeypot_exe", "/usr/local/bin/honeypot")
	viper.SetDefault("installer.honeypot_service_unit", "/etc/systemd/system/bb-honeypot.service")
	viper.SetDefault("installer.pam_bb_exe", "/usr/local/lib/pam_bb.so")
}

func LoadConfig() error {
	setupViper()
	viper.SafeWriteConfig()
	err := viper.ReadInConfig()
	return err
}
