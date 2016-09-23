package config_test

import (
	"os"
	"strings"
	"time"
	"volley/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {

	Describe("Parse", func() {
		It("returns error for empty TCAddresses", func() {
			json := `{"FirehoseCount":1}`
			_, err := config.Parse(strings.NewReader(json))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("At least one TrafficController URL is required"))
		})

		It("prefers AuthToken from env variable", func() {
			json := `{
				"TCAddresses":["1.1.1.1"],
				"AuthToken": "bearer authtoken"
			}`
			os.Setenv("AUTHTOKEN", "bearer preferred-authtoken")
			conf, err := config.Parse(strings.NewReader(json))
			Expect(err).ToNot(HaveOccurred())
			Expect(conf.AuthToken).To(Equal("bearer preferred-authtoken"))
		})

		It("parses min and max delay values", func() {
			json := `{
				"TCAddresses":["1.1.1.1"],
				"MinDelay": "1ms",
				"MaxDelay": "2ms"
			}`
			conf, err := config.Parse(strings.NewReader(json))
			Expect(err).ToNot(HaveOccurred())
			Expect(conf.MinDelay.Duration).To(Equal(time.Millisecond))
			Expect(conf.MaxDelay.Duration).To(Equal(2 * time.Millisecond))
		})

		Context("Sets Default", func() {
			It("SubscriptionID", func() {
				json := `{
					"TCAddresses":["1.1.1.1"],
					"FirehoseCount": 1
				}`
				conf, err := config.Parse(strings.NewReader(json))
				Expect(err).ToNot(HaveOccurred())
				Expect(conf.SubscriptionID).To(Equal("default"))
			})
		})
	})
})
