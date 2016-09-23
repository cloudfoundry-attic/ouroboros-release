package connectionmanager_test

import (
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
	"volley/config"
	"volley/connectionmanager"

	. "github.com/apoydence/eachers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
)

var _ = Describe("Connection", func() {
	Describe("Firehose", func() {
		It("creates a connection to the firehose endpoint", func() {
			handler := newTCServer()
			defer handler.stop()
			server := httptest.NewServer(handler)
			defer server.Close()

			conf := &config.Config{
				TCAddresses:    []string{strings.Replace(server.URL, "http", "ws", 1)},
				SubscriptionID: "some-sub-id",
			}
			mockAppIDStore := newMockAppIDStore()
			conn := connectionmanager.New(conf, mockAppIDStore)
			go conn.Firehose()

			Eventually(handler.firehoseSubs).Should(Receive(Equal(conf.SubscriptionID)))
			Consistently(handler.errs).ShouldNot(Receive())
		})

		It("is a slow consumer when delay is set", func() {
			handler := newTCServer()
			defer handler.stop()
			server := httptest.NewServer(handler)
			defer server.Close()

			conf := &config.Config{
				TCAddresses:    []string{strings.Replace(server.URL, "http", "ws", 1)},
				SubscriptionID: "some-sub-id",
				MinDelay:       config.Duration{99 * time.Millisecond},
				MaxDelay:       config.Duration{100 * time.Millisecond},
			}
			mockAppIDStore := newMockAppIDStore()
			conn := connectionmanager.New(conf, mockAppIDStore)
			go conn.Firehose()

			Eventually(handler.firehoseSubs).Should(Receive(Equal(conf.SubscriptionID)))
			go handler.sendLoop()
			var err error
			Eventually(handler.errs).Should(Receive(&err))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("i/o timeout"))
		})

		DescribeTable("app ID store event types", func(ev *events.Envelope, appID string) {
			handler := newTCServer()
			defer handler.stop()
			server := httptest.NewServer(handler)
			defer server.Close()

			conf := &config.Config{
				TCAddresses:    []string{strings.Replace(server.URL, "http", "ws", 1)},
				SubscriptionID: "some-sub-id",
			}
			mockAppIDStore := newMockAppIDStore()
			conn := connectionmanager.New(conf, mockAppIDStore)
			go conn.Firehose()

			Eventually(handler.firehoseSubs).Should(Receive(Equal(conf.SubscriptionID)))

			b, err := proto.Marshal(ev)
			Expect(err).ToNot(HaveOccurred())
			Expect(handler.ws.SetWriteDeadline(time.Now().Add(10 * time.Millisecond))).To(Succeed())
			Expect(handler.ws.WriteMessage(websocket.BinaryMessage, b)).To(Succeed())

			if appID == "" {
				Consistently(mockAppIDStore.AddInput).ShouldNot(BeCalled())
				return
			}
			Eventually(mockAppIDStore.AddInput).Should(BeCalled(With(appID)))
		},
			Entry("LogMessage", &events.Envelope{
				Origin:    proto.String("foo"),
				EventType: events.Envelope_LogMessage.Enum(),
				LogMessage: &events.LogMessage{
					AppId:       proto.String("i-am-an-app-id"),
					Message:     []byte("foo"),
					MessageType: events.LogMessage_OUT.Enum(),
					Timestamp:   proto.Int64(time.Now().UnixNano()),
				},
			}, "i-am-an-app-id"),
			Entry("LogMessage Without AppID", &events.Envelope{
				Origin:    proto.String("foo"),
				EventType: events.Envelope_LogMessage.Enum(),
				LogMessage: &events.LogMessage{
					Message:     []byte("foo"),
					MessageType: events.LogMessage_OUT.Enum(),
					Timestamp:   proto.Int64(time.Now().UnixNano()),
				},
			}, ""),
			Entry("HttpStartStop", &events.Envelope{
				Origin:    proto.String("foo"),
				EventType: events.Envelope_HttpStartStop.Enum(),
				HttpStartStop: &events.HttpStartStop{
					ApplicationId: &events.UUID{
						High: proto.Uint64(2),
						Low:  proto.Uint64(5),
					},
					StartTimestamp: proto.Int64(time.Now().UnixNano()),
					StopTimestamp:  proto.Int64(time.Now().UnixNano()),
					RequestId: &events.UUID{
						High: proto.Uint64(1),
						Low:  proto.Uint64(1),
					},
					PeerType:      events.PeerType_Client.Enum(),
					Method:        events.Method_GET.Enum(),
					Uri:           proto.String("/"),
					RemoteAddress: proto.String("foo"),
					UserAgent:     proto.String("n/a"),
					StatusCode:    proto.Int32(http.StatusTeapot),
					ContentLength: proto.Int64(1024),
				},
			}, formatUUID(&events.UUID{
				High: proto.Uint64(2),
				Low:  proto.Uint64(5),
			})),
		)
	})

	Describe("Stream", func() {
		It("creates a connection to the stream endpoint", func() {
			handler := newTCServer()
			defer handler.stop()
			server := httptest.NewServer(handler)
			defer server.Close()

			conf := &config.Config{
				TCAddresses: []string{strings.Replace(server.URL, "http", "ws", 1)},
			}
			mockAppIDStore := newMockAppIDStore()
			mockAppIDStore.GetOutput.Ret0 <- "some-app-id"
			conn := connectionmanager.New(conf, mockAppIDStore)
			go conn.Stream()

			Eventually(handler.streamApps).Should(Receive())
			Consistently(handler.errs).ShouldNot(Receive())
		})

		It("is a slow consumer when delay is set", func() {
			handler := newTCServer()
			defer handler.stop()
			server := httptest.NewServer(handler)
			defer server.Close()

			conf := &config.Config{
				TCAddresses:    []string{strings.Replace(server.URL, "http", "ws", 1)},
				SubscriptionID: "some-sub-id",
				MinDelay:       config.Duration{99 * time.Millisecond},
				MaxDelay:       config.Duration{100 * time.Millisecond},
			}
			mockAppIDStore := newMockAppIDStore()
			mockAppIDStore.GetOutput.Ret0 <- "some-app-id"
			conn := connectionmanager.New(conf, mockAppIDStore)
			go conn.Stream()

			Eventually(handler.streamApps).Should(Receive())
			go handler.sendLoop()
			Eventually(handler.errs).Should(Receive())
		})
	})
})

