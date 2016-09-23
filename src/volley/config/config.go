package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(v []byte) error {
	s := string(bytes.Trim(v, `"`))
	var err error
	d.Duration, err = time.ParseDuration(s)
	return err
}

type Config struct {
	TCAddresses    []string
	AuthToken      string
	FirehoseCount  int
	StreamCount    int
	SubscriptionID string
	MinDelay       Duration
	MaxDelay       Duration
}

func ParseFile(configFile string) (*Config, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Parse(file)
}

func Parse(reader io.Reader) (*Config, error) {
	config := &Config{}
	err := json.NewDecoder(reader).Decode(config)
	if err != nil {
		return nil, err
	}
	if len(config.TCAddresses) == 0 {
		return nil, errors.New("At least one TrafficController URL is required")
	}

	if config.SubscriptionID == "" {
		config.SubscriptionID = "default"
	}

	if os.Getenv("AUTHTOKEN") != "" {
		config.AuthToken = os.Getenv("AUTHTOKEN")
	}

	return config, nil
}
