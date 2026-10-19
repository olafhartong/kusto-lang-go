package kustolang

import "sync/atomic"

// Ported from src/Kusto.Language/Utils/SafeList.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type SafeList[T any] struct {
	items   []T
	length  int
	isOwner atomic.Int32
}

func NewSafeList[T any](items []T) *SafeList[T] {
	copied := append([]T{}, items...)
	return newSafeList(copied, true)
}

func EmptySafeList[T any]() *SafeList[T] {
	return newSafeList([]T{}, false)
}

func newSafeList[T any](items []T, isOwner bool) *SafeList[T] {
	list := &SafeList[T]{
		items:  items,
		length: len(items),
	}
	if isOwner {
		list.isOwner.Store(1)
	}
	return list
}

func (l *SafeList[T]) Count() int {
	if l == nil {
		return 0
	}
	return l.length
}

func (l *SafeList[T]) At(index int) T {
	if l == nil || index < 0 || index >= l.length {
		panic("index out of range")
	}
	return l.items[index]
}

func (l *SafeList[T]) IsOwner() bool {
	return l != nil && l.isOwner.Load() != 0
}

func (l *SafeList[T]) AddItem(item T) *SafeList[T] {
	if l == nil {
		return NewSafeList([]T{item})
	}
	if InterlockedCompareExchangeInt32(&l.isOwner, 0, 1) == 1 {
		items := append(l.items, item)
		return newSafeList(items, true)
	}

	items := make([]T, l.length, l.length+1)
	copy(items, l.items[:l.length])
	items = append(items, item)
	return newSafeList(items, true)
}

func (l *SafeList[T]) AddItems(items []T) *SafeList[T] {
	if l == nil {
		return NewSafeList(items)
	}
	if InterlockedCompareExchangeInt32(&l.isOwner, 0, 1) == 1 {
		newItems := append(l.items, items...)
		return newSafeList(newItems, true)
	}

	newItems := make([]T, l.length, l.length+len(items))
	copy(newItems, l.items[:l.length])
	newItems = append(newItems, items...)
	return newSafeList(newItems, true)
}

func (l *SafeList[T]) Values() []T {
	if l == nil {
		return nil
	}
	return append([]T{}, l.items[:l.length]...)
}
