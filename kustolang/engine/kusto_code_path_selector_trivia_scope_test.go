package kustolang

import (
	"strings"
	"testing"
)

func TestKustoCodePathScopeStartsAtSelectorTriviaProfiles(t *testing.T) {
	globals, currentTable, sampleTable := directiveContextGlobals()
	const text = `database("Samples").  SampleTable | take 1`
	dot := strings.Index(text, ".")
	selectorTrivia := dot + 1
	selectorText := strings.Index(text, "SampleTable")

	for _, profile := range kustoCodePathClearingProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			code := profile.parse(text, globals)
			if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid qualified path diagnostics = %#v", diagnostics)
			}
			assertScopeNames(t,
				code.GetSymbolsInScope(dot, SymbolMatchTable, IncludeFunctionKindNone),
				[]string{currentTable.SymbolName()},
			)
			for _, position := range []int{selectorTrivia, selectorText} {
				assertScopeNames(t,
					code.GetSymbolsInScope(position, SymbolMatchTable, IncludeFunctionKindNone),
					[]string{sampleTable.SymbolName()},
				)
			}
			items := NewKustoCodeServiceForCode(code).GetCompletionItems(selectorText, CompletionOptionsDefault, CancellationToken{}).Items
			requireCompletionKind(t, items, sampleTable.SymbolName(), CompletionKindTable)
			requireCompletionKind(t, items, "SampleFn()", CompletionKindDatabaseFunction)
			requireCompletionKind(t, items, CompletionDisplayText(Functions.ExternalTable), CompletionKindBuiltInFunction)
			assertCompletionExcludes(t, items, currentTable.SymbolName(), "CurrentFn()")
		})
	}
}

func TestKustoCodeMissingPathSelectorRetainsDiagnosticsAndTriviaScopeProfiles(t *testing.T) {
	globals, currentTable, sampleTable := directiveContextGlobals()
	const text = `database("Samples"). `

	for _, profile := range kustoCodePathClearingProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			code := profile.parse(text, globals)
			if diagnostics := code.GetDiagnostics(); len(diagnostics) == 0 {
				t.Fatal("missing path selector should retain syntax diagnostics")
			}
			assertScopeNames(t,
				code.GetSymbolsInScope(len(text), SymbolMatchTable, IncludeFunctionKindNone),
				[]string{sampleTable.SymbolName()},
			)
			items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
			requireCompletionKind(t, items, sampleTable.SymbolName(), CompletionKindTable)
			requireCompletionKind(t, items, "SampleFn()", CompletionKindDatabaseFunction)
			assertCompletionExcludes(t, items, currentTable.SymbolName(), "CurrentFn()")
		})
	}
}

func TestKustoCodeNestedEntityPathScopeReconstructsAfterRollbackProfiles(t *testing.T) {
	globals, currentTable, remoteTable := nestedPathScopeGlobals()
	const validText = `cluster("remote").database("RemoteDB"). RemoteTable | take 1`
	validPosition := strings.Index(validText, "RemoteTable")
	const malformedText = `cluster("remote").database("RemoteDB"). `

	for _, profile := range kustoCodePathClearingProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			valid := profile.parse(validText, globals)
			if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid nested path diagnostics = %#v", diagnostics)
			}
			assertScopeNames(t,
				valid.GetSymbolsInScope(validPosition, SymbolMatchTable, IncludeFunctionKindNone),
				[]string{remoteTable.SymbolName()},
			)
			validItems := NewKustoCodeServiceForCode(valid).GetCompletionItems(validPosition, CompletionOptionsDefault, CancellationToken{}).Items
			requireCompletionKind(t, validItems, remoteTable.SymbolName(), CompletionKindTable)
			requireCompletionKind(t, validItems, "RemoteFn()", CompletionKindDatabaseFunction)
			assertCompletionExcludes(t, validItems, currentTable.SymbolName())

			malformed := profile.parse(malformedText, globals)
			if diagnostics := malformed.GetDiagnostics(); len(diagnostics) == 0 {
				t.Fatal("missing nested path selector should retain syntax diagnostics")
			}
			assertScopeNames(t,
				malformed.GetSymbolsInScope(len(malformedText), SymbolMatchTable, IncludeFunctionKindNone),
				[]string{remoteTable.SymbolName()},
			)
			items := NewKustoCodeServiceForCode(malformed).GetCompletionItems(len(malformedText), CompletionOptionsDefault, CancellationToken{}).Items
			requireCompletionKind(t, items, remoteTable.SymbolName(), CompletionKindTable)
			requireCompletionKind(t, items, "RemoteFn()", CompletionKindDatabaseFunction)
			assertCompletionExcludes(t, items, currentTable.SymbolName())
		})
	}
}

