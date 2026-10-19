package kustolang_test

import (
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicDictionaryExtensions(t *testing.T) {
	dictionary := map[string]int{"existing": 1}

	if got := kustolang.DictionaryGetOrAdd(dictionary, "existing", 2); got != 1 {
		t.Fatalf("DictionaryGetOrAdd existing = %d, want 1", got)
	}
	if got := kustolang.DictionaryGetOrAdd(dictionary, "added", 2); got != 2 {
		t.Fatalf("DictionaryGetOrAdd added = %d, want 2", got)
	}

	factoryCalls := 0
	if got := kustolang.DictionaryGetOrAddFunc(dictionary, "existing", func(string) int {
		factoryCalls++
		return 3
	}); got != 1 || factoryCalls != 0 {
		t.Fatalf("DictionaryGetOrAddFunc existing = %d, calls = %d; want 1, 0", got, factoryCalls)
	}

	kustolang.DictionaryAddOrUpdate(dictionary, "added", 4)
	kustolang.DictionaryAddOrUpdateFunc(dictionary, "factory", func(key string) int {
		return len(key)
	}, func(key string, old int) int {
		return old + len(key)
	})
	kustolang.DictionaryAddOrUpdateFunc(dictionary, "factory", func(string) int {
		return -1
	}, func(key string, old int) int {
		return old + len(key)
	})
	kustolang.DictionaryAddOrUpdateValueFunc(dictionary, "value", func(_ string, value int) int {
		return value + 1
	}, func(_ string, old, value int) int {
		return old + value
	}, 5)
	kustolang.DictionaryAddOrUpdateValueFunc(dictionary, "value", func(_ string, value int) int {
		return value
	}, func(_ string, old, value int) int {
		return old + value
	}, 5)

	if dictionary["added"] != 4 || dictionary["factory"] != 14 || dictionary["value"] != 11 {
		t.Fatalf("dictionary after updates = %#v", dictionary)
	}
}

func TestPublicEnsureAndHashSetExtensions(t *testing.T) {
	value := 1
	kustolang.EnsureIsTrue(true)
	kustolang.EnsureAreEqual("same", "same")
	kustolang.EnsureNotNil(&value)
	kustolang.EnsureIsNil((*int)(nil))
	kustolang.EnsureArgumentNotNil(&value, "value")
	kustolang.EnsureElementsNotNil([]*int{&value}, "values")

	assertPanicContains(t, "custom", func() {
		kustolang.EnsureIsTrue(false, "custom")
	})
	assertPanicContains(t, "Expected: 1 actual: 2", func() {
		kustolang.EnsureAreEqual(1, 2)
	})
	assertPanicContains(t, "Expected not null", func() {
		kustolang.EnsureNotNil((*int)(nil))
	})
	assertPanicContains(t, "Expected null", func() {
		kustolang.EnsureIsNil(&value)
	})
	assertPanicContains(t, "argument \"value\" is nil", func() {
		kustolang.EnsureArgumentNotNil((*int)(nil), "value")
	})
	assertPanicContains(t, "values[1] is nil", func() {
		kustolang.EnsureElementsNotNil([]*int{&value, nil}, "values")
	})

	set := map[string]struct{}{"a": {}}
	kustolang.HashSetAddRange(set, []string{"b", "a", "c"})
	if len(set) != 3 {
		t.Fatalf("HashSetAddRange size = %d, want 3", len(set))
	}
}

func TestPublicInterlockedExtensions(t *testing.T) {
	var integer atomic.Int32
	integer.Store(1)
	if got := kustolang.InterlockedCompareExchangeInt32(&integer, 2, 1); got != 1 || integer.Load() != 2 {
		t.Fatalf("matching int compare-exchange returned %d and stored %d", got, integer.Load())
	}
	if got := kustolang.InterlockedCompareExchangeInt32(&integer, 3, 1); got != 2 || integer.Load() != 2 {
		t.Fatalf("mismatched int compare-exchange returned %d and stored %d", got, integer.Load())
	}

	type box struct{ value string }
	first := &box{value: "first"}
	second := &box{value: "second"}
	third := &box{value: "third"}
	var pointer atomic.Pointer[box]
	if got := kustolang.InterlockedCompareExchangePointer(&pointer, first, nil); got != nil || pointer.Load() != first {
		t.Fatalf("initial pointer compare-exchange returned %#v and stored %#v", got, pointer.Load())
	}
	if got := kustolang.InterlockedCompareExchangePointer(&pointer, second, nil); got != first || pointer.Load() != first {
		t.Fatalf("mismatched pointer compare-exchange returned %#v and stored %#v", got, pointer.Load())
	}
	if got := kustolang.InterlockedExchangePointer(&pointer, third); got != first || pointer.Load() != third {
		t.Fatalf("pointer exchange returned %#v and stored %#v", got, pointer.Load())
	}
}

type publicListItem struct {
	key   string
	value int
}

func TestPublicListConversionAndDistinctExtensions(t *testing.T) {
	input := []string{"a", "b"}
	readOnly := kustolang.ListToReadOnly(input)
	safe := kustolang.ListToSafeList(input)
	input[0] = "changed"
	if !reflect.DeepEqual(readOnly, []string{"a", "b"}) {
		t.Fatalf("ListToReadOnly = %#v", readOnly)
	}
	if got := safe.Values(); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("ListToSafeList = %#v", got)
	}

	set := kustolang.ListToHashSetEx([]string{"a", "b", "a"})
	if len(set) != 2 {
		t.Fatalf("ListToHashSetEx size = %d, want 2", len(set))
	}

	items := []publicListItem{{key: "a", value: 1}, {key: "b", value: 2}, {key: "a", value: 3}}
	identity := kustolang.ListToDictionaryLastValues(items, func(item publicListItem) string { return item.key })
	values := kustolang.ListToDictionaryLast(items,
		func(item publicListItem) string { return item.key },
		func(item publicListItem) int { return item.value },
	)
	if identity["a"].value != 3 || values["a"] != 3 || values["b"] != 2 {
		t.Fatalf("last dictionaries = %#v, %#v", identity, values)
	}

	first := kustolang.ListDistinctFirst(items, func(item publicListItem) string { return item.key })
	last := kustolang.ListDistinctLast(items, func(item publicListItem) string { return item.key })
	if got := listValues(first); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("ListDistinctFirst values = %#v", got)
	}
	if got := listValues(last); !reflect.DeepEqual(got, []int{2, 3}) {
		t.Fatalf("ListDistinctLast values = %#v", got)
	}

	updated := kustolang.ListAddOrUpdate(items[:2], []publicListItem{{key: "a", value: 4}}, func(item publicListItem) string { return item.key })
	if got := listValues(updated); !reflect.DeepEqual(got, []int{2, 4}) {
		t.Fatalf("ListAddOrUpdate values = %#v", got)
	}
}

