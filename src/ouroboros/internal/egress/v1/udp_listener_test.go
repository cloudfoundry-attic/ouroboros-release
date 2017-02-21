package egress_test

import (
	"net"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testUDPListener struct {
	conn *net.UDPConn
	msgs chan *events.Envelope
}

func (t *testUDPListener) Listen() string {
	defer GinkgoRecover()
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:0")
	Expect(err).ToNot(HaveOccurred())

	conn, err := net.ListenUDP("udp", serverAddr)
	Expect(err).ToNot(HaveOccurred())

	t.conn = conn

	go func() {
		for {
			buf := make([]byte, 1024)
			n, _, err := conn.ReadFromUDP(buf)
			if isClosedRead(err) {
				return
			}
			Expect(err).ToNot(HaveOccurred())

			var e events.Envelope
			err = proto.Unmarshal(buf[:n], &e)
			Expect(err).ToNot(HaveOccurred())

			t.msgs <- &e
		}
	}()

	return conn.LocalAddr().String()
}

func (t *testUDPListener) Close() {
	t.conn.Close()
}

func isClosedRead(err error) bool {
	netErr, ok := err.(*net.OpError)
	return ok && netErr.Err.Error() == "use of closed network connection"
}
