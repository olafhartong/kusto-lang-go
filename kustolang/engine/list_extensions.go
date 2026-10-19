package kustolang

import (
	"reflect"
	"strings"
)

// Ported from src/Kusto.Language/Utils/ListExtensions.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func ListToReadOnly[T any](items []T) []T {
	if len(items) == 0 {
		return EmptyReadOnlyList[T]()
	}
	return append([]T{}, items...)
}

func ListToSafeList[T any](items []T) *SafeList[T] {
	if len(items) == 0 {
		return EmptySafeList[T]()
	}
	return NewSafeList(items)
}

func ListToHashSetEx[T comparable](items []T) map[T]struct{} {
	set := map[T]struct{}{}
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}

func ListToDictionaryLastValues[TKey comparable, TValue any](values []TValue, keySelector func(TValue) TKey) map[TKey]TValue {
	return ListToDictionaryLast(values, keySelector, func(value TValue) TValue {
		return value
	})
}

func ListToDictionaryLast[TItem any, TKey comparable, TValue any](items []TItem, keySelector func(TItem) TKey, valueSelector func(TItem) TValue) map[TKey]TValue {
	result := map[TKey]TValue{}
	for _, item := range items {
		result[keySelector(item)] = valueSelector(item)
	}
	return result
}

func ListAddOrUpdate[TItem any, TKey comparable](items []TItem, newOrUpdatedItems []TItem, keySelector func(TItem) TKey) []TItem {
	if len(newOrUpdatedItems) == 0 {
		return items
	}
	if len(newOrUpdatedItems) == 1 {
		item := newOrUpdatedItems[0]
		if listItemIsNil(item) {
			return items
		}
		key := keySelector(item)
		result := make([]TItem, 0, len(items)+1)
		for _, existing := range items {
			if keySelector(existing) != key {
				result = append(result, existing)
			}
		}
		result = append(result, item)
		return ListToReadOnly(result)
	}

	combined := make([]TItem, 0, len(items)+len(newOrUpdatedItems))
	for _, item := range items {
		if !listItemIsNil(item) {
			combined = append(combined, item)
		}
	}
	for _, item := range newOrUpdatedItems {
		if !listItemIsNil(item) {
			combined = append(combined, item)
		}
	}
	return ListToReadOnly(ListDistinctLast(combined, keySelector))
}

func ListDistinctFirst[TItem any, TKey comparable](items []TItem, keySelector func(TItem) TKey) []TItem {
	seen := map[TKey]struct{}{}
	result := make([]TItem, 0, len(items))
	for _, item := range items {
		key := keySelector(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result
}

func ListDistinctLast[TItem any, TKey comparable](items []TItem, keySelector func(TItem) TKey) []TItem {
	seen := map[TKey]struct{}{}
	result := make([]TItem, 0, len(items))
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		key := keySelector(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func ListBinarySearch[T any](items []T, comparer func(T) int) int {
	left := 0
	right := len(items) - 1
	for left <= right {
		mid := (left + right) / 2
		c := comparer(items[mid])
		if c == 0 {
			return mid
		}
		if c > 0 {
			right = mid - 1
		} else {
			left = mid + 1
		}
	}
	return -1
}

func ListIndexOf[T comparable](items []T, item T) int {
	for i, value := range items {
		if value == item {
			return i
		}
	}
	return -1
}

func ListRemoveAll[T any](items []T, selector func(T) bool) []T {
	result := items[:0]
	for _, item := range items {
		if !selector(item) {
			result = append(result, item)
		}
	}
	return result
}

func ListSetCount[T any](items []T, count int) []T {
	if len(items) > count {
		return items[:count]
	}
	return items
}

func ListJoin(items []string, separator string, finalSeparator ...string) string {
	lastSeparator := ""
	if len(finalSeparator) > 0 {
		lastSeparator = finalSeparator[0]
	}
	var builder strings.Builder
	for i, item := range items {
		if i > 0 {
			if i < len(items)-1 {
				builder.WriteString(separator)
			} else {
				builder.WriteString(lastSeparator)
			}
		}
		builder.WriteString(item)
	}
	return builder.String()
}

func DictionaryIsEquivalentTo[K comparable, V comparable](left map[K]V, right map[K]V) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		rightValue, ok := right[key]
		if !ok || rightValue != value {
			return false
		}
	}
	return true
}

func ListFirstIndex[T any](items []T, predicate func(T) bool) int {
	for i, item := range items {
		if predicate(item) {
			return i
		}
	}
	return -1
}

func ListLastIndex[T any](items []T, predicate func(T) bool) int {
	for i := len(items) - 1; i >= 0; i-- {
		if predicate(items[i]) {
			return i
		}
	}
	return -1
}

func ListAppendItem[T any](items []T, newItem T) []T {
	return ListReplaceRangeItem(items, len(items), 0, newItem)
}

func ListAppendItems[T any](items []T, newItems []T) []T {
	return ListReplaceRangeItems(items, len(items), 0, newItems)
}

func ListReplaceItem[T any](items []T, index int, newItem T) []T {
	return ListReplaceRangeItem(items, index, 1, newItem)
}

func ListReplaceRangeItem[T any](items []T, start int, length int, newItem T) []T {
	result := append([]T{}, items...)
	result = listRemoveRange(result, start, length)
	if !listItemIsNil(newItem) && start >= 0 && start <= len(result) {
		var zero T
		result = append(result, zero)
		copy(result[start+1:], result[start:])
		result[start] = newItem
	}
	return result
}

func ListReplaceRangeItems[T any](items []T, start int, length int, newItems []T) []T {
	result := append([]T{}, items...)
	result = listRemoveRange(result, start, length)
	if len(newItems) > 0 && start >= 0 && start <= len(result) {
		result = append(result, newItems...)
		copy(result[start+len(newItems):], result[start:len(result)-len(newItems)])
		copy(result[start:], newItems)
	}
	return result
}

func ReadOnlyListEquals[T comparable](left []T, right []T) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func ReadOnlyListHashCode[T any](items []T, hash func(T) int) int {
	if items == nil {
		return 0
	}
	if len(items) == 1 {
		return hash(items[0])
	}
	value := 0
	for _, item := range items {
		value += hash(item)
	}
	return value
}

func listRemoveRange[T any](items []T, start int, length int) []T {
	if length > 0 && start >= 0 && start < len(items) {
		removed := len(items) - start
		if length < removed {
			removed = length
		}
		copy(items[start:], items[start+removed:])
		var zero T
		for i := len(items) - removed; i < len(items); i++ {
			items[i] = zero
		}
		items = items[:len(items)-removed]
	}
	return items
}

func listItemIsNil[T any](item T) bool {
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
