package kustolang

import (
	"fmt"
	"sync"
	"testing"
)

func TestKustoCacheWithGlobalsAndTypedValues(t *testing.T) {
	globals := NewGlobalState(NewDatabaseSymbol("db"))
	cached := globals.WithCache()
	if cached == globals || cached.Cache == nil || cached.Cache.Globals != cached {
		t.Fatalf("WithCache did not create cache bound to clone")
	}
	if cached.WithCache() != cached {
		t.Fatal("WithCache should preserve identity when cache already exists")
	}

	first := KustoCacheGetOrCreate(cached.Cache, func() *GlobalBindingCache {
		return NewGlobalBindingCache()
	})
	second := KustoCacheGetOrCreate(cached.Cache, func() *GlobalBindingCache {
		t.Fatal("creator should not run for cached type")
		return nil
	})
	if first == nil || second != first {
		t.Fatalf("typed cache returned %#v then %#v", first, second)
	}
	if got, ok := KustoCacheTryGetValue[*GlobalBindingCache](cached.Cache); !ok || got != first {
		t.Fatalf("TryGetValue = %#v, %t", got, ok)
	}

	rebound := cached.WithDatabase(NewDatabaseSymbol("other"))
	if rebound == cached || rebound.Cache == nil || rebound.Cache == cached.Cache || rebound.Cache.Globals != rebound {
		t.Fatalf("cache should be rebound and cleared on changed globals")
	}
	if _, ok := KustoCacheTryGetValue[*GlobalBindingCache](rebound.Cache); ok {
		t.Fatal("rebinding cache should not retain old typed entries")
	}
}

func TestMostRecentlyUsedCacheMirrorsUpstreamRecencyAndExistingValue(t *testing.T) {
	value := NewParameter("value", ScalarTypes.String.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), value)
	NewFunctionSymbol("Resolve", signature)

	first := NewCallSiteInfo(signature, []*Parameter{value}, []any{"first"})
	second := NewCallSiteInfo(signature, []*Parameter{value}, []any{"second"})
	third := NewCallSiteInfo(signature, []*Parameter{value}, []any{"third"})

	cache := NewMostRecentlyUsedCache[CallSiteInfo, string](2)
	cache.AddOrUpdate(first, "one")
	cache.AddOrUpdate(second, "two")
	if got, ok := cache.TryGetValue(first); !ok || got != "one" {
		t.Fatalf("first lookup = %q, %t", got, ok)
	}
	cache.AddOrUpdate(third, "three")
	if _, ok := cache.TryGetValue(second); ok {
		t.Fatal("least recently used entry should be evicted")
	}
	if got, ok := cache.TryGetValue(first); !ok || got != "one" {
		t.Fatalf("recent first entry = %q, %t", got, ok)
	}
	cache.AddOrUpdate(first, "replacement")
	if got, ok := cache.TryGetValue(first); !ok || got != "one" {
		t.Fatalf("upstream AddOrUpdate keeps existing value, got %q, %t", got, ok)
	}
}

func TestBindingCachesUseUpstreamKeySemantics(t *testing.T) {
	value := NewParameter("value", ScalarTypes.String.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), value)
	NewFunctionSymbol("Resolve", signature)
	callSite := NewCallSiteInfo(signature, []*Parameter{value}, []any{"alice"})

	local := NewLocalBindingCache()
	local.CallSiteToResultTypeMap.Add(callSite, ScalarTypes.String.Type())
	if got, ok := local.CallSiteToResultTypeMap.TryGetValue(callSite); !ok || got != ScalarTypes.String.Type() {
		t.Fatalf("local callsite result = %#v, %t", got, ok)
	}
	local.SignaturesComputingExpansion[signature] = true
	local.FunctionExpansionCounts[signature]++
	if !local.SignaturesComputingExpansion[signature] || local.FunctionExpansionCounts[signature] != 1 {
		t.Fatalf("local signature state = %#v / %#v", local.SignaturesComputingExpansion, local.FunctionExpansionCounts)
	}

	global := NewGlobalBindingCache()
	tableA := NewTableSymbol("A", NewColumnSymbol("x", ScalarTypes.Long.Type()))
	tableB := NewTableSymbol("B", NewColumnSymbol("x", ScalarTypes.Long.Type()))
	result := NewTableSymbol("Result", NewColumnSymbol("x", ScalarTypes.Long.Type()))
	global.CommonColumnsMap.AddOrUpdate([]*TableSymbol{tableA, tableB}, result)
	if got, ok := global.CommonColumnsMap.TryGetValue([]*TableSymbol{tableA, tableB}); !ok || got != result {
		t.Fatalf("table-list cache = %#v, %t", got, ok)
	}
	if _, ok := global.CommonColumnsMap.TryGetValue([]*TableSymbol{tableB, tableA}); ok {
		t.Fatal("table-list cache should use ordered table identity")
	}
	global.DatabaseFunctionBodyFacts[signature] = FunctionBodyFactsDefault.WithHasSyntaxErrors(true)
	if facts := global.DatabaseFunctionBodyFacts[signature]; facts == nil || !facts.HasSyntaxErrors() {
		t.Fatalf("database function facts = %#v", facts)
	}
}

