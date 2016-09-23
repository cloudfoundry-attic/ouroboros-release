package connectionmanager_test

import (
	"volley/connectionmanager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppIDStore", func() {
	It("returns an app ID weighted on the number of times it has been added", func() {
		store := connectionmanager.NewIDStore(3)
		store.Add("some-id")
		store.Add("some-id")
		store.Add("some-more-id")

		firstCount, secondCount := 0, 0
		for i := 0; i < 1000; i++ {
			switch store.Get() {
			case "some-id":
				firstCount++
			case "some-more-id":
				secondCount++
			}
		}
		Expect(firstCount).Should(BeNumerically("~", 666, 50))
		Expect(secondCount).Should(BeNumerically("~", 333, 50))
	})

	It("blocks until it is full", func() {
		store := connectionmanager.NewIDStore(3)
		store.Add("some-id")
		store.Add("some-id")

		done := make(chan struct{})

		go func() {
			store.Get()
			close(done)
		}()

		Consistently(done).ShouldNot(BeClosed())

		store.Add("some-more-id")
		Eventually(done).Should(BeClosed())
	})
})
