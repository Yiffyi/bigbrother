package push

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type FeishuBot struct {
	c               *lark.Client
	hc              *http.Client
	templateId      string
	templateVersion string
	receiveIdType   string
	receiveId       string
}

type feishuWebhookResponse struct {
	Code int
	Msg  string
	Data interface{}
}

func NewFeishuBot(v *viper.Viper) *FeishuBot {
	a := FeishuBot{}
	app_id, app_secret := v.GetString("app_id"), v.GetString("app_secret")

	dsts := strings.SplitN(v.GetString("dst"), ":", 2)
	if len(dsts) < 2 {
		log.Error().Str("dst", v.GetString("dst")).Msg("incorrect dst")
		return nil
	}
	a.receiveIdType, a.receiveId = dsts[0], dsts[1]

	temp := strings.SplitN(v.GetString("template_id"), ":", 2)
	if len(temp) < 2 {
		log.Error().Str("template_id", v.GetString("template_id")).Msg("incorrect template_id")
		return nil
	}
	a.templateId, a.templateVersion = temp[0], temp[1]

	if a.receiveIdType == "webhook" {
		a.hc = &http.Client{
			Timeout: v.GetDuration("timeout"),
		}
	} else {
		a.c = lark.NewClient(app_id, app_secret)
	}

	return &a
}

func (a *FeishuBot) postWebhookReq(v any, url string) (status int, err error) {
	b, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var j feishuWebhookResponse
	err = json.NewDecoder(resp.Body).Decode(&j)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		return resp.StatusCode, errors.New("bad HTTP statusCode")
	}

	if j.Code != 0 {
		return j.Code, errors.New(j.Msg)
	}
	return j.Code, nil

}

func (a *FeishuBot) SendTemplateMessage(vars map[string]any) error {
	if a.receiveIdType == "webhook" {
		status, err := a.postWebhookReq(map[string]interface{}{
			"msg_type": "interactive",
			"card": map[string]any{
				"type": "template",
				"data": map[string]any{
					"template_id":           a.templateId,
					"template_version_name": a.templateVersion,
					"template_variable":     vars,
				},
			},
		}, a.receiveId)

		if status != 0 || err != nil {
			log.Error().Int("status", status).Err(err).Msg("webhook url returned error")
		}

		return err
	} else {
		b, err := json.Marshal(map[string]any{
			"type": "template",
			"data": map[string]any{
				"template_id":           a.templateId,
				"template_version_name": a.templateVersion,
				"template_variable":     vars,
			},
		})

		if err != nil {
			return err
		}
		req := larkim.NewCreateMessageReqBuilder().
			ReceiveIdType(a.receiveIdType).
			Body(larkim.NewCreateMessageReqBodyBuilder().
				ReceiveId(a.receiveId).
				MsgType(`interactive`).
				Content(string(b)).
				Build()).
			Build()

		resp, err := a.c.Im.V1.Message.Create(context.Background(), req)
		if err != nil {
			return err
		}

		if !resp.Success() {
			return fmt.Errorf("logId: %s, error response: \n%s", resp.RequestId(), larkcore.Prettify(resp.CodeError))
		}

	}
	log.Info().Str("receive_id_type", a.receiveIdType).Str("receive_id", a.receiveId).Msg("feishu message sent")

	return nil
}

func (a *FeishuBot) SendTextMessage(text string) error {

	if a.receiveIdType == "webhook" {
		status, err := a.postWebhookReq(map[string]interface{}{
			"msg_type": "text",
			"content": map[string]string{
				"text": text,
			},
		}, a.receiveId)

		if status != 0 || err != nil {
			log.Error().Int("status", status).Err(err).Msg("webhook url returned error")
		}

		return err
	} else {
		b, err := json.Marshal(map[string]string{
			"text": text,
		})

		if err != nil {
			return err
		}
		req := larkim.NewCreateMessageReqBuilder().
			ReceiveIdType(a.receiveIdType).
			Body(larkim.NewCreateMessageReqBodyBuilder().
				ReceiveId(a.receiveId).
				MsgType(`text`).
				Content(string(b)).
				Build()).
			Build()

		resp, err := a.c.Im.V1.Message.Create(context.Background(), req)
		if err != nil {
			return err
		}

		if !resp.Success() {
			return fmt.Errorf("logId: %s, error response: \n%s", resp.RequestId(), larkcore.Prettify(resp.CodeError))
		}

	}
	log.Info().Str("receive_id_type", a.receiveIdType).Str("receive_id", a.receiveId).Msg("feishu message sent")

	return nil
}

func (b *FeishuBot) NotifyNewSSHLogin(ruser, rhost, user string, t time.Time) error {

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
		log.Error().Err(err).Msg("could not get hostname")
	}

	vars := map[string]any{
		"time": t.Format(time.RFC1123),
		"src":  ruser + "@" + rhost,
		"dst":  user + "@" + hostname,
	}

	err = b.SendTemplateMessage(vars)
	if err != nil {
		log.Error().Err(err).Str("receive_id_type", b.receiveIdType).Str("receive_id", b.receiveId).Msg("SendTemplateMessage returned error")
	}
	return err
}

func (b *FeishuBot) NotifyPAMAuthenticate(items map[string]string) error {
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

func (b *FeishuBot) NotifyPAMOpenSession(items map[string]string) error {
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
