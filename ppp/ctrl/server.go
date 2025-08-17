package ctrl

import (
	"errors"
	"slices"

	"github.com/rs/zerolog/log"
	"github.com/yiffyi/bigbrother/ppp/model"
)

type ProxyServerController struct {
	hostnameTags           map[string][]string
	hostnameEndpointGroups map[string][]string
	hostnameCfgTemplate    map[string]ConfigTemplate

	activeServers map[string]bool

	groupBaseEndpoints map[string][]ProxyEndpointInfo
}

func (c *ProxyServerController) RegisterActiveServer(hostname string) {
	c.activeServers[hostname] = true
}

func (c *ProxyServerController) UnregisterActiveServer(hostname string) {
	delete(c.activeServers, hostname)
}

func (c *ProxyServerController) CollectActiveEndpoints(tag string, groups []string) (res []ProxyEndpointInfo, err error) {
	for hostname, ok := range c.activeServers {
		hostnameEndpoints := []ProxyEndpointInfo{}
		if ok && slices.Contains(c.hostnameTags[hostname], tag) { // include this server
			for _, g := range groups { // include this endpoint group
				hostnameEndpoints = append(hostnameEndpoints, c.groupBaseEndpoints[g]...)
			}
			for _, v := range hostnameEndpoints {
				v.FillInHostname(hostname)
			}

			res = append(res, hostnameEndpoints...)
		}
	}

	return
}

func (c *ProxyServerController) RenderServerConfigTemplate(programType model.ProgramType, hostname string) ([]byte, error) {
	switch programType {
	case model.PROGRAM_TYPE_SINGBOX:
		if groups, ok := c.hostnameEndpointGroups[hostname]; ok {
			if len(groups) <= 0 {
				// log.Warn().Strs("groups", groups).Msg()
				return nil, errors.New("no group assigned to this hostname")
			}

			endpoints := []ProxyEndpointInfo{}
			for _, g := range groups {
				if e, ok := c.groupBaseEndpoints[g]; ok {
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

			if t, ok := c.hostnameCfgTemplate[hostname]; ok {
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
