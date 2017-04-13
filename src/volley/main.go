package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"syscall"
	"time"
	"tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/metric_sender"
	"github.com/cloudfoundry/dropsonde/metricbatcher"
	"github.com/cloudfoundry/dropsonde/metrics"

	"volley/app"
	"volley/cups"
	"volley/v1"
	"volley/v2"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	log.Println("Volley started...")
	defer log.Println("Volley closing")
	config, err := app.LoadConfig()
	if err != nil {
		log.Panic(err)
	}
	idStore := v1.NewIDStore(config.StreamCount)

	udpEmitter, err := emitter.NewUdpEmitter(fmt.Sprintf("127.0.0.1:%d", config.MetronPort))
	if err != nil {
		log.Panic(err)
	}
	eventEmitter := emitter.NewEventEmitter(udpEmitter, "volley")

	metricSender := metric_sender.NewMetricSender(eventEmitter)
	metricBatcher := metricbatcher.New(metricSender, config.MetricBatchInterval)
	metrics.Initialize(metricSender, metricBatcher)

	cupsTLS, err := tls.NewMutualTLSConfig(
		config.CUPSServerCert,
		config.CUPSServerKey,
		config.CUPSServerCA,
		config.CUPSServerCN,
	)
	if err != nil {
		log.Printf("Failed to load CUPS TLS config: %s", err)
	}

	go cups.ListenAndServe(
		cupsTLS,
		config.CUPSPort,
		idStore,
		config.SyslogDrainURLs,
		config.SyslogDrains,
	)

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

	tlsConfig, err := tls.NewMutualTLSConfig(config.TLSCertPath, config.TLSKeyPath, config.TLSCAPath, "reverselogproxy")
	if err != nil {
		log.Panic(err)
	}

	v2ConnManager := v2.NewConnectionManager(
		config.RLPAddresses,
		config.ReceiveDelay,
		metricBatcher,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	)
	egressV2 := app.NewEgressV2(
		v2ConnManager,
		idStore,
		config.FirehoseCount,
		config.StreamCount,
		config.StreamCount,
	)
	go egressV2.Start()

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
		config.SyslogDrainURLs,
		config.ETCDAddresses,
		idStore,
	)
	go syslogRegistrar.Start()

	// Blocking on pprof
	if err := http.ListenAndServe("localhost:0", nil); err != nil {
		log.Printf("Error starting pprof server: %s", err)
	}
}
