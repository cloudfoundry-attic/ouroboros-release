package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/bradylove/envstruct"
	"github.com/cloudfoundry-incubator/uaago"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
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

	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	header := make(http.Header)
	header.Set("Origin", "http://localhost")
	header.Set("Authorization", token)
	url := conf.TCAddr + "/firehose/" + conf.SubID
	log.Printf("Connecting to traffic controller at %s", url)
	ws, resp, err := dialer.Dial(url, header)
	if resp != nil && resp.StatusCode != http.StatusSwitchingProtocols {
		panic(fmt.Sprintf("Unexpected status %s", resp.Status))
	}
	if err != nil {
		panic(err)
	}
	defer ws.Close()

	metronAddr := net.UDPAddr{
		Port: conf.MetronPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.DialUDP("udp", nil, &metronAddr)
	if err != nil {
		log.Panicf("could not connect to metron: %s", err)
	}
	defer conn.Close()
	for count := uint64(0); ; count++ {
		_, b, err := ws.ReadMessage()
		if err != nil {
			log.Panicf("could not read websocket message: %s", err)
		}
		_, err = conn.Write(b)
		if err != nil {
			log.Panicf("could not write to metron: %s", err)
		}
		if count%1000 == 0 {
			message, err := proto.Marshal(&events.Envelope{
				Origin:    proto.String("ouroboros"),
				Timestamp: proto.Int64(time.Now().UnixNano()),
				EventType: events.Envelope_CounterEvent.Enum(),
				CounterEvent: &events.CounterEvent{
					Name:  proto.String("ouroboros.forwardedMessages"),
					Delta: proto.Uint64(1000),
					Total: proto.Uint64(count),
				},
			})
			if err != nil {
				log.Panicf("Failed to marshal count: %s", err)
			}
			_, err = conn.Write(message)
			if err != nil {
				log.Panicf("could not write to metron: %s", err)
			}
		}
	}
}
