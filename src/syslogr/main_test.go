package main_test

import (
	"net"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

const testMsg = `<14>1 test message ts: 12345 src: App srcId: 123`

var _ = Describe("Main", func() {
	It("fails to start with an invalid port", func() {
		os.Setenv("DELAY", "10us-15us")
		os.Setenv("PORT", "foo")
		path, err := gexec.Build("syslogr")
		Expect(err).ToNot(HaveOccurred())
		cmd := exec.Command(path)
		Expect(cmd.Start()).To(Succeed())
		defer cmd.Process.Signal(os.Interrupt)
		errs := make(chan error)
		go func() {
			err := cmd.Wait()
			errs <- err
		}()
		Eventually(errs).Should(Receive(HaveOccurred()))
	})

	It("accepts tcp connections", func() {
		os.Setenv("DELAY", "10us-15us")
		os.Setenv("PORT", "1234")
		path, err := gexec.Build("syslogr")
		Expect(err).ToNot(HaveOccurred())
		cmd := exec.Command(path)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		cmd.Start()
		defer cmd.Process.Signal(os.Interrupt)

		By("creating multiple concurrent connections")
		cleanup1 := writeSyslog("1234")
		defer cleanup1()
		cleanup2 := writeSyslog("1234")
		defer cleanup2()
	})
})

func writeSyslog(port string) (cleanup func()) {
	var (
		conn net.Conn
		err  error
	)
	Eventually(func() error {
		conn, err = net.Dial("tcp", ":"+port)
		return err
	}).Should(Succeed())

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer GinkgoRecover()

		sendCount := 50000
		errs := make(chan error, sendCount+1)
		go func() {
			// Ensure that we're overflowing the kernel buffer and forcing a flush.
			for i := 0; i < sendCount; i++ {
				_, err = conn.Write([]byte(testMsg))
				errs <- err
			}
		}()

		for i := 0; i < sendCount; i++ {
			Eventually(errs).Should(Receive(BeNil()))
		}
	}()
	return func() {
		<-done
		conn.Close()
	}
}
