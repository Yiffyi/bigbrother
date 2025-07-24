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

	viper.SetDefault("ppp.ssh_keys", []string{"/etc/ssh/ssh_host_ed25519_key"})
	viper.SetDefault("ppp.ssh_authorized_keys", "authorized_keys")

	viper.SetDefault("ppp.agent.ctrl_addr", "127.0.0.1:8022")
	viper.SetDefault("ppp.agent.hostname", "localhost")
	viper.SetDefault("ppp.agent.report_interval", "1h")

	viper.SetDefault("ppp.ctrl.base_url", "http://127.0.0.1:8080")
	viper.SetDefault("ppp.ctrl.web_root", "public")
	viper.SetDefault("ppp.ctrl.http_listen_addr", ":8080")
	viper.SetDefault("ppp.ctrl.ssh_listen_addr", ":8022")
	viper.SetDefault("ppp.ctrl.dsn", "data.db")
	viper.SetDefault("ppp.ctrl.clash_sub_template", "clash_sub_template.yaml")
	viper.SetDefault("ppp.ctrl.singbox_base_json", "sing-box.base.json")
	viper.SetDefault("ppp.ctrl.key_rotate_interval", "144h")
	viper.SetDefault("ppp.ctrl.keep_last_keys", 1)
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
