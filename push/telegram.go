package push

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v4"
)

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

func (b *TelegramBot) NotifyNewSSHLogin(rhost, user string, t time.Time) error {
	v := viper.Sub("push.telegram")

	chatid := v.GetInt64("to_chatid")
	username := v.GetString("to_username")

	var c *tele.Chat
	var err error
	if chatid != 0 {
		c, err = b.bot.ChatByID(chatid)
		if err != nil {
			return err
		}
	} else if len(username) == 0 {
		c, err = b.bot.ChatByUsername(username)
		if err != nil {
			return err
		}
	}

	msg := fmt.Sprintf(`# New SSH Login
			HOST: %s
			RHOST: %s
			USER: %s
			T: %s`, "...", rhost, user, t.Format(time.RFC1123))

	_, err = b.bot.Send(c, msg, tele.ModeMarkdownV2)
	return err
}
