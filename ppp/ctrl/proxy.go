package ctrl

import (
	"errors"
	"fmt"
	"maps"

	"github.com/yiffyi/bigbrother/ppp/model"
)

type ProxyServerSupplementInfo interface {
	SpecializeClientConfig(clientType model.ProgramType, genericInfo map[string]any) (map[string]any, error)
	SpecializeServerConfig(serverType model.ProgramType, genericInfo map[string]any) (map[string]any, error)
}

type ProxyServerInfo struct {
	Protocol       string
	Tag            string
	Server         string
	ServerPort     int
	SupplementInfo ProxyServerSupplementInfo
}

type Hysteria2SupplementInfo struct {
	password      string
	up            int
	down          int
	tls           bool
	tlsServerName string
	acmeEmail     string
}

func (s *Hysteria2SupplementInfo) singBoxClient() map[string]any {
	return map[string]any{
		"password":  s.password,
		"up_mbps":   s.up,
		"down_mbps": s.down,
		"tls": map[string]any{
			"enabled":     s.tls,
			"server_name": s.tlsServerName,
		},
	}
}

func (s *Hysteria2SupplementInfo) singBoxServer() map[string]any {
	return map[string]any{
		"users": []map[string]any{
			{
				"name":     "sekai",
				"password": s.password,
			},
		},
		"up_mbps":                 s.up,
		"down_mbps":               s.down,
		"ignore_client_bandwidth": false,
		"tls": map[string]any{
			"enabled":     s.tls,
			"server_name": s.tlsServerName,
			"acme": map[string]any{
				"domain": s.tlsServerName,
				"email":  s.acmeEmail,
			},
			"alpn": []string{
				"h3",
			},
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
	uuid              string
	security          string
	alterId           int
	tls               bool
	tlsServerName     string
	utls              bool
	utlsFingerprint   string
	reality           bool
	realityPrivateKey string
	realityPublicKey  string
	realityShortId    string
	multiplex         bool
}

func (s *VmessSupplementInfo) singBoxClient() map[string]any {
	return map[string]any{
		"uuid":     s.uuid,
		"security": s.security,
		"alter_id": s.alterId,
		"tls": map[string]any{
			"enabled":     s.tls,
			"server_name": s.tlsServerName,
			"utls": map[string]any{
				"enabled":     s.utls,
				"fingerprint": s.utlsFingerprint,
			},
			"reality": map[string]any{
				"enabled":    s.reality,
				"public_key": s.realityPublicKey,
				"short_id":   s.realityShortId,
			},
		},
		"multiplex": map[string]any{
			"enabled":     s.multiplex,
			"max_streams": 3,
		},
	}
}

func (s *VmessSupplementInfo) singBoxServer() map[string]any {
	return map[string]any{
		"users": []map[string]any{
			{
				"name":    "sekai",
				"uuid":    s.uuid,
				"alterId": s.alterId,
			},
		},
		"tls": map[string]any{
			"enabled":     s.tls,
			"server_name": s.tlsServerName,
			"alpn": []string{
				"h2", "http/1.1", "h3",
			},
			"reality": map[string]any{
				"enabled": s.reality,
				"handshake": map[string]any{
					"server":      s.tlsServerName,
					"server_port": 443,
				},
				"private_key": s.realityPrivateKey,
				"short_id":    s.realityShortId,
			},
		},
		"multiplex": map[string]any{
			"enabled": s.multiplex,
		},
	}
}

func (s *VmessSupplementInfo) clash() map[string]any {
	return map[string]any{
		"udp":                true,
		"uuid":               s.uuid,
		"alterId":            s.alterId,
		"cipher":             s.security,
		"tls":                s.tls,
		"servername":         s.tlsServerName,
		"client-fingerprint": s.utlsFingerprint,
		"alpn": []string{
			"h2", "http/1.1", "h3",
		},
		"reality-opts": map[string]any{
			"public-key": s.realityPublicKey,
			"short-id":   s.realityShortId,
		},
		"smux": map[string]any{
			"enabled": s.multiplex,
		},
	}
}

func (s *VmessSupplementInfo) SpecializeClientConfig(clientType model.ProgramType, genericInfo map[string]any) (map[string]any, error) {
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
