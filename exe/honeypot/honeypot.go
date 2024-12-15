package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/cespare/xxhash"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/yiffyi/bigbrother/misc"
	"golang.org/x/crypto/ssh"
)

func main() {
	err := misc.LoadConfig()
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	misc.SetupLog()

	v := viper.Sub("honeypot")

	config := NewSSHServerConfig(v)
	hostKeys := LoadHostKey(v)

	for _, v := range hostKeys {
		config.AddHostKey(v)
	}

	for _, addr := range viper.GetStringSlice("listen_addrs") {
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
		if t := newChannel.ChannelType(); t != "session" {
			newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			log.Debug().
				Uint64("session_id", sid).
				Str("type", newChannel.ChannelType()).
				Msg("rejected session channel")
			continue
		}

		// only session channel will fall through
		ch, reqs, err := newChannel.Accept()
		if err != nil {
			log.Error().
				Uint64("session_id", sid).
				Str("type", newChannel.ChannelType()).
				Err(err).
				Msg("failed to accept new channel")
			continue
		}
		go serveSessionChannel(sid, sshConn, ch)
		go servePerChannelRequests(sid, sshConn, reqs)
		log.Debug().
			Uint64("session_id", sid).
			Str("type", newChannel.ChannelType()).
			Msg("accepted new channel")
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
				Str("user", c.User()).
				Str("password", string(pass)).
				Msg("password auth attempt")
			if viper.GetBool("AllowAnyCred") {
				return nil, nil
			} else {
				return nil, fmt.Errorf("password rejected for %q", c.User())
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
