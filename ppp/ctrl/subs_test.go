package ctrl_test

import (
	"encoding/json"
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/yiffyi/bigbrother/ppp/ctrl"
	"github.com/yiffyi/bigbrother/ppp/model"
	"gopkg.in/yaml.v3"
)

var singBoxUserBase = `{
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
}`

var clashUserBase = `
proxies: []

proxy-groups:
  - {
      name: "PROXY",
      type: select,
      proxies:
        [
          "Hysteria - Seoul, AWS",
          "Hysteria - Tokyo, BWH",
          "Hysteria - Tokyo, AWS",
          "VMESS - Tokyo, BWH",
          "VMESS - Seoul, AWS",
          "VMESS - LAX, BWH",
          "Hysteria - LAX, BWH",
        ],
    }
  # - { name: 'SPECIAL', type: select, proxies: ["VMESS - Tokyo, BWH", "VMESS - LAX, BWH", "Hysteria - LAX, BWH"] }

dns:
  enable: true
  proxy-server-nameserver:
    - 223.5.5.5
    - 223.6.6.6
    # - system
  direct-nameserver:
    - 223.5.5.5
    - 223.6.6.6
    # - system
  nameserver:
    - https://1.1.1.1/dns-query#PROXY
    - https://1.0.0.1/dns-query#PROXY

rules:
  - GEOIP,LAN,DIRECT
  - GEOIP,CN,DIRECT
  - IP-CIDR,100.64.0.0/10,DIRECT
  - IP-CIDR,198.19.0.0/16,DIRECT
  - MATCH,PROXY
`

var _ = Describe("Subs", func() {
	var tempDir string
	var singBoxBasePath string
	var clashBasePath string

	var exampleEndpointConfig = ctrl.ProxyEndpointInfo{
		Protocol:   "hysteria2",
		Tag:        "hy2,local",
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
	}

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "ppp-test-*")
		Expect(err).To(BeNil())

		singBoxBasePath = path.Join(tempDir, "sing-box.base.json")
		err = os.WriteFile(singBoxBasePath, []byte(singBoxUserBase), os.ModePerm)
		Expect(err).To(BeNil())

		clashBasePath = path.Join(tempDir, "clash.base.yaml")
		err = os.WriteFile(clashBasePath, []byte(clashUserBase), os.ModePerm)
		Expect(err).To(BeNil())
	})

	It("should produce sing-box subscription", func() {
		gen := ctrl.SingBoxSubscriptionTemplate{
			TemplatePath: singBoxBasePath,
		}
		Expect(gen.ProgramType()).To(Equal(model.PROGRAM_TYPE_SINGBOX))
		Expect(gen.ContentType()).To(Equal("application/json"))

		res, err := gen.RenderUserConfigTemplate([]ctrl.ProxyEndpointInfo{exampleEndpointConfig})
		Expect(err).To(BeNil())

		GinkgoWriter.Print("RenderTemplate output:", string(res))

		var j map[string]any
		err = json.Unmarshal(res, &j)
		Expect(err).To(BeNil())

		Expect(len(j["outbounds"].([]any)) > 1).To(BeTrue())
	})

	It("should serve subscriptions", func() {
		c, err := ctrl.NewSubscriptionController(
			[]ctrl.ConfigTemplate{
				&ctrl.SingBoxSubscriptionTemplate{TemplatePath: singBoxBasePath},
				&ctrl.ClashSubscriptionTemplate{TemplatePath: clashBasePath},
			},
			[]ctrl.ProxyEndpointInfo{exampleEndpointConfig},
		)
		Expect(err).To(BeNil())

		b, err := c.GetSubscription("sing-box")
		Expect(err).To(BeNil())

		var j map[string]any
		err = json.Unmarshal(b, &j)
		Expect(err).To(BeNil())

		b, err = c.GetSubscription("clash")
		Expect(err).To(BeNil())
		GinkgoWriter.Print("clash subscription:", string(b))

		err = yaml.Unmarshal(b, &j)
		Expect(err).To(BeNil())

	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).To(BeNil())
	})
})
