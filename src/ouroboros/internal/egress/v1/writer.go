package egress

import (
	"log"
	"net"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

type Writer struct {
	addr string
	conn *net.UDPConn
}

func NewWriter(addr string) *Writer {
	return &Writer{
		addr: addr,
	}
}

func (w *Writer) Write(e *events.Envelope) {
	w.setupConn()

	data, err := proto.Marshal(e)
	if err != nil {
		log.Fatalf("Unable to marshal envelope (%+v): %s", e, err)
	}

	_, err = w.conn.Write(data)
	if err != nil {
		log.Fatalf("Unable to write to UDP: %s", err)
	}
}

func (w *Writer) setupConn() {
	if w.conn != nil {
		return
	}

	ra, err := net.ResolveUDPAddr("udp", w.addr)
	if err != nil {
		log.Fatalf("Invalid addr (%s): %s", w.addr, err)
	}

	w.conn, err = net.DialUDP("udp", nil, ra)
	if err != nil {
		log.Fatalf("could not connect to metron: %s", err)
	}
}
