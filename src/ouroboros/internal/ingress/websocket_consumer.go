package ingress

import (
	"crypto/tls"
	"log"
	"time"

	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

var counterEvent = &events.Envelope{
	Origin:    proto.String("ouroboros"),
	Timestamp: proto.Int64(time.Now().UnixNano()),
	EventType: events.Envelope_CounterEvent.Enum(),
	CounterEvent: &events.CounterEvent{
		Name:  proto.String("ouroboros.forwardedMessages"),
		Delta: proto.Uint64(1000),
	},
}

type EnvelopeWriter interface {
	Write(e *events.Envelope)
}

func Consume(addr, subId, token string, w EnvelopeWriter) {
	consumer := consumer.New(addr, &tls.Config{InsecureSkipVerify: true}, nil)
	msgChan, errorChan := consumer.Firehose(subId, token)
	go func() {
		for err := range errorChan {
			log.Fatalf("Error received from firehose consumer: %s", err)
		}
	}()

	var msgCount uint64
	for msg := range msgChan {
		msgCount++

		w.Write(msg)

		if msgCount%1000 == 0 {
			w.Write(counterEvent)
		}
	}
}
