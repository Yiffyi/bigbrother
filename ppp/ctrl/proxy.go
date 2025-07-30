package ctrl

import (
	"errors"
	"fmt"

	"github.com/yiffyi/bigbrother/ppp/model"
)

type ProxyServerSupplementInfo interface {
	SpecializeClientConfig(clientType model.ProgramType, genericInfo map[string]any) (map[string]any, error)
	SpecializeServerConfig(serverType model.ProgramType, genericInfo map[string]any) (map[string]any, error)
}

type ProxyServerInfo struct {
	Protocol   string
	Tag        string
	Server     string
	ServerPort int
	Password   string

	SupplementInfo ProxyServerSupplementInfo
}

type Hysteria2SupplementInfo struct {
	up            int
	down          int
	tls           bool
	tlsServerName string
}

func (s *Hysteria2SupplementInfo) singBox() map[string]any {
	return map[string]any{
		"up_mbps":   s.up,
		"down_mbps": s.down,
		"tls": map[string]any{
			"enabled":     s.tls,
			"server_name": s.tlsServerName,
		},
	}
}

func (s *Hysteria2SupplementInfo) clash() map[string]any {
	return map[string]any{
		"tls":  s.tls,
		"sni":  s.tlsServerName,
		"up":   fmt.Sprintf("%d Mbps", s.up),
		"down": fmt.Sprintf("%d Mbps", s.down),
	}
}

func (s *Hysteria2SupplementInfo) SpecializeClientConfig(clientType model.ProgramType, genericInfo map[string]any) (map[string]any, error) {
	var supp map[string]any
	switch clientType {
	case "sing-box":
		supp = s.singBox()
	case "clash":
		supp = s.clash()
	default:
		return nil, errors.New("unsupported client type")
	}

	for k, v := range supp {
		genericInfo[k] = v
	}

	return genericInfo, nil
}
