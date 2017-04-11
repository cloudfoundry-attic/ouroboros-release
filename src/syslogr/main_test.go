package main_test

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

const testMsg = `<14>1 test message ts: 12345 src: App srcId: 123`

var _ = Describe("Main", func() {
	It("fails to start with an invalid port", func() {
		cert, key, cleanup := setupCertKey()
		defer cleanup()

		os.Setenv("DELAY", "10us-15us")
		os.Setenv("PORT", "foo")
		os.Setenv("HTTPS_PORT", "1235")
		os.Setenv("CERT", cert)
		os.Setenv("KEY", key)

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
		cert, key, cleanup := setupCertKey()
		defer cleanup()

		os.Setenv("DELAY", "10us-15us")
		os.Setenv("PORT", "1234")
		os.Setenv("HTTPS_PORT", "1235")
		os.Setenv("CERT", cert)
		os.Setenv("KEY", key)

		path, err := gexec.Build("syslogr")
		Expect(err).ToNot(HaveOccurred())

		cmd := exec.Command(path)
		cmd.Stdin = os.Stdin
		cmd.Stderr = GinkgoWriter
		cmd.Start()
		defer cmd.Process.Signal(os.Interrupt)

		By("creating multiple concurrent connections")
		cleanup1 := writeSyslog("1234")
		defer cleanup1()
		cleanup2 := writeSyslog("1234")
		defer cleanup2()
	})

	It("accepts post requests via https", func() {
		cert, key, cleanup := setupCertKey()
		defer cleanup()

		os.Setenv("DELAY", "10us-15us")
		os.Setenv("PORT", "1234")
		os.Setenv("HTTPS_PORT", "1235")
		os.Setenv("CERT", cert)
		os.Setenv("KEY", key)

		path, err := gexec.Build("syslogr")
		Expect(err).ToNot(HaveOccurred())

		cmd := exec.Command(path)
		cmd.Stdin = os.Stdin
		cmd.Stderr = GinkgoWriter
		cmd.Start()
		defer cmd.Process.Signal(os.Interrupt)

		f := func() error {
			conn, err := net.Dial("tcp", "localhost:1235")
			if err == nil {
				conn.Close()
			}
			return err
		}
		Eventually(f).Should(Succeed())

		By("making https request")
		r := strings.NewReader(testMsg)
		resp, err := insecureClient().Post("https://localhost:1235", "text/plain", r)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(BeNumerically(">=", 200))
		Expect(resp.StatusCode).To(BeNumerically("<", 300))
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

func setupCertKey() (cert, key string, cleanup func()) {
	dir, err := ioutil.TempDir("", "")
	Expect(err).ToNot(HaveOccurred())

	Expect(RestoreAsset(dir, "syslogr.crt")).To(Succeed())
	Expect(RestoreAsset(dir, "syslogr.key")).To(Succeed())

	return path.Join(dir, "syslogr.crt"), path.Join(dir, "syslogr.key"), func() {
		os.RemoveAll(dir)
	}
}

func insecureClient() *http.Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	return &http.Client{
		Transport: tr,
	}
}
