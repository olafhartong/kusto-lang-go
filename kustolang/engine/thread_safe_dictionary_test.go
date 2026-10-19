package kustolang

import (
	"sort"
	"sync"
	"testing"
)

func TestThreadSafeDictionaryTryAddGetKeysValues(t *testing.T) {
	dictionary := NewThreadSafeDictionary[string, int]()
	if dictionary.Count() != 0 {
		t.Fatalf("empty count = %d, want 0", dictionary.Count())
	}
	if !dictionary.TryAdd("a", 1) {
		t.Fatal("TryAdd should add missing key")
	}
	if dictionary.TryAdd("a", 2) {
		t.Fatal("TryAdd should reject existing key")
	}
	if got, ok := dictionary.TryGetValue("a"); !ok || got != 1 {
		t.Fatalf("TryGetValue(a) = %d, %t, want 1, true", got, ok)
	}
	if _, ok := dictionary.TryGetValue("missing"); ok {
		t.Fatal("TryGetValue should reject missing key")
	}
	if dictionary.Count() != 1 {
		t.Fatalf("count = %d, want 1", dictionary.Count())
	}

	dictionary.TryAdd("b", 2)
	keys := dictionary.Keys()
	sort.Strings(keys)
	if len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
		t.Fatalf("keys = %#v, want [a b]", keys)
	}
	values := dictionary.Values()
	sort.Ints(values)
	if len(values) != 2 || values[0] != 1 || values[1] != 2 {
		t.Fatalf("values = %#v, want [1 2]", values)
	}
}

func TestThreadSafeDictionaryGetOrAdd(t *testing.T) {
	dictionary := NewThreadSafeDictionary[string, string]()
	if got := dictionary.GetOrAdd("name", "first"); got != "first" {
		t.Fatalf("added value = %q, want first", got)
	}
	if got := dictionary.GetOrAdd("name", "second"); got != "first" {
		t.Fatalf("existing value = %q, want first", got)
	}

	called := false
	if got := dictionary.GetOrAddFunc("name", func(key string) string {
		called = true
		return "factory"
	}); got != "first" {
		t.Fatalf("existing factory value = %q, want first", got)
	}
	if called {
		t.Fatal("factory should not run for existing key")
	}
	if got := dictionary.GetOrAddFunc("other", func(key string) string {
		return key + "-value"
	}); got != "other-value" {
		t.Fatalf("added factory value = %q, want other-value", got)
	}
}

func TestThreadSafeDictionaryAddOrUpdate(t *testing.T) {
	dictionary := NewThreadSafeDictionary[string, int]()
	dictionary.AddOrUpdate("count", 1)
	dictionary.AddOrUpdate("count", 2)
	if got, _ := dictionary.TryGetValue("count"); got != 2 {
		t.Fatalf("updated value = %d, want 2", got)
	}

	dictionary.AddOrUpdateFunc("total", func(key string) int {
		return len(key)
	}, func(key string, old int) int {
		return old + len(key)
	})
	if got, _ := dictionary.TryGetValue("total"); got != 5 {
		t.Fatalf("factory add value = %d, want 5", got)
	}
	dictionary.AddOrUpdateFunc("total", func(key string) int {
		return 0
	}, func(key string, old int) int {
		return old + len(key)
	})
	if got, _ := dictionary.TryGetValue("total"); got != 10 {
		t.Fatalf("factory update value = %d, want 10", got)
	}

	dictionary.AddOrUpdateValueFunc("score", func(key string, value int) int {
		return value + 1
	}, func(key string, old int, value int) int {
		return old + value
	}, 4)
	if got, _ := dictionary.TryGetValue("score"); got != 5 {
		t.Fatalf("value-factory add value = %d, want 5", got)
	}
	dictionary.AddOrUpdateValueFunc("score", func(key string, value int) int {
		return value
	}, func(key string, old int, value int) int {
		return old + value
	}, 4)
	if got, _ := dictionary.TryGetValue("score"); got != 9 {
		t.Fatalf("value-factory update value = %d, want 9", got)
	}
}

func TestThreadSafeDictionaryConcurrentAccess(t *testing.T) {
	dictionary := NewThreadSafeDictionary[int, int]()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			dictionary.GetOrAddFunc(i%10, func(key int) int {
				return key
			})
			dictionary.AddOrUpdateFunc(i%10, func(key int) int {
				return key
			}, func(key int, old int) int {
				return old + 1
			})
			dictionary.TryGetValue(i % 10)
		}()
	}
	wg.Wait()
	if dictionary.Count() != 10 {
		t.Fatalf("concurrent count = %d, want 10", dictionary.Count())
	}
}
