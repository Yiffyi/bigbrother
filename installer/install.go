package installer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/yiffyi/bigbrother/misc"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var installPAMToServiceFlag string
var installPAMCustomSOFlag string
var installZstdBytes []byte
var installSaveConfig bool

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

		if installSaveConfig {
			viper.WriteConfigAs(misc.GlobalConfigFullPath)
		}
	},
}

func SetupInstallCmd(installZstd []byte) *cobra.Command {
	viper.SetDefault("installer.honeypot_path", "/usr/local/bin/honeypot")
	viper.SetDefault("installer.honeypot_service_unit", "/etc/systemd/system/bb-honeypot.service")
	viper.SetDefault("installer.pam_bb_path", "/usr/local/lib/pam_bb.so")

	installCmd.Flags().StringVar(&installPAMToServiceFlag, "pamService", "sshd", "Service to add pam_bb")
	installCmd.Flags().StringVar(&installPAMCustomSOFlag, "pamSo", "", "Use a differeent pam_bb.so file instead of the embedded one")
	installCmd.Flags().BoolVar(&installSaveConfig, "saveConfig", false, "Save current config to /etc/bb/config.toml")

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

func patchPAMConfig(service string) {
	orgFilename := "/etc/pam.d/" + service
	tmpFilename := "/etc/pam.d/" + service + ".new"

	orgInfo, err := os.Stat(orgFilename)
	if err != nil {
		fmt.Println("Could not check pam config for", service, err)
		return
	}

	file, err := os.Open(orgFilename)
	if err != nil {
		fmt.Println("Could not open pam config for", service, err)
		return
	}
	defer file.Close()

	fileNew, err := os.OpenFile(tmpFilename, os.O_RDWR|os.O_CREATE, orgInfo.Mode())
	if err != nil {
		fmt.Println("Could not create temporary pam config for", service, err)
		return
	}
	defer fileNew.Close()
	defer os.Remove(tmpFilename)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, "# installed by bb") {
			// skip previous installation lines
			continue
			// fmt.Println("PAM module already installed")
			// return
		} else {
			fileNew.WriteString(line + "\n")
		}
	}

	dstPath := viper.GetString("installer.pam_bb_path")
	fileNew.WriteString(fmt.Sprintf("session optional %s # installed by bb\n", dstPath))

	newContent, err := os.ReadFile(tmpFilename)
	if err != nil {
		fmt.Println("Could not read temporary pam config for", service, err)
		return
	}

	fmt.Println("The patched pam config for", service, "is as follows:\n===", tmpFilename, "===")
	fmt.Println(string(newContent))
	fmt.Print("Is that OK? (Y/n) ")
	ok := "Y"
	fmt.Scanln(&ok)
	if strings.ToLower(ok) == "y" {
		err := os.Rename(orgFilename, orgFilename+".old")
		if err != nil {
			fmt.Println("Could not rename old pam config for", service, err)
		} else {
			err = os.Rename(tmpFilename, orgFilename)
			if err != nil {
				fmt.Println("Could not rename temporary pam config for", service, err)
			}
		}

		if err == nil {
			fmt.Println("Configured PAM service", service)
			return
		}
	}

	fmt.Println("Operation abored.")
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
		r, err := NewReaderFromTarZstd(installZstdBytes, "pam_bb.so")
		if err != nil {
			fmt.Println("Could not decompress embedded resources", err)
			return
		}
		defer r.Close()

		if err := installPAMFromReader(r, dstPath); err == nil {
			fmt.Println("Installed PAM module from embedded resources to", dstPath)
		}
	}

	if installPAMToServiceFlag != "" {
		patchPAMConfig(installPAMToServiceFlag)
	}
}

func installHoneypot() {

}
