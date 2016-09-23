package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"volley/config"
	"volley/connectionmanager"
)

var (
	configFilePath = flag.String("config", "config/volley.json", "Location of the config json file")
)

func main() {
	flag.Parse()
	config, err := config.ParseFile(*configFilePath)
	if err != nil {
		panic(fmt.Errorf("Unable to parse config: %s", err))
	}
	idStore := connectionmanager.NewIDStore(config.StreamCount)
	conn := connectionmanager.New(config, idStore)

	for i := 0; i < config.StreamCount; i++ {
		go conn.Stream()
	}
	for i := 0; i < config.FirehoseCount; i++ {
		go conn.Firehose()
	}

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	for range terminate {
		log.Print("Closing connections")
		conn.Close()
	}
}
