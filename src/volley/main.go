package main

import (
	"conf"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"syscall"
	"time"
	"tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/bradylove/envstruct"
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/metric_sender"
	"github.com/cloudfoundry/dropsonde/metricbatcher"
	"github.com/cloudfoundry/dropsonde/metrics"

	"volley/syslogdrain"
	"volley/v1"
	"volley/v2"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	log.Println("Volley started...")
	defer log.Println("Volley closing")
	config, err := LoadConfig()
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

	go syslogdrain.ListenAndServe(
		cupsTLS,
		config.CUPSPort,
		idStore,
		config.SyslogDrainURLs,
		config.SyslogDrains,
	)

	egressV1 := v1.NewEgressV1(
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

	if len(config.RLPAddresses) > 0 {
		rlpTLSConfig, err := tls.NewMutualTLSConfig(
			config.TLSCertPath,
			config.TLSKeyPath,
			config.TLSCAPath,
			"reverselogproxy",
		)
		if err != nil {
			log.Panic(err)
		}

		v2ConnManager := v2.NewConnectionManager(
			config.RLPAddresses,
			config.ReceiveDelay,
			config.UsePreferredTags,
			metricBatcher,
			grpc.WithTransportCredentials(credentials.NewTLS(rlpTLSConfig)),
		)
		egressV2 := v2.NewEgressV2(
			v2ConnManager,
			idStore,
			config.FirehoseCount,
			config.StreamCount,
			config.StreamCount,
		)
		go egressV2.Start()
	}

	killer := NewKiller(
		config.KillDelay,
		func() {
			if err := syscall.Kill(os.Getpid(), syscall.SIGKILL); err != nil {
				log.Fatalf("I HAVE TOO MUCH TO LIVE FOR: %s!!!!", err)
			}
		},
	)
	go killer.Start()

	if len(config.SyslogDrainURLs) > 0 {
		syslogRegistrar := syslogdrain.NewSyslogRegistrar(
			config.SyslogTTL,
			config.SyslogDrains,
			config.SyslogDrainURLs,
			config.ETCDAddresses,
			idStore,
		)
		go syslogRegistrar.Start()
	}

	// Blocking on pprof
	if err := http.ListenAndServe("localhost:0", nil); err != nil {
		log.Printf("Error starting pprof server: %s", err)
	}
}

type Config struct {
	TCAddresses          []string           `env:"TC_ADDRS,           required"`
	MetronPort           int                `env:"METRON_PORT,        required"`
	RLPAddresses         []string           `env:"RLP_ADDRS"`
	MetricBatchInterval  time.Duration      `env:"METRIC_BATCH_INTERVAL"`
	ETCDAddresses        []string           `env:"ETCD_ADDRS"`
	SyslogDrainURLs      []string           `env:"SYSLOG_DRAIN_URLS"`
	AuthToken            string             `env:"AUTH_TOKEN"`
	FirehoseCount        int                `env:"FIREHOSE_COUNT"`
	StreamCount          int                `env:"STREAM_COUNT"`
	RecentLogCount       int                `env:"RECENT_LOG_COUNT"`
	ContainerMetricCount int                `env:"CONTAINER_METRIC_COUNT"`
	SyslogDrains         int                `env:"SYSLOG_DRAINS"`
	SyslogTTL            time.Duration      `env:"SYSLOG_TTL"`
	SubscriptionID       string             `env:"SUB_ID"`
	ReceiveDelay         conf.DurationRange `env:"RECV_DELAY"`
	AsyncRequestDelay    conf.DurationRange `env:"ASYNC_REQUEST_DELAY"`
	KillDelay            conf.DurationRange `env:"KILL_DELAY"`
	UsePreferredTags     bool               `env:"USE_PREFERRED_TAGS"`
	TLSCertPath          string             `env:"V2_TLS_CERT_PATH"`
	TLSKeyPath           string             `env:"V2_TLS_KEY_PATH"`
	TLSCAPath            string             `env:"V2_TLS_CA_PATH"`

	CUPSPort       int16  `env:"CUPS_PORT,        required"`
	CUPSServerCert string `env:"CUPS_SERVER_CERT, required"`
	CUPSServerKey  string `env:"CUPS_SERVER_KEY,  required"`
	CUPSServerCA   string `env:"CUPS_SERVER_CA,   required"`
	CUPSServerCN   string `env:"CUPS_SERVER_CN,   required"`
}

func LoadConfig() (Config, error) {
	var c Config
	c.MetricBatchInterval = 5 * time.Second
	err := envstruct.Load(&c)
	return c, err
}

type Killer struct {
	killDelay conf.DurationRange
	kill      func()
}

// NewKiller calls a function after a random delay
func NewKiller(killDelay conf.DurationRange, kill func()) *Killer {
	return &Killer{
		killDelay: killDelay,
		kill:      kill,
	}
}

func (k *Killer) Start() {
	k.killAfterRandomDelay()
}

func (v *Killer) killAfterRandomDelay() {
	delta := int(v.killDelay.Max - v.killDelay.Min)
	killDelay := v.killDelay.Min + time.Duration(rand.Intn(delta))
	time.AfterFunc(killDelay, v.kill)
}
