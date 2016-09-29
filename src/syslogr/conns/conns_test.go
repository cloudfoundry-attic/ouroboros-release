package conns_test

import (
	"errors"
	"syslogr/conns"
	"time"

	. "github.com/apoydence/eachers"
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

		done := make(chan struct{})
		go func() {
			defer close(done)
			conns.Handle(mockReader, mockRanger, newMockMetricBatcher())
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

	It("reports received byte count", func() {
		mockReader, mockBatcher, cleanup := startHandle(0)
		defer cleanup()

		testMsg := []byte("I AM A MESSAGE")
		write(mockReader, testMsg, nil)

		Eventually(mockBatcher.BatchAddCounterInput).Should(BeCalled(
			With("syslogr.receivedBytes", uint64(len(testMsg))),
		))
	})
})

func startHandle(delay time.Duration) (*mockReader, *mockMetricBatcher, func()) {
	mockReader := newMockReader()
	mockRanger := newMockRanger()
	mockRanger.DelayRangeOutput.Min <- delay
	mockRanger.DelayRangeOutput.Max <- delay + 1
	mockBatcher := newMockMetricBatcher()
	done := make(chan struct{})
	go func() {
		defer close(done)
		conns.Handle(mockReader, mockRanger, mockBatcher)
	}()
	return mockReader, mockBatcher, func() {
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
