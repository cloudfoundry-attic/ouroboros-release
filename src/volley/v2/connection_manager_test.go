package v2_test

import (
	"conf"
	"errors"
	"log"
	loggregator "loggregator/v2"
	"net"
	"volley/v2"

	"google.golang.org/grpc"

	"github.com/cloudfoundry/dropsonde/metricbatcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConnectionManager", func() {
	var (
		reqs    chan *loggregator.EgressRequest
		addrs   []string
		spies   []*spyLoggregator
		c       *v2.ConnectionManager
		batcher *spyBatcher
	)

	BeforeEach(func() {
		batcher = &spyBatcher{}
		reqs = make(chan *loggregator.EgressRequest, 100)

		addrs = nil
		spies = nil

		for i := 0; i < 3; i++ {
			spy, addr := startSpyLoggregator(reqs)
			addrs = append(addrs, addr)
			spies = append(spies, spy)
		}
		c = v2.NewConnectionManager(addrs, conf.DurationRange{}, batcher, grpc.WithInsecure())
	})

	Context("without an error", func() {
		BeforeEach(func() {
			for _, s := range spies {
				close(s.errs)
			}
		})

		It("connects to RLP with the given filter", func() {
			f := &loggregator.Filter{SourceId: "some-id"}
			go c.Assault(f)

			var req *loggregator.EgressRequest
			Eventually(reqs).Should(Receive(&req))
			Expect(req.GetFilter()).To(Equal(f))
		})

		It("makes a request to every RLP", func() {
			for i := 0; i < 10; i++ {
				f := &loggregator.Filter{SourceId: "some-id"}
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
			f := &loggregator.Filter{SourceId: "some-id"}
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
	reqs           chan *loggregator.EgressRequest
}

func startSpyLoggregator(reqs chan *loggregator.EgressRequest) (*spyLoggregator, string) {
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
	loggregator.RegisterEgressServer(s, spy)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Panicf("failed to serve: %s", err)
		}
	}()

	return spy, lis.Addr().String()
}

func (s *spyLoggregator) Receiver(req *loggregator.EgressRequest, rx loggregator.Egress_ReceiverServer) error {
	s.receiverCalled <- true
	s.reqs <- req
	return <-s.errs
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
