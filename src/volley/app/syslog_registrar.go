package app

import (
	"crypto/sha1"
	"log"
	"math/rand"
	"path"
	"time"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
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
		AdvertiseRandom(r.idGetter, c, r.drainURLs, r.ttl)
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

type ETCDSetter interface {
	Set(ctx context.Context, key, value string, opts *client.SetOptions) (*client.Response, error)
}

// AdvertiseRandom advertises a random drain URL for the first app ID returned
// from ids.
func AdvertiseRandom(ids IDGetter, etcd ETCDSetter, drainURLs []string, ttl time.Duration) {
	drain := drainURLs[rand.Intn(len(drainURLs))]
	drainHash := sha1.Sum([]byte(drain))
	id := ids.Get()
	key := path.Join("/loggregator", "services", id, string(drainHash[:]))
	_, err := etcd.Set(context.Background(), key, drain, &client.SetOptions{TTL: ttl})
	if err != nil {
		log.Printf("etcd failed: %s", err)
	}
}
