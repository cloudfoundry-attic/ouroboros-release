package app

import (
	loggregator "loggregator/v2"
)

type ConnectionManager interface {
	Assault(filter *loggregator.Filter)
}

type EgressV2 struct {
	connManager   ConnectionManager
	idStore       IDGetter
	firehoses     int
	appStreams    int
	appLogStreams int
}

func NewEgressV2(
	c ConnectionManager,
	s IDGetter,
	firehoseCount, appStreamCount, appLogStreamCount int,
) *EgressV2 {
	return &EgressV2{
		connManager:   c,
		idStore:       s,
		firehoses:     firehoseCount,
		appStreams:    appStreamCount,
		appLogStreams: appLogStreamCount,
	}
}

func (e *EgressV2) Start() {
	firehoseFilter := &loggregator.Filter{}

	for i := 0; i < e.firehoses; i++ {
		go e.connManager.Assault(firehoseFilter)
	}

	for i := 0; i < e.appStreams; i++ {
		f := &loggregator.Filter{
			SourceId: e.idStore.Get(),
		}
		go e.connManager.Assault(f)
	}

	for i := 0; i < e.appLogStreams; i++ {
		f := &loggregator.Filter{
			SourceId: e.idStore.Get(),
			Message:  &loggregator.Filter_Log{&loggregator.LogFilter{}},
		}
		go e.connManager.Assault(f)
	}
}
