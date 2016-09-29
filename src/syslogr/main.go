package main

import (
	"conf"
	"net"
	"strconv"
	"syslogr/conns"
	"syslogr/ranger"
	"time"

	"github.com/bradylove/envstruct"
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/metric_sender"
	"github.com/cloudfoundry/dropsonde/metricbatcher"
)

type Config struct {
	Port       int                `env:"port"`
	Delay      conf.DurationRange `env:"delay"`
	MetronPort int                `env:"metron_port"`
}

func main() {
	var conf Config
	if err := envstruct.Load(&conf); err != nil {
		panic(err)
	}
	l, err := net.Listen("tcp", ":"+strconv.Itoa(conf.Port))
	if err != nil {
		panic(err)
	}
	defer l.Close()
	ranger, err := ranger.New(conf.Delay.Min, conf.Delay.Max)
	if err != nil {
		panic(err)
	}
	e, err := emitter.NewUdpEmitter(":" + strconv.Itoa(conf.MetronPort))
	if err != nil {
		panic(err)
	}
	sender := metric_sender.NewMetricSender(emitter.NewEventEmitter(e, "syslogr"))
	batcher := metricbatcher.New(sender, time.Second)
	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go conns.Handle(conn, ranger, batcher)
	}

}

func handleConn(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		conn.Read(buf)
	}
}
