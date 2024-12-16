package cmd

import (
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/yiffyi/bigbrother/misc"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var installPAMToServiceFlag string
var installPAMCustomSOFlag string
var installZstdBytes []byte

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Big Brother executables",
	Long:  `Choose to install pam module or honeypot, or both.`,
	Run: func(cmd *cobra.Command, args []string) {
		if slices.Contains(args, "pam") {
			installPAM()
		}

		if slices.Contains(args, "honeypot") {
			installHoneypot()
		}
	},
}

func SetupInstallCmd(installZstd []byte) *cobra.Command {
	installCmd.Flags().StringVar(&installPAMToServiceFlag, "pamService", "", "Service to add pam_bb")
	installCmd.Flags().StringVar(&installPAMCustomSOFlag, "pamSo", "", "Use a differeent pam_bb.so file instead of the embedded one")

	installZstdBytes = installZstd
	return installCmd
}

func installPAMFromReader(src io.Reader, dstPath string) error {
	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("Could not create file at", dstPath, err)
		return err
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		fmt.Println("Could not copy pam_bb.so to", dstPath, err)
		return err
	}

	return nil
}

func installPAMFromFile(srcPath, dstPath string) error {
	src, err := os.OpenFile(srcPath, os.O_RDONLY, 0)
	if err != nil {
		fmt.Println("Could not open pam_bb.so at", srcPath, err)
		return err
	}

	return installPAMFromReader(src, dstPath)
}

func installPAM() {
	dstPath := viper.GetString("installer.pam_bb_path")
	if installPAMCustomSOFlag != "" {
		if _, err := os.Stat(installPAMCustomSOFlag); err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Could not find pam_bb.so at", installPAMCustomSOFlag, err)
			}
		}

		// install by Link
		err := os.Link(installPAMCustomSOFlag, dstPath)

		if err != nil {
			// or install by Copy
			installPAMFromFile(installPAMCustomSOFlag, dstPath)
		}
		fmt.Println("Installed PAM module", installPAMCustomSOFlag, "to", dstPath)
	} else {
		r, err := misc.NewReaderFromTarZstd(installZstdBytes, "pam_bb.so")
		if err != nil {
			fmt.Println("Could not decompress embedded resources", err)
			return
		}
		defer r.Close()

		if err := installPAMFromReader(r, dstPath); err == nil {
			fmt.Println("Installed PAM module from embedded resources to", dstPath)
		}
	}
}

func installHoneypot() {

}
