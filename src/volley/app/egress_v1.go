package app

import (
	"conf"
	"math/rand"
	"time"
	"volley/connectionmanager"
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
	idStore              connectionmanager.AppIDStore
	batcher              connectionmanager.Batcher
}

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
	idStore connectionmanager.AppIDStore,
	batcher connectionmanager.Batcher,
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

func (e EgressV1) Start() {
	conn := connectionmanager.New(
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