func TestPublicListSearchAndReplacementExtensions(t *testing.T) {
	items := []int{1, 3, 5, 7}
	if got := kustolang.ListBinarySearch(items, func(value int) int {
		return value - 5
	}); got != 2 {
		t.Fatalf("ListBinarySearch = %d, want 2", got)
	}
	if got := kustolang.ListIndexOf(items, 3); got != 1 {
		t.Fatalf("ListIndexOf = %d, want 1", got)
	}
	if got := kustolang.ListFirstIndex(items, func(value int) bool { return value > 3 }); got != 2 {
		t.Fatalf("ListFirstIndex = %d, want 2", got)
	}
	if got := kustolang.ListLastIndex(items, func(value int) bool { return value < 7 }); got != 2 {
		t.Fatalf("ListLastIndex = %d, want 2", got)
	}

	assertSliceEqual(t, "ListAppendItem", kustolang.ListAppendItem([]string{"a"}, "b"), []string{"a", "b"})
	assertSliceEqual(t, "ListAppendItems", kustolang.ListAppendItems([]string{"a"}, []string{"b", "c"}), []string{"a", "b", "c"})
	assertSliceEqual(t, "ListReplaceItem", kustolang.ListReplaceItem([]string{"a", "b"}, 1, "B"), []string{"a", "B"})
	assertSliceEqual(t, "ListReplaceRangeItem", kustolang.ListReplaceRangeItem([]string{"a", "b", "c"}, 1, 2, "x"), []string{"a", "x"})
	assertSliceEqual(t, "ListReplaceRangeItems", kustolang.ListReplaceRangeItems([]string{"a", "b", "c"}, 1, 1, []string{"x", "y"}), []string{"a", "x", "y", "c"})
}

func TestNewCancellationTokenFunc(t *testing.T) {
	canceled := false
	checks := 0
	token := kustolang.NewCancellationTokenFunc(func() bool {
		checks++
		return canceled
	})
	if err := token.Err(); err != nil || checks != 1 {
		t.Fatalf("uncanceled Err = %v, checks = %d", err, checks)
	}
	canceled = true
	if err := token.Err(); err == nil || checks != 2 {
		t.Fatalf("canceled Err = %v, checks = %d", err, checks)
	}
}

func listValues(items []publicListItem) []int {
	values := make([]int, len(items))
	for i, item := range items {
		values[i] = item.value
	}
	return values
}

func assertPanicContains(t *testing.T, want string, action func()) {
	t.Helper()
	defer func() {
		value := recover()
		if value == nil {
			t.Fatalf("expected panic containing %q", want)
		}
		if got := fmt.Sprint(value); !strings.Contains(got, want) {
			t.Fatalf("panic = %q, want substring %q", got, want)
		}
	}()
	action()
}

func assertSliceEqual[T comparable](t *testing.T, name string, got, want []T) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}
