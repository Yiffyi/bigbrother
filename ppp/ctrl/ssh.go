package ctrl

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/cespare/xxhash"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

type AuthorizedKeysList struct {
	watcher *fsnotify.Watcher
	lock    *sync.RWMutex
	path    string
	keys    map[uint64]bool
}

func NewAuthorizedKeysList(path string) (list *AuthorizedKeysList, err error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal().Err(err).Str("path", path).Msg("could not start fsnotify")
		return nil, err
	}

	lock := sync.RWMutex{}

	list = &AuthorizedKeysList{watcher, &lock, path, map[uint64]bool{}}

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// log.Println("event:", event)
				if event.Has(fsnotify.Write) {
					err := list.Reload()
					if err != nil {
						log.Error().Err(err).Str("path", path).Msg("authorized_keys changed, but could not reload")
					} else {
						log.Info().Str("path", path).Int("size", len(list.keys)).Msg("authorized_keys reloaded")
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error().Err(err).Msg("fsnotify watcher got error")
				// log.Println("error:", err)
			}
		}
	}()
	return
}

func (l *AuthorizedKeysList) Reload() error {
	b, err := os.ReadFile(l.path)
	if err != nil {
		return err
	}

	next := map[uint64]bool{}
	for key, _, _, rest, err := ssh.ParseAuthorizedKey(b); err != nil || len(rest) > 0; b = rest {
		next[xxhash.Sum64(key.Marshal())] = true
	}

	if err != nil {
		return err
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	l.keys = next
	return nil
}

func (l *AuthorizedKeysList) OK(keyBytes []byte) bool {
	l.lock.RLock()
	defer l.lock.RUnlock()

	k := xxhash.Sum64(keyBytes)

	v, ok := l.keys[k]
	return v && ok
}

func (l *AuthorizedKeysList) Close() {
	l.watcher.Close()
}

func NewSSHServerConfig(serverVersion string, keyList AuthorizedKeysList) *ssh.ServerConfig {
	return &ssh.ServerConfig{
		ServerVersion: serverVersion,
		PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			sid := xxhash.Sum64(c.SessionID())
			// Should use constant-time compare (or better, salt+hash) in a production setting.
			log.Info().
				Uint64("session_id", sid).
				Str("addr", c.RemoteAddr().String()).
				Str("user", c.User()).
				Bytes("key", ssh.MarshalAuthorizedKey(key)).
				Msg("public key auth attempt")

			if keyList.OK(key.Marshal()) {
				return nil, nil
			} else {
				return nil, errors.New("could not found provided key")
			}
		},
		//Define a function to run when a client attempts a password login

	}
}

func LoadHostKey(hostKeyPath []string) []ssh.Signer {
	if len(hostKeyPath) == 0 {
		panic("no host keys found")
	}

	ret := make([]ssh.Signer, len(hostKeyPath))
	for i, path := range hostKeyPath {
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

func ListenSSH(listenAddr string, serverConfig *ssh.ServerConfig) {
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
		if t := newChannel.ChannelType(); t == "lklk/ppp" {
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
