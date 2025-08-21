package main

import (
	"github.com/yiffyi/bigbrother/honeypot"
	"github.com/yiffyi/bigbrother/misc"
)

func main() {
	cmd := honeypot.SetupHoneyDCmd()

	if err := misc.LoadConfig([]string{"."}); err != nil {
		panic(err)
	}

	misc.SetupLog()

	cmd.Execute()
}
