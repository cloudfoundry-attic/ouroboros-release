package syslogdrain_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCups(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cups Suite")
}
