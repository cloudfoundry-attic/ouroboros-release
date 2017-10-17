package v2_test

import (
	"conf"
	"errors"
	"log"
	"net"
	"volley/v2"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/cloudfoundry/dropsonde/metricbatcher"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConnectionManager", func() {
	var (
		reqs    chan *loggregator_v2.EgressRequest
		addrs   []string
		spies   []*spyLoggregator
		c       *v2.ConnectionManager
		batcher *spyBatcher
	)

	BeforeEach(func() {
		batcher = &spyBatcher{}
		reqs = make(chan *loggregator_v2.EgressRequest, 100)

		addrs = nil
		spies = nil

		for i := 0; i < 3; i++ {
			spy, addr := startSpyLoggregator(reqs)
			addrs = append(addrs, addr)
			spies = append(spies, spy)
		}
		c = v2.NewConnectionManager(addrs, conf.DurationRange{}, true, batcher, grpc.WithInsecure())
	})

	Context("without an error", func() {
		BeforeEach(func() {
			for _, s := range spies {
				close(s.errs)
			}
		})

		It("connects to RLP with the given selector", func() {
			f := &loggregator_v2.Selector{SourceId: "some-id"}
			go c.Assault(f)

			var req *loggregator_v2.EgressRequest
			Eventually(reqs).Should(Receive(&req))
			Expect(req.GetSelectors()[0]).To(Equal(f))
			Expect(req.UsePreferredTags).To(BeTrue())
		})

		It("makes a request to every RLP", func() {
			for i := 0; i < 10; i++ {
				f := &loggregator_v2.Selector{SourceId: "some-id"}
				go c.Assault(f)
			}

			for _, spy := range spies {
				Eventually(spy.receiverCalled).Should(Receive())
			}
		})
	})

	Context("when an error occurs", func() {
		BeforeEach(func() {
			for _, s := range spies {
				s.errs <- errors.New("some-error")
				close(s.errs)
			}
		})
		It("retries on an error", func() {
			f := &loggregator_v2.Selector{SourceId: "some-id"}
			go c.Assault(f)

			for _, s := range spies {
				Eventually(func() int { return len(s.receiverCalled) }).Should(BeNumerically(">", 1))
			}
		})
	})
})

type spyLoggregator struct {
	receiverCalled chan bool
	errs           chan error
	reqs           chan *loggregator_v2.EgressRequest
}

func startSpyLoggregator(reqs chan *loggregator_v2.EgressRequest) (*spyLoggregator, string) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Panicf("failed to listen: %s", err)
	}

	spy := &spyLoggregator{
		reqs:           reqs,
		receiverCalled: make(chan bool, 100),
		errs:           make(chan error, 100),
	}

	s := grpc.NewServer()
	loggregator_v2.RegisterEgressServer(s, spy)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Panicf("failed to serve: %s", err)
		}
	}()

	return spy, lis.Addr().String()
}

func (s *spyLoggregator) Receiver(
	req *loggregator_v2.EgressRequest,
	rx loggregator_v2.Egress_ReceiverServer,
) error {
	s.receiverCalled <- true
	s.reqs <- req
	return <-s.errs
}

func (s *spyLoggregator) BatchedReceiver(
	req *loggregator_v2.EgressBatchRequest,
	rx loggregator_v2.Egress_BatchedReceiverServer,
) error {
	return grpc.Errorf(codes.Unimplemented, "Not yet implemented")
}

type spyBatcher struct{}

func (s *spyBatcher) BatchCounter(string) metricbatcher.BatchCounterChainer {
	return s
}
func (s *spyBatcher) SetTag(string, string) metricbatcher.BatchCounterChainer {
	return s
}
func (s *spyBatcher) Add(uint64) {}
func (s *spyBatcher) Increment() {}
