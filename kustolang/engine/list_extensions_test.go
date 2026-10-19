package kustolang

import "testing"

type listExtensionItem struct {
	name  string
	value int
}

func TestListToReadOnlyAndSafeListCopyInput(t *testing.T) {
	items := []string{"a", "b"}
	readOnly := ListToReadOnly(items)
	safe := ListToSafeList(items)
	items[0] = "changed"

	if readOnly[0] != "a" {
		t.Fatalf("read-only copy = %#v, want first item a", readOnly)
	}
	if safe.At(0) != "a" {
		t.Fatalf("safe list first item = %q, want a", safe.At(0))
	}
	if empty := ListToSafeList([]string{}); empty.Count() != 0 || empty.IsOwner() {
		t.Fatalf("empty safe list count/owner = %d/%t, want 0/false", empty.Count(), empty.IsOwner())
	}
}

func TestListToHashSetAndDictionaryLast(t *testing.T) {
	set := ListToHashSetEx([]string{"a", "b", "a"})
	if len(set) != 2 {
		t.Fatalf("set count = %d, want 2", len(set))
	}

	items := []listExtensionItem{
		{name: "a", value: 1},
		{name: "b", value: 2},
		{name: "a", value: 3},
	}
	values := ListToDictionaryLast(items, func(item listExtensionItem) string {
		return item.name
	}, func(item listExtensionItem) int {
		return item.value
	})
	if values["a"] != 3 || values["b"] != 2 {
		t.Fatalf("dictionary last = %#v, want a=3 b=2", values)
	}

	identity := ListToDictionaryLastValues([]string{"one", "two"}, func(value string) int {
		return len(value)
	})
	if identity[3] != "two" {
		t.Fatalf("dictionary last values = %#v, want length 3 -> two", identity)
	}
}

func TestListAddOrUpdateDistinctFirstAndLast(t *testing.T) {
	itemA := &listExtensionItem{name: "a", value: 1}
	itemB := &listExtensionItem{name: "b", value: 2}
	itemAUpdated := &listExtensionItem{name: "a", value: 3}

	updated := ListAddOrUpdate([]*listExtensionItem{itemA, itemB}, []*listExtensionItem{itemAUpdated}, func(item *listExtensionItem) string {
		return item.name
	})
	if len(updated) != 2 || updated[0] != itemB || updated[1] != itemAUpdated {
		t.Fatalf("single add/update = %#v, want [b aUpdated]", updated)
	}
	if same := ListAddOrUpdate(updated, []*listExtensionItem{nil}, func(item *listExtensionItem) string {
		return item.name
	}); len(same) != len(updated) || same[0] != updated[0] || same[1] != updated[1] {
		t.Fatalf("nil single update changed list: %#v", same)
	}

	multiple := ListAddOrUpdate([]*listExtensionItem{itemA, itemB}, []*listExtensionItem{nil, itemAUpdated}, func(item *listExtensionItem) string {
		return item.name
	})
	if len(multiple) != 2 || multiple[0] != itemB || multiple[1] != itemAUpdated {
		t.Fatalf("multi add/update = %#v, want [b aUpdated]", multiple)
	}

	distinctFirst := ListDistinctFirst([]listExtensionItem{
		{name: "a", value: 1},
		{name: "b", value: 2},
		{name: "a", value: 3},
	}, func(item listExtensionItem) string {
		return item.name
	})
	if len(distinctFirst) != 2 || distinctFirst[0].value != 1 || distinctFirst[1].value != 2 {
		t.Fatalf("distinct first = %#v, want first a and b", distinctFirst)
	}

	distinctLast := ListDistinctLast([]listExtensionItem{
		{name: "a", value: 1},
		{name: "b", value: 2},
		{name: "a", value: 3},
	}, func(item listExtensionItem) string {
		return item.name
	})
	if len(distinctLast) != 2 || distinctLast[0].value != 2 || distinctLast[1].value != 3 {
		t.Fatalf("distinct last = %#v, want b and last a", distinctLast)
	}
}

