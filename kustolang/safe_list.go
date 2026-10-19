package kustolang

// Ported from src/Kusto.Language/Utils/SafeList.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// SafeList is an immutable snapshot over an append-only list. Adding items
// returns a new snapshot and never changes the visible contents of an older
// snapshot.
type SafeList[T any] struct {
	value *internal.SafeList[T]
}

// NewSafeList creates a snapshot containing a defensive copy of items.
func NewSafeList[T any](items []T) *SafeList[T] {
	return wrapSafeList(internal.NewSafeList(items))
}

// EmptySafeList creates the source-compatible non-owning empty snapshot.
func EmptySafeList[T any]() *SafeList[T] {
	return wrapSafeList(internal.EmptySafeList[T]())
}

func wrapSafeList[T any](value *internal.SafeList[T]) *SafeList[T] {
	return &SafeList[T]{value: value}
}

func (l *SafeList[T]) internalValue() *internal.SafeList[T] {
	if l == nil {
		return nil
	}
	return l.value
}

// Count returns the number of elements visible in this snapshot.
func (l *SafeList[T]) Count() int {
	return l.internalValue().Count()
}

// At returns the element at index and panics when index is outside the
// snapshot, matching the source indexer.
func (l *SafeList[T]) At(index int) T {
	return l.internalValue().At(index)
}

// AddItem returns a new snapshot with item appended.
func (l *SafeList[T]) AddItem(item T) *SafeList[T] {
	return wrapSafeList(l.internalValue().AddItem(item))
}

// AddItems returns a new snapshot with items appended in order.
func (l *SafeList[T]) AddItems(items []T) *SafeList[T] {
	return wrapSafeList(l.internalValue().AddItems(items))
}

// Values returns a defensive copy of the elements visible in this snapshot.
func (l *SafeList[T]) Values() []T {
	return l.internalValue().Values()
}
