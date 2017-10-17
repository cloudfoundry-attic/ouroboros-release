package v1_test

import (
	"conf"
	"encoding/binary"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"volley/v1"

	. "github.com/apoydence/eachers"
	"github.com/apoydence/eachers/testhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/dropsonde/envelope_extensions"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
)

var _ = Describe("Connection", func() {
	var (
		handler     *tcServer
		server      *httptest.Server
		mockBatcher *mockBatcher
		mockChainer *mockBatchCounterChainer
		mockIDStore *mockAppIDStore
		conn        *v1.ConnectionManager
	)

	BeforeEach(func() {
		handler = newTCServer()
		server = httptest.NewServer(handler)

		mockIDStore = newMockAppIDStore()
		mockIDStore.GetOutput.Ret0 <- "some-app-id"
		mockBatcher = newMockBatcher()
		mockChainer = newMockBatchCounterChainer()
		testhelpers.AlwaysReturn(mockBatcher.BatchCounterOutput, mockChainer)
		testhelpers.AlwaysReturn(mockChainer.SetTagOutput, mockChainer)

		conn = v1.NewConnectionManager(
			[]string{strings.Replace(server.URL, "http", "ws", 1)},
			"some-auth",
			"some-sub-id",
			conf.DurationRange{},
			mockIDStore,
			mockBatcher,
		)
	})

	AfterEach(func() {
		select {
		case <-handler.done:
			server.Close()
		default:
			handler.stop()
			server.Close()
		}
	})

	Describe("Firehose", func() {
		It("creates a connection to the firehose endpoint", func() {
			go conn.Firehose()

			Eventually(handler.firehoseSubs).Should(Receive(Equal("some-sub-id")))
			Consistently(handler.errs).ShouldNot(Receive())
		})

		It("increments an openConnections metric when a new connection is made", func() {
			go conn.Firehose()

			Eventually(handler.firehoseSubs).Should(Receive(Equal("some-sub-id")))
			Eventually(mockBatcher.BatchCounterInput).Should(BeCalled(With("volley.openConnections")))
			Eventually(mockChainer.SetTagInput).Should(BeCalled(With("conn_type", "firehose")))
			Eventually(mockChainer.IncrementCalled).Should(BeCalled())
		})

		It("increments a closedConnections metric when an error occurs", func() {
			go conn.Firehose()

			Eventually(handler.firehoseSubs).Should(Receive(Equal("some-sub-id")))
			Eventually(mockBatcher.BatchCounterInput).Should(BeCalled(With("volley.openConnections")))
			Eventually(mockChainer.SetTagInput).Should(BeCalled(With("conn_type", "firehose")))
			Eventually(mockChainer.IncrementCalled).Should(BeCalled())

			handler.stop()
			Eventually(mockBatcher.BatchCounterInput).Should(BeCalled(With("volley.closedConnections")))
			Eventually(mockChainer.SetTagInput).Should(BeCalled(With("conn_type", "firehose")))
			Eventually(mockChainer.IncrementCalled).Should(BeCalled())
		})

		It("increments a receivedEnvelopes metric when new envelopes are received", func() {
			go conn.Firehose()

			Eventually(handler.firehoseSubs).Should(Receive(Equal("some-sub-id")))

			done := make(chan struct{})
			defer close(done)
			go func(mockIDStore *mockAppIDStore, done chan struct{}) {
				for {
					select {
					case <-done:
						return
					default:
						<-mockIDStore.AddInput.AppID
						<-mockIDStore.AddCalled
					}
				}
			}(mockIDStore, done)
			go handler.sendLoop(2000)

			// Drain volley.openConnections
			<-mockBatcher.BatchCounterInput.Name
			<-mockChainer.SetTagInput.Key
			<-mockChainer.SetTagInput.Value

			Eventually(mockBatcher.BatchCounterInput).Should(BeCalled(With("volley.receivedEnvelopes")))
			Eventually(mockChainer.SetTagInput).Should(BeCalled(With("conn_type", "firehose")))
			Eventually(mockChainer.AddInput).Should(BeCalled(With(uint64(1000))))

			By("batching by 1000")
			Eventually(mockBatcher.BatchCounterInput).Should(BeCalled(With("volley.receivedEnvelopes")))
			Eventually(mockChainer.SetTagInput).Should(BeCalled(With("conn_type", "firehose")))
			Eventually(mockChainer.AddInput).Should(BeCalled(With(uint64(1000))))
		})

		It("is a slow consumer when delay is set", func() {
			slowConn := v1.NewConnectionManager(
				[]string{strings.Replace(server.URL, "http", "ws", 1)},
				"some-auth",
				"some-sub-id",
				conf.DurationRange{
					Min: 99 * time.Millisecond,
					Max: 100 * time.Millisecond,
				},
				mockIDStore,
				mockBatcher,
			)

			go slowConn.Firehose()

			Eventually(handler.firehoseSubs).Should(Receive(Equal("some-sub-id")))
			go handler.sendLoop(10000000)
			var err error
			Eventually(handler.errs).Should(Receive(&err))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("i/o timeout"))
		})

		It("ignores system app IDs", func() {
			go conn.Firehose()

			Eventually(handler.firehoseSubs).Should(Receive(Equal("some-sub-id")))

			ev := &events.Envelope{
				Origin:    proto.String("foo"),
				EventType: events.Envelope_LogMessage.Enum(),
				LogMessage: &events.LogMessage{
					AppId:       proto.String(envelope_extensions.SystemAppId),
					Message:     []byte("foo"),
					MessageType: events.LogMessage_OUT.Enum(),
					Timestamp:   proto.Int64(time.Now().UnixNano()),
				},
			}

			b, err := proto.Marshal(ev)
			Expect(err).ToNot(HaveOccurred())
			Expect(handler.ws.SetWriteDeadline(time.Now().Add(10 * time.Millisecond))).To(Succeed())
			Expect(handler.ws.WriteMessage(websocket.BinaryMessage, b)).To(Succeed())

			Consistently(mockIDStore.AddInput).ShouldNot(BeCalled())
		})

		DescribeTable("app ID store event types", func(ev *events.Envelope, appID string) {
			go conn.Firehose()

			Eventually(handler.firehoseSubs).Should(Receive(Equal("some-sub-id")))

			b, err := proto.Marshal(ev)
			Expect(err).ToNot(HaveOccurred())
			Expect(handler.ws.SetWriteDeadline(time.Now().Add(10 * time.Millisecond))).To(Succeed())
			Expect(handler.ws.WriteMessage(websocket.BinaryMessage, b)).To(Succeed())

			if appID == "" {
				Consistently(mockIDStore.AddInput).ShouldNot(BeCalled())
				return
			}
			Eventually(mockIDStore.AddInput).Should(BeCalled(With(appID)))
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
			go conn.Stream()

			Eventually(handler.streamApps).Should(Receive())
			Consistently(handler.errs).ShouldNot(Receive())
		})

		It("increments an openConnections metric when a new connection is made", func() {
			go conn.Stream()

			Eventually(handler.streamApps).Should(Receive())
			Consistently(handler.errs).ShouldNot(Receive())
			Eventually(mockBatcher.BatchCounterInput).Should(BeCalled(With("volley.openConnections")))
			Eventually(mockChainer.SetTagInput).Should(BeCalled(With("conn_type", "stream")))
			Eventually(mockChainer.IncrementCalled).Should(BeCalled())
		})

		It("increments a closedConnections metric when an error occurs", func() {
			go conn.Stream()

			Eventually(handler.streamApps).Should(Receive())
			Consistently(handler.errs).ShouldNot(Receive())
			Eventually(mockBatcher.BatchCounterInput).Should(BeCalled(With("volley.openConnections")))
			Eventually(mockChainer.SetTagInput).Should(BeCalled(With("conn_type", "stream")))
			Eventually(mockChainer.IncrementCalled).Should(BeCalled())

			handler.stop()
			Eventually(mockBatcher.BatchCounterInput).Should(BeCalled(With("volley.closedConnections")))
			Eventually(mockChainer.SetTagInput).Should(BeCalled(With("conn_type", "stream")))
			Eventually(mockChainer.IncrementCalled).Should(BeCalled())
		})

		It("increments a receivedEnvelopes metric when new envelopes are received", func() {
			go conn.Stream()

			Eventually(handler.streamApps).Should(Receive())

			done := make(chan struct{})
			defer close(done)
			go func(mockIDStore *mockAppIDStore, done chan struct{}) {
				for {
					select {
					case <-done:
						return
					default:
						<-mockIDStore.AddInput.AppID
						<-mockIDStore.AddCalled
					}
				}
			}(mockIDStore, done)
			go handler.sendLoop(2000)

			// Drain volley.openConnections
			<-mockBatcher.BatchCounterInput.Name
			<-mockChainer.SetTagInput.Key
			<-mockChainer.SetTagInput.Value

			Eventually(mockBatcher.BatchCounterInput).Should(BeCalled(With("volley.receivedEnvelopes")))
			Eventually(mockChainer.SetTagInput).Should(BeCalled(With("conn_type", "stream")))
			Eventually(mockChainer.AddInput).Should(BeCalled(With(uint64(1000))))

			Eventually(mockBatcher.BatchCounterInput).Should(BeCalled(With("volley.receivedEnvelopes")))
			Eventually(mockChainer.SetTagInput).Should(BeCalled(With("conn_type", "stream")))
			Eventually(mockChainer.AddInput).Should(BeCalled(With(uint64(1000))))
		})

		It("is a slow consumer when delay is set", func() {
			slowConn := v1.NewConnectionManager(
				[]string{strings.Replace(server.URL, "http", "ws", 1)},
				"some-auth",
				"some-sub-id",
				conf.DurationRange{
					Min: 99 * time.Millisecond,
					Max: 100 * time.Millisecond,
				},
				mockIDStore,
				mockBatcher,
			)

			go slowConn.Stream()

			Eventually(handler.streamApps).Should(Receive())
			go handler.sendLoop(10000000)
			Eventually(handler.errs).Should(Receive())
		})
	})

	Describe("Recent Logs", func() {
		It("sends a request to the recentlogs endpoint", func() {
			go conn.RecentLogs()

			Eventually(handler.recentLogReqs).Should(Receive())
			Consistently(handler.errs).ShouldNot(Receive())
		})

		It("increments a connection metric for recentlogs", func() {
			msg := &events.Envelope{
				Origin:    proto.String("foo"),
				EventType: events.Envelope_LogMessage.Enum(),
				LogMessage: &events.LogMessage{
					Message:     []byte("some-log"),
					MessageType: events.LogMessage_OUT.Enum(),
					Timestamp:   proto.Int64(time.Now().UnixNano()),
					AppId:       proto.String("some-app"),
				},
			}
			b, err := proto.Marshal(msg)
			Expect(err).ToNot(HaveOccurred())
			handler.setResponse(response{
				data:       b,
				statusCode: 200,
			})

			go conn.RecentLogs()
			close(handler.done)

			Eventually(handler.recentLogReqs).Should(Receive())
			Consistently(handler.errs).ShouldNot(Receive())
			Eventually(mockBatcher.BatchCounterInput, 2).Should(BeCalled(With("volley.numberOfRequests")))
			Eventually(mockChainer.SetTagInput).Should(BeCalled(With("conn_type", "recentlogs")))
			Eventually(mockChainer.IncrementCalled).Should(BeCalled())
		})

		It("increments an error metric if request errors out", func() {
			handler.setResponse(response{data: nil, statusCode: 500})
			go conn.RecentLogs()
			close(handler.done)

			Eventually(handler.recentLogReqs).Should(Receive())
			Consistently(handler.errs).ShouldNot(Receive())
			Eventually(mockBatcher.BatchCounterInput, 2).Should(BeCalled(With("volley.numberOfRequestErrors")))
			Eventually(mockChainer.SetTagInput).Should(BeCalled(With("conn_type", "recentlogs")))
			Eventually(mockChainer.IncrementCalled).Should(BeCalled())
		})
	})

	Describe("Container Metrics", func() {
		It("sends a request to the container metrics endpoint", func() {
			go conn.ContainerMetrics()

			Eventually(handler.containerMetricsReqs).Should(Receive())
			Consistently(handler.errs).ShouldNot(Receive())
		})

		It("increments a connection metric for containerMetrics", func() {
			env := &events.Envelope{
				Origin:    proto.String("foo"),
				EventType: events.Envelope_ContainerMetric.Enum(),
				ContainerMetric: &events.ContainerMetric{
					ApplicationId: proto.String("some-app"),
					InstanceIndex: proto.Int32(1),
					CpuPercentage: proto.Float64(0.1),
					MemoryBytes:   proto.Uint64(1),
					DiskBytes:     proto.Uint64(1),
				},
			}
			b, err := proto.Marshal(env)
			Expect(err).ToNot(HaveOccurred())
			handler.setResponse(response{
				data:       b,
				statusCode: 200,
			})

			go conn.ContainerMetrics()
			close(handler.done)

			Eventually(handler.containerMetricsReqs).Should(Receive())
			Consistently(handler.errs).ShouldNot(Receive())
			Eventually(mockBatcher.BatchCounterInput, 2).Should(BeCalled(With("volley.numberOfRequests")))
			Eventually(mockChainer.SetTagInput).Should(BeCalled(With("conn_type", "containermetrics")))
			Eventually(mockChainer.IncrementCalled).Should(BeCalled())
		})

		It("increments an error metric if request errors out", func() {
			handler.setResponse(response{data: nil, statusCode: 500})
			go conn.ContainerMetrics()
			close(handler.done)

			Eventually(handler.containerMetricsReqs).Should(Receive())
			Consistently(handler.errs).ShouldNot(Receive())
			Eventually(mockBatcher.BatchCounterInput, 2).Should(BeCalled(With("volley.numberOfRequestErrors")))
			Eventually(mockChainer.SetTagInput).Should(BeCalled(With("conn_type", "containermetrics")))
			Eventually(mockChainer.IncrementCalled).Should(BeCalled())
		})
	})
})

