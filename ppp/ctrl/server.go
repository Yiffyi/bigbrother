package ctrl

import (
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/yiffyi/bigbrother/ppp/model"
)

type ProxyServerController struct {
	serverGroups   map[string][]string
	groupEndpoints map[string][]ProxyEndpointInfo
	groupTemplates map[string]ConfigTemplate
}

func (c *ProxyServerController) RenderServerConfigTemplate(programType model.ProgramType, hostname string) ([]byte, error) {
	switch programType {
	case model.PROGRAM_TYPE_SINGBOX:
		if groups, ok := c.serverGroups[hostname]; ok {
			if len(groups) <= 0 {
				// log.Warn().Strs("groups", groups).Msg()
				return nil, errors.New("no group assigned to this hostname")
			}

			endpoints := []ProxyEndpointInfo{}
			for _, g := range groups {
				if e, ok := c.groupEndpoints[g]; ok {
					endpoints = append(endpoints, e...)
				} else {
					log.Warn().Str("group", g).Msg("could not find group")
				}
			}

			for _, e := range endpoints {
				e.FillInHostname(hostname)
			}

			if len(endpoints) <= 0 {
				log.Warn().Msg("no endpoints added over base config")
			}

			if t, ok := c.groupTemplates[groups[0]]; ok {
				return t.RenderServerConfigTemplate(endpoints)
			} else {
				log.Error().Msg("no template attached primary group")
			}

		} else {
			return nil, errors.New("could not find server hostname in configs")
		}
		return nil, nil
	default:
		return nil, errors.New("unsupported program type")
	}
}
