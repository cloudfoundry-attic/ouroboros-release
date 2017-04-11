package main

import (
	"conf"
	"fmt"
	"net"
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
	Delay      conf.DurationRange `env:"DELAY"`
	MetronPort int                `env:"METRON_PORT"`
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

	serviceSyslog(conf.Port, ranger, batcher)
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
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
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