func TestListSearchIndexRemoveSetAndJoin(t *testing.T) {
	items := []int{1, 3, 5, 7, 9}
	if got := ListBinarySearch(items, func(value int) int {
		switch {
		case 5 < value:
			return 1
		case 5 > value:
			return -1
		default:
			return 0
		}
	}); got != 2 {
		t.Fatalf("binary search index = %d, want 2", got)
	}
	if got := ListBinarySearch(items, func(value int) int {
		if 4 < value {
			return 1
		}
		return -1
	}); got != -1 {
		t.Fatalf("missing binary search index = %d, want -1", got)
	}
	if got := ListIndexOf([]string{"a", "b"}, "b"); got != 1 {
		t.Fatalf("index = %d, want 1", got)
	}
	if got := ListFirstIndex(items, func(value int) bool { return value > 4 }); got != 2 {
		t.Fatalf("first index = %d, want 2", got)
	}
	if got := ListLastIndex(items, func(value int) bool { return value < 8 }); got != 3 {
		t.Fatalf("last index = %d, want 3", got)
	}

	removed := ListRemoveAll([]int{1, 2, 3, 4}, func(value int) bool { return value%2 == 0 })
	if len(removed) != 2 || removed[0] != 1 || removed[1] != 3 {
		t.Fatalf("removed = %#v, want [1 3]", removed)
	}
	if got := ListSetCount([]int{1, 2, 3}, 2); len(got) != 2 || got[1] != 2 {
		t.Fatalf("set count = %#v, want [1 2]", got)
	}
	if got := ListJoin([]string{"a", "b", "c"}, ",", " and "); got != "a,b and c" {
		t.Fatalf("join = %q, want a,b and c", got)
	}
	if got := ListJoin([]string{"a", "b"}, ","); got != "ab" {
		t.Fatalf("join default final separator = %q, want ab", got)
	}
}

func TestListReplaceAppendAndEquivalence(t *testing.T) {
	appended := ListAppendItem([]string{"a", "b"}, "c")
	if len(appended) != 3 || appended[2] != "c" {
		t.Fatalf("append item = %#v, want [a b c]", appended)
	}
	appendedItems := ListAppendItems([]string{"a"}, []string{"b", "c"})
	if len(appendedItems) != 3 || appendedItems[1] != "b" || appendedItems[2] != "c" {
		t.Fatalf("append items = %#v, want [a b c]", appendedItems)
	}
	replaced := ListReplaceItem([]string{"a", "b", "c"}, 1, "B")
	if len(replaced) != 3 || replaced[1] != "B" {
		t.Fatalf("replace item = %#v, want [a B c]", replaced)
	}
	rangeItem := ListReplaceRangeItem([]string{"a", "b", "c"}, 1, 2, "x")
	if len(rangeItem) != 2 || rangeItem[0] != "a" || rangeItem[1] != "x" {
		t.Fatalf("replace range item = %#v, want [a x]", rangeItem)
	}
	rangeItems := ListReplaceRangeItems([]string{"a", "b", "c"}, 1, 1, []string{"x", "y"})
	if len(rangeItems) != 4 || rangeItems[1] != "x" || rangeItems[2] != "y" || rangeItems[3] != "c" {
		t.Fatalf("replace range items = %#v, want [a x y c]", rangeItems)
	}
	noInsert := ListReplaceRangeItem([]*listExtensionItem{{name: "a"}}, 1, 0, nil)
	if len(noInsert) != 1 || noInsert[0].name != "a" {
		t.Fatalf("nil item insert = %#v, want unchanged", noInsert)
	}

	if !DictionaryIsEquivalentTo(map[string]int{"a": 1}, map[string]int{"a": 1}) {
		t.Fatal("equivalent dictionaries should compare true")
	}
	if DictionaryIsEquivalentTo(map[string]int{"a": 1}, map[string]int{"a": 2}) {
		t.Fatal("different dictionaries should compare false")
	}
	if !ReadOnlyListEquals([]int{1, 2}, []int{1, 2}) || ReadOnlyListEquals([]int{1, 2}, []int{2, 1}) {
		t.Fatal("read-only list equality mismatch")
	}
	if got := ReadOnlyListHashCode([]int{1, 2, 3}, func(value int) int { return value }); got != 6 {
		t.Fatalf("read-only list hash = %d, want 6", got)
	}
}
