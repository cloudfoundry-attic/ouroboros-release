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
