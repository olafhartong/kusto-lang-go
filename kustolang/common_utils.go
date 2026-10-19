package kustolang

// Ported from src/Kusto.Language/Utils/DictionaryExtensions.cs,
// Ensure.cs, HashSetExtensions.cs, Interlocked.cs, and ListExtensions.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"sync/atomic"

	internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"
)

func DictionaryGetOrAdd[K comparable, V any](dictionary map[K]V, key K, value V) V {
	return internal.DictionaryGetOrAdd(dictionary, key, value)
}

func DictionaryGetOrAddFunc[K comparable, V any](dictionary map[K]V, key K, valueFactory func(K) V) V {
	return internal.DictionaryGetOrAddFunc(dictionary, key, valueFactory)
}

func DictionaryAddOrUpdate[K comparable, V any](dictionary map[K]V, key K, value V) {
	internal.DictionaryAddOrUpdate(dictionary, key, value)
}

func DictionaryAddOrUpdateFunc[K comparable, V any](dictionary map[K]V, key K, addFactory func(K) V, updateFactory func(K, V) V) {
	internal.DictionaryAddOrUpdateFunc(dictionary, key, addFactory, updateFactory)
}

func DictionaryAddOrUpdateValueFunc[K comparable, V any](dictionary map[K]V, key K, addFactory func(K, V) V, updateFactory func(K, V, V) V, value V) {
	internal.DictionaryAddOrUpdateValueFunc(dictionary, key, addFactory, updateFactory, value)
}

func EnsureIsTrue(value bool, message ...string) {
	internal.EnsureIsTrue(value, message...)
}

func EnsureAreEqual[T comparable](expected T, actual T, message ...string) {
	internal.EnsureAreEqual(expected, actual, message...)
}

func EnsureNotNil(value any, message ...string) {
	internal.EnsureNotNil(value, message...)
}

func EnsureIsNil(value any, message ...string) {
	internal.EnsureIsNil(value, message...)
}

func EnsureArgumentNotNil(value any, paramName string) {
	internal.EnsureArgumentNotNil(value, paramName)
}

func EnsureElementsNotNil[T any](list []T, listName string) {
	internal.EnsureElementsNotNil(list, listName)
}

func HashSetAddRange[T comparable](hashSet map[T]struct{}, items []T) {
	internal.HashSetAddRange(hashSet, items)
}

func InterlockedCompareExchangeInt32(value *atomic.Int32, newValue int32, comparand int32) int32 {
	return internal.InterlockedCompareExchangeInt32(value, newValue, comparand)
}

func InterlockedCompareExchangePointer[T any](value *atomic.Pointer[T], newValue *T, comparand *T) *T {
	return internal.InterlockedCompareExchangePointer(value, newValue, comparand)
}

func InterlockedExchangePointer[T any](value *atomic.Pointer[T], newValue *T) *T {
	return internal.InterlockedExchangePointer(value, newValue)
}

func ListToReadOnly[T any](items []T) []T {
	return internal.ListToReadOnly(items)
}

func ListToSafeList[T any](items []T) *SafeList[T] {
	return wrapSafeList(internal.ListToSafeList(items))
}

func ListToHashSetEx[T comparable](items []T) map[T]struct{} {
	return internal.ListToHashSetEx(items)
}

func ListToDictionaryLastValues[TKey comparable, TValue any](values []TValue, keySelector func(TValue) TKey) map[TKey]TValue {
	return internal.ListToDictionaryLastValues(values, keySelector)
}

func ListToDictionaryLast[TItem any, TKey comparable, TValue any](items []TItem, keySelector func(TItem) TKey, valueSelector func(TItem) TValue) map[TKey]TValue {
	return internal.ListToDictionaryLast(items, keySelector, valueSelector)
}

func ListAddOrUpdate[TItem any, TKey comparable](items []TItem, newOrUpdatedItems []TItem, keySelector func(TItem) TKey) []TItem {
	return internal.ListAddOrUpdate(items, newOrUpdatedItems, keySelector)
}

func ListDistinctFirst[TItem any, TKey comparable](items []TItem, keySelector func(TItem) TKey) []TItem {
	return internal.ListDistinctFirst(items, keySelector)
}

func ListDistinctLast[TItem any, TKey comparable](items []TItem, keySelector func(TItem) TKey) []TItem {
	return internal.ListDistinctLast(items, keySelector)
}

func ListBinarySearch[T any](items []T, comparer func(T) int) int {
	return internal.ListBinarySearch(items, comparer)
}

func ListIndexOf[T comparable](items []T, item T) int {
	return internal.ListIndexOf(items, item)
}

func ListFirstIndex[T any](items []T, predicate func(T) bool) int {
	return internal.ListFirstIndex(items, predicate)
}

func ListLastIndex[T any](items []T, predicate func(T) bool) int {
	return internal.ListLastIndex(items, predicate)
}

func ListAppendItem[T any](items []T, newItem T) []T {
	return internal.ListAppendItem(items, newItem)
}

func ListAppendItems[T any](items []T, newItems []T) []T {
	return internal.ListAppendItems(items, newItems)
}

func ListReplaceItem[T any](items []T, index int, newItem T) []T {
	return internal.ListReplaceItem(items, index, newItem)
}

func ListReplaceRangeItem[T any](items []T, start int, length int, newItem T) []T {
	return internal.ListReplaceRangeItem(items, start, length, newItem)
}

func ListReplaceRangeItems[T any](items []T, start int, length int, newItems []T) []T {
	return internal.ListReplaceRangeItems(items, start, length, newItems)
}

// NewCancellationTokenFunc creates a token that queries canceled whenever its
// cancellation state is observed.
func NewCancellationTokenFunc(canceled func() bool) CancellationToken {
	return internal.NewCancellationTokenFunc(canceled)
}
