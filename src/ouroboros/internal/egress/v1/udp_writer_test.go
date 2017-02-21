package egress_test

import (
	egress "ouroboros/internal/egress/v1"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UdpWriter", func() {
	var (
		udpListener *testUDPListener
		udpWriter   *egress.UDPWriter
	)

	BeforeEach(func() {
		udpListener = &testUDPListener{
			msgs: make(chan *events.Envelope),
		}
		udpAddr := udpListener.Listen()

		udpWriter = egress.NewUDPWriter(udpAddr)
	})

	It("sends the envelope over UDP", func() {
		e := &events.Envelope{
			Origin:    proto.String("some-origin"),
			Timestamp: proto.Int64(99),
			EventType: events.Envelope_LogMessage.Enum(),
		}

		udpWriter.Write(e)

		var outEnv *events.Envelope
		Eventually(udpListener.msgs).Should(Receive(&outEnv))
		Expect(outEnv.GetOrigin()).To(Equal("some-origin"))
	})
})
