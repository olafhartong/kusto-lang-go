package kustolang

import "sync"

// Ported from src/Kusto.Language/Utils/MostRecentlyUsedCache.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type equaler[T any] interface {
	Equals(T) bool
}

type MostRecentlyUsedCache[K equaler[K], V any] struct {
	mu      sync.Mutex
	pairs   []mruPair[K, V]
	maxSize int
}

type mruPair[K equaler[K], V any] struct {
	key   K
	value V
}

func NewMostRecentlyUsedCache[K equaler[K], V any](maxSize ...int) *MostRecentlyUsedCache[K, V] {
	size := 10
	if len(maxSize) > 0 {
		size = maxSize[0]
	}
	return &MostRecentlyUsedCache[K, V]{maxSize: size}
}

func (c *MostRecentlyUsedCache[K, V]) Count() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.pairs)
}

func (c *MostRecentlyUsedCache[K, V]) TryGetValue(key K) (V, bool) {
	var zero V
	if c == nil {
		return zero, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, pair := range c.pairs {
		if pair.key.Equals(key) {
			if i > 0 {
				c.pairs = append(c.pairs[:i], c.pairs[i+1:]...)
				c.pairs = append([]mruPair[K, V]{pair}, c.pairs...)
			}
			return pair.value, true
		}
	}
	return zero, false
}

func (c *MostRecentlyUsedCache[K, V]) AddOrUpdate(key K, value V) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, pair := range c.pairs {
		if pair.key.Equals(key) {
			// Upstream keeps the existing value while refreshing recency.
			c.pairs = append(c.pairs[:i], c.pairs[i+1:]...)
			c.pairs = append([]mruPair[K, V]{{key: key, value: pair.value}}, c.pairs...)
			return
		}
	}
	c.pairs = append([]mruPair[K, V]{{key: key, value: value}}, c.pairs...)
	for c.maxSize >= 0 && len(c.pairs) > c.maxSize {
		c.pairs = c.pairs[:len(c.pairs)-1]
	}
}
