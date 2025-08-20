package ctrl

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yiffyi/bigbrother/ppp"
)

var ctrlCmd = &cobra.Command{
	Use: "ctrl",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ctrlMain()
		// return agentMain(viper.GetString("ppp.agent.ctrl_addr"), viper.GetString("ppp.agent.ssh_known_hosts"), viper.GetString("ppp.ssh_keys"))
	},
}

func SetupCtrlCmd(v *viper.Viper) *cobra.Command {
	v.SetDefault("ppp.ctrl.base_url", "http://127.0.0.1:8080")
	v.SetDefault("ppp.ctrl.web_root", "public")
	v.SetDefault("ppp.ctrl.http_listen_addr", ":8080")
	v.SetDefault("ppp.ctrl.ssh_listen_addr", ":8022")
	v.SetDefault("ppp.ctrl.ssh_authorized_keys", "authorized_keys")
	v.SetDefault("ppp.ctrl.dsn", "data.db")
	v.SetDefault("ppp.ctrl.clash_sub_template", "clash_sub_template.yaml")
	v.SetDefault("ppp.ctrl.singbox_base_json", "sing-box.base.json")
	v.SetDefault("ppp.ctrl.key_rotate_interval", "144h")
	v.SetDefault("ppp.ctrl.keep_last_keys", 1)
	return ctrlCmd
}

func ctrlMain() error {
	authList, err := NewAuthorizedKeysList(viper.GetString("ppp.ctrl.ssh_authorized_keys"))
	if err != nil {
		return err
	}

	serverConfig := NewSSHServerConfig(ppp.CTRL_SSH_SERVER_VERSION, authList)
	for _, hostKey := range LoadHostKey(viper.GetStringSlice("ppp.ssh_keys")) {
		serverConfig.AddHostKey(hostKey)
	}

	proxyCtrl, err := NewSubscriptionController([]ConfigTemplate{
		&ClashSubscriptionTemplate{
			TemplatePath: viper.GetString("ppp.ctrl.clash_sub_template"),
		},
		&SingBoxSubscriptionTemplate{
			TemplatePath: viper.GetString("ppp.ctrl.singbox_base_json"),
		},
	})

	if err != nil {
		return err
	}

	ListenSSH(viper.GetString("ppp.ctrl.ssh_listen_addr"), serverConfig, proxyCtrl)

	return nil
}
