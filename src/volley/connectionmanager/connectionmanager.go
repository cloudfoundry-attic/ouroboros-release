package connectionmanager

import (
	"crypto/tls"
	"log"
	"math/rand"
	"sync"
	"time"
	"volley/config"

	"github.com/cloudfoundry/dropsonde/envelope_extensions"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
)

type AppIDStore interface {
	Add(appID string)
	Get() string
}

type ConnectionManager struct {
	consumers    []*consumer.Consumer
	consumerLock sync.Mutex
	conf         config.Config
	appStore     AppIDStore
}

func New(conf config.Config, appStore AppIDStore) *ConnectionManager {
	var consumers []*consumer.Consumer
	for _, tcAddress := range conf.TCAddresses {
		c := consumer.New(tcAddress, &tls.Config{InsecureSkipVerify: true}, nil)
		consumers = append(consumers, c)
	}

	return &ConnectionManager{
		conf:      conf,
		consumers: consumers,
		appStore:  appStore,
	}
}

func (c *ConnectionManager) pick() *consumer.Consumer {
	pos := rand.Intn(len(c.consumers))

	c.consumerLock.Lock()
	defer c.consumerLock.Unlock()
	return c.consumers[pos]
}

func (c *ConnectionManager) Firehose() {
	consumer := c.pick()
	msgs, errs := consumer.Firehose(c.conf.SubscriptionID, c.conf.AuthToken)
	go c.consume(msgs)
	for err := range errs {
		log.Printf("Error from %s: %v\n", c.conf.SubscriptionID, err.Error())
	}
}

func (c *ConnectionManager) Stream() {
	consumer := c.pick()
	appID := c.appStore.Get()
	msgs, errs := consumer.Stream(appID, c.conf.AuthToken)
	go c.consume(msgs)
	for err := range errs {
		log.Printf("Error from %s: %v\n", appID, err.Error())
	}
}

func (c *ConnectionManager) consume(msgs <-chan *events.Envelope) {
	delta := int(c.conf.ReceiveDelay.Max - c.conf.ReceiveDelay.Min)
	for msg := range msgs {
		appID := envelope_extensions.GetAppId(msg)
		if appID != "" {
			c.appStore.Add(appID)
		}
		if delta == 0 {
			continue
		}
		delay := c.conf.ReceiveDelay.Min + time.Duration(rand.Intn(delta))
		time.Sleep(delay)
	}
}

func (c *ConnectionManager) Close() {
	for _, consumer := range c.consumers {
		consumer.Close()
	}
}
