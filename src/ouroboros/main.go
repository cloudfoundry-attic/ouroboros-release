package main

import (
	"fmt"
	"log"
	egressv1 "ouroboros/internal/egress/v1"
	"ouroboros/internal/ingress"

	"github.com/bradylove/envstruct"
	"github.com/cloudfoundry-incubator/uaago"
)

type config struct {
	TCAddr       string `env:"TC_ADDRESS"`
	SubID        string `env:"SUBSCRIPTION_ID"`
	MetronPort   int    `env:"METRON_PORT"`
	UAAAddr      string `env:"UAA_ADDRESS"`
	ClientID     string `env:"CLIENT_ID"`
	ClientSecret string `env:"CLIENT_SECRET"`
}

func main() {
	conf := loadConfig()
	token := fetchUaaToken(conf)

	log.Println("Starting ouroboros V1 egress")
	v1Writer := egressv1.NewWriter(fmt.Sprintf("localhost:%d", conf.MetronPort))

	log.Println("Starting ouroboros ingress")
	ingress.Consume(conf.TCAddr, conf.SubID, token, v1Writer)
}

func loadConfig() config {
	var conf config
	if err := envstruct.Load(&conf); err != nil {
		log.Fatalf("ouroboros is not happy with your environment: %s", err)
	}

	return conf
}

func fetchUaaToken(conf config) string {
	uaaClient, err := uaago.NewClient(conf.UAAAddr)
	if err != nil {
		log.Panicf("Error creating uaa client: %s", err)
	}

	token, err := uaaClient.GetAuthToken(conf.ClientID, conf.ClientSecret, true)
	if err != nil {
		log.Panicf("Error getting token from uaa: %s", err)
	}

	return token
}