func TestBindingCacheHelpersRouteDatabaseSignaturesGlobally(t *testing.T) {
	value := NewParameter("value", ScalarTypes.String.Type())
	dbFunction := NewComputedFunctionSymbol("Resolve", "T", TabularityTabular, value)
	dbSignature := dbFunction.Signatures()[0]
	globals := GlobalStateWithProperty(
		NewGlobalState(NewDatabaseSymbol("db", dbFunction)),
		Properties.MaxCachedResultTypes,
		2,
	)
	global := NewGlobalBindingCache()
	local := NewLocalBindingCache()

	if !IsDatabaseSymbolSignature(globals, dbSignature) {
		t.Fatal("database computed function should be a database-symbol signature")
	}
	facts := FunctionBodyFactsDefault.AddDependentParameter(value)
	SetFunctionBodyFacts(globals, global, local, dbSignature, facts)
	if got, ok := TryGetFunctionBodyFacts(global, local, dbSignature); !ok || got != facts {
		t.Fatalf("facts = %#v, %t", got, ok)
	}
	if len(global.DatabaseFunctionBodyFacts) != 1 || len(local.NonDatabaseFunctionBodyFacts) != 0 {
		t.Fatalf("facts should be routed globally, global=%d local=%d", len(global.DatabaseFunctionBodyFacts), len(local.NonDatabaseFunctionBodyFacts))
	}

	first := NewCallSiteInfo(dbSignature, []*Parameter{value}, []any{"first"})
	second := NewCallSiteInfo(dbSignature, []*Parameter{value}, []any{"second"})
	third := NewCallSiteInfo(dbSignature, []*Parameter{value}, []any{"third"})
	AddResultTypeToCache(globals, global, local, first, ScalarTypes.String.Type())
	AddResultTypeToCache(globals, global, local, second, ScalarTypes.Long.Type())
	if got, ok := TryGetResultTypeFromCache(global, local, first); !ok || got != ScalarTypes.String.Type() {
		t.Fatalf("first result type = %#v, %t", got, ok)
	}
	AddResultTypeToCache(globals, global, local, third, ScalarTypes.Bool.Type())
	if _, ok := TryGetResultTypeFromCache(global, local, second); ok {
		t.Fatal("global result-type cache should evict the least recently used call site")
	}
	if local.CallSiteToResultTypeMap.Count() != 0 {
		t.Fatalf("database result type should not be cached locally, count=%d", local.CallSiteToResultTypeMap.Count())
	}
}

func TestBindingCacheHelpersRouteNonDatabaseSignaturesLocally(t *testing.T) {
	value := NewParameter("value", ScalarTypes.String.Type())
	localFunction := NewComputedFunctionSymbol("LocalResolve", "T", TabularityTabular, value)
	localSignature := localFunction.Signatures()[0]
	globals := NewGlobalState(NewDatabaseSymbol("db"))
	global := NewGlobalBindingCache()
	local := NewLocalBindingCache()

	if IsDatabaseSymbolSignature(globals, localSignature) {
		t.Fatal("function outside the database should not be a database-symbol signature")
	}
	facts := FunctionBodyFactsDefault.WithHasSyntaxErrors(true)
	SetFunctionBodyFacts(globals, global, local, localSignature, facts)
	if got, ok := TryGetFunctionBodyFacts(global, local, localSignature); !ok || got != facts {
		t.Fatalf("facts = %#v, %t", got, ok)
	}
	if len(global.DatabaseFunctionBodyFacts) != 0 || len(local.NonDatabaseFunctionBodyFacts) != 1 {
		t.Fatalf("facts should be routed locally, global=%d local=%d", len(global.DatabaseFunctionBodyFacts), len(local.NonDatabaseFunctionBodyFacts))
	}

	callSite := NewCallSiteInfo(localSignature, []*Parameter{value}, []any{"first"})
	expansion := NewFunctionCallExpansion(NewSyntaxNode(SyntaxKindNameReference, nil), nil, 0)
	AddExpansionToCache(globals, global, local, callSite, expansion)
	if got, ok := TryGetExpansionFromCache(global, local, callSite); !ok || got != expansion {
		t.Fatalf("local expansion = %#v, %t", got, ok)
	}
	if len(global.CallSiteToExpansionMap) != 0 || local.CallSiteToExpansionMap.Count() != 1 {
		t.Fatalf("expansion should be routed locally, global=%d local=%d", len(global.CallSiteToExpansionMap), local.CallSiteToExpansionMap.Count())
	}
}

