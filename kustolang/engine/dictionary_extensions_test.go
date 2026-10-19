package kustolang

import "testing"

func TestDictionaryGetOrAdd(t *testing.T) {
	dictionary := map[string]int{"a": 1}
	if got := DictionaryGetOrAdd(dictionary, "a", 2); got != 1 {
		t.Fatalf("existing value = %d, want 1", got)
	}
	if dictionary["a"] != 1 {
		t.Fatalf("existing key was overwritten: %d", dictionary["a"])
	}
	if got := DictionaryGetOrAdd(dictionary, "b", 2); got != 2 {
		t.Fatalf("added value = %d, want 2", got)
	}
	if dictionary["b"] != 2 {
		t.Fatalf("missing added key value: %d", dictionary["b"])
	}
}

func TestDictionaryGetOrAddFunc(t *testing.T) {
	dictionary := map[string]string{"name": "old"}
	called := false
	if got := DictionaryGetOrAddFunc(dictionary, "name", func(key string) string {
		called = true
		return key + "-new"
	}); got != "old" {
		t.Fatalf("existing factory value = %q, want old", got)
	}
	if called {
		t.Fatal("factory should not be called for existing key")
	}
	if got := DictionaryGetOrAddFunc(dictionary, "other", func(key string) string {
		return key + "-new"
	}); got != "other-new" {
		t.Fatalf("added factory value = %q, want other-new", got)
	}
}

func TestDictionaryAddOrUpdate(t *testing.T) {
	dictionary := map[string]int{}
	DictionaryAddOrUpdate(dictionary, "count", 1)
	DictionaryAddOrUpdate(dictionary, "count", 2)
	if got := dictionary["count"]; got != 2 {
		t.Fatalf("updated value = %d, want 2", got)
	}

	DictionaryAddOrUpdateFunc(dictionary, "total", func(key string) int {
		return len(key)
	}, func(key string, old int) int {
		return old + len(key)
	})
	if got := dictionary["total"]; got != 5 {
		t.Fatalf("factory add value = %d, want 5", got)
	}
	DictionaryAddOrUpdateFunc(dictionary, "total", func(key string) int {
		return 0
	}, func(key string, old int) int {
		return old + len(key)
	})
	if got := dictionary["total"]; got != 10 {
		t.Fatalf("factory update value = %d, want 10", got)
	}

	DictionaryAddOrUpdateValueFunc(dictionary, "score", func(key string, value int) int {
		return value + 1
	}, func(key string, old int, value int) int {
		return old + value
	}, 4)
	if got := dictionary["score"]; got != 5 {
		t.Fatalf("value-factory add value = %d, want 5", got)
	}
	DictionaryAddOrUpdateValueFunc(dictionary, "score", func(key string, value int) int {
		return value
	}, func(key string, old int, value int) int {
		return old + value
	}, 4)
	if got := dictionary["score"]; got != 9 {
		t.Fatalf("value-factory update value = %d, want 9", got)
	}
}
