package conns_test

import (
	"errors"
	"syslogr/conns"
	"time"

	"github.com/cloudfoundry/dropsonde/metricbatcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Conns", func() {
	It("reads multiple messages from a connection", func() {
		mockReader, _, cleanup := startHandle(0)
		defer cleanup()

		write(mockReader, []byte("first"), nil)
		write(mockReader, []byte("second"), nil)
	})

	It("returns when it receives a read error", func() {
		mockReader := newMockReader()
		mockRanger := newMockRanger()
		mockRanger.DelayRangeOutput.Min <- 0
		mockRanger.DelayRangeOutput.Max <- 0
		batcher := &SpyBatcher{}

		done := make(chan struct{})
		go func() {
			defer close(done)
			conns.Handle(mockReader, mockRanger, batcher)
		}()
		mockReader.ReadOutput.Len <- 0
		mockReader.ReadOutput.Err <- errors.New("boom")
		Eventually(done).Should(BeClosed())
	})

	It("chooses a delay range and pauses between reads", func() {
		mockReader, _, cleanup := startHandle(time.Second / 2)
		defer cleanup()

		write(mockReader, []byte("foo"), nil)
		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(done)
			write(mockReader, []byte("bar"), nil)
		}()
		Consistently(done, 490*time.Millisecond).ShouldNot(BeClosed())
		Eventually(done, 50*time.Millisecond).Should(BeClosed())
	})
})

func startHandle(delay time.Duration) (*mockReader, *SpyBatcher, func()) {
	mockReader := newMockReader()
	mockRanger := newMockRanger()
	mockRanger.DelayRangeOutput.Min <- delay
	mockRanger.DelayRangeOutput.Max <- delay + 1
	batcher := &SpyBatcher{}
	done := make(chan struct{})
	go func() {
		defer close(done)
		conns.Handle(mockReader, mockRanger, batcher)
	}()
	return mockReader, batcher, func() {
		mockReader.ReadOutput.Len <- 0
		mockReader.ReadOutput.Err <- errors.New("stop")
		Eventually(done).Should(BeClosed())
	}
}

func write(r *mockReader, msg []byte, err error) {
	var buf []byte
	Eventually(r.ReadInput.Buf).Should(Receive(&buf))
	Expect(len(buf)).To(BeNumerically(">", len(msg)))
	copy(buf, msg)
	r.ReadOutput.Len <- len(msg)
	r.ReadOutput.Err <- err
}

type SpyBatcher struct{}

func (s *SpyBatcher) BatchCounter(name string) metricbatcher.BatchCounterChainer {
	return s
}
func (s *SpyBatcher) SetTag(key, value string) metricbatcher.BatchCounterChainer {
	return s
}
func (s *SpyBatcher) Increment()       {}
func (s *SpyBatcher) Add(value uint64) {}
