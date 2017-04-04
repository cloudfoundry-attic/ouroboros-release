//go:generate hel

package v1_test

import (
	"log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConnection(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Connection Suite")
}

var _ = BeforeSuite(func() {
	if !testing.Verbose() {
		log.SetOutput(GinkgoWriter)
	}
})
