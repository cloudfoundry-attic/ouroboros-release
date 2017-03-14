package v2

import (
	"conf"
	"context"
	"log"
	loggregator "loggregator/v2"
	"math/rand"
	"time"

	"github.com/cloudfoundry/dropsonde/metricbatcher"

	"google.golang.org/grpc"
)

type Batcher interface {
	BatchCounter(name string) metricbatcher.BatchCounterChainer
}

type ConnectionManager struct {
	addrs        []string
	receiveDelay conf.DurationRange
	batcher      Batcher
	dialOpts     []grpc.DialOption
}

func NewConnectionManager(
	addrs []string,
	receiveDelay conf.DurationRange,
	batcher Batcher,
	dialOpts ...grpc.DialOption,
) *ConnectionManager {
	return &ConnectionManager{
		addrs:        addrs,
		receiveDelay: receiveDelay,
		batcher:      batcher,
		dialOpts:     dialOpts,
	}
}

func (m *ConnectionManager) Assault(filter *loggregator.Filter) {
	addr := m.addrs[rand.Intn(len(m.addrs))]
	conn, err := grpc.Dial(addr, m.dialOpts...)
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()
	c := loggregator.NewEgressClient(conn)

	r, err := c.Receiver(context.Background(), &loggregator.EgressRequest{Filter: filter})
	if err != nil {
		log.Fatalf("could not receive stream: %s", err)
	}

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
