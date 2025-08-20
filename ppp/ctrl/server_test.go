package ctrl_test

import (
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/yiffyi/bigbrother/ppp/ctrl"
)

var singBoxServerBase = `
{
    "log": {
        "level": "info"
    },
    "dns": {
        "servers": [
            {
                "tag": "cloudflare-1",
                "address": "https://1.1.1.1/dns-query",
                "detour": "direct-out"
            }
        ],
        "strategy": "prefer_ipv4"
    },
    "inbounds": [
    ],
    "outbounds": [
        {
            "type": "direct",
            "tag": "direct-out"
        }
    ],
    "route": {
        "rules": [
            {
                "action": "sniff",
                "timeout": "1s"
            },
            {
                "protocol": "quic",
                "action": "reject"
            }
        ],
        "rule_set": []
    }
}
`

var _ = Describe("Server", func() {
	var tempDir string
	var singBoxBasePath string
	var exampleServers []ctrl.ProxyServerConfig
	var exampleEndpoints []ctrl.ProxyEndpointInfo
	var exampleController *ctrl.ProxyServerController

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "ppp-test-*")
		Expect(err).To(BeNil())

		singBoxBasePath = path.Join(tempDir, "sing-box.base.json")
		err = os.WriteFile(singBoxBasePath, []byte(singBoxServerBase), os.ModePerm)
		Expect(err).To(BeNil())

		exampleServers = []ctrl.ProxyServerConfig{
			{
				Hostname:       "test01.doesn-t.work",
				Tags:           []string{"tag01", "tag99"},
				EndpointGroups: []string{"group01", "group99"},

				ProgramType:        "sing-box",
				ConfigTemplatePath: singBoxBasePath,
			},
			{
				Hostname:       "test02.doesn-t.work",
				Tags:           []string{"tag02", "tag99"},
				EndpointGroups: []string{"group02", "group99"},

				ProgramType:        "sing-box",
				ConfigTemplatePath: singBoxBasePath,
			},
		}

		exampleEndpoints = []ctrl.ProxyEndpointInfo{
			{
				Group:      "group01",
				Protocol:   "hysteria2",
				Tag:        "hy2,local,group01",
				Server:     "127.0.0.1",
				ServerPort: 8443,
				SupplementInfo: &ctrl.Hysteria2SupplementInfo{
					Passwords:     []string{"lo,pw0", "lo,pw1"},
					Up:            0,
					Down:          0,
					TLS:           false,
					TLSServerName: "localhost",
					ACMEEmail:     "localhost@localdomain",
				},
			},
			{
				Group:      "group02",
				Protocol:   "hysteria2",
				Tag:        "hy2,local,group02",
				Server:     "127.0.0.2",
				ServerPort: 8443,
				SupplementInfo: &ctrl.Hysteria2SupplementInfo{
					Passwords:     []string{"lo,pw1", "lo,pw2"},
					Up:            0,
					Down:          0,
					TLS:           false,
					TLSServerName: "localhost",
					ACMEEmail:     "localhost@localdomain",
				},
			},
			{
				Group:      "group99",
				Protocol:   "vmess",
				Tag:        "vmess,local,group99",
				Server:     "127.0.0.99",
				ServerPort: 8443,
				SupplementInfo: &ctrl.VmessSupplementInfo{
					UUIDs: []string{
						"65a9c572-8227-4473-9308-eaa707195525", "e2bbf8d3-bff8-4f68-bda6-190bb3973404",
					},
					Security:          "auto",
					AlterId:           0,
					TLS:               true,
					TLSServerName:     "localhost",
					UTLS:              true,
					UTLSFingerprint:   "ios",
					Reality:           true,
					RealityPrivateKey: "some base64 string",
					RealityPublicKey:  "some base64 string",
					RealityShortId:    "some hex string",
					Multiplex:         true,
				},
			},
		}

		exampleController = ctrl.NewProxyServerController(exampleServers, exampleEndpoints)
		Expect(exampleController).ToNot(BeNil())
	})

	It("should match 1 host for tag01, tag02", func() {
		var err error
		err = exampleController.RegisterActiveServer("test01.doesn-t.work")
		Expect(err).ShouldNot(HaveOccurred())
		err = exampleController.RegisterActiveServer("test02.doesn-t.work")
		Expect(err).ShouldNot(HaveOccurred())

		eps, err := exampleController.CollectActiveEndpoints("tag01", []string{"group99"})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(len(eps)).To(Equal(1))
		Expect(eps[0].Server).To(Equal("test01.doesn-t.work"))

		eps, err = exampleController.CollectActiveEndpoints("tag02", []string{"group99"})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(len(eps)).To(Equal(1))
		Expect(eps[0].Server).To(Equal("test02.doesn-t.work"))
	})

	It("should get two endpoints for group01+group99", func() {
		var err error
		err = exampleController.RegisterActiveServer("test01.doesn-t.work")
		Expect(err).ShouldNot(HaveOccurred())

		eps, err := exampleController.CollectActiveEndpoints("tag01", []string{"group01", "group99"})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(len(eps)).To(Equal(2))
		Expect(eps[0].Server).To(Equal("test01.doesn-t.work"))
		Expect(eps[1].Server).To(Equal("test01.doesn-t.work"))
	})

})
