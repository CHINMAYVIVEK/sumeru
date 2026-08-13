// Package cache provides a simple in-process TTL cache for the engine.
package cache

import (
	"sync"
	"time"
)

type entry struct {
	val       interface{}
	expiresAt time.Time
}

var (
	mu    sync.RWMutex
	store = map[string]entry{}
)

// Get returns a cached value if present and not expired.
func Get(key string) (interface{}, bool) {
	mu.RLock()
	e, ok := store[key]
	mu.RUnlock()
	if !ok {
		return nil, false
	}
	if !e.expiresAt.IsZero() && time.Now().After(e.expiresAt) {
		mu.Lock()
		delete(store, key)
		mu.Unlock()
		return nil, false
	}
	return e.val, true
}

// Set stores val with optional TTL (0 = no expiry).
func Set(key string, val interface{}, ttl time.Duration) {
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	mu.Lock()
	store[key] = entry{val: val, expiresAt: exp}
	mu.Unlock()
}

// Delete removes a key.
func Delete(key string) {
	mu.Lock()
	delete(store, key)
	mu.Unlock()
}

// Clear empties the cache.
func Clear() {
	mu.Lock()
	store = map[string]entry{}
	mu.Unlock()
}

// DeletePrefix removes all keys with the given prefix.
func DeletePrefix(prefix string) {
	if prefix == "" {
		return
	}
	mu.Lock()
	for k := range store {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(store, k)
		}
	}
	mu.Unlock()
}
