package kustolang

import (
	"sort"
	"testing"
)

func TestTextKeyedDictionaryRangeLookupMatchesDifferentBackingText(t *testing.T) {
	dictionary := NewTextKeyedDictionary[string]()
	added := dictionary.GetOrAddValueRange("xxTargetyy", 2, 6, "stored")
	if added != "stored" {
		t.Fatalf("added value = %q, want stored", added)
	}

	if got, ok := dictionary.TryGetValue("Target", 0, 6); !ok || got != "stored" {
		t.Fatalf("range lookup = %q, %t, want stored, true", got, ok)
	}
	if !dictionary.ContainsKey("Target") {
		t.Fatal("ContainsKey should match equivalent full text")
	}
	if !dictionary.ContainsKeyRange("__Target__", 2, 6) {
		t.Fatal("ContainsKeyRange should match equivalent subrange")
	}
}

func TestTextKeyedDictionaryGetOrAddFuncIsLazyForExistingKey(t *testing.T) {
	dictionary := NewTextKeyedDictionary[int]()
	if got := dictionary.GetOrAddValue("Name", 1); got != 1 {
		t.Fatalf("added value = %d, want 1", got)
	}
	called := false
	if got := dictionary.GetOrAddValueFunc("Name", func(string, int, int) int {
		called = true
		return 2
	}); got != 1 {
		t.Fatalf("existing value = %d, want 1", got)
	}
	if called {
		t.Fatal("evaluator should not run for an existing key")
	}
	if got := dictionary.GetOrAddValueRangeFunc("xxOther", 2, 5, func(text string, start int, length int) int {
		return len(text[start : start+length])
	}); got != 5 {
		t.Fatalf("new range value = %d, want 5", got)
	}
}

func TestTextKeyedDictionaryContainsRejectsInvalidRanges(t *testing.T) {
	dictionary := NewTextKeyedDictionary[string]()
	dictionary.GetOrAddValue("", "empty")
	if got, ok := dictionary.TryGetValue("", 0, 0); !ok || got != "empty" {
		t.Fatalf("empty TryGetValue = %q, %t, want empty, true", got, ok)
	}
	if dictionary.ContainsKey("") {
		t.Fatal("ContainsKey should reject zero-length text like upstream ContainsKey")
	}
	if dictionary.ContainsKeyRange("abc", -1, 1) {
		t.Fatal("ContainsKeyRange should reject negative start")
	}
	if dictionary.ContainsKeyRange("abc", 0, 0) {
		t.Fatal("ContainsKeyRange should reject zero length")
	}
	if dictionary.ContainsKeyRange("abc", 2, 2) {
		t.Fatal("ContainsKeyRange should reject ranges past the text end")
	}
}

func TestTextKeyedDictionaryValuesAndProjection(t *testing.T) {
	type item struct {
		key   string
		value string
	}
	dictionary := ToTextKeyedDictionary([]item{
		{key: "a", value: "first"},
		{key: "b", value: "second"},
		{key: "a", value: "ignored"},
	}, func(item item) string {
		return item.key
	}, func(item item) string {
		return item.value
	})

	values := dictionary.Values()
	sort.Strings(values)
	if len(values) != 2 || values[0] != "first" || values[1] != "second" {
		t.Fatalf("values = %#v, want [first second]", values)
	}
}
