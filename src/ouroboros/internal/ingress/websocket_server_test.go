package ingress_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(*http.Request) bool { return true },
}

type testWebsocketHandler struct {
	conn          *websocket.Conn
	token         string
	started, done chan struct{}
}

func (h *testWebsocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer GinkgoRecover()
	Expect(h.conn).To(BeNil())

	Expect(r.Header.Get("Authorization")).To(Equal("Bearer " + h.token))

	var err error
	h.conn, err = upgrader.Upgrade(w, r, nil)
	Expect(err).ToNot(HaveOccurred())
	close(h.started)

	defer func() {
		h.conn.WriteControl(websocket.CloseNormalClosure, nil, time.Now().Add(time.Second))
		h.conn.Close()
	}()
	<-h.done
}

func (h *testWebsocketHandler) Send(msg []byte) error {
	<-h.started
	return h.conn.WriteMessage(websocket.BinaryMessage, msg)
}
