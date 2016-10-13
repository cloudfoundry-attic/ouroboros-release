package connectionmanager

import (
	"crypto/tls"
	"log"
	"math/rand"
	"sync"
	"time"
	"volley/config"

	"github.com/cloudfoundry/dropsonde/envelope_extensions"
	"github.com/cloudfoundry/dropsonde/metricbatcher"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
)

var (
	openConnectionsMetric   = "openConnections"
	closedConnectionsMetric = "closedConnections"
)

type AppIDStore interface {
	Add(appID string)
	Get() string
}

type Batcher interface {
	BatchCounter(name string) metricbatcher.BatchCounterChainer
}

type ConnectionManager struct {
	consumers    []*consumer.Consumer
	consumerLock sync.Mutex
	conf         config.Config
	appStore     AppIDStore
	batcher      Batcher
}

func New(conf config.Config, appStore AppIDStore, batcher Batcher) *ConnectionManager {
	var consumers []*consumer.Consumer
	for _, tcAddress := range conf.TCAddresses {
		c := consumer.New(tcAddress, &tls.Config{InsecureSkipVerify: true}, nil)
		consumers = append(consumers, c)
	}

	return &ConnectionManager{
		conf:      conf,
		consumers: consumers,
		appStore:  appStore,
		batcher:   batcher,
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
	c.batcher.BatchCounter("volley.openConnections").SetTag("conn_type", "firehose").Increment()
	go c.consume(msgs, "firehose")
	for err := range errs {
		c.batcher.BatchCounter("volley.closedConnections").SetTag("conn_type", "firehose").Increment()
		log.Printf("Error from %s: %v\n", c.conf.SubscriptionID, err.Error())
	}
}

func (c *ConnectionManager) Stream() {
	consumer := c.pick()
	appID := c.appStore.Get()
	msgs, errs := consumer.Stream(appID, c.conf.AuthToken)
	c.batcher.BatchCounter("volley.openConnections").SetTag("conn_type", "stream").Increment()
	go c.consume(msgs, "stream")
	for err := range errs {
		c.batcher.BatchCounter("volley.closedConnections").SetTag("conn_type", "stream").Increment()
		log.Printf("Error from %s: %v\n", appID, err.Error())
	}
}

func (c *ConnectionManager) RecentLogs() {
	consumer := c.pick()
	appID := c.appStore.Get()
	_, err := consumer.RecentLogs(appID, c.conf.AuthToken)
	if err != nil {
		c.batcher.BatchCounter("volley.numberOfRequestErrors").SetTag("conn_type", "recentlogs").Increment()
		log.Printf("Error from %s: %v\n", appID, err.Error())
		return
	}
	c.batcher.BatchCounter("volley.numberOfRequests").SetTag("conn_type", "recentlogs").Increment()
}

func (c *ConnectionManager) ContainerMetrics() {
	consumer := c.pick()
	appID := c.appStore.Get()
	_, err := consumer.ContainerMetrics(appID, c.conf.AuthToken)
	if err != nil {
		c.batcher.BatchCounter("volley.numberOfRequestErrors").SetTag("conn_type", "containermetrics").Increment()
		log.Printf("Error from %s: %v\n", appID, err.Error())
		return
	}
	c.batcher.BatchCounter("volley.numberOfRequests").SetTag("conn_type", "containermetrics").Increment()
}

func (c *ConnectionManager) consume(msgs <-chan *events.Envelope, connType string) {
	delta := int(c.conf.ReceiveDelay.Max - c.conf.ReceiveDelay.Min)
	var count int
	for msg := range msgs {
		count++
		if count%1000 == 0 {
			c.batcher.BatchCounter("volley.receivedEnvelopes").SetTag("conn_type", connType).Add(1000)
		}
		appID := envelope_extensions.GetAppId(msg)
		if appID != "" && appID != envelope_extensions.SystemAppId {
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
