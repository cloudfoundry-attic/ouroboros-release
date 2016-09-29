package conns

import (
	"math/rand"
	"time"
)

type Reader interface {
	Read(buf []byte) (len int, err error)
}

type Ranger interface {
	DelayRange() (min, max time.Duration)
}

type MetricBatcher interface {
	BatchAddCounter(name string, delta uint64)
}

func Handle(reader Reader, ranger Ranger, batcher MetricBatcher) {
	min, max := ranger.DelayRange()
	delta := int(max - min)
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			return
		}
		batcher.BatchAddCounter("syslogr.receivedBytes", uint64(n))
		delay := min + time.Duration(rand.Intn(delta))
		time.Sleep(delay)
	}
}
