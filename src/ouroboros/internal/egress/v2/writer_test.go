package egress_test

import (
	"fmt"
	"log"
	"net"
	egress "ouroboros/internal/egress/v2"
	loggregator "ouroboros/internal/loggregator/v2"

	"google.golang.org/grpc"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Writer", func() {
	var (
		ingressAddr   string
		mockServer    *mockIngressServer
		mockConverter *mockConverter
		v2Writer      *egress.Writer
		v2e           *loggregator.Envelope
	)

	BeforeEach(func() {
		mockServer, ingressAddr = startIngressServer()
		mockConverter = newMockConverter()

		v2e = &loggregator.Envelope{
			Timestamp: 99,
		}
		go func() {
			for {
				mockConverter.ToV2Output.V2e <- v2e
			}
		}()

		v2Writer = egress.NewWriter(ingressAddr, mockConverter, grpc.WithInsecure())
	})

	It("sends the envelope to the Sender stream", func(done Done) {
		defer close(done)

		e := &events.Envelope{
			Origin:    proto.String("some-origin"),
			Timestamp: proto.Int64(99),
			EventType: events.Envelope_LogMessage.Enum(),
		}

		go func() {
			for i := 0; i < 100; i++ {
				v2Writer.Write(e)
			}
		}()

		fmt.Println("Getting sender")
		var sender loggregator.Ingress_SenderServer
		Eventually(mockServer.SenderInput.Arg0).Should(Receive(&sender))

		outEnv, err := sender.Recv()
		Expect(err).ToNot(HaveOccurred())
		Expect(outEnv.Timestamp).To(Equal(int64(99)))
	})
})

func startIngressServer() (*mockIngressServer, string) {
	lis, err := net.Listen("tcp", "localhost:0")
	Expect(err).ToNot(HaveOccurred())

	mockIngress := newMockIngressServer()
	grpcServer := grpc.NewServer()

	loggregator.RegisterIngressServer(grpcServer, mockIngress)

	go func() {
		log.Fatal(grpcServer.Serve(lis))
	}()

	return mockIngress, lis.Addr().String()
}
