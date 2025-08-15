package push

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

type PushChannel interface {
	NotifyNewSSHLogin(ruser, rhost, user string, t time.Time) error
	NotifyPAMAuthenticate(items map[string]string) error
	NotifyPAMOpenSession(items map[string]string) error
}

func GetPushChannel(name string) (ch PushChannel, err error) {
	v := viper.Sub("push")

	if v.IsSet(name) {
		v = v.Sub(name)
		switch v.GetString("type") {
		case "telegram":
			ch, err = NewTelegramBot(v)
		case "feishu":
			ch, err = NewFeishuBot(v), nil
		default:
			ch, err = nil, errors.New("unsupported push channel")
		}
	} else {
		return nil, errors.New("could not found push channel")
	}

	if ch != nil && err == nil {
		return ch, err
	}

	return NewDummyChannel(), err
}
