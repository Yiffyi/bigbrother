package ctrl

import (
	"errors"

	"github.com/yiffyi/bigbrother/ppp/model"
)

type SubscriptionController struct {
	genMap  map[model.ProgramType]ConfigTemplate
	servers []ProxyEndpointInfo
}

func NewSubscriptionController(generators []ConfigTemplate, servers []ProxyEndpointInfo) (*SubscriptionController, error) {
	c := &SubscriptionController{
		genMap:  map[model.ProgramType]ConfigTemplate{},
		servers: nil,
	}
	for _, v := range generators {
		if _, ok := c.genMap[v.ProgramType()]; !ok {
			c.genMap[v.ProgramType()] = v
		} else {
			return nil, errors.New("conflict ClientType found in subscription generators")
		}
	}

	c.servers = servers

	return c, nil
}

func (c *SubscriptionController) GetSubscription(clientType string) ([]byte, error) {
	if gen, ok := c.genMap[model.ProgramType(clientType)]; ok {
		return gen.RenderUserConfigTemplate(c.servers)
	} else {
		return nil, errors.New("unsupported proxy type")
	}

}
