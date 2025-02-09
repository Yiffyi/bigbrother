package honeypot_test

import (
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"gorm.io/gorm"

	"github.com/yiffyi/bigbrother/honeypot"
)

var _ = Describe("Rec", func() {
	var db *gorm.DB
	var err error

	BeforeEach(func() {
		v := viper.New()
		v.SetDefault("db", `file::memory:?cache=shared`)
		db, err = honeypot.OpenDatabase(v)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can decide whether to accept user & pass", func() {
		addr := &net.TCPAddr{net.ParseIP("127.0.0.1"), 65432, ""}
		err = honeypot.RecordAuthAttempt(db, "root", "123456", addr)
		Expect(err).NotTo(HaveOccurred())
		err = honeypot.RecordAuthAttempt(db, "root", "345678", addr)
		Expect(err).NotTo(HaveOccurred())
		err = honeypot.RecordAuthAttempt(db, "root", "123456", addr)
		Expect(err).NotTo(HaveOccurred())

		allow, err := honeypot.ShallAllowConnection(db, "root", "123456", addr)
		Expect(err).NotTo(HaveOccurred())
		Expect(allow).To(BeTrue())

		allow, err = honeypot.ShallAllowConnection(db, "root", "345678", addr)
		Expect(err).NotTo(HaveOccurred())
		Expect(allow).NotTo(BeTrue())
	})
})
