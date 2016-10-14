package config

import (
	"conf"
	"time"

	"github.com/bradylove/envstruct"
)

type Config struct {
	TCAddresses          []string           `env:"tc_addrs,required"`
	MetronPort           int                `env:"metron_port,required"`
	MetricBatchInterval  time.Duration      `env:"metric_batch_interval"`
	ETCDAddresses        []string           `env:"etcd_addrs"`
	SyslogAddresses      []string           `env:"syslog_addrs"`
	AuthToken            string             `env:"auth_token"`
	FirehoseCount        int                `env:"firehose_count"`
	StreamCount          int                `env:"stream_count"`
	RecentLogCount       int                `env:"recent_log_count"`
	ContainerMetricCount int                `env:"container_metric_count"`
	SyslogDrains         int                `env:"syslog_drains"`
	SyslogTTL            time.Duration      `env:"syslog_ttl"`
	SubscriptionID       string             `env:"sub_id"`
	ReceiveDelay         conf.DurationRange `env:"recv_delay"`
	AsyncRequestDelay    conf.DurationRange `env:"async_request_delay"`
	KillDelay            conf.DurationRange `env:"kill_delay"`
}

func Load() (Config, error) {
	var c Config
	c.MetricBatchInterval = 5 * time.Second
	err := envstruct.Load(&c)
	return c, err
}