func TestKustoCodeMissingClusterSelectorOffersDatabaseAccessorProfiles(t *testing.T) {
	globals, _, _ := nestedPathScopeGlobals()
	const text = `cluster("remote"). `

	for _, profile := range kustoCodePathClearingProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			code := profile.parse(text, globals)
			if diagnostics := code.GetDiagnostics(); len(diagnostics) == 0 {
				t.Fatal("missing cluster selector should retain syntax diagnostics")
			}
			assertScopeNames(t, code.GetSymbolsInScope(len(text), SymbolMatchDatabase, IncludeFunctionKindNone), []string{})
			items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
			requireCompletionKind(t, items, CompletionDisplayText(Functions.Database), CompletionKindBuiltInFunction)
			assertCompletionExcludes(t, items, "RemoteDB", "CurrentDB")
		})
	}
}

func TestKustoCodeMissingTableResultSelectorSuppressesColumnsProfiles(t *testing.T) {
	globals, currentTable, sampleTable := directiveContextGlobals()
	const text = `database("Samples").table("SampleTable"). `

	for _, profile := range kustoCodePathClearingProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			code := profile.parse(text, globals)
			if diagnostics := code.GetDiagnostics(); len(diagnostics) == 0 {
				t.Fatal("missing table-result selector should retain syntax diagnostics")
			}
			assertScopeNames(t, code.GetSymbolsInScope(len(text), SymbolMatchColumn, IncludeFunctionKindNone), []string{})
			items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionExcludes(t, items, "SampleColumn", "CurrentColumn", currentTable.SymbolName(), sampleTable.SymbolName())
		})
	}
}

