package drains

import (
	"crypto/sha1"
	"math/rand"
	"path"
	"time"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type IDGetter interface {
	Get() (id string)
}

type ETCDSetter interface {
	Set(ctx context.Context, key, value string, opts *client.SetOptions) (*client.Response, error)
}

// AdvertiseRandom advertises a random drain URL for the first app ID
// returned from ids.
func AdvertiseRandom(ids IDGetter, etcd ETCDSetter, drains []string, ttl time.Duration) {
	drain := drains[rand.Intn(len(drains))]
	drainHash := sha1.Sum([]byte(drain))
	id := ids.Get()
	key := path.Join("/loggregator", "services", id, string(drainHash[:]))
	etcd.Set(context.Background(), key, drain, &client.SetOptions{TTL: ttl})
}
