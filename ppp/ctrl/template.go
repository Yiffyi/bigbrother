package ctrl

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/yiffyi/bigbrother/ppp/model"
	"gopkg.in/yaml.v3"
)

type ConfigTemplate interface {
	ProgramType() model.ProgramType
	ContentType() string
	RenderUserConfigTemplate(outbounds []ProxyEndpointInfo) ([]byte, error)
	RenderServerConfigTemplate(inbounds []ProxyEndpointInfo) ([]byte, error)
}

type SingBoxSubscriptionTemplate struct {
	TemplatePath string
}

func (g *SingBoxSubscriptionTemplate) ProgramType() model.ProgramType {
	return model.PROGRAM_TYPE_SINGBOX
}

func (g *SingBoxSubscriptionTemplate) ContentType() string {
	return "application/json"
}

func (g *SingBoxSubscriptionTemplate) RenderUserConfigTemplate(servers []ProxyEndpointInfo) ([]byte, error) {
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
			info, err := s.GenerateUserConfig(model.PROGRAM_TYPE_SINGBOX)
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

func (g *SingBoxSubscriptionTemplate) RenderServerConfigTemplate(servers []ProxyEndpointInfo) ([]byte, error) {
	return nil, nil
}

type ClashSubscriptionTemplate struct {
	TemplatePath string
}

func (g *ClashSubscriptionTemplate) ProgramType() model.ProgramType {
	return model.PROGRAM_TYPE_CLASH
}

func (g *ClashSubscriptionTemplate) ContentType() string {
	return "application/yaml"
}

func (g *ClashSubscriptionTemplate) RenderUserConfigTemplate(servers []ProxyEndpointInfo) ([]byte, error) {
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
			info, err := s.GenerateUserConfig(model.PROGRAM_TYPE_CLASH)
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

func (g *ClashSubscriptionTemplate) RenderServerConfigTemplate(servers []ProxyEndpointInfo) ([]byte, error) {
	return nil, nil
}
