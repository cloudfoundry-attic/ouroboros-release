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
	config, err := config.ParseConfig(*configFilePath)
	if err != nil {
		panic(fmt.Errorf("Unable to parse config: %s", err))
	}
	conn := connectionmanager.New(config)

	for i := 0; i < config.StreamCount; i++ {
		go conn.NewStream()
	}
	for i := 0; i < config.FirehoseCount; i++ {
		go conn.NewFirehose()
	}

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	for range terminate {
		log.Print("Closing connections")
		conn.Close()
	}
}
