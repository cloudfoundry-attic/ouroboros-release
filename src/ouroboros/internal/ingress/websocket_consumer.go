package ingress

import (
	"crypto/tls"
	"log"

	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
)

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

	for msg := range msgChan {
		w.Write(msg)
	}
}
