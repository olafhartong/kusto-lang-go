package kustolang

import (
	"reflect"
	"sync"
)

// Ported from src/Kusto.Language/KustoCache.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type KustoCache struct {
	Globals *GlobalState
	mu      sync.RWMutex
	values  map[reflect.Type]any
}

func NewKustoCache(globals *GlobalState) *KustoCache {
	return &KustoCache{Globals: globals, values: map[reflect.Type]any{}}
}

func (c *KustoCache) WithGlobals(globals *GlobalState) *KustoCache {
	if c == nil {
		return nil
	}
	if c.Globals == globals {
		return c
	}
	return NewKustoCache(globals)
}

func KustoCacheGetOrCreate[T any](cache *KustoCache, creator func() T) T {
	var zero T
	if cache == nil {
		return zero
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	key := reflect.TypeOf((*T)(nil)).Elem()
	value := DictionaryGetOrAddFunc(cache.values, key, func(reflect.Type) any {
		return creator()
	})
	return value.(T)
}

func KustoCacheTryGetValue[T any](cache *KustoCache) (T, bool) {
	var zero T
	if cache == nil {
		return zero, false
	}
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	key := reflect.TypeOf((*T)(nil)).Elem()
	if value, ok := cache.values[key]; ok {
		return value.(T), true
	}
	return zero, false
}
