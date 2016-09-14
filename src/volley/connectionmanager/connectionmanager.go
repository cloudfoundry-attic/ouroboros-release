package connectionmanager

import (
	"crypto/tls"
	"log"
	"math/rand"
	"sync"
	"volley/config"

	"github.com/cloudfoundry/noaa/consumer"
)

type ConnectionManager struct {
	consumers    []*consumer.Consumer
	consumerLock sync.Mutex
	conf         *config.Config
}

func New(conf *config.Config) *ConnectionManager {
	var consumers []*consumer.Consumer
	for _, tcAddress := range conf.TCAddresses {
		c := consumer.New(tcAddress, &tls.Config{InsecureSkipVerify: true}, nil)
		consumers = append(consumers, c)
	}

	return &ConnectionManager{
		conf:      conf,
		consumers: consumers,
	}
}

func (c *ConnectionManager) pick() *consumer.Consumer {
	pos := rand.Intn(len(c.consumers))

	c.consumerLock.Lock()
	defer c.consumerLock.Unlock()
	return c.consumers[pos]
}

func (c *ConnectionManager) NewFirehose() {
	log.Print("Creating New Firehose Connection")
	consumer := c.pick()
	msgs, errs := consumer.FirehoseWithoutReconnect(c.conf.SubscriptionId, c.conf.AuthToken)
	go func() {
		for range msgs {
		}
	}()
	for err := range errs {
		log.Printf("Error from %s: %v\n", c.conf.SubscriptionId, err.Error())
	}
}

func (c *ConnectionManager) NewStream() {
	log.Print("Creating New Stream Connection")
	consumer := c.pick()
	msgs, errs := consumer.StreamWithoutReconnect(c.conf.AppID, c.conf.AuthToken)
	go func() {
		for range msgs {
		}
	}()
	for err := range errs {
		log.Printf("Error from %s: %v\n", c.conf.AppID, err.Error())
	}
}

func (c *ConnectionManager) Close() {
	for _, consumer := range c.consumers {
		consumer.Close()
	}
}
