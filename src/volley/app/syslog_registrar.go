package app

import (
	"log"
	"time"
	"volley/drains"

	"github.com/coreos/etcd/client"
)

type IDGetter interface {
	Get() (id string)
}

type SyslogRegistrar struct {
	etcdAddrs  []string
	drainURLs  []string
	drainCount int
	ttl        time.Duration
	idGetter   IDGetter
}

// NewSyslogRegistrar creates a SyslogRegistrar which will write various syslog
// drain configuration details into etcd
func NewSyslogRegistrar(
	ttl time.Duration,
	drainCount int,
	drainURLs []string,
	etcdAddrs []string,
	idGetter IDGetter,
) *SyslogRegistrar {
	return &SyslogRegistrar{
		etcdAddrs:  etcdAddrs,
		drainURLs:  drainURLs,
		drainCount: drainCount,
		ttl:        ttl,
		idGetter:   idGetter,
	}
}

func (r *SyslogRegistrar) Start() {
	if len(r.etcdAddrs) == 0 {
		return
	}

	c := r.setupClient()
	for i := 0; i < r.drainCount; i++ {
		drains.AdvertiseRandom(r.idGetter, c, r.drainURLs, r.ttl)
	}
}

func (r *SyslogRegistrar) setupClient() client.KeysAPI {
	c, err := client.New(client.Config{
		Endpoints: r.etcdAddrs,
	})
	if err != nil {
		log.Panic(err)
	}

	return client.NewKeysAPI(c)
}
