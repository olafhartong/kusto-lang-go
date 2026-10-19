package kustolang_test

import (
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestConnectionInfoFacadePreservesSourceParsing(t *testing.T) {
	text := `Data Source="https://cluster;with;semi";Initial Catalog= db ;User=' name '`
	info := kustolang.ParseConnectionInfo(text)
	if info.Text() != text || info.DataSource() != "https://cluster;with;semi" {
		t.Fatalf("connection = text:%q data-source:%q", info.Text(), info.DataSource())
	}
	if got := info.GetPart("initial catalog"); got != "db" {
		t.Fatalf("catalog = %q", got)
	}
	if got := info.GetPart("USER"); got != " name " {
		t.Fatalf("user = %q", got)
	}
	if got := kustolang.ParseConnectionInfo(" https://cluster ").DataSource(); got != "https://cluster" {
		t.Fatalf("unnamed data source = %q", got)
	}
	if got := kustolang.ParseConnectionInfo("A=\u200b'x'").GetPart("A"); got != "x" {
		t.Fatalf("zero-width-space value = %q", got)
	}
	if got := kustolang.ParseConnectionInfo("A=\u2028'x'").GetPart("A"); got != "'x'" {
		t.Fatalf("line-separator value = %q", got)
	}
}

func TestOptionalAndUniqueNameTableFacades(t *testing.T) {
	var absent kustolang.Optional[int]
	if absent.HasValue() || absent.Value() != 0 {
		t.Fatalf("zero optional = value:%d has:%t", absent.Value(), absent.HasValue())
	}
	present := kustolang.NewOptional(0)
	if !present.HasValue() || present.Value() != 0 {
		t.Fatalf("present optional = value:%d has:%t", present.Value(), present.HasValue())
	}

	names := kustolang.NewUniqueNameTable()
	if got := names.GetOrAddName("x"); got != "x" {
		t.Fatalf("first name = %q", got)
	}
	if got := names.GetOrAddName("x"); got != "x1" {
		t.Fatalf("second name = %q", got)
	}
	if got := names.GetOrAddName("x", "column"); got != "column2" {
		t.Fatalf("alternate base name = %q", got)
	}
	names.Clear()
	if got := names.GetOrAddName("x"); got != "x" {
		t.Fatalf("name after clear = %q", got)
	}
}

func TestTextKeyedDictionaryUsesUTF16Ranges(t *testing.T) {
	dictionary := kustolang.NewTextKeyedDictionary[string]()
	text := "A😀BC"
	if got := dictionary.GetOrAddValueRange(text, 3, 1, "bee"); got != "bee" {
		t.Fatalf("insert = %q", got)
	}
	if got, ok := dictionary.TryGetValue("xxB", 2, 1); !ok || got != "bee" {
		t.Fatalf("equivalent range = %q, %t", got, ok)
	}
	if dictionary.ContainsKeyRange(text, 1, 0) || dictionary.ContainsKeyRange(text, -1, 1) {
		t.Fatal("invalid/empty ranges must not be reported as keys")
	}

	dictionary.GetOrAddValueRange(text, 1, 1, "high-surrogate")
	dictionary.GetOrAddValueRange(text, 2, 1, "low-surrogate")
	if high, _ := dictionary.TryGetValue(text, 1, 1); high != "high-surrogate" {
		t.Fatalf("high surrogate = %q", high)
	}
	if low, _ := dictionary.TryGetValue(text, 2, 1); low != "low-surrogate" {
		t.Fatalf("low surrogate = %q", low)
	}

	called := 0
	if got := dictionary.GetOrAddValueRangeFunc("B", 0, 1, func(string, int, int) string {
		called++
		return "changed"
	}); got != "bee" || called != 0 {
		t.Fatalf("existing evaluator result = %q, calls = %d", got, called)
	}

	values := dictionary.Values()
	values[0] = "mutated"
	if got, _ := dictionary.TryGetValue("B", 0, 1); got != "bee" {
		t.Fatalf("values leaked mutable storage: %q", got)
	}
}

func TestTextKeyedDictionaryInvalidLookupPanicsLikeSourceKeyConstruction(t *testing.T) {
	dictionary := kustolang.NewTextKeyedDictionary[int]()
	defer func() {
		if recover() == nil {
			t.Fatal("invalid UTF-16 range did not panic")
		}
	}()
	dictionary.TryGetValue("x", 2, 1)
}

func TestStringTableInternsUTF16Subranges(t *testing.T) {
	table := kustolang.NewStringTable()
	text := "A😀BC"
	if got := table.AddRange(text, 1, 2); got != "😀" {
		t.Fatalf("astral range = %q", got)
	}
	if !table.ContainsKey("x😀y", 1, 2) || !table.Contains("😀") {
		t.Fatal("UTF-16 subrange was not interned")
	}
	if got := table.AddRange(text, 3, 1); got != "B" {
		t.Fatalf("post-surrogate range = %q", got)
	}
	values := table.Values()
	values[0] = "changed"
	if !table.Contains("😀") {
		t.Fatal("Values returned mutable table storage")
	}
}

func TestThreadSafeDictionaryComparerFactoriesAndSnapshots(t *testing.T) {
	dictionary := kustolang.NewThreadSafeDictionaryWithComparer[string, int](strings.EqualFold)
	if !dictionary.TryAdd("Alpha", 1) || dictionary.TryAdd("alpha", 2) {
		t.Fatal("custom comparer did not define key identity")
	}
	if got := dictionary.GetOrAdd("ALPHA", 3); got != 1 {
		t.Fatalf("GetOrAdd existing = %d", got)
	}
	dictionary.AddOrUpdateFunc("alpha", func(string) int { return 10 }, func(_ string, old int) int { return old + 1 })
	if got, ok := dictionary.TryGetValue("ALPHA"); !ok || got != 2 {
		t.Fatalf("updated value = %d, %t", got, ok)
	}
	keys := dictionary.Keys()
	keys[0] = "changed"
	if _, ok := dictionary.TryGetValue("alpha"); !ok {
		t.Fatal("Keys returned mutable dictionary storage")
	}
	if got := dictionary.GetOrAddFunc("outer", func(string) int {
		if !dictionary.TryAdd("inner", 7) {
			t.Fatal("re-entrant factory could not add a key")
		}
		return 6
	}); got != 6 {
		t.Fatalf("re-entrant factory result = %d", got)
	}
	dictionary.AddOrUpdateFunc("outer", func(string) int { return 0 }, func(_ string, old int) int {
		dictionary.AddOrUpdate("side-effect", 9)
		return old + 1
	})
	if got, _ := dictionary.TryGetValue("outer"); got != 7 {
		t.Fatalf("re-entrant update result = %d", got)
	}

	concurrent := kustolang.NewThreadSafeDictionary[int, int]()
	var group sync.WaitGroup
	for i := 0; i < 64; i++ {
		group.Add(1)
		go func(value int) {
			defer group.Done()
			concurrent.AddOrUpdate(value%8, 1)
		}(i)
	}
	group.Wait()
	if concurrent.Count() != 8 {
		t.Fatalf("concurrent count = %d", concurrent.Count())
	}
}

func TestThreadSafeDictionaryRejectsNilKeys(t *testing.T) {
	dictionary := kustolang.NewThreadSafeDictionary[*int, int]()
	defer func() {
		if recover() == nil {
			t.Fatal("nil key did not panic")
		}
	}()
	dictionary.TryAdd(nil, 1)
}

func TestObjectPoolAndMostRecentlyUsedCacheSourceBehavior(t *testing.T) {
	type pooled struct{ value int }
	var created atomic.Int32
	pool := kustolang.NewObjectPool(func() *pooled {
		created.Add(1)
		return &pooled{}
	}, func(value *pooled) {
		value.value = 0
	}, 1)
	item := pool.AllocateFromPool()
	item.value = 42
	pool.ReturnToPool(item)
	if reused := pool.AllocateFromPool(); reused != item || reused.value != 0 {
		t.Fatalf("reused item = %#v, original = %#v", reused, item)
	}
	if created.Load() != 1 {
		t.Fatalf("created = %d", created.Load())
	}

	cache := kustolang.NewMostRecentlyUsedCache[string, int](2)
	cache.AddOrUpdate("a", 1)
	cache.AddOrUpdate("b", 2)
	if value, ok := cache.TryGetValue("a"); !ok || value != 1 {
		t.Fatalf("cached a = %d, %t", value, ok)
	}
	cache.AddOrUpdate("c", 3)
	if _, ok := cache.TryGetValue("b"); ok {
		t.Fatal("least-recently-used entry was not evicted")
	}
	cache.AddOrUpdate("a", 99)
	if value, _ := cache.TryGetValue("a"); value != 1 {
		t.Fatalf("source refresh quirk was not preserved: %d", value)
	}
}

func TestValueAndStringNumberComparerStaticFacades(t *testing.T) {
	if !kustolang.ValueComparer.AreEquivalent("True", true) {
		t.Fatal("cross-type string equivalence failed")
	}
	if kustolang.ValueComparer.AreEquivalent("Alpha", "alpha", true) {
		t.Fatal("case-sensitive equivalence ignored")
	}
	left := []int{1, 2}
	if kustolang.ValueComparer.AreEquivalent(left, []int{1, 2}) || !kustolang.ValueComparer.AreEquivalent(left, left) {
		t.Fatal("reference identity equivalence failed")
	}
	if !kustolang.ValueComparer.AreEquivalent(math.NaN(), math.NaN()) {
		t.Fatal("NaN equivalence failed")
	}
	if got := kustolang.StringAndNumberComparer.Ordinal.Compare("item2", "item10"); got >= 0 {
		t.Fatalf("numeric order = %d", got)
	}
	if got := kustolang.StringAndNumberComparer.OrdinalIgnoreCase.Compare("Alpha02", "alpha2"); got <= 0 {
		t.Fatalf("case-insensitive leading-zero order = %d", got)
	}
	if got := kustolang.StringAndNumberComparer.Ordinal.Compare("𝟡2", "𝟙10"); got <= 0 {
		t.Fatalf("supplementary digit order = %d", got)
	}
	if got := kustolang.StringAndNumberComparer.Ordinal.Compare("𐀀", "\ue000"); got >= 0 {
		t.Fatalf("UTF-16 ordinal order = %d", got)
	}
}

func TestToTextKeyedDictionaryKeepsFirstDuplicateValue(t *testing.T) {
	type item struct {
		key   string
		value int
	}
	dictionary := kustolang.ToTextKeyedDictionary(
		[]item{{key: "a", value: 1}, {key: "a", value: 2}, {key: "b", value: 3}},
		func(item item) string { return item.key },
		func(item item) int { return item.value },
	)
	if value, _ := dictionary.TryGetValue("a", 0, 1); value != 1 {
		t.Fatalf("duplicate value = %d", value)
	}
}

func TestUtilityFacadeIncrementRetainsMalformedKQLRecovery(t *testing.T) {
	if diagnostics := kustolang.ParseAndAnalyze(`print x = 1 + 2`).GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid diagnostics = %#v", diagnostics)
	}
	malformed := kustolang.ParseAndAnalyze(`print x = 1 +`)
	if diagnostics := malformed.GetDiagnostics(); !hasDiagnostic(diagnostics, "KS006") {
		t.Fatalf("malformed diagnostics = %#v", diagnostics)
	}
	if malformed.Syntax().GetLastToken(true) == nil {
		t.Fatal("malformed recovery did not retain a syntax tree")
	}
}
