package app

import (
	"conf"
	"time"

	"github.com/bradylove/envstruct"
)

type Config struct {
	TCAddresses          []string           `env:"TC_ADDRS,           required"`
	MetronPort           int                `env:"METRON_PORT,        required"`
	RLPAddresses         []string           `env:"RLP_ADDRS,          required"`
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
