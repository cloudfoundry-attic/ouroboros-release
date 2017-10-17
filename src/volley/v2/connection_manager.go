package v2

import (
	"conf"
	"context"
	"log"
	"math/rand"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/cloudfoundry/dropsonde/metricbatcher"

	"google.golang.org/grpc"
)

type Batcher interface {
	BatchCounter(name string) metricbatcher.BatchCounterChainer
}

type ConnectionManager struct {
	addrs            []string
	receiveDelay     conf.DurationRange
	usePreferredTags bool
	batcher          Batcher
	dialOpts         []grpc.DialOption
}

// NewConnectionManager manages the gRPC connections to
// the Loggregator V2 API
func NewConnectionManager(
	addrs []string,
	receiveDelay conf.DurationRange,
	usePreferredTags bool,
	batcher Batcher,
	dialOpts ...grpc.DialOption,
) *ConnectionManager {
	return &ConnectionManager{
		addrs:            addrs,
		receiveDelay:     receiveDelay,
		usePreferredTags: usePreferredTags,
		batcher:          batcher,
		dialOpts:         dialOpts,
	}
}

// Assault repeatedly establishes connections to the Loggregator V2 API
// and reads from those connections for a random length of time
func (m *ConnectionManager) Assault(s *loggregator_v2.Selector) {
	for {
		m.establishConnection(s)
	}
}

func (m *ConnectionManager) establishConnection(s *loggregator_v2.Selector) {
	addr := m.addrs[rand.Intn(len(m.addrs))]
	conn, err := grpc.Dial(addr, m.dialOpts...)
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()
	c := loggregator_v2.NewEgressClient(conn)

	ctx, _ := context.WithTimeout(context.Background(), time.Minute+(time.Duration(rand.Intn(30000))*time.Millisecond))
	r, err := c.Receiver(ctx, &loggregator_v2.EgressRequest{
		UsePreferredTags: m.usePreferredTags,
		Selectors: []*loggregator_v2.Selector{
			s,
		},
	})
	if err != nil {
		log.Printf("could not receive stream: %s", err)
		return
	}

	m.connect(r)
}

func (m *ConnectionManager) connect(r loggregator_v2.Egress_ReceiverClient) {
	delta := int(m.receiveDelay.Max - m.receiveDelay.Min)
	var count int
	for {
		_, err := r.Recv()
		if err != nil {
			return
		}

		count++
		if count%1000 == 0 {
			m.batcher.BatchCounter("volley.receivedEnvelopes").SetTag("version", "v2").Add(1000)
		}
		if delta == 0 {
			continue
		}
		delay := m.receiveDelay.Min + time.Duration(rand.Intn(delta))
		time.Sleep(delay)
	}
}
