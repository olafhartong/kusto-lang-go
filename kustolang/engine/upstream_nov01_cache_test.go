package kustolang

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
)

const nov01Commit = "82267c7628e70898cd8bec54c20fc4e61c0a523a"

type nov01CacheFixture struct {
	Commit                                                string
	SourceCalls, KustoCalls, RetryCalls, DistinctGrammars int
	SourceReused, KustoReused, RetryStored                bool
	Grammars                                              []struct {
		Index                                                                 int
		ReusedAfterOptionSwitch, ReusedAcrossGlobals, ReusedAcrossParserKinds bool
	}
}

func readNov01CacheFixture(t *testing.T) nov01CacheFixture {
	t.Helper()
	data, err := os.ReadFile("testdata/upstream_82267c7_caching.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture nov01CacheFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Commit != nov01Commit || len(fixture.Grammars) != 8 || fixture.DistinctGrammars != 8 {
		t.Fatal("incorrectly pinned or incomplete cache source oracle")
	}
	return fixture
}

func nov01Options(index int) ParseOptions {
	return DefaultParseOptions.WithAlwaysProduceEndTokens(index&1 != 0).
		WithAllowLiteralsWithLineBreaks(index&2 != 0).WithAllowNonAdjacentWildcardParts(index&4 != 0)
}

func TestUpstreamNov01CacheFactories(t *testing.T) {
	want := readNov01CacheFixture(t)
	source := NewSourceCache()
	sourceCalls := 0
	creator := func() *int { sourceCalls++; value := sourceCalls; return &value }
	first := SourceCacheGetOrCreateFunc(source, creator)
	again := SourceCacheGetOrCreateFunc(source, creator)
	if sourceCalls != want.SourceCalls || (first == again) != want.SourceReused {
		t.Fatalf("source calls=%d reused=%v; want %d/%v", sourceCalls, first == again, want.SourceCalls, want.SourceReused)
	}
	cache := NewKustoCache(DefaultGlobalState)
	kustoCalls := 0
	kustoCreator := func() *int { kustoCalls++; value := kustoCalls; return &value }
	kustoFirst := KustoCacheGetOrCreate(cache, kustoCreator)
	kustoAgain := KustoCacheGetOrCreate(cache, kustoCreator)
	if kustoCalls != want.KustoCalls || (kustoFirst == kustoAgain) != want.KustoReused {
		t.Fatalf("KustoCache calls=%d reused=%v; want %d/%v", kustoCalls, kustoFirst == kustoAgain, want.KustoCalls, want.KustoReused)
	}
	retry := NewSourceCache()
	retryCalls := 0
	func() {
		defer func() {
			if recover() == nil {
				t.Error("factory panic was swallowed")
			}
		}()
		SourceCacheGetOrCreateFunc(retry, func() *int { retryCalls++; panic("factory failure") })
	}()
	value := SourceCacheGetOrCreateFunc(retry, func() *int { retryCalls++; value := 42; return &value })
	stored, ok := SourceCacheTryGetValue[*int](retry)
	if retryCalls != want.RetryCalls || (ok && stored == value) != want.RetryStored {
		t.Fatalf("retry calls=%d stored=%v; want %d/%v", retryCalls, ok && stored == value, want.RetryCalls, want.RetryStored)
	}
}

func TestUpstreamNov01GrammarCacheVariants(t *testing.T) {
	fixture := readNov01CacheFixture(t)
	distinct := map[*QueryGrammar]bool{}
	for _, want := range fixture.Grammars {
		t.Run(fmt.Sprint(want.Index), func(t *testing.T) {
			options := nov01Options(want.Index)
			first := QueryGrammarFrom(DefaultGlobalState.WithParseOptions(options.WithParserKind(ParserKindGrammar)))
			QueryGrammarFrom(DefaultGlobalState.WithParseOptions(options.WithAlwaysProduceEndTokens(want.Index&1 == 0)))
			if got := first == QueryGrammarFrom(DefaultGlobalState.WithParseOptions(options)); got != want.ReusedAfterOptionSwitch {
				t.Errorf("reuse after switching=%v, want %v", got, want.ReusedAfterOptionSwitch)
			}
			globals := NewGlobalState(NewDatabaseSymbol("other")).WithCache().WithParseOptions(options)
			if got := first == QueryGrammarFrom(globals); got != want.ReusedAcrossGlobals {
				t.Errorf("reuse across globals=%v, want %v", got, want.ReusedAcrossGlobals)
			}
			if got := first == QueryGrammarFrom(globals.WithParseOptions(options.WithParserKind(ParserKindGrammar))); got != want.ReusedAcrossParserKinds {
				t.Errorf("reuse across parser kinds=%v, want %v", got, want.ReusedAcrossParserKinds)
			}
			if first.Options != options {
				t.Errorf("grammar options=%+v, want canonical %+v", first.Options, options)
			}
			if len(globals.Cache.values) != 0 {
				t.Error("query grammar was stored in the schema's KustoCache")
			}
			distinct[first] = true
		})
	}
	if len(distinct) != fixture.DistinctGrammars {
		t.Fatalf("distinct grammars=%d, want %d", len(distinct), fixture.DistinctGrammars)
	}
}

func TestUpstreamNov01GrammarCacheConcurrentPublication(t *testing.T) {
	type result struct {
		index   int
		grammar *QueryGrammar
	}
	start := make(chan struct{})
	results := make(chan result, 128)
	var group sync.WaitGroup
	for i := range 128 {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			index := i % 8
			globals := NewGlobalState(NewDatabaseSymbol(fmt.Sprintf("db%d", i))).WithParseOptions(nov01Options(index).WithParserKind(ParserKind(i % 2)))
			if i%3 == 0 {
				globals = globals.WithCache()
			}
			results <- result{index, QueryGrammarFrom(globals)}
		}()
	}
	close(start)
	group.Wait()
	close(results)
	var published [8]*QueryGrammar
	for result := range results {
		if published[result.index] == nil {
			published[result.index] = result.grammar
		}
		if result.grammar != published[result.index] {
			t.Errorf("multiple grammars published for variant %d", result.index)
		}
		if result.grammar.Options != nov01Options(result.index) {
			t.Errorf("incorrect options for variant %d", result.index)
		}
	}
	if QueryGrammarFrom(nil) != published[1] {
		t.Error("nil globals did not reuse the default option variant")
	}
}
