//go:generate hel

package conns_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConns(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Conns Suite")
}
