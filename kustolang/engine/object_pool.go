package kustolang

import (
	"reflect"
	"sync"
)

// Ported from src/Kusto.Language/Utils/ObjectPool.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type ObjectPool[T any] struct {
	creator  func() T
	resetter func(T)
	items    []T
	used     []bool
	mu       sync.Mutex
}

func NewObjectPool[T any](creator func() T, resetter func(T), size ...int) *ObjectPool[T] {
	poolSize := 10
	if len(size) > 0 {
		poolSize = size[0]
	}
	return &ObjectPool[T]{
		creator:  creator,
		resetter: resetter,
		items:    make([]T, poolSize),
		used:     make([]bool, poolSize),
	}
}

func (p *ObjectPool[T]) AllocateFromPool() T {
	p.mu.Lock()
	for i, isUsed := range p.used {
		if isUsed {
			item := p.items[i]
			var zero T
			p.items[i] = zero
			p.used[i] = false
			p.mu.Unlock()
			return item
		}
	}
	p.mu.Unlock()
	return p.creator()
}

func (p *ObjectPool[T]) ReturnToPool(item T) {
	p.resetter(item)
	if objectPoolValueIsNil(item) {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	for i, isUsed := range p.used {
		if !isUsed {
			p.items[i] = item
			p.used[i] = true
			break
		}
	}
}

func objectPoolValueIsNil[T any](item T) bool {
	value := reflect.ValueOf(item)
	if !value.IsValid() {
		return true
	}
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
