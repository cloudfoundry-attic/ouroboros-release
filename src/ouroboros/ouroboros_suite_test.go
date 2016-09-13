package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestOuroboros(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ouroboros Suite")
}
