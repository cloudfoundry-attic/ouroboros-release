package config_test

import (
	"os"
	"time"
	"volley/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("Load", func() {
		It("returns an error if TC_ADDRS is empty", func() {
			_, err := config.Load()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("TC_ADDRS is required but was empty"))
		})

		It("returns an error if RECV_DELAY isn't a range", func() {
			os.Setenv("TC_ADDRS", "foo,bar")
			defer os.Unsetenv("TC_ADDRS")
			os.Setenv("RECV_DELAY", "1us")
			defer os.Unsetenv("RECV_DELAY")

			_, err := config.Load()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected DurationRange to be of format {min}-{max}"))
		})

		It("returns an error if RECV_DELAY.Min can't be parsed", func() {
			os.Setenv("TC_ADDRS", "foo,bar")
			defer os.Unsetenv("TC_ADDRS")
			os.Setenv("RECV_DELAY", "foo-1us")
			defer os.Unsetenv("RECV_DELAY")

			_, err := config.Load()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error parsing DurationRange.Min: time: invalid duration foo"))
		})

		It("returns an error if RECV_DELAY.Max can't be parsed", func() {
			os.Setenv("TC_ADDRS", "foo,bar")
			defer os.Unsetenv("TC_ADDRS")
			os.Setenv("RECV_DELAY", "1us-foo")
			defer os.Unsetenv("RECV_DELAY")

			_, err := config.Load()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error parsing DurationRange.Max: time: invalid duration foo"))
		})

		It("returns an error if KILL_DELAY isn't a range", func() {
			os.Setenv("TC_ADDRS", "foo,bar")
			defer os.Unsetenv("TC_ADDRS")
			os.Setenv("KILL_DELAY", "1us")
			defer os.Unsetenv("KILL_DELAY")

			_, err := config.Load()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected DurationRange to be of format {min}-{max}"))
		})

		It("returns an error if KILL_DELAY.Min can't be parsed", func() {
			os.Setenv("TC_ADDRS", "foo,bar")
			defer os.Unsetenv("TC_ADDRS")
			os.Setenv("KILL_DELAY", "foo-1us")
			defer os.Unsetenv("KILL_DELAY")

			_, err := config.Load()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error parsing DurationRange.Min: time: invalid duration foo"))
		})

		It("returns an error if KILL_DELAY.Max can't be parsed", func() {
			os.Setenv("TC_ADDRS", "foo,bar")
			defer os.Unsetenv("TC_ADDRS")
			os.Setenv("KILL_DELAY", "1us-foo")
			defer os.Unsetenv("KILL_DELAY")

			_, err := config.Load()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error parsing DurationRange.Max: time: invalid duration foo"))
		})

		It("loads environment variables", func() {
			os.Setenv("TC_ADDRS", "foo,bar")
			defer os.Unsetenv("TC_ADDRS")
			os.Setenv("AUTH_TOKEN", "token")
			defer os.Unsetenv("AUTH_TOKEN")
			os.Setenv("FIREHOSE_COUNT", "1")
			defer os.Unsetenv("FIREHOSE_COUNT")
			os.Setenv("STREAM_COUNT", "2")
			defer os.Unsetenv("STREAM_COUNT")
			os.Setenv("SUB_ID", "subscription")
			defer os.Unsetenv("SUB_ID")
			os.Setenv("RECV_DELAY", "1us-10ms")
			defer os.Unsetenv("RECV_DELAY")
			os.Setenv("KILL_DELAY", "10us-100ms")
			defer os.Unsetenv("KILL_DELAY")

			c, err := config.Load()
			Expect(err).ToNot(HaveOccurred())
			Expect(c.TCAddresses).To(ConsistOf("foo", "bar"))
			Expect(c.AuthToken).To(Equal("token"))
			Expect(c.FirehoseCount).To(Equal(1))
			Expect(c.StreamCount).To(Equal(2))
			Expect(c.SubscriptionID).To(Equal("subscription"))
			Expect(c.ReceiveDelay.Min).To(Equal(time.Microsecond))
			Expect(c.ReceiveDelay.Max).To(Equal(10 * time.Millisecond))
			Expect(c.KillDelay.Min).To(Equal(10 * time.Microsecond))
			Expect(c.KillDelay.Max).To(Equal(100 * time.Millisecond))
		})
	})
})
