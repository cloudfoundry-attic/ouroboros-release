package app_test

import (
	"conf"
	"time"
	"volley/app"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Killer", func() {
	var (
		killed chan struct{}
		k      *app.Killer
	)

	BeforeEach(func() {
		killed = make(chan struct{})
		k = app.NewKiller(
			conf.DurationRange{Min: time.Millisecond, Max: 5 * time.Millisecond},
			func() { close(killed) },
		)
		go k.Start()
	})

	It("kills the app", func() {
		Eventually(killed).Should(BeClosed())
	})
})