type tcServer struct {
	ws           *websocket.Conn
	streamApps   chan string
	firehoseSubs chan string
	errs         chan error
	done         chan struct{}
}

func newTCServer() *tcServer {
	return &tcServer{
		streamApps:   make(chan string, 100),
		firehoseSubs: make(chan string, 100),
		errs:         make(chan error, 100),
		done:         make(chan struct{}),
	}
}

func (tc *tcServer) stop() {
	tc.ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Time{})
	tc.ws.Close()
	close(tc.done)
}

func (tc *tcServer) sendLoop() {
	msg := &events.Envelope{
		Origin:    proto.String("foo"),
		EventType: events.Envelope_CounterEvent.Enum(),
		CounterEvent: &events.CounterEvent{
			Total: proto.Uint64(10),
			Delta: proto.Uint64(3),
			Name:  proto.String("foo"),
		},
	}
	b, err := proto.Marshal(msg)
	Expect(err).ToNot(HaveOccurred())
	for {
		Expect(tc.ws.SetWriteDeadline(time.Now().Add(10 * time.Millisecond))).To(Succeed())
		err := tc.ws.WriteMessage(websocket.BinaryMessage, b)
		if err != nil {
			tc.errs <- err
			break
		}
	}
}

func (tc *tcServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer GinkgoRecover()
	var err error
	tc.ws, err = websocket.Upgrade(w, r, nil, 1024, 1024)
	Expect(err).ToNot(HaveOccurred())

	switch {
	case strings.Contains(r.URL.Path, "stream"):
		idStart := strings.Index(r.URL.Path, "stream/") + len("stream/")
		idEnd := idStart + strings.Index(r.URL.Path[idStart:], "/")
		tc.streamApps <- r.URL.Path[idStart:idEnd]
	case strings.Contains(r.URL.Path, "firehose"):
		idStart := strings.Index(r.URL.Path, "firehose/") + len("firehose/")
		tc.firehoseSubs <- r.URL.Path[idStart:]
	}
	log.Printf("Waiting on done")
	<-tc.done
	log.Printf("Done")
}

func formatUUID(uuid *events.UUID) string {
	var uuidBytes [16]byte
	binary.LittleEndian.PutUint64(uuidBytes[:8], uuid.GetLow())
	binary.LittleEndian.PutUint64(uuidBytes[8:], uuid.GetHigh())
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuidBytes[0:4], uuidBytes[4:6], uuidBytes[6:8], uuidBytes[8:10], uuidBytes[10:])
}
