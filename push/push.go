package push

import (
	"errors"
	"time"
)

type PushChannel interface {
	NotifyNewSSHLogin(rhost, user string, t time.Time) error
}

func GetPushChannel(name string) (PushChannel, error) {
	switch name {
	case "telegram":
		return NewTelegramBot()
	default:
		return nil, errors.New("unsupported push channel")
	}
}
