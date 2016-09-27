package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"volley/config"
	"volley/connectionmanager"
)

func main() {
	println("started")
	config, err := config.Load()
	if err != nil {
		panic(err)
	}
	idStore := connectionmanager.NewIDStore(config.StreamCount)
	conn := connectionmanager.New(config, idStore)

	for i := 0; i < config.FirehoseCount; i++ {
		go conn.Firehose()
	}
	for i := 0; i < config.StreamCount; i++ {
		go conn.Stream()
	}

	go killAfterRandomDelay(config.KillDelay)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Print("Closing connections")
	conn.Close()
}

func killAfterRandomDelay(delayRange config.DurationRange) {
	delta := int(delayRange.Max - delayRange.Min)
	killDelay := delayRange.Min + time.Duration(rand.Intn(delta))
	killAfter := time.After(killDelay)
	<-killAfter
	if err := syscall.Kill(os.Getpid(), syscall.SIGKILL); err != nil {
		log.Fatalf("I HAVE TOO MUCH TO LIVE FOR: %s!!!!", err)
	}
}
