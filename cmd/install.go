package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var installPAMFlag bool
var installHoneypotFlag bool
var installPAMToServiceFlag string

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Big Brother executables",
	Long:  `Choose to install pam module or honeypot, or both.`,
	Run: func(cmd *cobra.Command, args []string) {
		if installPAMFlag {
			installPAM()
		}

		if installHoneypotFlag {
			installHoneypot()
		}
	},
}

func SetupInstallCmd() *cobra.Command {
	installCmd.Flags().BoolVar(&installPAMFlag, "pam", true, "Install PAM module")
	installCmd.Flags().StringVar(&installPAMToServiceFlag, "pamService", "", "Service to add pam_bb")
	installCmd.Flags().BoolVar(&installHoneypotFlag, "honeypot", true, "Install honeypot")

	return installCmd
}

func installPAM() {
	if _, err := os.Stat("./pam_bb.so"); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Could not find pam_bb.so", err)
		}
	}

	dstPath := viper.GetString("installer.pam_bb_path")
	err := os.Link("./pam_bb.so", dstPath)
	if err != nil {
		src, err := os.OpenFile("./pam_bb.so", os.O_RDONLY, 0)
		if err != nil {
			fmt.Println("Could not open pam_bb.so", err)
		}

		dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			fmt.Println("Could not create file at ", dstPath, err)
		}

		_, err = io.Copy(dst, src)
		if err != nil {
			fmt.Println("Could not copy pam_bb.so to ", dstPath, err)
		}
	}

	fmt.Println("Installed pam_bb.so to ", dstPath)
}

func installHoneypot() {

}
