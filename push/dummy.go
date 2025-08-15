package push

import (
	"errors"
	"time"

	"github.com/rs/zerolog/log"
)

type DummyChannel struct {
}

func NewDummyChannel() *DummyChannel {
	return &DummyChannel{}
}

func (b *DummyChannel) NotifyNewSSHLogin(ruser, rhost, user string, t time.Time) error {
	log.Info().Str("ruser", ruser).Str("rhost", rhost).Str("user", user).Time("t", t)
	return nil
}

func (b *DummyChannel) NotifyPAMAuthenticate(items map[string]string) error {
	if v, ok := items["service"]; ok && v == "sshd" {
		return b.NotifyNewSSHLogin(items["ruser"], items["rhost"], items["user"], time.Now())
	} else {
		err := errors.New("unsupported caller service from PAM")
		log.Info().
			Str("service", v).
			Err(err).
			Msg("new PAM session opened")
		return err
	}
}

func (b *DummyChannel) NotifyPAMOpenSession(items map[string]string) error {
	if v, ok := items["service"]; ok && v == "sshd" {
		return b.NotifyNewSSHLogin(items["ruser"], items["rhost"], items["user"], time.Now())
	} else {
		err := errors.New("unsupported caller service from PAM")
		log.Info().
			Str("service", v).
			Err(err).
			Msg("new PAM session opened")
		return err
	}
}
