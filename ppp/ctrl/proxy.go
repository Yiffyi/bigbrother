package ctrl

import (
	"encoding/json"
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type SubscriptionGenerator interface {
	ProxyType() string
	ContentType() string
	RenderTemplate() ([]byte, error)
}

type SingBoxSubscriptionGenerator struct {
	templatePath string
}

func (g *SingBoxSubscriptionGenerator) ProxyType() string {
	return "sing-box"
}

func (g *SingBoxSubscriptionGenerator) ContentType() string {
	return "application/json"
}

func (g *SingBoxSubscriptionGenerator) RenderTemplate() ([]byte, error) {
	b, err := os.ReadFile(g.templatePath)
	if err != nil {
		return nil, err
	}

	var p map[string]interface{}
	err = json.Unmarshal(b, &p)
	if err != nil {
		return nil, err
	}

	b, err = json.MarshalIndent(p, "", "    ")
	if err != nil {
		return nil, err
	}

	return b, nil
}

type ClashSubscriptionGenerator struct {
	templatePath string
}

func (g *ClashSubscriptionGenerator) ProxyType() string {
	return "clash"
}

func (g *ClashSubscriptionGenerator) ContentType() string {
	return "application/yaml"
}

func (g *ClashSubscriptionGenerator) RenderTemplate() ([]byte, error) {
	b, err := os.ReadFile(g.templatePath)
	if err != nil {
		return nil, err
	}

	var p map[string]interface{}
	err = yaml.Unmarshal(b, &p)
	if err != nil {
		return nil, err
	}

	b, err = yaml.Marshal(p)
	if err != nil {
		return nil, err
	}

	return b, nil
}

type ProxyController struct {
	clashSub   SubscriptionGenerator
	singBoxSub SubscriptionGenerator
}

func (c *ProxyController) GetSubscription(proxyType string) ([]byte, error) {
	switch proxyType {
	case "sing-box":
		return c.singBoxSub.RenderTemplate()
	case "clash":
		return c.clashSub.RenderTemplate()
	default:
		return nil, errors.New("unsupported proxy type")
	}
}
