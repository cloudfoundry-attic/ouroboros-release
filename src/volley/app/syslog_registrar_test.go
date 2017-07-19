package app_test

import (
	"crypto/sha1"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"
	"volley/app"

	"github.com/coreos/etcd/client"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"

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

	It("advertises drain URLs for apps", func() {
		syslogURL := "some-syslog-url"
		syslogHash := sha1.Sum([]byte(syslogURL))
		spySetter := &SpySetter{}
		app.AdvertiseRandom(&SpyIDGetter{}, spySetter, []string{syslogURL}, time.Minute)

		Expect(spySetter.key).To(Equal(
			"/loggregator/services/app-id/" + string(syslogHash[:]),
		))
		Expect(spySetter.value).To(Equal(syslogURL))
		Expect(spySetter.options).To(Equal(&client.SetOptions{TTL: time.Minute}))
	})

	It("picks a random drain URL", func() {
		spySetter := &SpySetter{}
		syslogURLs := []string{"syslog1", "syslog2"}

		advertised := make(map[string]struct{})
		for tries := 0; tries < 100 && len(advertised) < len(syslogURLs); tries++ {
			app.AdvertiseRandom(&SpyIDGetter{}, spySetter, syslogURLs, time.Second)

			Expect(syslogURLs).To(ContainElement(spySetter.value))
			advertised[spySetter.value] = struct{}{}
		}
		Expect(advertised).To(HaveLen(2))
		Expect(advertised).To(HaveKey(syslogURLs[0]))
		Expect(advertised).To(HaveKey(syslogURLs[1]))
	})
})

type SpySetter struct {
	key     string
	value   string
	options *client.SetOptions
}

func (s *SpySetter) Set(
	ctx context.Context,
	key string,
	value string,
	opts *client.SetOptions,
) (*client.Response, error) {
	s.key = key
	s.value = value
	s.options = opts

	return nil, nil
}

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
