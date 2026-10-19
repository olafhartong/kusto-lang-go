package kustolang

import (
	"sync/atomic"
	"testing"
)

func TestInterlockedCompareExchangeInt32ReturnsOriginalAndSwapsOnMatch(t *testing.T) {
	var value atomic.Int32
	value.Store(1)

	if got := InterlockedCompareExchangeInt32(&value, 2, 1); got != 1 {
		t.Fatalf("CompareExchange first result = %d, want original 1", got)
	}
	if got := value.Load(); got != 2 {
		t.Fatalf("value after matching compare-exchange = %d, want 2", got)
	}

	if got := InterlockedCompareExchangeInt32(&value, 3, 1); got != 2 {
		t.Fatalf("CompareExchange mismatch result = %d, want original 2", got)
	}
	if got := value.Load(); got != 2 {
		t.Fatalf("value after mismatched compare-exchange = %d, want unchanged 2", got)
	}
}

func TestInterlockedPointerHelpersUseReferenceIdentity(t *testing.T) {
	type box struct {
		value string
	}

	first := &box{value: "first"}
	second := &box{value: "second"}
	third := &box{value: "third"}

	var slot atomic.Pointer[box]
	if got := InterlockedCompareExchangePointer(&slot, first, nil); got != nil {
		t.Fatalf("initial pointer compare-exchange = %#v, want nil", got)
	}
	if got := slot.Load(); got != first {
		t.Fatalf("slot after initial compare-exchange = %#v, want first", got)
	}

	if got := InterlockedCompareExchangePointer(&slot, second, nil); got != first {
		t.Fatalf("mismatched pointer compare-exchange = %#v, want first", got)
	}
	if got := slot.Load(); got != first {
		t.Fatalf("slot after mismatched compare-exchange = %#v, want first", got)
	}

	if got := InterlockedExchangePointer(&slot, third); got != first {
		t.Fatalf("pointer exchange = %#v, want first", got)
	}
	if got := slot.Load(); got != third {
		t.Fatalf("slot after exchange = %#v, want third", got)
	}
}
