package agent

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yiffyi/bigbrother/ppp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var agentCmd = &cobra.Command{
	Use: "agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		return agentMain(viper.GetString("ppp.agent.ctrl_addr"), viper.GetString("ppp.agent.ssh_known_hosts"), viper.GetString("ppp.ssh_user"), viper.GetStringSlice("ppp.ssh_keys")[0])
	},
}

func SetupAgentCmd() *cobra.Command {
	return agentCmd
}

func handleReqs(in <-chan *ssh.Request) {
	for req := range in {
		switch req.Type {
		case "updateProxyConfig":

		}
	}
}

func agentMain(ctrlAddr string, knownHostsPath string, sshUser string, sshPrivKeyPath string) (err error) {
	var checkHostKey ssh.HostKeyCallback
	if len(knownHostsPath) > 0 {
		checkHostKey, err = knownhosts.New(knownHostsPath)
		if err != nil {
			return
		}
	} else {
		log.Warn().Msg("known_hosts not specified, host key check is disabled")
		checkHostKey, err = ssh.InsecureIgnoreHostKey(), nil
	}

	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(func() (signers []ssh.Signer, err error) {
				privKeyBytes, err := os.ReadFile(sshPrivKeyPath)
				if err != nil {
					log.Error().Err(err).Str("path", sshPrivKeyPath).Msg("could not read SSH private key for auth")
					return nil, err
				}
				privKey, err := ssh.ParsePrivateKey(privKeyBytes)
				if err != nil {
					log.Error().Err(err).Str("path", sshPrivKeyPath).Msg("could not parse SSH private key for auth")
					return nil, err
				}

				return []ssh.Signer{privKey}, nil
			}),
		},
		HostKeyCallback: checkHostKey,
	}

	client, err := ssh.Dial("tcp", ctrlAddr, config)
	if err != nil {
		log.Fatal().Err(err).Str("ctrlAddr", ctrlAddr).Msg("could not dial control server")
	}
	defer client.Close()

	pppChan, pppReqs, err := client.OpenChannel(ppp.PPP_SSH_CHANNEL_V1, nil)
	if err != nil {
		log.Fatal().Err(err).Str("chan", ppp.PPP_SSH_CHANNEL_V1).Str("ctrlAddr", ctrlAddr).Msg("could not open channel, maybe that's not our control server")
	}
	defer pppChan.Close()

	// We don't need to send anything
	pppChan.CloseWrite()
	handleReqs(pppReqs)

	return nil
}
