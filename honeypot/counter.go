package honeypot

import (
	"net"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

type CounterStrikeController struct {
	db *gorm.DB
}

func NewCounterStrikeController(db *gorm.DB) *CounterStrikeController {
	return &CounterStrikeController{
		db: db,
	}
}

func (c *CounterStrikeController) ctWorker(remote string, user string, password string) {
	checkHostKey := ssh.InsecureIgnoreHostKey()

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PasswordCallback(func() (secret string, err error) {
				return password, nil
			}),
		},
		HostKeyCallback: checkHostKey,
		Timeout:         time.Minute * 1,
	}

	client, err := ssh.Dial("tcp", remote, config)
	if err != nil {
		log.Info().Str("remote", remote).Str("user", user).Str("password", password).Err(err).Msg("remote didn't accept the connection")
		return
	} else {
		log.Info().Str("remote", remote).Str("user", user).Str("password", password).Msg("Counter-Terrorists WIN")
		client.Close()
		return
	}

}

func (c *CounterStrikeController) Strike(srcAddr net.Addr, user string, password string) {
	remote := net.TCPAddr{}
	switch addr := srcAddr.(type) {
	// case *net.UDPAddr: // though this is impossible
	// 	a.RemoteIP = addr.IP.String()
	// 	a.RemotePort = addr.Port
	case *net.TCPAddr:
		remote = *addr
		remote.Port = 22
	}

	go c.ctWorker(remote.String(), user, password)
}
