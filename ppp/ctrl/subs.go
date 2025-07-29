package ctrl

import (
	"encoding/json"
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type SubscriptionGenerator interface {
	ClientType() string
	ContentType() string
	RenderTemplate(servers []ProxyServerInfo) ([]byte, error)
}

type SingBoxSubscriptionGenerator struct {
	templatePath string
}

func (g *SingBoxSubscriptionGenerator) ClientType() string {
	return "sing-box"
}

func (g *SingBoxSubscriptionGenerator) ContentType() string {
	return "application/json"
}

func (g *SingBoxSubscriptionGenerator) RenderTemplate(servers []ProxyServerInfo) ([]byte, error) {
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

func (g *ClashSubscriptionGenerator) ClientType() string {
	return "clash"
}

func (g *ClashSubscriptionGenerator) ContentType() string {
	return "application/yaml"
}

func (g *ClashSubscriptionGenerator) RenderTemplate(servers []ProxyServerInfo) ([]byte, error) {
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

type SubscriptionController struct {
	genMap  map[string]SubscriptionGenerator
	servers []ProxyServerInfo
}

func NewSubscriptionController(generators []SubscriptionGenerator) (*SubscriptionController, error) {
	c := &SubscriptionController{}
	for _, v := range generators {
		if _, ok := c.genMap[v.ClientType()]; !ok {
			c.genMap[v.ClientType()] = v
		} else {
			return nil, errors.New("conflict ClientType found in subscription generators")
		}
	}

	return c, nil
}

func (c *SubscriptionController) GetSubscription(clientType string) ([]byte, error) {
	if gen, ok := c.genMap[clientType]; ok {
		return gen.RenderTemplate(c.servers)
	} else {
		return nil, errors.New("unsupported proxy type")
	}

}
