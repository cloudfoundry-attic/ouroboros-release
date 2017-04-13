package drains

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
