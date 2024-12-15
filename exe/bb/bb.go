package main

import "github.com/spf13/cobra"

func main() {
	var rootCmd = &cobra.Command{
		Use:   "bb",
		Short: "bb - Big Brother is WATCHING you(r server)",
	}

	rootCmd.Execute()
}
