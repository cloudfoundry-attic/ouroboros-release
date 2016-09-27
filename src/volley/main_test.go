package main_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/gorilla/websocket"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Volley", func() {
	var (
		server *httptest.Server
		conn   *websocket.Conn
	)

	BeforeEach(func() {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			var err error
			conn, err = websocket.Upgrade(w, r, nil, 1024, 1024)
			Expect(err).ToNot(HaveOccurred())
		}))
	})

	AfterEach(func() {
		if conn != nil {
			conn.Close()
		}
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
})
