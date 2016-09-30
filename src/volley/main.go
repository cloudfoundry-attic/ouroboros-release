package main

import (
	"conf"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/metric_sender"
	"github.com/cloudfoundry/dropsonde/metricbatcher"
	"github.com/cloudfoundry/dropsonde/metrics"
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

	udpEmitter, err := emitter.NewUdpEmitter(fmt.Sprintf("127.0.0.1:%d", config.MetronPort))
	if err != nil {
		panic(err)
	}
	eventEmitter := emitter.NewEventEmitter(udpEmitter, "volley")

	metricSender := metric_sender.NewMetricSender(eventEmitter)
	metricBatcher := metricbatcher.New(metricSender, config.MetricBatchInterval)
	metrics.Initialize(metricSender, metricBatcher)

	conn := connectionmanager.New(config, idStore, metricBatcher)

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
