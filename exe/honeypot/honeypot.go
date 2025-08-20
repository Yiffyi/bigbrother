package main

import (
	"github.com/spf13/viper"
	"github.com/yiffyi/bigbrother/honeypot"
	"github.com/yiffyi/bigbrother/misc"
)

func main() {
	cmd := honeypot.SetupHoneyDCmd(viper.Sub("honeypot"))

	if err := misc.LoadConfig([]string{"."}); err != nil {
		panic(err)
	}

	misc.SetupLog()

	cmd.Execute()
}
