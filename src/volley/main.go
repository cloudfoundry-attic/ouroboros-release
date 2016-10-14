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

func init() {
	rand.Seed(time.Now().UnixNano())
}

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

	go syncRequest(config.FirehoseCount, conn.Firehose)
	go syncRequest(config.StreamCount, conn.Stream)
	go asyncRequest(config.AsyncRequestDelay, config.RecentLogCount, conn.RecentLogs)
	go asyncRequest(config.AsyncRequestDelay, config.ContainerMetricCount, conn.ContainerMetrics)

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
			drains.AdvertiseRandom(idStore, api, config.SyslogAddresses, config.SyslogTTL)
		}
	}

	go killAfterRandomDelay(config.KillDelay)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Print("Closing connections")
	conn.Close()
}

func syncRequest(count int, endpoint func()) {
	for i := 0; i < count; i++ {
		go endpoint()
	}
}

func asyncRequest(delay conf.DurationRange, count int, endpoint func()) {
	delta := int(delay.Max - delay.Min)
	for i := 0; i < count; i++ {
		delay := delay.Min + time.Duration(rand.Intn(delta))
		go endpoint()
		time.Sleep(delay)
	}
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
