package honeypot_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHoneypot(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Honeypot Suite")
}
