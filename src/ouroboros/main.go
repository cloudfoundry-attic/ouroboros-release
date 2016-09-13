package main

import (
	"crypto/tls"
	"log"
	"net"

	"github.com/bradylove/envstruct"
	"github.com/cloudfoundry-incubator/uaago"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/gogo/protobuf/proto"
)

type config struct {
	TCAddr       string `env:"tc_address"`
	SubID        string `env:"subscription_id"`
	MetronPort   int    `env:"metron_port"`
	UAAAddr      string `env:"uaa_address"`
	ClientID     string `env:"client_id"`
	ClientSecret string `env:"client_secret"`
}

func main() {
	var conf config
	err := envstruct.Load(&conf)
	if err != nil {
		log.Fatalf("ouroboros is not happy with your environment: %s", err)
	}

	uaaClient, err := uaago.NewClient(conf.UAAAddr)
	if err != nil {
		log.Panicf("Error creating uaa client: %s", err)
	}
	token, err := uaaClient.GetAuthToken(conf.ClientID, conf.ClientSecret, true)
	if err != nil {
		log.Panicf("Error getting token from uaa: %s", err)
	}
	consumer := consumer.New(conf.TCAddr, &tls.Config{InsecureSkipVerify: true}, nil)
	msgs, errs := consumer.Firehose(conf.SubID, token)
	go func() {
		err := <-errs
		log.Panicf("received %s", err)
	}()

	metronAddr := net.UDPAddr{
		Port: conf.MetronPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.DialUDP("udp", nil, &metronAddr)
	if err != nil {
		log.Panicf("could not connect to metron: %s", err)
	}
	defer conn.Close()
	for msg := range msgs {
		b, err := proto.Marshal(msg)
		if err != nil {
			log.Panicf("could not marshal envelope: %s", err)
		}

		_, err = conn.Write(b)
		if err != nil {
			log.Panicf("could not write to metron: %s", err)
		}
	}
}
