//go:generate hel

package connectionmanager_test

import (
	"io/ioutil"
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
	log.SetOutput(ioutil.Discard)
})
