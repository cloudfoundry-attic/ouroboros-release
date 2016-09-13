package config_test

import (
	"os"
	"strings"
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

		It("returns error if StreamCount is > 0 and AppID is not set", func() {
			json := `{
					"TCAddresses":["1.1.1.1"],
				  "StreamCount": 1}`
			_, err := config.Parse(strings.NewReader(json))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("AppID is required to make stream connections"))
		})

		It("prefers AuthToken from env variable", func() {
			json := `{
					"TCAddresses":["1.1.1.1"],
				  "AuthToken": "bearer authtoken"}`
			os.Setenv("AUTHTOKEN", "bearer preferred-authtoken")
			conf, err := config.Parse(strings.NewReader(json))
			Expect(err).ToNot(HaveOccurred())
			Expect(conf.AuthToken).To(Equal("bearer preferred-authtoken"))
		})

		Context("Sets Default", func() {
			It("SubscriptionId", func() {
				json := `{
					"TCAddresses":["1.1.1.1"],
				  "FirehoseCount": 1}`
				conf, err := config.Parse(strings.NewReader(json))
				Expect(err).ToNot(HaveOccurred())
				Expect(conf.SubscriptionId).To(Equal("default"))
			})
		})
	})
})