func TestBindingCachePolicyRejectsLocalTabularVariableCases(t *testing.T) {
	tableName := NewParameter("tableName", ScalarTypes.String.Type())
	other := NewParameter("other", ScalarTypes.String.Type())
	dbFunction := NewComputedFunctionSymbol("ExpandTable", "table(tableName)", TabularityTabular, tableName, other)
	signature := dbFunction.Signatures()[0]
	globals := NewGlobalState(NewDatabaseSymbol("db", dbFunction))
	global := NewGlobalBindingCache()
	local := NewLocalBindingCache()
	scope := NewLocalScope(nil)
	scope.AddSymbol(NewVariableSymbol("LocalRows", NewTableSymbol("LocalRows", NewColumnSymbol("x", ScalarTypes.Long.Type())), false, nil, nil))

	resultFacts := FunctionBodyFactsDefault.
		WithHasUnqualifiedTableCall(true).
		AddDependentParameter(tableName)
	SetFunctionBodyFacts(globals, global, local, signature, resultFacts)
	localCallSite := NewCallSiteInfo(signature, []*Parameter{tableName, other}, []any{"LocalRows", "safe"})
	if CanCacheResultType(scope, global, local, localCallSite) {
		t.Fatal("result-type caching should reject local tabular variable names in call-site values")
	}
	if CanCacheExpansion(scope, global, local, localCallSite) {
		t.Fatal("expansion caching should reject dependent parameter values that name local tabular variables")
	}

	safeResultCallSite := NewCallSiteInfo(signature, []*Parameter{tableName, other}, []any{"RemoteRows", "OtherRows"})
	if !CanCacheResultType(scope, global, local, safeResultCallSite) {
		t.Fatal("result-type caching should allow non-local table names")
	}
	nonDependentLocalCallSite := NewCallSiteInfo(signature, []*Parameter{tableName, other}, []any{"RemoteRows", "LocalRows"})
	if !CanCacheExpansion(scope, global, local, nonDependentLocalCallSite) {
		t.Fatal("expansion caching should ignore non-dependent string values")
	}

	localNameFacts := FunctionBodyFactsDefault.
		WithHasUnqualifiedTableCall(true).
		WithUnqualifiedTableNames([]string{"LocalRows"})
	SetFunctionBodyFacts(globals, global, local, signature, localNameFacts)
	if CanCacheResultType(scope, global, local, safeResultCallSite) || CanCacheExpansion(scope, global, local, nonDependentLocalCallSite) {
		t.Fatal("caching should reject function bodies with unqualified calls to local tabular variables")
	}
}

func TestGlobalBindingCacheSupportsConcurrentDatabaseFunctionCalls(t *testing.T) {
	value := NewParameter("value", ScalarTypes.String.Type())
	databaseFunction := NewComputedFunctionSymbol("Resolve", "T", TabularityTabular, value)
	signature := databaseFunction.Signatures()[0]
	globals := GlobalStateWithProperty(
		NewGlobalState(NewDatabaseSymbol("db", databaseFunction)),
		Properties.MaxCachedResultTypes,
		64,
	)
	globals = GlobalStateWithProperty(globals, Properties.MaxCachedExpansions, 64)
	global := NewGlobalBindingCache()

	const workers = 32
	var group sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		group.Add(1)
		go func(worker int) {
			defer group.Done()
			local := NewLocalBindingCache()
			SetFunctionBodyFacts(globals, global, local, signature, FunctionBodyFactsDefault.AddDependentParameter(value))
			callSite := NewCallSiteInfo(signature, []*Parameter{value}, []any{fmt.Sprintf("value-%d", worker)})
			AddResultTypeToCache(globals, global, local, callSite, ScalarTypes.String.Type())
			if result, ok := TryGetResultTypeFromCache(global, local, callSite); !ok || result != ScalarTypes.String.Type() {
				t.Errorf("worker %d result type = %#v, %t", worker, result, ok)
			}
			resultSymbol := NewTableSymbol(fmt.Sprintf("Result-%d", worker), NewColumnSymbol("A", ScalarTypes.String.Type()))
			AddResultSymbolToCache(globals, global, local, callSite, resultSymbol)
			if result, ok := TryGetResultSymbolFromCache(global, local, callSite); !ok || result != resultSymbol {
				t.Errorf("worker %d result symbol = %#v, %t", worker, result, ok)
			}
			expansion := NewFunctionCallExpansion(NewSyntaxNode(SyntaxKindNameReference, nil), nil, worker)
			AddExpansionToCache(globals, global, local, callSite, expansion)
			if result, ok := TryGetExpansionFromCache(global, local, callSite); !ok || result == nil {
				t.Errorf("worker %d expansion = %#v, %t", worker, result, ok)
			}
		}(worker)
	}
	group.Wait()

	if facts, ok := TryGetFunctionBodyFacts(global, nil, signature); !ok || facts == nil {
		t.Fatalf("database function facts = %#v, %t", facts, ok)
	}
	if cache, ok := global.tryGetResultTypeCache(signature); !ok || cache.Count() == 0 || cache.Count() > 64 {
		t.Fatalf("result-type cache = %#v, %t", cache, ok)
	}
	if cache, ok := global.tryGetResultSymbolCache(signature); !ok || cache.Count() == 0 || cache.Count() > 64 {
		t.Fatalf("result-symbol cache = %#v, %t", cache, ok)
	}
	if cache, ok := global.tryGetExpansionCache(signature); !ok || cache.Count() == 0 || cache.Count() > 64 {
		t.Fatalf("expansion cache = %#v, %t", cache, ok)
	}
}
