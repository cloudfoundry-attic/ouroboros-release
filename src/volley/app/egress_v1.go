package app

import (
	"conf"
	"math/rand"
	"time"
	"volley/v1"
)

type EgressV1 struct {
	firehoseCount        int
	streamCount          int
	recentLogCount       int
	containerMetricCount int
	tcAddrs              []string
	authToken            string
	subscriptionID       string
	receiveDelay         conf.DurationRange
	asyncRequestDelay    conf.DurationRange
	idStore              v1.AppIDStore
	batcher              v1.Batcher
}

// NewEgressV1 creates consumers of Loggregator. Consumers may be firehose
// connections, application streams, recent logs requests, or container
// metrics. Note that the number of consumers of a particular type are also
// configurable, e.g., we may have 10 firehose consumers, 5 application
// streams, 15 recent log requests, etc.
func NewEgressV1(
	firehoseCount int,
	streamCount int,
	recentLogCount int,
	containerMetricCount int,
	tcAddrs []string,
	authToken string,
	subscriptionID string,
	receiveDelay conf.DurationRange,
	asyncRequestDelay conf.DurationRange,
	idStore v1.AppIDStore,
	batcher v1.Batcher,
) *EgressV1 {
	return &EgressV1{
		firehoseCount:        firehoseCount,
		streamCount:          streamCount,
		recentLogCount:       recentLogCount,
		containerMetricCount: containerMetricCount,
		tcAddrs:              tcAddrs,
		authToken:            authToken,
		subscriptionID:       subscriptionID,
		receiveDelay:         receiveDelay,
		asyncRequestDelay:    asyncRequestDelay,
		idStore:              idStore,
		batcher:              batcher,
	}
}

// Start initiates all the configured consumers
func (e EgressV1) Start() {
	conn := v1.New(
		e.tcAddrs,
		e.authToken,
		e.subscriptionID,
		e.receiveDelay,
		e.idStore,
		e.batcher,
	)
	defer conn.Close()

	go e.syncRequest(e.firehoseCount, conn.Firehose)
	go e.syncRequest(e.streamCount, conn.Stream)
	go e.asyncRequest(e.asyncRequestDelay, e.recentLogCount, conn.RecentLogs)
	go e.asyncRequest(e.asyncRequestDelay, e.containerMetricCount, conn.ContainerMetrics)
}

func (e *EgressV1) syncRequest(count int, endpoint func()) {
	for i := 0; i < count; i++ {
		go endpoint()
	}
}

func (e *EgressV1) asyncRequest(delay conf.DurationRange, count int, endpoint func()) {
	delta := int(delay.Max - delay.Min)
	for i := 0; i < count; i++ {
		delay := delay.Min + time.Duration(rand.Intn(delta))
		go endpoint()
		time.Sleep(delay)
	}
}
