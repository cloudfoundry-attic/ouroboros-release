package config_test

import (
	"os"
	"time"
	"volley/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	WithEnv := func(env map[string]string, f func()) {
		for k, v := range env {
			os.Setenv(k, v)
		}
		f()
		for k, _ := range env {
			os.Unsetenv(k)
		}
	}

	Describe("Load", func() {
		It("returns an error if RLP_ADDRS is empty", func() {
			WithEnv(map[string]string{
				"METRON_PORT": "12345",
				"TC_ADDRS":    "foo,bar",
			}, func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("RLP_ADDRS is required but was empty"))
			})
		})

		It("returns an error if TC_ADDRS is empty", func() {
			WithEnv(map[string]string{
				"METRON_PORT": "12345",
			}, func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("TC_ADDRS is required but was empty"))
			})
		})

		It("returns an error if METRON_PORT is empty", func() {
			WithEnv(map[string]string{
				"TC_ADDRS": "foo,bar",
			}, func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("METRON_PORT is required but was empty"))
			})
		})

		It("returns an error if RECV_DELAY isn't a range", func() {
			WithEnv(map[string]string{
				"TC_ADDRS":    "foo,bar",
				"METRON_PORT": "12345",
				"RLP_ADDRS":   "foo,bar",
				"RECV_DELAY":  "1us",
			}, func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Expected DurationRange to be of format {min}-{max}"))
			})
		})

		It("returns an error if RECV_DELAY.Min can't be parsed", func() {
			WithEnv(map[string]string{
				"TC_ADDRS":    "foo,bar",
				"METRON_PORT": "12345",
				"RLP_ADDRS":   "foo,bar",
				"RECV_DELAY":  "foo-1us",
			}, func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Error parsing DurationRange.Min: time: invalid duration foo"))
			})
		})

		It("returns an error if RECV_DELAY.Max can't be parsed", func() {
			WithEnv(map[string]string{
				"TC_ADDRS":    "foo,bar",
				"RLP_ADDRS":   "foo,bar",
				"METRON_PORT": "12345",
				"RECV_DELAY":  "1us-foo",
			}, func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Error parsing DurationRange.Max: time: invalid duration foo"))
			})
		})

		It("returns an error if ASYNC_REQUEST_DELAY isn't a range", func() {
			WithEnv(map[string]string{
				"TC_ADDRS":            "foo,bar",
				"RLP_ADDRS":           "foo,bar",
				"METRON_PORT":         "12345",
				"ASYNC_REQUEST_DELAY": "1us",
			}, func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Expected DurationRange to be of format {min}-{max}"))
			})
		})

		It("returns an error if ASYNC_REQUEST_DELAY.Min can't be parsed", func() {
			WithEnv(map[string]string{
				"TC_ADDRS":            "foo,bar",
				"RLP_ADDRS":           "foo,bar",
				"METRON_PORT":         "12345",
				"ASYNC_REQUEST_DELAY": "foo-1us",
			}, func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Error parsing DurationRange.Min: time: invalid duration foo"))
			})
		})

		It("returns an error if ASYNC_REQUEST_DELAY.Max can't be parsed", func() {
			WithEnv(map[string]string{
				"TC_ADDRS":            "foo,bar",
				"RLP_ADDRS":           "foo,bar",
				"METRON_PORT":         "12345",
				"ASYNC_REQUEST_DELAY": "1us-foo",
			}, func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Error parsing DurationRange.Max: time: invalid duration foo"))
			})
		})

		It("returns an error if KILL_DELAY isn't a range", func() {
			WithEnv(map[string]string{
				"TC_ADDRS":    "foo,bar",
				"RLP_ADDRS":   "foo,bar",
				"METRON_PORT": "12345",
				"KILL_DELAY":  "1us",
			}, func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Expected DurationRange to be of format {min}-{max}"))
			})
		})

		It("returns an error if KILL_DELAY.Min can't be parsed", func() {
			WithEnv(map[string]string{
				"TC_ADDRS":    "foo,bar",
				"RLP_ADDRS":   "foo,bar",
				"METRON_PORT": "12345",
				"KILL_DELAY":  "foo-1us",
			}, func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Error parsing DurationRange.Min: time: invalid duration foo"))
			})
		})

		It("returns an error if KILL_DELAY.Max can't be parsed", func() {
			WithEnv(map[string]string{
				"TC_ADDRS":    "foo,bar",
				"RLP_ADDRS":   "foo,bar",
				"METRON_PORT": "12345",
				"KILL_DELAY":  "1us-foo",
			}, func() {
				_, err := config.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Error parsing DurationRange.Max: time: invalid duration foo"))
			})
		})

		It("loads environment variables", func() {
			WithEnv(map[string]string{
				"TC_ADDRS":               "foo,bar",
				"RLP_ADDRS":              "foo,bar",
				"AUTH_TOKEN":             "token",
				"FIREHOSE_COUNT":         "1",
				"STREAM_COUNT":           "2",
				"RECENT_LOG_COUNT":       "2",
				"CONTAINER_METRIC_COUNT": "2",
				"SUB_ID":                 "subscription",
				"RECV_DELAY":             "1us-10ms",
				"KILL_DELAY":             "10us-100ms",
				"METRON_PORT":            "12345",
				"METRIC_BATCH_INTERVAL":  "100ms",
			}, func() {

				c, err := config.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(c.TCAddresses).To(ConsistOf("foo", "bar"))
				Expect(c.AuthToken).To(Equal("token"))
				Expect(c.FirehoseCount).To(Equal(1))
				Expect(c.StreamCount).To(Equal(2))
				Expect(c.RecentLogCount).To(Equal(2))
				Expect(c.ContainerMetricCount).To(Equal(2))
				Expect(c.SubscriptionID).To(Equal("subscription"))
				Expect(c.ReceiveDelay.Min).To(Equal(time.Microsecond))
				Expect(c.ReceiveDelay.Max).To(Equal(10 * time.Millisecond))
				Expect(c.KillDelay.Min).To(Equal(10 * time.Microsecond))
				Expect(c.KillDelay.Max).To(Equal(100 * time.Millisecond))
				Expect(c.MetronPort).To(Equal(12345))
				Expect(c.MetricBatchInterval).To(Equal(100 * time.Millisecond))
			})
		})
	})
})
