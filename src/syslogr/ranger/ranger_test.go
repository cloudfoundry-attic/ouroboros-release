package ranger_test

import (
	"syslogr/ranger"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ranger", func() {
	It("returns an error when min and max don't provide enough room", func() {
		min := time.Duration(10)
		max := time.Duration(11)
		_, err := ranger.New(min, max)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("ranger: min must be at least two less than max"))
	})

	It("returns a random range between its min and max", func() {
		min := time.Duration(10)
		max := time.Duration(20)
		ranger, err := ranger.New(min, max)
		Expect(err).ToNot(HaveOccurred())

		dmin, dmax := ranger.DelayRange()
		Expect(dmin).To(BeNumerically(">=", min))
		Expect(dmax).To(BeNumerically("<", max))
		Expect(dmin).To(BeNumerically("<", dmax))
	})
})
