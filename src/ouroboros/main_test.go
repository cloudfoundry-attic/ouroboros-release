package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("ouroboros", func() {
	It("receives from a firehose and writes to metron", func() {
		authHandler := &testAuthHandler{
			ClientID:     "gandalf",
			ClientSecret: "keep it secret, keep it safe.",
			token:        "some-good-token",
			requests:     make(chan *http.Request),
		}
		authServer := httptest.NewServer(authHandler)
		defer authServer.Close()

		wsHandler := &testWebsocketHandler{
			token:   authHandler.token,
			started: make(chan struct{}),
			done:    make(chan struct{}),
		}
		wsServer := httptest.NewServer(wsHandler)
		defer wsServer.Close()

		listener := &testUDPListener{
			msgs: make(chan []byte),
		}
		go listener.Listen()
		defer listener.Close()

		path, err := gexec.Build("ouroboros")
		Expect(err).ToNot(HaveOccurred())
		cmd := exec.Command(path)
		cmd.Env = []string{
			"METRON_PORT=3456",
			"SUBSCRIPTION_ID=test",
			fmt.Sprintf("TC_ADDRESS=ws://%s", strings.TrimPrefix(wsServer.URL, "http://")),
			fmt.Sprintf("UAA_ADDRESS=%s", authServer.URL),
			fmt.Sprintf("CLIENT_ID=%s", authHandler.ClientID),
			fmt.Sprintf("CLIENT_SECRET=%s", authHandler.ClientSecret),
		}

		session, err := gexec.Start(cmd, os.Stdout, os.Stderr)
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			session.Terminate().Wait()
		}()

		Eventually(authHandler.requests).Should(Receive())
		Eventually(wsHandler.started).Should(BeClosed())

		msg, err := proto.Marshal(&events.Envelope{
			Origin:    proto.String("test"),
			EventType: events.Envelope_LogMessage.Enum(),
			LogMessage: &events.LogMessage{
				Message:     []byte("I AM A BANANA!"),
				MessageType: events.LogMessage_OUT.Enum(),
				Timestamp:   proto.Int64(time.Now().UnixNano()),
			},
		})
		Expect(err).ToNot(HaveOccurred())

		wsHandler.Send(msg)
		Eventually(listener.msgs).Should(Receive(Equal(msg)))
	})
})
