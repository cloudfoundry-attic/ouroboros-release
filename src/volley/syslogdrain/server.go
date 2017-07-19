package syslogdrain

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
)

// ListenAndServe starts a TCP listener which emulates CAPI
// and will provide the specified number of syslog drains
// with corresponding URLs
func ListenAndServe(
	tlsConfig *tls.Config,
	port int16,
	idGetter idGetter,
	drainURLs []string,
	drainCount int,
) {
	l, err := tls.Listen("tcp", fmt.Sprintf(":%d", port), tlsConfig)
	if err != nil {
		log.Panicf("Failed to start CUPS provider: %s", err)
	}

	handler := NewCUPSHandler(idGetter, drainURLs, drainCount)
	log.Println(http.Serve(l, handler))
}
