package ctrl

import (
	"errors"

	"github.com/yiffyi/bigbrother/ppp/model"
)

type SubscriptionController struct {
	genMap map[model.ProgramType]ConfigTemplate
}

func NewSubscriptionController(generators []ConfigTemplate) (*SubscriptionController, error) {
	c := &SubscriptionController{
		genMap: map[model.ProgramType]ConfigTemplate{},
	}
	for _, v := range generators {
		if _, ok := c.genMap[v.ProgramType()]; !ok {
			c.genMap[v.ProgramType()] = v
		} else {
			return nil, errors.New("conflict ClientType found in subscription generators")
		}
	}

	return c, nil
}

func (c *SubscriptionController) GetSubscription(clientType string, endpoints []ProxyEndpointInfo) ([]byte, error) {
	if gen, ok := c.genMap[model.ProgramType(clientType)]; ok {
		return gen.RenderUserConfigTemplate(endpoints)
	} else {
		return nil, errors.New("unsupported proxy type")
	}

}
