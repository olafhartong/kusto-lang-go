package kustolang

import "sync/atomic"

// Ported from src/Kusto.Language/Utils/Interlocked.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func InterlockedCompareExchangeInt32(value *atomic.Int32, newValue int32, comparand int32) int32 {
	if value == nil {
		return 0
	}
	for {
		original := value.Load()
		if original != comparand {
			return original
		}
		if value.CompareAndSwap(comparand, newValue) {
			return original
		}
	}
}

func InterlockedCompareExchangePointer[T any](value *atomic.Pointer[T], newValue *T, comparand *T) *T {
	if value == nil {
		return nil
	}
	for {
		original := value.Load()
		if original != comparand {
			return original
		}
		if value.CompareAndSwap(comparand, newValue) {
			return original
		}
	}
}

func InterlockedExchangePointer[T any](value *atomic.Pointer[T], newValue *T) *T {
	if value == nil {
		return nil
	}
	return value.Swap(newValue)
}
