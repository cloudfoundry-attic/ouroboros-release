package main

import (
	"conf"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coreos/etcd/client"

	"volley/config"
	"volley/connectionmanager"
	"volley/drains"
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

	if len(config.ETCDAddresses) > 0 {
		cfg := client.Config{
			Endpoints: config.ETCDAddresses,
		}
		c, err := client.New(cfg)
		if err != nil {
			panic(err)
		}
		api := client.NewKeysAPI(c)
		for i := 0; i < config.SyslogDrains; i++ {
			drains.AdvertiseRandom(idStore, api, config.SyslogAddresses)
		}
	}

	go killAfterRandomDelay(config.KillDelay)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Print("Closing connections")
	conn.Close()
}

func killAfterRandomDelay(delayRange conf.DurationRange) {
	delta := int(delayRange.Max - delayRange.Min)
	killDelay := delayRange.Min + time.Duration(rand.Intn(delta))
	killAfter := time.After(killDelay)
	<-killAfter
	if err := syscall.Kill(os.Getpid(), syscall.SIGKILL); err != nil {
		log.Fatalf("I HAVE TOO MUCH TO LIVE FOR: %s!!!!", err)
	}
}
