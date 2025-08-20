package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yiffyi/bigbrother/misc"
	"github.com/yiffyi/bigbrother/ppp/agent"
	"github.com/yiffyi/bigbrother/ppp/ctrl"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "ppp",
	}

	rootCmd.AddCommand(agent.SetupAgentCmd(viper.GetViper()))
	rootCmd.AddCommand(ctrl.SetupCtrlCmd(viper.GetViper()))

	if err := misc.LoadConfig([]string{"."}); err != nil {
		panic(err)
	}

	misc.SetupLog()

	rootCmd.Execute()
}
