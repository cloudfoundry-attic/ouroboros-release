package main_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/gorilla/websocket"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Volley", func() {
	var (
		connected   chan bool
		server      *httptest.Server
		connections chan *websocket.Conn
	)

	BeforeEach(func() {
		connections = make(chan *websocket.Conn, 100)
		connected = make(chan bool, 100)
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
			Expect(err).ToNot(HaveOccurred())
			connected <- true
			connections <- conn
		}))
	})

	AfterEach(func() {
		closeAll(connections)
		server.Close()
	})

	It("kills itself", func() {
		path, err := gexec.Build("volley")
		Expect(err).ToNot(HaveOccurred())
		cmd := exec.Command(path)
		cmd.Env = []string{
			"TC_ADDRS=" + server.URL,
			"KILL_DELAY=100ms-150ms",
		}

		Expect(cmd.Start()).To(Succeed())
		defer cmd.Process.Signal(os.Interrupt)
		errs := make(chan error)
		go func() {
			errs <- cmd.Wait()
		}()
		Eventually(errs).Should(Receive(&err))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("signal: killed"))
	})

	It("writes addresses to etcd", func() {
		path, err := gexec.Build("volley")
		Expect(err).ToNot(HaveOccurred())

		syslogURL := "some-syslog-url"

		etcdhandler := handler{
			reqs:   make(chan *http.Request),
			bodies: make(chan []byte),
		}
		etcdserver := httptest.NewServer(etcdhandler)
		cmd := exec.Command(path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = []string{
			"TC_ADDRS=" + strings.Replace(server.URL, "http", "ws", 1),
			"KILL_DELAY=10s-20s",
			"ETCD_ADDRS=" + etcdserver.URL,
			"FIREHOSE_COUNT=1",
			"STREAM_COUNT=1",
			"SYSLOG_ADDRS=" + syslogURL,
			"SYSLOG_DRAINS=5",
		}

		Expect(cmd.Start()).To(Succeed())
		defer cmd.Process.Signal(os.Interrupt)

		Eventually(connected).Should(Receive())

		env := &events.Envelope{
			Origin:    proto.String("test"),
			EventType: events.Envelope_LogMessage.Enum(),
			LogMessage: &events.LogMessage{
				Message:     []byte("foo"),
				MessageType: events.LogMessage_OUT.Enum(),
				Timestamp:   proto.Int64(time.Now().UnixNano()),
				AppId:       proto.String("app-id"),
			},
		}
		b, err := proto.Marshal(env)
		Expect(err).ToNot(HaveOccurred())
		firehose := <-connections
		defer firehose.Close()
		Expect(firehose.WriteMessage(websocket.BinaryMessage, b)).To(Succeed())

		var req *http.Request
		for i := 0; i < 5; i++ {
			Eventually(etcdhandler.reqs).Should(Receive(&req))
			Expect(req.Method).To(Equal("PUT"))
			Expect(req.URL.Path).To(HavePrefix("/v2/keys/loggregator/services/app-id/"))
			body := <-etcdhandler.bodies
			Expect(string(body)).To(Equal("value=" + syslogURL))
		}
	})
})

type handler struct {
	reqs   chan *http.Request
	bodies chan []byte
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer GinkgoRecover()
	h.reqs <- r
	body, err := ioutil.ReadAll(r.Body)
	Expect(err).ToNot(HaveOccurred())
	h.bodies <- body
}

func closeAll(conns chan *websocket.Conn) {
	for {
		select {
		case conn := <-conns:
			conn.Close()
		default:
			return
		}
	}
}
