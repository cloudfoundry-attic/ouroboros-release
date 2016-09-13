package main_test

import (
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testUDPListener struct {
	conn *net.UDPConn
	msgs chan []byte
}

func (t *testUDPListener) Listen() {
	defer GinkgoRecover()
	serverAddr, err := net.ResolveUDPAddr("udp", ":3456")
	Expect(err).ToNot(HaveOccurred())
	conn, err := net.ListenUDP("udp", serverAddr)
	Expect(err).ToNot(HaveOccurred())
	t.conn = conn

	for {
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if isClosedRead(err) {
			return
		}
		Expect(err).ToNot(HaveOccurred())
		t.msgs <- buf[:n]
	}
}

func (t *testUDPListener) Close() {
	t.conn.Close()
}

func isClosedRead(err error) bool {
	netErr, ok := err.(*net.OpError)
	return ok && netErr.Err.Error() == "use of closed network connection"
}
