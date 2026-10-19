package kustolang_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicOptionCatalogMatchesPinnedUpstream(t *testing.T) {
	all := kustolang.Options.All()
	if len(all) != 58 {
		t.Fatalf("option count = %d, want 58", len(all))
	}
	if all[0] != kustolang.Options.BestEffort || all[len(all)-1] != kustolang.Options.ValidatePermissions {
		t.Fatalf("option order = first:%v last:%v", all[0], all[len(all)-1])
	}
	if kustolang.Options.QueryNow.SymbolName() != "query_now" ||
		len(kustolang.Options.QueryNow.Types()) != 1 ||
		kustolang.Options.QueryNow.Types()[0] != kustolang.ScalarTypes.DateTime {
		t.Fatalf("query_now metadata = %#v", kustolang.Options.QueryNow)
	}
	if got := kustolang.DefaultGlobalState.GetOption("query_now"); got != kustolang.Options.QueryNow {
		t.Fatalf("default query_now = %p, catalog = %p", got, kustolang.Options.QueryNow)
	}

	all[0] = nil
	if kustolang.Options.All()[0] != kustolang.Options.BestEffort {
		t.Fatal("Options.All exposed mutable catalog storage")
	}
}

func TestPublicGlobalStatePropertiesAreTypedAndImmutable(t *testing.T) {
	if got := kustolang.GlobalStateGetProperty(kustolang.DefaultGlobalState, kustolang.Properties.MaxParseTextSize); got != 4*1024*1024 {
		t.Fatalf("MaxParseTextSize = %d", got)
	}
	if kustolang.Properties.MaxAnalysisDepth.Name() != "MaxAnalysisDepth" || kustolang.Properties.MaxAnalysisDepth.DefaultValue() != 500 {
		t.Fatalf("MaxAnalysisDepth = %q/%d", kustolang.Properties.MaxAnalysisDepth.Name(), kustolang.Properties.MaxAnalysisDepth.DefaultValue())
	}

	custom := kustolang.NewGlobalStateProperty[int]("", 7)
	if custom.Name() != "" || kustolang.GlobalStateGetProperty[int](nil, custom) != 7 {
		t.Fatalf("custom property = %q/%d", custom.Name(), kustolang.GlobalStateGetProperty[int](nil, custom))
	}
	changed := kustolang.GlobalStateWithProperty(kustolang.DefaultGlobalState, custom, 42)
	if changed == kustolang.DefaultGlobalState || kustolang.GlobalStateGetProperty(changed, custom) != 42 {
		t.Fatalf("changed property state = same:%t value:%d", changed == kustolang.DefaultGlobalState, kustolang.GlobalStateGetProperty(changed, custom))
	}
	if kustolang.GlobalStateGetProperty(kustolang.DefaultGlobalState, custom) != 7 {
		t.Fatal("GlobalStateWithProperty mutated the source state")
	}
	if same := kustolang.GlobalStateWithProperty(changed, custom, 42); same != changed {
		t.Fatal("assigning the existing property value did not preserve identity")
	}
}

type publicCachedValue struct {
	value int
}

func TestPublicKustoCacheCreatesOneValuePerTypeConcurrently(t *testing.T) {
	cache := kustolang.NewKustoCache(kustolang.DefaultGlobalState)
	var calls atomic.Int32
	values := make(chan *publicCachedValue, 32)
	var group sync.WaitGroup
	for range 32 {
		group.Add(1)
		go func() {
			defer group.Done()
			values <- kustolang.KustoCacheGetOrCreate(cache, func() *publicCachedValue {
				calls.Add(1)
				return &publicCachedValue{value: 42}
			})
		}()
	}
	group.Wait()
	close(values)

	var first *publicCachedValue
	for value := range values {
		if first == nil {
			first = value
		}
		if value != first || value.value != 42 {
			t.Fatalf("cache value = %p/%d, first = %p", value, value.value, first)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("creator calls = %d, want 1", calls.Load())
	}
	if got, ok := kustolang.KustoCacheTryGetValue[*publicCachedValue](cache); !ok || got != first {
		t.Fatalf("TryGetValue = %p/%t, want %p/true", got, ok, first)
	}
	if cache.WithGlobals(kustolang.DefaultGlobalState) != cache {
		t.Fatal("WithGlobals replaced cache for identical globals")
	}
	if replacement := cache.WithGlobals(kustolang.NewGlobalState(nil)); replacement == cache {
		t.Fatal("WithGlobals reused cache for different globals")
	}
}

func TestPublicFunctionBodyFactsMatchImmutableUpstreamBehavior(t *testing.T) {
	parameter := kustolang.NewParameter("value", kustolang.ScalarTypes.String.Type())
	facts := kustolang.FunctionBodyFactsDefault.
		AddDependentParameter(nil).
		AddDependentParameter(parameter).
		AddUnqualifiedTableName("").
		AddUnqualifiedTableName("Events").
		WithHasClusterCall(true).
		WithNonVariableReturnType(kustolang.ScalarTypes.String.Type())

	if facts == kustolang.FunctionBodyFactsDefault {
		t.Fatal("function facts mutation reused the default value")
	}
	dependentParameters := facts.DependentParameters()
	if len(dependentParameters) != 2 || dependentParameters[0] != nil || dependentParameters[1] != parameter {
		t.Fatalf("dependent parameters = %#v", dependentParameters)
	}
	unqualifiedTableNames := facts.UnqualifiedTableNames()
	if len(unqualifiedTableNames) != 2 || unqualifiedTableNames[0] != "" || unqualifiedTableNames[1] != "Events" {
		t.Fatalf("unqualified table names = %#v", unqualifiedTableNames)
	}
	if !facts.HasVariableReturnType() || !facts.HasClusterCall() || !facts.IsInteresting() {
		t.Fatalf("facts flags = variable:%t cluster:%t interesting:%t", facts.HasVariableReturnType(), facts.HasClusterCall(), facts.IsInteresting())
	}
	if facts.NonVariableReturnType() != kustolang.ScalarTypes.String.Type() {
		t.Fatalf("non-variable return type = %#v", facts.NonVariableReturnType())
	}
	if same := facts.AddDependentParameter(parameter); same != facts {
		t.Fatal("adding a duplicate dependent parameter did not preserve identity")
	}
	if len(kustolang.FunctionBodyFactsDefault.DependentParameters()) != 0 || len(kustolang.FunctionBodyFactsDefault.UnqualifiedTableNames()) != 0 {
		t.Fatal("function-facts operations mutated the default value")
	}
	dependentParameters[0] = parameter
	unqualifiedTableNames[0] = "mutated"
	if facts.DependentParameters()[0] != nil || facts.UnqualifiedTableNames()[0] != "" {
		t.Fatal("function facts exposed mutable list storage")
	}
}
