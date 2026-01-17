package store

import (
	"sync"
	"time"

	"github.com/IsaacDSC/auditory/pkg/clock"
)

type MemIdempotency struct {
	ttl   time.Duration
	mu    sync.RWMutex
	store map[string]time.Time
}

func NewMemIdempotency(ttl time.Duration) *MemIdempotency {
	return &MemIdempotency{
		ttl:   ttl,
		store: make(map[string]time.Time),
	}
}

func (mi *MemIdempotency) Set(key string) {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	mi.store[key] = clock.Now().Add(mi.ttl)
}

func (mi *MemIdempotency) Get(key string) (time.Time, bool) {
	mi.mu.RLock()
	defer mi.mu.RUnlock()
	ttl, ok := mi.store[key]
	return ttl, ok
}

// reset expired keys every ttl duration
func (mi *MemIdempotency) Reset() {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	now := clock.Now()
	for key, expiresAt := range mi.store {
		// if current time is after the expiration time, delete the key
		if now.After(expiresAt) {
			delete(mi.store, key)
		}
	}
}
