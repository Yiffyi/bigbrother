package push

import (
	"errors"
	"fmt"
	"html"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
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

func (b *TelegramBot) NotifyNewSSHLogin(ruser, rhost, user string, t time.Time) error {
	v := viper.Sub("push.telegram")

	chatid := v.GetInt64("to_chatid")
	username := v.GetString("to_username")

	var c *tele.Chat
	var err error
	if chatid != 0 {
		c, err = b.bot.ChatByID(chatid)
	} else if len(username) > 0 {
		c, err = b.bot.ChatByUsername(username)
	} else {
		err = errors.New("could not found valid telegram config")
	}

	if err != nil {
		log.Error().Err(err).Str("username", username).Int64("chatid", chatid).Msg("could not get chat recipient")
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
		log.Error().Err(err).Msg("could not get hostname")
	}

	msg_format := `<b>New SSH Login.</b>
At <code>%s</code>, a new SSH connection to managed machine has established:
<code>%s@%s</code> --> <code>%s@%s</code>`

	msg := fmt.Sprintf(msg_format, html.EscapeString(t.Format(time.RFC1123)), html.EscapeString(ruser), html.EscapeString(rhost), html.EscapeString(hostname), html.EscapeString(user))

	_, err = b.bot.Send(c, msg, tele.ModeHTML)
	if err != nil {
		log.Error().Err(err).Str("username", username).Int64("chatid", chatid).Msg("bot.Send returned error")
	}
	return err
}
