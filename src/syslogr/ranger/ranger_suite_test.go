package ranger_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRanger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ranger Suite")
}
