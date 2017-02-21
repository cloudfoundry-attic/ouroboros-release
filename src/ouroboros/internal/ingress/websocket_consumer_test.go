package ingress_test

import (
	"log"
	"net/http/httptest"
	"ouroboros/internal/ingress"
	"strings"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("WebsocketConsumer", func() {
	var (
		wsHandler *testWebsocketHandler
		wsServer  *httptest.Server

		spyEnvelopeWriter *spyEnvelopeWriter
	)

	BeforeEach(func() {
		wsHandler = &testWebsocketHandler{
			token:   "some-good-token",
			started: make(chan struct{}),
			done:    make(chan struct{}),
		}
		wsServer = httptest.NewServer(wsHandler)
		spyEnvelopeWriter = newSpyEnvelopeWriter()

		e := &events.Envelope{
			Origin:    proto.String("some-origin"),
			Timestamp: proto.Int64(99),
			EventType: events.Envelope_LogMessage.Enum(),
		}

		data, err := proto.Marshal(e)
		Expect(err).ToNot(HaveOccurred())

		go serveDataUp(wsHandler, data)
		go ingress.Consume(
			strings.Replace(wsServer.URL, "http", "ws", -1),
			"sub-id",
			"Bearer some-good-token",
			spyEnvelopeWriter,
		)
	})

	It("reads data from the websocket and writes it to the EnvelopeWriter", func() {
		var e *events.Envelope
		Eventually(spyEnvelopeWriter.envelope).Should(Receive(&e))
		Expect(e.GetOrigin()).To(Equal("some-origin"))
		Expect(e.GetTimestamp()).To(Equal(int64(99)))
	})

	It("writes a counter metric for the number of consumed envelopes", func() {
		var e *events.Envelope
		f := func() bool {
			select {
			case e = <-spyEnvelopeWriter.envelope:
				return e.GetEventType() == events.Envelope_CounterEvent
			default:
				return false
			}
		}

		Eventually(f, 1, "1ns").Should(BeTrue())
		Expect(e.GetOrigin()).To(Equal("ouroboros"))

		counterEvent := e.GetCounterEvent()
		Expect(counterEvent.GetName()).To(Equal("ouroboros.forwardedMessages"))
		Expect(counterEvent.GetDelta()).To(Equal(uint64(1000)))
	})
})

func serveDataUp(server *testWebsocketHandler, data []byte) {
	defer GinkgoRecover()
	for {
		if err := server.Send(data); err != nil {
			log.Println("WS SEND", err)
			return
		}
	}
}

type spyEnvelopeWriter struct {
	envelope chan *events.Envelope
}

func newSpyEnvelopeWriter() *spyEnvelopeWriter {
	return &spyEnvelopeWriter{
		envelope: make(chan *events.Envelope, 1000),
	}
}

func (s *spyEnvelopeWriter) Write(e *events.Envelope) {
	s.envelope <- e
}