func TestKustoCodeDatabasePathUsesMembersAndAccessorFunctions(t *testing.T) {
	table := NewTableSymbol("Events", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	external := NewExternalTableSymbol("Archive", NewColumnSymbol("Archived", ScalarTypes.Bool.Type()))
	view := NewMaterializedViewSymbol("Daily", []*ColumnSymbol{NewColumnSymbol("Day", ScalarTypes.DateTime.Type())}, "")
	stored := NewStoredQueryResultSymbol("Saved", NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	group := NewEntityGroupSymbol("Targets", table)
	graph := NewGraphModelSymbol("SecurityGraph", nil, nil, nil)
	database := NewDatabaseSymbol("db", table, external, view, stored, group, graph)
	globals := NewGlobalState(database)
	const text = `database("db"). `
	code := ParseAndAnalyze(text, globals)
	if len(code.GetDiagnostics()) == 0 {
		t.Fatal("missing database selector produced no diagnostics")
	}
	match := SymbolMatchTable | SymbolMatchExternalTable | SymbolMatchMaterializedView |
		SymbolMatchStoredQueryResult | SymbolMatchEntityGroup | SymbolMatchGraph | SymbolMatchFunction
	assertScopeNames(t, code.GetSymbolsInScope(len(text), match, IncludeFunctionKindNone), []string{
		"Events", "Daily", "Targets", "table", "external_table", "materialized_view", "entity_group", "stored_query_result", "graph",
	})

	items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, items, "Events", CompletionKindTable)
	requireCompletionKind(t, items, "Daily", CompletionKindMaterialiedView)
	requireCompletionKind(t, items, "Targets", CompletionKindEntityGroup)
	for _, function := range []*FunctionSymbol{
		Functions.Table, Functions.ExternalTable, Functions.MaterializedView,
		Functions.EntityGroup, Functions.StoredQueryResult, Functions.Graph,
	} {
		requireCompletionKind(t, items, CompletionDisplayText(function), CompletionKindBuiltInFunction)
	}
	assertCompletionExcludes(t, items, "Archive", "Saved", "SecurityGraph")
}

func TestKustoCodeMacroExpandPathIncludesSpecialMembers(t *testing.T) {
	const text = `macro-expand Targets as Scope (Scope. )`
	code := ParseAndAnalyze(text, kustoCodeOperatorLocalGlobals())
	if len(code.GetDiagnostics()) == 0 {
		t.Fatal("missing macro-expand path selector produced no diagnostics")
	}
	position := strings.Index(text, ". ") + 2
	for _, name := range []string{"$current_database", "$current_cluster_endpoint"} {
		if symbol := symbolInScope(code, position, name); symbol == nil {
			t.Fatalf("macro-expand path is missing %q", name)
		}
	}
	items := NewKustoCodeServiceForCode(code).GetCompletionItems(position, CompletionOptionsDefault, CancellationToken{}).Items
	assertCompletionIncludes(t, items, "$current_database", "$current_cluster_endpoint")
}

func TestKustoCodePathScopePassThroughGroupAggregatesDatabaseMembers(t *testing.T) {
	first := NewDatabaseSymbol("first", NewTableSymbol("FirstTable"))
	second := NewDatabaseSymbol("second", NewTableSymbol("SecondTable"))
	code := ParseAndAnalyze("")
	collector := newKustoCodeScopeCollector(SymbolMatchTable|SymbolMatchFunction, IncludeFunctionKindNone)
	code.addPathSymbolsInScope(0, NewGroupSymbol(first, second), collector)
	assertScopeNames(t, collector.symbols, []string{
		"FirstTable", "table", "external_table", "materialized_view", "entity_group", "stored_query_result", "graph",
		"SecondTable", "table", "external_table", "materialized_view", "entity_group", "stored_query_result", "graph",
	})
}

func TestKustoCodeUnresolvedIntrinsicPathUsesOpenDatabaseScope(t *testing.T) {
	const text = `database("missing"). `
	code := ParseAndAnalyze(text, NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("Events"))))
	if len(code.GetDiagnostics()) == 0 {
		t.Fatal("unresolved intrinsic path produced no diagnostics")
	}
	assertScopeNames(t, code.GetSymbolsInScope(len(text), SymbolMatchAny, IncludeFunctionKindAll), []string{
		"table", "external_table", "materialized_view", "entity_group", "stored_query_result", "graph",
	})
	items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	assertCompletionIncludes(t, items, "table(name, [query_data_scope])")
	assertCompletionExcludes(t, items, "Events")
}

func nestedPathScopeGlobals() (*GlobalState, *TableSymbol, *TableSymbol) {
	currentTable := NewTableSymbol("CurrentTable", NewColumnSymbol("CurrentColumn", ScalarTypes.String.Type()))
	currentDatabase := NewDatabaseSymbol("CurrentDB", currentTable)
	currentCluster := NewClusterSymbol("current.kusto.windows.net", currentDatabase)
	remoteTable := NewTableSymbol("RemoteTable", NewColumnSymbol("RemoteColumn", ScalarTypes.Long.Type()))
	remoteDatabase := NewDatabaseSymbol("RemoteDB", remoteTable, NewComputedFunctionSymbol("RemoteFn", "print x=1", TabularityTabular))
	remoteCluster := NewClusterSymbol("remote.kusto.windows.net", remoteDatabase)
	globals := NewGlobalState(currentDatabase).
		WithClusterList([]*ClusterSymbol{currentCluster, remoteCluster}).
		WithCluster(currentCluster).
		WithDatabase(currentDatabase)
	return globals, currentTable, remoteTable
}
