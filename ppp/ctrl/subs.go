package ctrl

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/yiffyi/bigbrother/ppp/model"
	"gopkg.in/yaml.v3"
)

type SubscriptionTemplate interface {
	ClientType() model.ProgramType
	ContentType() string
	RenderTemplate(servers []ProxyEndpointInfo) ([]byte, error)
}

type SingBoxSubscriptionTemplate struct {
	TemplatePath string
}

func (g *SingBoxSubscriptionTemplate) ClientType() model.ProgramType {
	return model.PROGRAM_TYPE_SINGBOX
}

func (g *SingBoxSubscriptionTemplate) ContentType() string {
	return "application/json"
}

func (g *SingBoxSubscriptionTemplate) RenderTemplate(servers []ProxyEndpointInfo) ([]byte, error) {
	b, err := os.ReadFile(g.TemplatePath)
	if err != nil {
		return nil, err
	}

	var p map[string]interface{}
	err = json.Unmarshal(b, &p)
	if err != nil {
		return nil, err
	}

	if outbounds, ok := p["outbounds"].([]any); ok {
		for _, s := range servers {
			info := map[string]any{
				"type":        s.Protocol,
				"tag":         s.Tag,
				"server":      s.Server,
				"server_port": s.ServerPort,
			}

			info, err := s.SupplementInfo.SpecializeUserConfig(model.PROGRAM_TYPE_SINGBOX, info)
			if err != nil {
				return nil, err
			}
			outbounds = append(outbounds, info)
		}
		p["outbounds"] = outbounds
	} else {
		return nil, errors.New("could not found outbounds section in sing-box base config")
	}

	b, err = json.MarshalIndent(p, "", "    ")
	if err != nil {
		return nil, err
	}

	return b, nil
}

type ClashSubscriptionTemplate struct {
	TemplatePath string
}

func (g *ClashSubscriptionTemplate) ClientType() model.ProgramType {
	return model.PROGRAM_TYPE_CLASH
}

func (g *ClashSubscriptionTemplate) ContentType() string {
	return "application/yaml"
}

func (g *ClashSubscriptionTemplate) RenderTemplate(servers []ProxyEndpointInfo) ([]byte, error) {
	b, err := os.ReadFile(g.TemplatePath)
	if err != nil {
		return nil, err
	}

	var p map[string]interface{}
	err = yaml.Unmarshal(b, &p)
	if err != nil {
		return nil, err
	}

	if outbounds, ok := p["proxies"].([]any); ok {
		for _, s := range servers {
			info := map[string]any{
				"name":   s.Tag,
				"type":   s.Protocol,
				"server": s.Server,
				"port":   s.ServerPort,
			}

			info, err := s.SupplementInfo.SpecializeUserConfig(model.PROGRAM_TYPE_CLASH, info)
			if err != nil {
				return nil, err
			}
			outbounds = append(outbounds, info)
		}
		p["proxies"] = outbounds
	} else {
		return nil, errors.New("could not found proxies section in clash base config")
	}

	b, err = yaml.Marshal(p)
	if err != nil {
		return nil, err
	}

	return b, nil
}

type SubscriptionController struct {
	genMap  map[model.ProgramType]SubscriptionTemplate
	servers []ProxyEndpointInfo
}

func NewSubscriptionController(generators []SubscriptionTemplate, servers []ProxyEndpointInfo) (*SubscriptionController, error) {
	c := &SubscriptionController{
		genMap:  map[model.ProgramType]SubscriptionTemplate{},
		servers: nil,
	}
	for _, v := range generators {
		if _, ok := c.genMap[v.ClientType()]; !ok {
			c.genMap[v.ClientType()] = v
		} else {
			return nil, errors.New("conflict ClientType found in subscription generators")
		}
	}

	c.servers = servers

	return c, nil
}

func (c *SubscriptionController) GetSubscription(clientType string) ([]byte, error) {
	if gen, ok := c.genMap[model.ProgramType(clientType)]; ok {
		return gen.RenderTemplate(c.servers)
	} else {
		return nil, errors.New("unsupported proxy type")
	}

}
