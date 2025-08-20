package ctrl

import (
	"errors"
	"slices"

	"github.com/rs/zerolog/log"
	"github.com/yiffyi/bigbrother/ppp/model"
)

type ProxyServerConfig struct {
	Hostname       string
	Tags           []string
	EndpointGroups []string

	ProgramType        model.ProgramType
	ConfigTemplatePath string
}

type ProxyServerController struct {
	servers    map[string]ProxyServerConfig
	srvCfgTmpl map[string]ConfigTemplate

	activeServers map[string]bool

	grpEpConfig map[string][]ProxyEndpointInfo
}

func NewProxyServerController(servers []ProxyServerConfig, endpoints []ProxyEndpointInfo) *ProxyServerController {
	c := &ProxyServerController{
		servers:       map[string]ProxyServerConfig{},
		srvCfgTmpl:    map[string]ConfigTemplate{},
		activeServers: map[string]bool{},
		grpEpConfig:   map[string][]ProxyEndpointInfo{},
	}

	for _, srv := range servers {
		if srv.ProgramType == model.PROGRAM_TYPE_SINGBOX {
			c.servers[srv.Hostname] = srv
			c.srvCfgTmpl[srv.Hostname] = &SingBoxSubscriptionTemplate{TemplatePath: srv.ConfigTemplatePath}
		} else {
			log.Error().Str("program_type", string(srv.ProgramType)).Msg("unsupported server program type")
		}
	}

	for _, ep := range endpoints {
		c.grpEpConfig[ep.Group] = append(c.grpEpConfig[ep.Group], ep)
	}

	return c
}

func (c *ProxyServerController) RegisterActiveServer(hostname string) error {
	if _, ok := c.servers[hostname]; ok {
		c.activeServers[hostname] = true
		return nil
	} else {
		return errors.New("could not found hostname")
	}
}

func (c *ProxyServerController) UnregisterActiveServer(hostname string) {
	delete(c.activeServers, hostname)
}

func (c *ProxyServerController) CollectActiveEndpoints(tag string, groups []string) (res []ProxyEndpointInfo, err error) {
	for hostname, ok := range c.activeServers {
		hostnameEndpoints := []ProxyEndpointInfo{}
		if ok && slices.Contains(c.servers[hostname].Tags, tag) { // include this server
			for _, g := range groups { // include this endpoint group
				hostnameEndpoints = append(hostnameEndpoints, c.grpEpConfig[g]...)
			}
			for _, v := range hostnameEndpoints {
				v.FillInHostname(hostname)
				res = append(res, v)
			}

		}
	}

	return
}

func (c *ProxyServerController) RenderServerConfigTemplate(programType model.ProgramType, hostname string) ([]byte, error) {
	switch programType {
	case model.PROGRAM_TYPE_SINGBOX:
		if srv, ok := c.servers[hostname]; ok {
			groups := srv.EndpointGroups
			if len(groups) <= 0 {
				// log.Warn().Strs("groups", groups).Msg()
				return nil, errors.New("no group assigned to this hostname")
			}

			endpoints := []ProxyEndpointInfo{}
			for _, g := range groups {
				if e, ok := c.grpEpConfig[g]; ok {
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

			if t, ok := c.srvCfgTmpl[hostname]; ok {
				return t.RenderServerConfigTemplate(endpoints)
			} else {
				log.Error().Msg("no template attached to server")
			}

		} else {
			return nil, errors.New("could not find server hostname in configs")
		}
		return nil, nil
	default:
		return nil, errors.New("unsupported program type")
	}
}
