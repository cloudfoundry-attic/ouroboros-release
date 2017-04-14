package conns

import (
	"math/rand"
	"time"

	"github.com/cloudfoundry/dropsonde/metricbatcher"
)

type Reader interface {
	Read(buf []byte) (len int, err error)
}

type Ranger interface {
	DelayRange() (min, max time.Duration)
}

type MetricBatcher interface {
	BatchCounter(name string) metricbatcher.BatchCounterChainer
}

func Handle(reader Reader, ranger Ranger, batcher MetricBatcher) {
	batcher.BatchCounter("handleConn").
		SetTag("protocol", "syslog").
		Increment()
	min, max := ranger.DelayRange()
	delta := int(max - min)
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			return
		}
		batcher.BatchCounter("receivedBytes").
			SetTag("protocol", "syslog").
			Add(uint64(n))
		delay := min + time.Duration(rand.Intn(delta))
		time.Sleep(delay)
	}
}
