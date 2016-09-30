//go:generate hel

package drains_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDrains(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Drains Suite")
}
