package ingress_test

import (
	"ouroboros/internal/ingress"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MetricCounter", func() {
	var (
		envelope *events.Envelope
	)

	BeforeEach(func() {
		envelope = &events.Envelope{
			Origin:    proto.String("some-origin"),
			Timestamp: proto.Int64(99),
			EventType: events.Envelope_LogMessage.Enum(),
		}
	})

	It("emits a counter envelope for a given number of writes", func() {
		writer := newSpyEnvelopeWriter()
		mc := ingress.NewMetricCounter(
			"deployment-name",
			"job-name",
			"instance-index",
			"instance-ip",
			10,
			writer,
		)

		for i := 0; i < 10; i++ {
			mc.Write(envelope)
		}

		s := toSlice(writer.envelope)
		Expect(s).To(HaveLen(11))
		Expect(s[10].GetTimestamp()).ToNot(Equal(int64(0)))
		Expect(s[10].GetDeployment()).To(Equal("deployment-name"))
		Expect(s[10].GetJob()).To(Equal("job-name"))
		Expect(s[10].GetIndex()).To(Equal("instance-index"))
		Expect(s[10].GetIp()).To(Equal("instance-ip"))
		Expect(s[10].GetCounterEvent().GetDelta()).To(Equal(uint64(10)))
		Expect(s[10].GetCounterEvent().GetName()).To(Equal("ingress"))
	})
})

func toSlice(c <-chan *events.Envelope) (results []*events.Envelope) {
	for {
		select {
		case x := <-c:
			results = append(results, x)
		default:
			return results
		}
	}
}
