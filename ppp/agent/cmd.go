package agent

import (
	"bytes"
	"encoding/gob"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yiffyi/bigbrother/ppp"
	"github.com/yiffyi/bigbrother/ppp/model"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var agentCmd = &cobra.Command{
	Use: "agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		return agentMain(viper.GetString("ppp.agent.ctrl_addr"), viper.GetString("ppp.agent.ssh_known_hosts"), viper.GetString("ppp.ssh_user"), viper.GetStringSlice("ppp.ssh_keys")[0])
	},
}

func SetupAgentCmd(v *viper.Viper) *cobra.Command {
	v.SetDefault("ppp.agent.ctrl_addr", "127.0.0.1:8022")
	v.SetDefault("ppp.agent.ssh_known_hosts", "known_hosts")
	v.SetDefault("ppp.agent.hostname", "localhost")
	v.SetDefault("ppp.agent.report_interval", "1h")
	v.SetDefault("ppp.agent.proxy_type", "sing-box")
	v.SetDefault("ppp.agent.proxy_program", "sing-box")
	v.SetDefault("ppp.agent.proxy_args", []string{"-c", "stdin", "run"})
	v.SetDefault("ppp.agent.proxy_share_console", false)
	return agentCmd
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

	pppChan, pppReqs, err := client.OpenChannel(ppp.SSH_CHANNEL_V1, nil)
	if err != nil {
		log.Fatal().Err(err).Str("chan", ppp.SSH_CHANNEL_V1).Str("ctrlAddr", ctrlAddr).Msg("could not open channel, maybe that's not our control server")
	}
	defer pppChan.Close()

	// We don't need to send anything
	pppChan.CloseWrite()

	proxy, err := NewServer(model.ProgramType(viper.GetString("ppp.agent.proxy_type")), viper.GetString("ppp.agent.proxy_program"), viper.GetStringSlice("ppp.agent.proxy_args"), nil, viper.GetBool("ppp.agent.proxy_share_console"))
	if err != nil {
		return err
	}

	for sshReq := range pppReqs {
		switch sshReq.Type {
		case "updateProxyConfig":
			reqDec := gob.NewDecoder(bytes.NewReader(sshReq.Payload))
			var req model.UpdateServerConfigRequest
			err = reqDec.Decode(&req)
			if err != nil {
				sshReq.Reply(false, nil)
				continue
			}
			log.Info().Bytes("config", req.Config).Bool("restart", req.Restart).Msg("received new proxy config")
			err = proxy.UpdateServerConfig(req.Config, req.Restart)
			sshReq.Reply(err == nil, nil)
		}
	}

	return nil
}
