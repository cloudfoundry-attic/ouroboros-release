package v1

import (
	"math/rand"
	"sync/atomic"
	"time"
	"unsafe"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type IDStore struct {
	ids      []unsafe.Pointer
	writeIDX int64
	filled   chan struct{}
}

func NewIDStore(len int) *IDStore {
	return &IDStore{
		ids:      make([]unsafe.Pointer, len),
		writeIDX: -1,
		filled:   make(chan struct{}),
	}
}

func (i *IDStore) Add(id string) {
	count := atomic.AddInt64(&i.writeIDX, 1)
	if count == int64(len(i.ids))-1 {
		close(i.filled)
	}
	idx := count % int64(len(i.ids))
	atomic.StorePointer(&i.ids[idx], unsafe.Pointer(&id))
}

func (i *IDStore) Get() string {
	<-i.filled
	idx := rand.Intn(len(i.ids))
	v := (*string)(atomic.LoadPointer(&i.ids[idx]))
	return *v
}

func (i *IDStore) GetN(n int) []string {
	wid := atomic.LoadInt64(&i.writeIDX)

	if wid >= int64(len(i.ids)) {
		wid = int64(len(i.ids) - 1)
	}
	if int64(n) > wid+1 {
		n = int(wid + 1)
	}

	ids := make([]string, n)
	for j := 0; j < n; j++ {
		ids[j] = *(*string)(atomic.LoadPointer(&i.ids[j]))
	}
	shuffle(ids)
	return ids
}

func shuffle(a []string) {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}
