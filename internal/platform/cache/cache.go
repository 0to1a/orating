// Package cache is a typed in-process cache with optional per-entry TTL.
// Two modes: Replace(map) for atomic snapshot swaps, SetWithTTL for
// individual entries. Janitor goroutine starts on New(ctx).
package cache

import (
	"context"
	"sync"
	"time"
)

const JanitorInterval = 5 * time.Minute

type item[V any] struct {
	value     V
	expiresAt time.Time // zero = no expiry
}

func (i item[V]) expired(now time.Time) bool {
	return !i.expiresAt.IsZero() && now.After(i.expiresAt)
}

type Cache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]item[V]
}

// New starts the janitor goroutine. Cancel ctx to stop it.
func New[K comparable, V any](ctx context.Context) *Cache[K, V] {
	c := &Cache[K, V]{items: map[K]item[V]{}}
	go c.runJanitor(ctx)
	return c
}

// Set stores v with no expiry.
func (c *Cache[K, V]) Set(k K, v V) {
	c.mu.Lock()
	c.items[k] = item[V]{value: v}
	c.mu.Unlock()
}

// SetWithTTL stores v with the given TTL. ttl <= 0 means no expiry.
func (c *Cache[K, V]) SetWithTTL(k K, v V, ttl time.Duration) {
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	c.mu.Lock()
	c.items[k] = item[V]{value: v, expiresAt: exp}
	c.mu.Unlock()
}

func (c *Cache[K, V]) Get(k K) (V, bool) {
	c.mu.RLock()
	it, ok := c.items[k]
	c.mu.RUnlock()
	var zero V
	if !ok {
		return zero, false
	}
	if it.expired(time.Now()) {
		return zero, false
	}
	return it.value, true
}

func (c *Cache[K, V]) Delete(k K) {
	c.mu.Lock()
	delete(c.items, k)
	c.mu.Unlock()
}

// Replace atomically swaps the entire dataset. Used for snapshot-mode refresh.
func (c *Cache[K, V]) Replace(items map[K]V) {
	next := make(map[K]item[V], len(items))
	for k, v := range items {
		next[k] = item[V]{value: v}
	}
	c.mu.Lock()
	c.items = next
	c.mu.Unlock()
}

func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// DeleteExpired removes expired entries and returns the count.
func (c *Cache[K, V]) DeleteExpired() int {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	removed := 0
	for k, it := range c.items {
		if it.expired(now) {
			delete(c.items, k)
			removed++
		}
	}
	return removed
}

func (c *Cache[K, V]) runJanitor(ctx context.Context) {
	t := time.NewTicker(JanitorInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.DeleteExpired()
		}
	}
}
