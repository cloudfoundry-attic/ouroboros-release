package ingress

import (
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

type MetricCounter struct {
	deploymentName string
	jobName        string
	instanceIndex  string
	instanceIP     string
	reportCount    uint64
	counter        uint64
	writer         EnvelopeWriter
}

func NewMetricCounter(deployment, job, idx, ip string, reportCount uint64, w EnvelopeWriter) *MetricCounter {
	return &MetricCounter{
		deploymentName: deployment,
		jobName:        job,
		instanceIndex:  idx,
		instanceIP:     ip,
		reportCount:    reportCount,
		writer:         w,
	}
}

func (m *MetricCounter) Write(e *events.Envelope) {
	m.counter++

	m.writer.Write(e)

	if m.counter%m.reportCount == 0 {
		m.emitIngressCounter(m.reportCount)
	}
}

func (m *MetricCounter) emitIngressCounter(delta uint64) {
	env := &events.Envelope{
		Origin:     proto.String("ouroboros"),
		Timestamp:  proto.Int64(time.Now().UnixNano()),
		Deployment: proto.String(m.deploymentName),
		Job:        proto.String(m.jobName),
		Index:      proto.String(m.instanceIndex),
		Ip:         proto.String(m.instanceIP),
		EventType:  events.Envelope_CounterEvent.Enum(),
		CounterEvent: &events.CounterEvent{
			Name:  proto.String("ingress"),
			Delta: proto.Uint64(delta),
		},
	}

	m.writer.Write(env)
}
