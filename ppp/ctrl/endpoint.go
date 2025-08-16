package ctrl

import (
	"errors"
	"fmt"
	"maps"

	"github.com/yiffyi/bigbrother/ppp/model"
)

type ProxySupplementInfo interface {
	FillInHostname(hostname string)
	SpecializeUserConfig(clientType model.ProgramType, genericInfo map[string]any) (map[string]any, error)
	SpecializeServerConfig(serverType model.ProgramType, genericInfo map[string]any) (map[string]any, error)
}

type ProxyEndpointInfo struct {
	Protocol       string
	Tag            string
	Server         string
	ServerPort     int
	SupplementInfo ProxySupplementInfo
}

func (s *ProxyEndpointInfo) FillInHostname(hostname string) {
	s.Server = hostname
	s.SupplementInfo.FillInHostname(hostname)
}

func (s *ProxyEndpointInfo) GenerateUserConfig(clientType model.ProgramType) (map[string]any, error) {
	var info map[string]any
	switch clientType {
	case model.PROGRAM_TYPE_CLASH:
		info = map[string]any{
			"name":   s.Tag,
			"type":   s.Protocol,
			"server": s.Server,
			"port":   s.ServerPort,
		}
	case model.PROGRAM_TYPE_SINGBOX:
		info = map[string]any{
			"type":        s.Protocol,
			"tag":         s.Tag,
			"server":      s.Server,
			"server_port": s.ServerPort,
		}
	default:
		return nil, errors.New("unsupported clientType")
	}

	info, err := s.SupplementInfo.SpecializeUserConfig(clientType, info)
	if err != nil {
		return nil, err
	}

	return info, err
}

func (s *ProxyEndpointInfo) GenerateServerConfig(serverType model.ProgramType) (map[string]any, error) {
	var info map[string]any
	switch serverType {
	case model.PROGRAM_TYPE_SINGBOX:
		info = map[string]any{
			"type":        s.Protocol,
			"tag":         s.Tag,
			"listen":      "::", // deal with it later
			"listen_port": s.ServerPort,
		}
	default:
		return nil, errors.New("unsupported serverType")
	}

	info, err := s.SupplementInfo.SpecializeServerConfig(serverType, info)
	if err != nil {
		return nil, err
	}

	return info, err

}

type Hysteria2SupplementInfo struct {
	Passwords     []string
	Up            int
	Down          int
	TLS           bool
	TLSServerName string
	ACMEEmail     string
}

func (s *Hysteria2SupplementInfo) singBoxClient() map[string]any {
	return map[string]any{
		"password":  s.Passwords[0], // client always takes the first
		"up_mbps":   s.Up,
		"down_mbps": s.Down,
		"tls": map[string]any{
			"enabled":     s.TLS,
			"server_name": s.TLSServerName,
		},
	}
}

func (s *Hysteria2SupplementInfo) singBoxServer() map[string]any {
	users := []map[string]string{}
	for k, v := range s.Passwords {
		users = append(users,
			map[string]string{
				"name":     fmt.Sprintf("hy2:%d", k),
				"password": v,
			},
		)
	}

	return map[string]any{
		"users":                   users,
		"up_mbps":                 s.Up,
		"down_mbps":               s.Down,
		"ignore_client_bandwidth": false,
		"tls": map[string]any{
			"enabled":     s.TLS,
			"server_name": s.TLSServerName,
			// "acme": map[string]any{
			// 	"domain": s.TLSServerName,
			// 	"email":  s.ACMEEmail,
			// },
			"alpn": []string{
				"h3",
			},
		},
	}
}

func (s *Hysteria2SupplementInfo) clash() map[string]any {
	return map[string]any{
		"tls":      s.TLS,
		"sni":      s.TLSServerName,
		"password": s.Passwords[0],
		"up":       fmt.Sprintf("%d Mbps", s.Up),
		"down":     fmt.Sprintf("%d Mbps", s.Down),
	}
}

func (s *Hysteria2SupplementInfo) FillInHostname(hostname string) {
	s.TLSServerName = hostname
}

func (s *Hysteria2SupplementInfo) SpecializeUserConfig(clientType model.ProgramType, genericInfo map[string]any) (map[string]any, error) {
	var supp map[string]any
	switch clientType {
	case "sing-box":
		supp = s.singBoxClient()
	case "clash":
		supp = s.clash()
	default:
		return nil, errors.New("unsupported client type")
	}

	maps.Copy(genericInfo, supp)

	return genericInfo, nil
}

