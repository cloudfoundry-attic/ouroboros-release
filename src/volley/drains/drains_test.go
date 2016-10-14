package drains_test

import (
	"crypto/sha1"
	"time"
	"volley/drains"

	"golang.org/x/net/context"

	. "github.com/apoydence/eachers"
	"github.com/coreos/etcd/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Drains", func() {
	It("advertises drain URLs for apps", func() {
		mockIDGetter := newMockIDGetter()
		mockIDGetter.GetOutput.Id <- "some-id"
		mockETCDSetter := newMockETCDSetter()
		close(mockETCDSetter.SetOutput.Ret0)
		close(mockETCDSetter.SetOutput.Ret1)
		syslogURL := "some-syslog-url"
		syslogHash := sha1.Sum([]byte(syslogURL))
		drains.AdvertiseRandom(mockIDGetter, mockETCDSetter, []string{syslogURL}, time.Minute)
		Eventually(mockETCDSetter.SetInput).Should(BeCalled(
			With(
				context.Background(),
				"/loggregator/services/some-id/"+string(syslogHash[:]),
				syslogURL,
				&client.SetOptions{TTL: time.Minute},
			),
		))
	})

	It("picks a random drain URL", func() {
		mockIDGetter := newMockIDGetter()
		mockETCDSetter := newMockETCDSetter()
		close(mockETCDSetter.SetOutput.Ret0)
		close(mockETCDSetter.SetOutput.Ret1)
		syslogURLs := []string{"syslog1", "syslog2"}

		advertised := make(map[string]struct{})
		for tries := 0; tries < 100 && len(advertised) < len(syslogURLs); tries++ {
			mockIDGetter.GetOutput.Id <- "some-id"
			drains.AdvertiseRandom(mockIDGetter, mockETCDSetter, syslogURLs, time.Second)

			var chosen string
			Eventually(mockETCDSetter.SetInput.Value).Should(Receive(&chosen))
			Expect(syslogURLs).To(ContainElement(chosen))
			advertised[chosen] = struct{}{}
		}
		Expect(advertised).To(HaveLen(2))
		Expect(advertised).To(HaveKey(syslogURLs[0]))
		Expect(advertised).To(HaveKey(syslogURLs[1]))
	})
})