type tcServer struct {
	ws                   *websocket.Conn
	response             response
	streamApps           chan string
	firehoseSubs         chan string
	containerMetricsReqs chan string
	recentLogReqs        chan string
	errs                 chan error
	done                 chan struct{}
}

type response struct {
	data       []byte
	statusCode int
}

func newTCServer() *tcServer {
	return &tcServer{
		streamApps:           make(chan string, 100),
		firehoseSubs:         make(chan string, 100),
		recentLogReqs:        make(chan string, 100),
		containerMetricsReqs: make(chan string, 100),
		errs:                 make(chan error, 100),
		done:                 make(chan struct{}),
	}
}

func (tc *tcServer) stop() {
	if tc.ws != nil {
		tc.ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Time{})
		tc.ws.Close()
	}
	close(tc.done)
}

func (tc *tcServer) sendLoop(sendCount int) {
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

	for i := 0; i < sendCount; i++ {
		Expect(tc.ws.SetWriteDeadline(time.Now().Add(10 * time.Millisecond))).To(Succeed())
		err := tc.ws.WriteMessage(websocket.BinaryMessage, b)
		if err != nil {
			tc.errs <- err
			break
		}
	}
}

func (tc *tcServer) setResponse(resp response) {
	tc.response = resp
}

func (tc *tcServer) createMultiPartResp(rw http.ResponseWriter) {
	if tc.response.statusCode != http.StatusOK {
		http.Error(rw, "bad request", tc.response.statusCode)
		return
	}

	mp := multipart.NewWriter(rw)
	defer mp.Close()

	rw.Header().Set("Content-Type", `multipart/x-protobuf; boundary=`+mp.Boundary())

	partWriter, err := mp.CreatePart(nil)
	if err != nil {
		tc.errs <- err
	}
	partWriter.Write(tc.response.data)
}

