package v2

import (
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

type Assaulter interface {
	Assault(filter *loggregator_v2.Selector)
}

type IDGetter interface {
	Get() (id string)
}

// EgressV2 initiates the configured number of connections to Loggregator's V2
// API and uses the Assaulter to simulate hostile consumers
type EgressV2 struct {
	connManager   Assaulter
	idStore       IDGetter
	firehoses     int
	appStreams    int
	appLogStreams int
}

// NewEgressV2 creates a new EgressV2 with a configurable number of firehose
// connections, app streams, and app log streams
func NewEgressV2(
	c Assaulter,
	s IDGetter,
	firehoseCount int,
	appStreamCount int,
	appLogStreamCount int,
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
	firehoseFilter := &loggregator_v2.Selector{}

	for i := 0; i < e.firehoses; i++ {
		go e.connManager.Assault(firehoseFilter)
	}

	for i := 0; i < e.appStreams; i++ {
		f := &loggregator_v2.Selector{
			SourceId: e.idStore.Get(),
		}
		go e.connManager.Assault(f)
	}

	for i := 0; i < e.appLogStreams; i++ {
		f := &loggregator_v2.Selector{
			SourceId: e.idStore.Get(),
			Message: &loggregator_v2.Selector_Log{
				Log: &loggregator_v2.LogSelector{},
			},
		}
		go e.connManager.Assault(f)
	}
}
