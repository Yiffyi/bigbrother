package honeypot

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/cespare/xxhash"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

var db *gorm.DB
var vHoneypot *viper.Viper
var honeydCmd = &cobra.Command{
	Use: "honeyd",
	Run: func(cmd *cobra.Command, args []string) {
		honeydMain()
	},
}

func SetupHoneyDCmd(v *viper.Viper) *cobra.Command {
	v.SetDefault("enabled", false)
	v.SetDefault("server_version", "SSH-2.0-OpenSSH_8.4p1 Debian-5+deb11u3")
	v.SetDefault("server_host_keys", []string{"id_rsa"})
	v.SetDefault("listen_addrs", []string{"0.0.0.0:2022"})
	v.SetDefault("allow_any_creds", false)

	vHoneypot = v
	return honeydCmd
}

func honeydMain() {
	v := vHoneypot

	var err error
	db, err = OpenDatabase(v)
	if err != nil {
		panic(err)
	}

	config := NewSSHServerConfig(v)
	hostKeys := LoadHostKey(v)

	for _, v := range hostKeys {
		config.AddHostKey(v)
	}

	for _, addr := range v.GetStringSlice("listen_addrs") {
		go sshConnHandler(addr, config)
	}

	// Setup a channel to receive a signal
	done := make(chan os.Signal, 1)

	// Notify this channel when a SIGINT is received
	signal.Notify(done, os.Interrupt)

	<-done
	log.Warn().Msg("shutting down")
}

func sshConnHandler(listenAddr string, serverConfig *ssh.ServerConfig) {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Error().Err(err).Str("listenAddr", listenAddr).Msg("failed to listen")
		panic(err)
	}

	log.Info().
		Str("addr", listenAddr).
		Msg("tcp listener started")

	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Error().Err(err).Msg("failed to accept incoming connection")
			continue
		}

		// Before use, a handshake must be performed on the incoming net.Conn.
		sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, serverConfig)
		if err != nil {
			log.Error().Err(err).Msg("failed to handshake")
			continue
		}

		sid := xxhash.Sum64(sshConn.SessionID())
		log.Info().
			Uint64("session_id", sid).
			// Str("session_id", string(sshConn.SessionID())).
			Str("client_version", string(sshConn.ClientVersion())).
			Str("user", sshConn.User()).
			Str("addr", sshConn.RemoteAddr().String()).
			Msg("new incoming connection")

		// Discard all global out-of-band Requests
		go serveGlobalRequests(sid, sshConn, reqs)
		go serveNewChannels(sid, sshConn, chans)
	}
}

func serveGlobalRequests(sid uint64, _ *ssh.ServerConn, in <-chan *ssh.Request) {
	for req := range in {
		if req.WantReply {
			req.Reply(false, nil)
		}
		log.Debug().
			Uint64("session_id", sid).
			Str("type", req.Type).
			Bool("want_reply", req.WantReply).
			Bytes("payload", req.Payload).
			Msg("rejected global request")
	}
}

func serveNewChannels(sid uint64, sshConn *ssh.ServerConn, in <-chan ssh.NewChannel) {
	for newChannel := range in {
		if t := newChannel.ChannelType(); t == "session" {
			ch, reqs, err := newChannel.Accept()
			if err != nil {
				log.Error().
					Uint64("session_id", sid).
					Str("type", newChannel.ChannelType()).
					Err(err).
					Msg("failed to accept new channel")
				continue
			}

			log.Debug().
				Uint64("session_id", sid).
				Str("type", newChannel.ChannelType()).
				Msg("accepted new channel")

			go serveSessionChannel(sid, sshConn, ch)
			go servePerChannelRequests(sid, sshConn, reqs)
			continue
		}

		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", newChannel.ChannelType()))
		log.Debug().
			Uint64("session_id", sid).
			Str("type", newChannel.ChannelType()).
			Msg("rejected session channel")
	}
}

func serveSessionChannel(sid uint64, _ *ssh.ServerConn, channel ssh.Channel) {
	scanner := bufio.NewScanner(channel)
	for scanner.Scan() {
		line := scanner.Text() // Gets the current line as string
		log.Debug().
			Uint64("session_id", sid).
			Str("line", line).
			Msg("remote sent command")
	}
}

func servePerChannelRequests(sid uint64, _ *ssh.ServerConn, in <-chan *ssh.Request) {
	for req := range in {
		switch req.Type {
		case "shell":
			if req.WantReply {
				req.Reply(true, nil)
			}
			log.Info().
				Uint64("session_id", sid).
				Str("type", req.Type).
				Str("payload", string(req.Payload)).
				Bool("want_reply", req.WantReply).
				Msg("shell spawn request accepted")
		case "exec":
			if req.WantReply {
				req.Reply(true, nil)
			}
			var payload = struct{ Value string }{}
			err := ssh.Unmarshal(req.Payload, &payload)
			if err != nil {
				log.Error().
					Err(err).
					Uint64("session_id", sid).
					Str("type", req.Type).
					Str("payload", string(req.Payload)).
					Bool("want_reply", req.WantReply).
					Msg("could not parse exec request payload")
			} else {
				log.Info().
					Uint64("session_id", sid).
					Str("type", req.Type).
					Str("payload", payload.Value).
					Bool("want_reply", req.WantReply).
					Msg("exec request accepted")
			}
		default:
			if req.WantReply {
				req.Reply(false, nil)
			}
			log.Debug().
				Uint64("session_id", sid).
				Str("type", req.Type).
				Bool("want_reply", req.WantReply).
				Bytes("payload", req.Payload).
				Msg("rejected per-channel request")
		}
	}
}

func NewSSHServerConfig(v *viper.Viper) *ssh.ServerConfig {
	return &ssh.ServerConfig{
		ServerVersion: v.GetString("server_version"),
		//Define a function to run when a client attempts a password login
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			sid := xxhash.Sum64(c.SessionID())
			// Should use constant-time compare (or better, salt+hash) in a production setting.
			log.Info().
				Uint64("session_id", sid).
				Str("addr", c.RemoteAddr().String()).
				Str("user", c.User()).
				Str("password", string(pass)).
				Msg("password auth attempt")

			RecordAuthAttempt(db, c.User(), string(pass), c.RemoteAddr())

			if v.GetBool("allow_any_creds") {
				return nil, nil
			} else {
				allow, err := ShallAllowConnection(db, c.User(), string(pass), c.RemoteAddr())
				if err != nil {
					return nil, fmt.Errorf("password rejected due to internal error: %w", err)
				}

				if allow {
					return nil, nil
				} else {
					return nil, fmt.Errorf("password rejected, please try harder")
				}
			}
		},
	}
}

func LoadHostKey(v *viper.Viper) []ssh.Signer {
	paths := v.GetStringSlice("server_host_keys")
	if len(paths) == 0 {
		panic("no host keys found")
	}

	ret := make([]ssh.Signer, len(paths))
	for i, path := range paths {
		buf, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}

		private, err := ssh.ParsePrivateKey(buf)
		if err != nil {
			panic(err)
		}

		log.Debug().Str("path", path).Msg("host key loaded")
		ret[i] = private
	}
	return ret
}
