package egress

import (
	"context"
	"log"

	"github.com/cloudfoundry/sonde-go/events"

	loggregator "loggregator/v2"

	"google.golang.org/grpc"
)

type Converter interface {
	ToV2(v1e *events.Envelope) (v2e *loggregator.Envelope)
}

type Writer struct {
	sender    loggregator.Ingress_SenderClient
	converter Converter
	count     int
}

func NewWriter(addr string, c Converter, dialOpts ...grpc.DialOption) *Writer {
	conn, err := grpc.Dial(addr, dialOpts...)
	if err != nil {
		log.Fatalf("Failed to dial Loggregator V2 API: %s", err)
	}

	client := loggregator.NewIngressClient(conn)
	sender, err := client.Sender(context.Background(), grpc.FailFast(true))
	if err != nil {
		log.Fatalf("Failed to open Loggregator V2 ingress stream: %s", err)
	}

	return &Writer{sender: sender, converter: c}
}

func (w *Writer) Write(msg *events.Envelope) {
	if err := w.sender.Send(w.converter.ToV2(msg)); err != nil {
		log.Fatalf("Failed to send V2 envelope: %s", err)
	}

	w.count++
	if w.count%1000 == 0 {
		log.Print("Egressed 1000 envelopes")
	}
}
