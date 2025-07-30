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

func SetupCtrlCmd() *cobra.Command {
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

	proxyCtrl, err := NewSubscriptionController([]SubscriptionTemplate{
		&ClashSubscriptionTemplate{
			templatePath: viper.GetString("ppp.ctrl.clash_sub_template"),
		},
		&SingBoxSubscriptionTemplate{
			templatePath: viper.GetString("ppp.ctrl.singbox_base_json"),
		},
	})

	if err != nil {
		return err
	}

	ListenSSH(viper.GetString("ppp.ctrl.ssh_listen_addr"), serverConfig, proxyCtrl)

	return nil
}
