package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/bradylove/envstruct"
)

type DurationRange struct {
	Min, Max time.Duration
}

func (d *DurationRange) UnmarshalEnv(v string) error {
	values := strings.Split(v, "-")
	if len(values) != 2 {
		return fmt.Errorf("Expected DurationRange to be of format {min}-{max}")
	}
	var err error
	d.Min, err = time.ParseDuration(values[0])
	if err != nil {
		return fmt.Errorf("Error parsing DurationRange.Min: %s", err)
	}
	d.Max, err = time.ParseDuration(values[1])
	if err != nil {
		return fmt.Errorf("Error parsing DurationRange.Max: %s", err)
	}
	return nil
}

type Config struct {
	TCAddresses    []string      `env:"tc_addrs,required"`
	AuthToken      string        `env:"auth_token"`
	FirehoseCount  int           `env:"firehose_count"`
	StreamCount    int           `env:"stream_count"`
	SubscriptionID string        `env:"sub_id"`
	ReceiveDelay   DurationRange `env:"recv_delay"`
	KillDelay      DurationRange `env:"kill_delay"`
}

func Load() (Config, error) {
	var c Config
	err := envstruct.Load(&c)
	return c, err
}
