package main

import (
	_ "embed"

	"github.com/spf13/cobra"
	"github.com/yiffyi/bigbrother/cmd"
	"github.com/yiffyi/bigbrother/misc"
)

//go:embed install.tar.zst
var installZstdBytes []byte

func main() {
	if err := misc.LoadConfig(); err != nil {
		panic(err)
	}

	misc.SetupLog()

	var rootCmd = &cobra.Command{
		Use:   "bb",
		Short: "bb - Big Brother is WATCHING you(r server)",
	}

	rootCmd.AddCommand(cmd.SetupInstallCmd(installZstdBytes))

	rootCmd.Execute()
}