func (s *Hysteria2SupplementInfo) SpecializeServerConfig(serverType model.ProgramType, genericInfo map[string]any) (map[string]any, error) {
	var supp map[string]any
	switch serverType {
	case model.PROGRAM_TYPE_SINGBOX:
		supp = s.singBoxServer()
	default:
		return nil, errors.New("unsupported server type")
	}

	maps.Copy(genericInfo, supp)

	return genericInfo, nil
}

type VmessSupplementInfo struct {
	UUIDs             []string
	Security          string
	AlterId           int
	TLS               bool
	TLSServerName     string
	UTLS              bool
	UTLSFingerprint   string
	Reality           bool
	RealityPrivateKey string
	RealityPublicKey  string
	RealityShortId    string
	Multiplex         bool
}

func (s *VmessSupplementInfo) singBoxClient() map[string]any {
	return map[string]any{
		"uuid":     s.UUIDs[0],
		"security": s.Security,
		"alter_id": s.AlterId,
		"tls": map[string]any{
			"enabled":     s.TLS,
			"server_name": s.TLSServerName,
			"utls": map[string]any{
				"enabled":     s.UTLS,
				"fingerprint": s.UTLSFingerprint,
			},
			"reality": map[string]any{
				"enabled":    s.Reality,
				"public_key": s.RealityPublicKey,
				"short_id":   s.RealityShortId,
			},
		},
		"multiplex": map[string]any{
			"enabled":     s.Multiplex,
			"max_streams": 3,
		},
	}
}

func (s *VmessSupplementInfo) singBoxServer() map[string]any {
	users := []map[string]any{}
	for k, v := range s.UUIDs {
		users = append(users,
			map[string]any{
				"name":    fmt.Sprintf("vmess:%d", k),
				"uuid":    v,
				"alterId": s.AlterId,
			},
		)
	}
	return map[string]any{
		"users": users,
		"tls": map[string]any{
			"enabled":     s.TLS,
			"server_name": s.TLSServerName,
			"alpn": []string{
				"h2", "http/1.1", "h3",
			},
			"reality": map[string]any{
				"enabled": s.Reality,
				"handshake": map[string]any{
					"server":      s.TLSServerName,
					"server_port": 443,
				},
				"private_key": s.RealityPrivateKey,
				"short_id":    s.RealityShortId,
			},
		},
		"multiplex": map[string]any{
			"enabled": s.Multiplex,
		},
	}
}

func (s *VmessSupplementInfo) clash() map[string]any {
	return map[string]any{
		"udp":                true,
		"uuid":               s.UUIDs[0],
		"alterId":            s.AlterId,
		"cipher":             s.Security,
		"tls":                s.TLS,
		"servername":         s.TLSServerName,
		"client-fingerprint": s.UTLSFingerprint,
		"alpn": []string{
			"h2", "http/1.1", "h3",
		},
		"reality-opts": map[string]any{
			"public-key": s.RealityPublicKey,
			"short-id":   s.RealityShortId,
		},
		"smux": map[string]any{
			"enabled": s.Multiplex,
		},
	}
}

func (s *VmessSupplementInfo) FillInHostname(hostname string) {
	s.TLSServerName = hostname
}

func (s *VmessSupplementInfo) SpecializeUserConfig(clientType model.ProgramType, genericInfo map[string]any) (map[string]any, error) {
	var supp map[string]any
	switch clientType {
	case model.PROGRAM_TYPE_SINGBOX:
		supp = s.singBoxClient()
	case model.PROGRAM_TYPE_CLASH:
		supp = s.clash()
	default:
		return nil, errors.New("unsupported client type")
	}

	maps.Copy(genericInfo, supp)

	return genericInfo, nil
}

func (s *VmessSupplementInfo) SpecializeServerConfig(serverType model.ProgramType, genericInfo map[string]any) (map[string]any, error) {
	var supp map[string]any
	switch serverType {
	case model.PROGRAM_TYPE_SINGBOX:
		supp = s.singBoxServer()
	default:
		return nil, errors.New("unsupported server type")
	}

	maps.Copy(genericInfo, supp)

	return genericInfo, nil
}
