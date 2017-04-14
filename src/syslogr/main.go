package main

import (
	"conf"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"syslogr/conns"
	"syslogr/ranger"
	"time"

	"github.com/bradylove/envstruct"
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/metric_sender"
	"github.com/cloudfoundry/dropsonde/metricbatcher"
)

type Config struct {
	Port       int                `env:"PORT"`
	HTTPSPort  int                `env:"HTTPS_PORT"`
	Delay      conf.DurationRange `env:"DELAY"`
	MetronPort int                `env:"METRON_PORT"`
	Cert       string             `env:"CERT"`
	Key        string             `env:"KEY"`
}

func main() {
	var conf Config
	if err := envstruct.Load(&conf); err != nil {
		panic(err)
	}

	batcher := metricBatcher(conf.MetronPort)
	ranger, err := ranger.New(conf.Delay.Min, conf.Delay.Max)
	if err != nil {
		panic(err)
	}

	go serviceSyslog(conf.Port, ranger, batcher)
	serviceHTTPS(conf.HTTPSPort, conf.Cert, conf.Key, batcher)
}

func metricBatcher(port int) *metricbatcher.MetricBatcher {
	udpEmitter, err := emitter.NewUdpEmitter(fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	eventEmitter := emitter.NewEventEmitter(udpEmitter, "syslogr")
	sender := metric_sender.NewMetricSender(eventEmitter)
	return metricbatcher.New(sender, time.Second)
}

func serviceSyslog(port int, r *ranger.Ranger, b *metricbatcher.MetricBatcher) {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("listening for tcp on: %s", addr)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go conns.Handle(conn, r, b)
	}
}

func serviceHTTPS(port int, cert, key string, b *metricbatcher.MetricBatcher) {
	addr := fmt.Sprintf(":%d", port)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b.BatchCounter("receivedRequest").
			SetTag("protocol", "https").
			Increment()
		d, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		b.BatchCounter("receivedBytes").
			SetTag("protocol", "https").
			Add(uint64(len(d)))
		w.WriteHeader(http.StatusOK)
	})
	log.Printf("listening for https on: %s", addr)
	log.Fatal(http.ListenAndServeTLS(addr, cert, key, handler))
}
