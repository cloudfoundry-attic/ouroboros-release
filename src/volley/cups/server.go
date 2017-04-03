package cups

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
)

func ListenAndServe(
	tlsConfig *tls.Config,
	port int16,
	idGetter idGetter,
	drains []string,
	drainCount int,
) {
	handler := NewCUPSHandler(idGetter, drains, drainCount)

	l, err := tls.Listen("tcp", fmt.Sprintf(":%d", port), tlsConfig)
	if err != nil {
		log.Panicf("Failed to start CUPS provider: %s", err)
	}

	log.Println(http.Serve(l, handler))
}
