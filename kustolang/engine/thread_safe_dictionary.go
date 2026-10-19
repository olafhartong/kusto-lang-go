package kustolang

import "sync"

// Ported from src/Kusto.Language/Utils/ThreadSafeDictionary.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type ThreadSafeDictionary[K comparable, V any] struct {
	mu     sync.RWMutex
	values map[K]V
}

func NewThreadSafeDictionary[K comparable, V any]() *ThreadSafeDictionary[K, V] {
	return &ThreadSafeDictionary[K, V]{values: map[K]V{}}
}

func (d *ThreadSafeDictionary[K, V]) Count() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.values)
}

func (d *ThreadSafeDictionary[K, V]) Keys() []K {
	d.mu.RLock()
	defer d.mu.RUnlock()
	keys := make([]K, 0, len(d.values))
	for key := range d.values {
		keys = append(keys, key)
	}
	return keys
}

func (d *ThreadSafeDictionary[K, V]) Values() []V {
	d.mu.RLock()
	defer d.mu.RUnlock()
	values := make([]V, 0, len(d.values))
	for _, value := range d.values {
		values = append(values, value)
	}
	return values
}

func (d *ThreadSafeDictionary[K, V]) TryGetValue(key K) (V, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	value, ok := d.values[key]
	return value, ok
}

func (d *ThreadSafeDictionary[K, V]) TryAdd(key K, value V) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.values[key]; ok {
		return false
	}
	d.values[key] = value
	return true
}

func (d *ThreadSafeDictionary[K, V]) GetOrAdd(key K, value V) V {
	d.mu.Lock()
	defer d.mu.Unlock()
	return DictionaryGetOrAdd(d.values, key, value)
}

func (d *ThreadSafeDictionary[K, V]) GetOrAddFunc(key K, valueFactory func(K) V) V {
	d.mu.Lock()
	defer d.mu.Unlock()
	return DictionaryGetOrAddFunc(d.values, key, valueFactory)
}

func (d *ThreadSafeDictionary[K, V]) AddOrUpdate(key K, value V) {
	d.mu.Lock()
	defer d.mu.Unlock()
	DictionaryAddOrUpdate(d.values, key, value)
}

func (d *ThreadSafeDictionary[K, V]) AddOrUpdateFunc(key K, addFactory func(K) V, updateFactory func(K, V) V) {
	d.mu.Lock()
	defer d.mu.Unlock()
	DictionaryAddOrUpdateFunc(d.values, key, addFactory, updateFactory)
}

func (d *ThreadSafeDictionary[K, V]) AddOrUpdateValueFunc(key K, addFactory func(K, V) V, updateFactory func(K, V, V) V, value V) {
	d.mu.Lock()
	defer d.mu.Unlock()
	DictionaryAddOrUpdateValueFunc(d.values, key, addFactory, updateFactory, value)
}
