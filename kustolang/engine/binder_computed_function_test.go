package kustolang

import (
	"sync"
	"testing"
)

func TestNativeComputedFunctionExpansionBindsOwnerContextAndResult(t *testing.T) {
	remoteTable := NewTableSymbol("RemoteTable", NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	remoteFunction := NewComputedFunctionSymbol("RemoteRows", `RemoteTable | project Value`, TabularityTabular)
	remoteDatabase := NewDatabaseSymbol("Remote", remoteTable, remoteFunction)
	currentDatabase := NewDatabaseSymbol("Current")
	currentCluster := NewClusterSymbol("current.kusto.windows.net", currentDatabase)
	remoteCluster := NewClusterSymbol("remote.kusto.windows.net", remoteDatabase)
	globals := NewGlobalState(currentDatabase).
		WithClusterList([]*ClusterSymbol{currentCluster, remoteCluster}).
		WithCluster(currentCluster).
		WithDatabase(currentDatabase)

	query := ParseAndAnalyze(`cluster("remote").database("Remote").RemoteRows() | project Value`, globals)
	if diagnostics := query.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("computed function diagnostics = %#v", diagnostics)
	}
	call := namedFunctionCall(t, query.Syntax(), "RemoteRows")
	if call.GetCalledFunctionBody() == nil || call.GetCalledFunctionFacts() == nil {
		t.Fatal("computed function call did not retain its bound expansion")
	}
	project := firstBoundNode[*ProjectOperator](t, query.Syntax())
	if table := semanticTable(t, project); table.GetColumn("Value") == nil || table.GetColumn("Value").Type != ScalarTypes.Long.Type() {
		t.Fatalf("computed function downstream result = %#v", table)
	}
}

func TestNativeComputedFunctionExpansionBindsConstantParametersAndStopsRecursion(t *testing.T) {
	parameter := NewParameter("databaseName", ScalarTypes.String.Type())
	function := NewComputedFunctionSymbol("Rows", `database(databaseName).T`, TabularityTabular, parameter)
	database := NewDatabaseSymbol("db", NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type())), function)
	globals := NewGlobalState(database)

	code := ParseAndAnalyze(`Rows("db")`, globals)
	call := namedFunctionCall(t, code.Syntax(), "Rows")
	body := call.GetCalledFunctionBody()
	if body == nil {
		t.Fatal("parameterized computed function has no expansion")
	}
	var databaseCall *FunctionCallExpression
	walkSyntaxWithTriviaStartAndAncestors(body, 0, func(element SyntaxElement, _ int, _ []SyntaxElement) {
		if candidate, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression); ok && NameReferenceSimpleName(candidate.Name()) == "database" {
			databaseCall = candidate
		}
	})
	if databaseCall == nil {
		t.Fatal("expanded database call not found")
	}
	argument := databaseCall.ArgumentList().Expressions().At(0).Element()
	if !ExpressionIsConstant(argument) || ExpressionConstantValue(argument) != "db" {
		t.Fatalf("expanded parameter semantics = constant:%t value:%#v", ExpressionIsConstant(argument), ExpressionConstantValue(argument))
	}

	loop := NewComputedFunctionSymbol("Loop", `Loop()`, TabularityTabular)
	loopGlobals := NewGlobalState(NewDatabaseSymbol("db", loop))
	loopCode := ParseAndAnalyze(`Loop()`, loopGlobals)
	loopCall := namedFunctionCall(t, loopCode.Syntax(), "Loop")
	if loopCall.GetCalledFunctionBody() == nil {
		t.Fatal("outer recursive call should retain one bounded expansion")
	}
}

func TestNativeComputedFunctionCustomReturnContextCarriesCurrentFunction(t *testing.T) {
	var current *FunctionSymbol
	capture := NewFunctionSymbol("capture", NewCustomSignature(func(context CustomReturnTypeContext) Symbol {
		current = context.CurrentFunction
		return ScalarTypes.Long.Type()
	}, TabularityScalar))
	outer := NewComputedFunctionSymbol("Outer", `print x=capture()`, TabularityTabular)
	globals := NewGlobalState(NewDatabaseSymbol("db", capture, outer))
	code := ParseAndAnalyze(`Outer()`, globals)
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("computed custom-return diagnostics = %#v", diagnostics)
	}
	if current != outer {
		t.Fatalf("custom-return current function = %#v, want Outer", current)
	}
}

