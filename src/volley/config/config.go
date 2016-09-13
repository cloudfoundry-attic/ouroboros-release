package config

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

type Config struct {
	TCAddresses    []string
	AuthToken      string
	FirehoseCount  int
	StreamCount    int
	SubscriptionId string
	AppID          string
}

func ParseConfig(configFile string) (*Config, error) {
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

	if config.StreamCount > 0 && config.AppID == "" {
		return nil, errors.New("AppID is required to make stream connections")
	}

	if config.SubscriptionId == "" {
		config.SubscriptionId = "default"
	}

	if os.Getenv("AUTHTOKEN") != "" {
		config.AuthToken = os.Getenv("AUTHTOKEN")
	}

	return config, nil
}
