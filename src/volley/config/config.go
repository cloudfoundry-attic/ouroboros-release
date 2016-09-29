package config

import (
	"conf"

	"github.com/bradylove/envstruct"
)

type Config struct {
	TCAddresses    []string           `env:"tc_addrs,required"`
	AuthToken      string             `env:"auth_token"`
	FirehoseCount  int                `env:"firehose_count"`
	StreamCount    int                `env:"stream_count"`
	SubscriptionID string             `env:"sub_id"`
	ReceiveDelay   conf.DurationRange `env:"recv_delay"`
	KillDelay      conf.DurationRange `env:"kill_delay"`
}

func Load() (Config, error) {
	var c Config
	err := envstruct.Load(&c)
	return c, err
}
