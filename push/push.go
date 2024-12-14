package push

import (
	"errors"
	"time"
)

type PushChannel interface {
	NotifyNewSSHLogin(ruser, rhost, user string, t time.Time) error
	NotifyPAMAuthenticate(items map[string]string) error
	NotifyPAMOpenSession(items map[string]string) error
}

func GetPushChannel(name string) (PushChannel, error) {
	switch name {
	case "telegram":
		return NewTelegramBot()
	default:
		return nil, errors.New("unsupported push channel")
	}
}
