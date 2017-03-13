package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/metric_sender"
	"github.com/cloudfoundry/dropsonde/metricbatcher"
	"github.com/cloudfoundry/dropsonde/metrics"

	"volley/app"
	"volley/config"
	"volley/connectionmanager"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	log.Println("Volley started...")
	defer log.Println("Volley closing")
	config, err := config.Load()
	if err != nil {
		log.Panic(err)
	}
	idStore := connectionmanager.NewIDStore(config.StreamCount)

	udpEmitter, err := emitter.NewUdpEmitter(fmt.Sprintf("127.0.0.1:%d", config.MetronPort))
	if err != nil {
		log.Panic(err)
	}
	eventEmitter := emitter.NewEventEmitter(udpEmitter, "volley")

	metricSender := metric_sender.NewMetricSender(eventEmitter)
	metricBatcher := metricbatcher.New(metricSender, config.MetricBatchInterval)
	metrics.Initialize(metricSender, metricBatcher)

	egressV1 := app.NewEgressV1(
		config.FirehoseCount,
		config.StreamCount,
		config.RecentLogCount,
		config.ContainerMetricCount,
		config.TCAddresses,
		config.AuthToken,
		config.SubscriptionID,
		config.ReceiveDelay,
		config.AsyncRequestDelay,
		idStore,
		metricBatcher,
	)
	go egressV1.Start()

	killer := app.NewKiller(
		config.KillDelay,
		func() {
			if err := syscall.Kill(os.Getpid(), syscall.SIGKILL); err != nil {
				log.Fatalf("I HAVE TOO MUCH TO LIVE FOR: %s!!!!", err)
			}
		},
	)
	go killer.Start()

	syslogRegistrar := app.NewSyslogRegistrar(
		config.SyslogTTL,
		config.SyslogDrains,
		config.SyslogAddresses,
		config.ETCDAddresses,
		idStore,
	)
	go syslogRegistrar.Start()

	// Blocking on pprof
	if err := http.ListenAndServe("localhost:0", nil); err != nil {
		log.Printf("Error starting pprof server: %s", err)
	}
}