func (tc *tcServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer GinkgoRecover()
	var err error

	switch {
	case strings.Contains(r.URL.Path, "stream"):
		tc.ws, err = websocket.Upgrade(w, r, nil, 1024, 1024)
		Expect(err).ToNot(HaveOccurred())
		idStart := strings.Index(r.URL.Path, "stream/") + len("stream/")
		idEnd := idStart + strings.Index(r.URL.Path[idStart:], "/")
		tc.streamApps <- r.URL.Path[idStart:idEnd]
	case strings.Contains(r.URL.Path, "firehose"):
		tc.ws, err = websocket.Upgrade(w, r, nil, 1024, 1024)
		Expect(err).ToNot(HaveOccurred())
		idStart := strings.Index(r.URL.Path, "firehose/") + len("firehose/")
		tc.firehoseSubs <- r.URL.Path[idStart:]
	case strings.Contains(r.URL.Path, "recentlogs"):
		idStart := strings.Index(r.URL.Path, "recentlogs/") + len("recentlogs/")
		idEnd := idStart + strings.Index(r.URL.Path[idStart:], "/")
		tc.recentLogReqs <- r.URL.Path[idStart:idEnd]
		tc.createMultiPartResp(w)
	case strings.Contains(r.URL.Path, "containermetrics"):
		idStart := strings.Index(r.URL.Path, "containermetrics/") + len("containermetrics/")
		idEnd := idStart + strings.Index(r.URL.Path[idStart:], "/")
		tc.containerMetricsReqs <- r.URL.Path[idStart:idEnd]
		tc.createMultiPartResp(w)
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