func TestNativeComputedFunctionConcurrentFactsCacheAndMalformedCalls(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	function := NewComputedFunctionSymbol("Rows", `T | project A`, TabularityTabular)
	globals := NewGlobalState(NewDatabaseSymbol("db", table, function))
	globals = globals.WithParseOptions(globals.ParseOptions.WithAlwaysProduceEndTokens(true)).WithCache()

	tests := []struct {
		query           string
		wantDiagnostics bool
	}{
		{query: `Rows()`},
		{query: `Rows(`, wantDiagnostics: true},
	}

	const workers = 24
	var group sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		for _, test := range tests {
			group.Add(1)
			go func(worker int, testCase struct {
				query           string
				wantDiagnostics bool
			}) {
				defer group.Done()
				code := ParseAndAnalyze(testCase.query, globals)
				diagnostics := code.GetDiagnostics()
				if (len(diagnostics) != 0) != testCase.wantDiagnostics {
					t.Errorf("worker %d %q diagnostics = %#v", worker, testCase.query, diagnostics)
				}
				if !testCase.wantDiagnostics {
					call := namedFunctionCall(t, code.Syntax(), "Rows")
					if call.GetCalledFunctionFacts() == nil {
						t.Errorf("worker %d valid call did not retain function facts", worker)
					}
				}
			}(worker, test)
		}
	}
	group.Wait()

	cache, ok := KustoCacheTryGetValue[*GlobalBindingCache](globals.Cache)
	if !ok || cache == nil {
		t.Fatal("computed binding did not create a global binding cache")
	}
	if _, ok := cache.tryGetDatabaseFunctionBodyFacts(function.Signatures()[0]); !ok {
		t.Fatal("computed binding did not cache database function facts")
	}
	first := ParseAndAnalyze(`Rows()`, globals)
	second := ParseAndAnalyze(`Rows()`, globals)
	firstCall := namedFunctionCall(t, first.Syntax(), "Rows")
	secondCall := namedFunctionCall(t, second.Syntax(), "Rows")
	if firstCall.GetCalledFunctionBody() == nil || firstCall.GetCalledFunctionBody() != secondCall.GetCalledFunctionBody() {
		t.Fatal("repeated database function calls did not reuse their cached expansion")
	}
	if expansion, ok := cache.tryGetExpansionCache(function.Signatures()[0]); !ok || expansion.Count() != 1 {
		t.Fatalf("function expansion cache = %#v, %t", expansion, ok)
	}
}

func TestNativeComputedFunctionCachesTableResultWithDeferredExpansion(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	otherTable := NewTableSymbol("U", NewColumnSymbol("B", ScalarTypes.Long.Type()))
	tableName := NewParameter("tableName", ScalarTypes.String.Type())
	function := NewComputedFunctionSymbol("Rows", `table(tableName)`, TabularityTabular, tableName)
	globals := NewGlobalState(NewDatabaseSymbol("db", table, otherTable, function))
	globals = GlobalStateWithProperty(globals, Properties.MaxCachedResultTypes, 2)
	globals = GlobalStateWithProperty(globals, Properties.MaxCachedExpansions, 0)
	globals = globals.WithParseOptions(globals.ParseOptions.WithAlwaysProduceEndTokens(true)).WithCache()

	first := ParseAndAnalyze(`Rows("T") | project A`, globals)
	if diagnostics := first.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("first call diagnostics = %#v", diagnostics)
	}
	firstCall := namedFunctionCall(t, first.Syntax(), "Rows")
	firstResult, ok := firstCall.SyntaxNode.Payload(functionCallResultSymbolPayload)
	if !ok {
		t.Fatal("first call did not retain its result symbol")
	}
	firstTable, ok := firstResult.(*TableSymbol)
	if !ok || firstTable.GetColumn("A") == nil || firstTable.GetColumn("A").Type != ScalarTypes.String.Type() {
		t.Fatalf("first computed result = %#v", firstResult)
	}

	different := ParseAndAnalyze(`Rows("U") | project B`, globals)
	if diagnostics := different.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("different argument diagnostics = %#v", diagnostics)
	}
	differentCall := namedFunctionCall(t, different.Syntax(), "Rows")
	differentResult, ok := differentCall.SyntaxNode.Payload(functionCallResultSymbolPayload)
	differentTable, isTable := differentResult.(*TableSymbol)
	if !ok || !isTable || differentResult == firstResult || differentTable.GetColumn("B") == nil || differentTable.GetColumn("B").Type != ScalarTypes.Long.Type() {
		t.Fatalf("different computed result = %#v, want separate U table schema", differentResult)
	}

	second := ParseAndAnalyze(`Rows("T") | project A`, globals)
	if diagnostics := second.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("cached call diagnostics = %#v", diagnostics)
	}
	secondCall := namedFunctionCall(t, second.Syntax(), "Rows")
	secondResult, ok := secondCall.SyntaxNode.Payload(functionCallResultSymbolPayload)
	if !ok || secondResult != firstResult {
		t.Fatalf("cached result = %#v, want original table symbol %#v", secondResult, firstResult)
	}
	info, ok := secondCall.GetSemanticInfo()
	if !ok || info.CalledFunctionInfo == nil || info.CalledFunctionInfo.expansion != nil {
		t.Fatalf("cached call should defer expansion, semantic=%#v", info)
	}
	project := firstBoundNode[*ProjectOperator](t, second.Syntax())
	if result := semanticTable(t, project); result.GetColumn("A") == nil || result.GetColumn("A").Type != ScalarTypes.String.Type() {
		t.Fatalf("cached downstream result = %#v", result)
	}
	if secondCall.GetCalledFunctionBody() == nil {
		t.Fatal("deferred computed expansion did not bind on demand")
	}

	cache, ok := KustoCacheTryGetValue[*GlobalBindingCache](globals.Cache)
	if !ok || cache == nil {
		t.Fatal("computed binding did not create a global binding cache")
	}
	if resultCache, ok := cache.tryGetResultSymbolCache(function.Signatures()[0]); !ok || resultCache.Count() != 2 {
		t.Fatalf("result-symbol cache = %#v, %t", resultCache, ok)
	}

	malformed := ParseAndAnalyze(`Rows(`, globals)
	if diagnostics := malformed.GetDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("malformed computed-function call should retain parse diagnostics")
	}
}
