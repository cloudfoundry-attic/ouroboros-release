package app_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"
	"volley/app"

	"github.com/gorilla/websocket"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SyslogRegistrar", func() {
	var (
		etcdserver  *httptest.Server
		etcdhandler handler
	)

	BeforeEach(func() {
		etcdhandler = handler{
			reqs:   make(chan *http.Request),
			bodies: make(chan []byte),
		}
		etcdserver = httptest.NewServer(etcdhandler)
	})

	It("adds syslog drain bindings to etcd", func() {
		r := app.NewSyslogRegistrar(
			time.Hour,
			5,
			[]string{"some-url"},
			[]string{etcdserver.URL},
			SpyIDGetter{},
		)

		go r.Start()

		var req *http.Request
		for i := 0; i < 5; i++ {
			Eventually(etcdhandler.reqs).Should(Receive(&req))
			Expect(req.Method).To(Equal("PUT"))
			Expect(req.URL.Path).To(HavePrefix("/v2/keys/loggregator/services/app-id/"))

			var body []byte
			Eventually(etcdhandler.bodies).Should(Receive(&body))

			params, err := url.ParseQuery(string(body))
			Expect(err).ToNot(HaveOccurred())
			Expect(params.Get("value")).To(Equal("some-url"))
			Expect(params.Get("ttl")).To(Equal("3600"))
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

type SpyIDGetter struct{}

func (s SpyIDGetter) Get() string {
	return "app-id"
}
