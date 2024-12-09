package push

import (
	"net/http"
	"time"

	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v4"
)

func init() {
	v := viper.Sub("push.telegram")
	v.SetDefault("token", "")
	v.SetDefault("timeout", time.Second*5)
}

type TelegramBot struct {
	bot *tele.Bot
}

func NewTelegramBot() (*TelegramBot, error) {
	v := viper.Sub("push.telegram")

	bot, err := tele.NewBot(tele.Settings{
		Token: v.GetString("token"),
		Client: &http.Client{
			Timeout: v.GetDuration("timeout"),
		},
	})

	if err != nil {
		return nil, err
	}

	return &TelegramBot{
		bot: bot,
	}, nil
}
