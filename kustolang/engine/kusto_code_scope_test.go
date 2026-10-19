package kustolang

import (
	"reflect"
	"strings"
	"testing"
)

func TestKustoCodeSymbolsInScopeFollowTopLevelStatementOrder(t *testing.T) {
	text := `declare query_parameters(p:string); let alpha=1; let beta=2; print value=beta`
	code := ParseAndAnalyze(text)
	parameterSeparator := strings.Index(text, ";")
	assertScopeNames(t,
		code.GetSymbolsInScope(parameterSeparator, SymbolMatchLocal, IncludeFunctionKindAll),
		[]string{},
	)
	assertScopeNames(t,
		code.GetSymbolsInScope(parameterSeparator+1, SymbolMatchLocal, IncludeFunctionKindAll),
		[]string{"p"},
	)
	firstSeparator := strings.Index(text[parameterSeparator+1:], ";") + parameterSeparator + 1
	assertScopeNames(t,
		code.GetSymbolsInScope(firstSeparator, SymbolMatchLocal, IncludeFunctionKindAll),
		[]string{"p"},
	)
	assertScopeNames(t,
		code.GetSymbolsInScope(firstSeparator+1, SymbolMatchLocal, IncludeFunctionKindAll),
		[]string{"p", "alpha"},
	)

	insideSecond := strings.Index(text, "beta=2") + len("beta")
	assertScopeNames(t,
		code.GetSymbolsInScope(insideSecond, SymbolMatchLocal, IncludeFunctionKindNone),
		[]string{"p", "alpha"},
	)

	finalExpression := strings.LastIndex(text, "beta")
	assertScopeNames(t,
		code.GetSymbolsInScope(finalExpression, SymbolMatchLocal, IncludeFunctionKindNone),
		[]string{"p", "alpha", "beta"},
	)
}

func TestKustoCodeSymbolsInScopeKeepFunctionScopesLexical(t *testing.T) {
	text := `let outer=1; let f=(p:string) { let inner=2; print a=p, b=inner, c=outer }; print result=outer`
	code := ParseAndAnalyze(text)

	insideInnerDeclaration := strings.Index(text, "inner=2") + len("inner")
	assertScopeNames(t,
		code.GetSymbolsInScope(insideInnerDeclaration, SymbolMatchLocal|SymbolMatchFunction, IncludeFunctionKindLocalFunctions|IncludeFunctionKindLocalViews),
		[]string{"outer", "p"},
	)

	insideBodyExpression := strings.Index(text, "a=p") + len("a=")
	assertScopeNames(t,
		code.GetSymbolsInScope(insideBodyExpression, SymbolMatchLocal|SymbolMatchFunction, IncludeFunctionKindLocalFunctions|IncludeFunctionKindLocalViews),
		[]string{"outer", "p", "inner"},
	)

	outsideFunction := strings.LastIndex(text, "outer")
	assertScopeNames(t,
		code.GetSymbolsInScope(outsideFunction, SymbolMatchLocal|SymbolMatchFunction, IncludeFunctionKindLocalFunctions|IncludeFunctionKindLocalViews),
		[]string{"outer", "f"},
	)
}

func TestKustoCodeSymbolsInScopeIncludeNestedFunctionCapturesWithoutLeaks(t *testing.T) {
	text := `let f=(outer:string) { let before=outer; let g=(inner:long) { print a=outer, b=before, c=inner }; print x=before }; print done=1`
	code := ParseAndAnalyze(text)

	insideNested := strings.Index(text, "a=outer") + len("a=")
	assertScopeNames(t,
		code.GetSymbolsInScope(insideNested, SymbolMatchLocal|SymbolMatchFunction, IncludeFunctionKindLocalFunctions|IncludeFunctionKindLocalViews),
		[]string{"outer", "before", "inner"},
	)

	insideOuterAfterNested := strings.Index(text, "x=before") + len("x=")
	assertScopeNames(t,
		code.GetSymbolsInScope(insideOuterAfterNested, SymbolMatchLocal|SymbolMatchFunction, IncludeFunctionKindLocalFunctions|IncludeFunctionKindLocalViews),
		[]string{"outer", "before", "g"},
	)

	outside := strings.Index(text, "done=1")
	assertScopeNames(t,
		code.GetSymbolsInScope(outside, SymbolMatchLocal|SymbolMatchFunction, IncludeFunctionKindLocalFunctions|IncludeFunctionKindLocalViews),
		[]string{"f"},
	)
}

func TestKustoCodeColumnsInScopeUseBoundPipelineInput(t *testing.T) {
	table := NewTableSymbol("Events",
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("Score", ScalarTypes.Long.Type()),
	)
	text := `Events | project Account | where Account != ""`
	code := ParseAndAnalyze(text, NewGlobalState(NewDatabaseSymbol("db", table)))

	projectPosition := strings.Index(text, "project") + len("project")
	projectInput := code.GetColumnsInScope(projectPosition)
	if projectInput != table {
		t.Fatalf("project input = %#v, want metadata table", projectInput)
	}

	wherePosition := strings.Index(text, "where") + len("where")
	whereInput := code.GetColumnsInScope(wherePosition)
	if whereInput == nil {
		t.Fatal("where input is nil")
	}
	assertScopeColumnNames(t, whereInput, []string{"Account"})
	assertScopeNames(t,
		code.GetSymbolsInScope(wherePosition, SymbolMatchColumn, IncludeFunctionKindNone),
		[]string{"Account"},
	)

	if sourceScope := code.GetColumnsInScope(strings.Index(text, "Events")); sourceScope != nil {
		t.Fatalf("source row scope = %#v, want nil", sourceScope)
	}
}

func TestKustoCodeSymbolsInScopeRespectGlobalMatchesAndFunctionFlags(t *testing.T) {
	table := NewTableSymbol("Events", NewColumnSymbol("A", ScalarTypes.String.Type()))
	external := NewExternalTableSymbol("Archive", NewColumnSymbol("Raw", ScalarTypes.Dynamic.Type()))
	view := NewMaterializedViewSymbol("Daily", []*ColumnSymbol{NewColumnSymbol("Day", ScalarTypes.DateTime.Type())}, "Events | summarize count()")
	databaseFunction := NewComputedFunctionSymbol("DbFn", "print x=1", TabularityTabular)
	databaseView := NewComputedFunctionSymbol("DbView", "Events", TabularityTabular).WithIsView(true)
	database := NewDatabaseSymbol("Primary", table, external, view, databaseFunction, databaseView)
	otherDatabase := NewDatabaseSymbol("Other")
	primaryCluster := NewClusterSymbol("primary.kusto.windows.net", database)
	otherCluster := NewClusterSymbol("other.kusto.windows.net", otherDatabase)
	builtIn := NewScalarFunctionSymbol("BuiltinFn", ScalarTypes.String.Type())
	globals := NewGlobalState(database).WithFunctions([]*FunctionSymbol{builtIn}).WithAggregates(nil).WithPlugIns(nil)
	globals.Clusters = []*ClusterSymbol{primaryCluster, otherCluster}
	globals.Cluster = primaryCluster
	globals.Database = database

	code := ParseAndAnalyze("print x=1", globals)
	metadataMatch := SymbolMatchTable | SymbolMatchExternalTable | SymbolMatchMaterializedView | SymbolMatchDatabase | SymbolMatchCluster
	assertScopeNames(t,
		code.GetSymbolsInScope(0, metadataMatch, IncludeFunctionKindNone),
		[]string{"Events", "Archive", "Daily", "Primary", "primary.kusto.windows.net", "other.kusto.windows.net"},
	)

	assertScopeNames(t,
		code.GetSymbolsInScope(0, SymbolMatchFunction|SymbolMatchView, IncludeFunctionKindBuiltInFunctions),
		[]string{"BuiltinFn"},
	)
	assertScopeNames(t,
		code.GetSymbolsInScope(0, SymbolMatchFunction|SymbolMatchView, IncludeFunctionKindDatabaseFunctions),
		[]string{"DbFn", "DbView"},
	)

	first := code.GetSymbolsInScope(0, metadataMatch|SymbolMatchFunction|SymbolMatchView, IncludeFunctionKindAll)
	second := code.GetSymbolsInScope(0, metadataMatch|SymbolMatchFunction|SymbolMatchView, IncludeFunctionKindAll)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("scope order changed between calls:\nfirst:  %v\nsecond: %v", scopeNames(first), scopeNames(second))
	}
}

func TestKustoCodeSpeculativeReferencedSymbolUsesBoundContext(t *testing.T) {
	column := NewColumnSymbol("Shared", ScalarTypes.String.Type())
	table := NewTableSymbol("Events", column)
	databaseFunction := NewScalarFunctionSymbol("DbFn", ScalarTypes.Long.Type())
	builtIn := NewScalarFunctionSymbol("BuiltinFn", ScalarTypes.Bool.Type())
	database := NewDatabaseSymbol("Primary", table, databaseFunction)
	globals := NewGlobalState(database).WithFunctions([]*FunctionSymbol{builtIn}).WithAggregates(nil).WithPlugIns(nil)
	text := `let Shared=1; Events | where Shared != "" | project Shared`
	code := ParseAndAnalyze(text, globals)
	position := strings.Index(text, "Shared !=")

	if got := code.GetSpeculativeReferencedSymbol(position, "Shared", SymbolMatchColumn|SymbolMatchLocal); got != column {
		t.Fatalf("row-scope Shared = %#v, want column", got)
	}
	if got := code.GetSpeculativeReferencedSymbol(position, "Events", SymbolMatchTable); got != table {
		t.Fatalf("table = %#v, want Events", got)
	}
	if got := code.GetSpeculativeReferencedSymbol(position, "DbFn", SymbolMatchFunction); got != databaseFunction {
		t.Fatalf("database function = %#v, want DbFn", got)
	}
	if got := code.GetSpeculativeReferencedSymbol(position, "BuiltinFn", SymbolMatchFunction); got != builtIn {
		t.Fatalf("built-in function = %#v, want BuiltinFn", got)
	}
	if got := code.GetSpeculativeReferencedSymbol(position, "Events", SymbolMatchFunction); got != nil {
		t.Fatalf("match-filtered table = %#v, want nil", got)
	}
	if got := code.GetSpeculativeReferencedSymbol(position, "Missing", SymbolMatchAny); got != nil {
		t.Fatalf("missing symbol = %#v, want nil", got)
	}
}

func TestKustoCodeSpeculativeReferencedSymbolUsesLexicalLocalWithoutRowCollision(t *testing.T) {
	text := `let Threshold=10; print Result=Threshold`
	code := ParseAndAnalyze(text)
	position := strings.LastIndex(text, "Threshold")
	declaration := parsedNameDeclarationBySimpleName(t, code, "Threshold")
	if got := code.GetSpeculativeReferencedSymbol(position, "Threshold", SymbolMatchLocal); got != declaration.ReferencedSymbol() {
		t.Fatalf("local = %#v, want %#v", got, declaration.ReferencedSymbol())
	}
}

func TestKustoCodeSymbolsInScopeDistinguishLocalFunctionsAndViews(t *testing.T) {
	text := `let fn=() { print x=1 }; let localView=view() { print x=1 }; print done=1`
	code := ParseAndAnalyze(text)
	position := strings.Index(text, "done=1")

	assertScopeNames(t,
		code.GetSymbolsInScope(position, SymbolMatchFunction|SymbolMatchView, IncludeFunctionKindLocalFunctions),
		[]string{"fn"},
	)
	assertScopeNames(t,
		code.GetSymbolsInScope(position, SymbolMatchFunction|SymbolMatchView, IncludeFunctionKindLocalViews),
		[]string{"localView"},
	)
}

func TestKustoCodeSymbolsInScopePreserveCategoriesWithoutSuppressingLexicalEntries(t *testing.T) {
	table := NewTableSymbol("Shared")
	databaseFunction := NewComputedFunctionSymbol("Shared", "print x=1", TabularityTabular)
	database := NewDatabaseSymbol("db", table, databaseFunction)
	globals := NewGlobalState(database).WithFunctions(nil).WithAggregates(nil).WithPlugIns(nil)
	code := ParseAndAnalyze(`let Shared=1; print x=Shared`, globals)

	beforeLocal := strings.Index(code.Text, "let Shared")
	before := code.GetSymbolsInScope(beforeLocal, SymbolMatchTable|SymbolMatchFunction, IncludeFunctionKindDatabaseFunctions)
	assertScopeNames(t, before, []string{"Shared", "Shared"})
	if _, ok := before[0].(*TableSymbol); !ok {
		t.Fatalf("first Shared symbol = %T, want table", before[0])
	}
	if _, ok := before[1].(*FunctionSymbol); !ok {
		t.Fatalf("second Shared symbol = %T, want function", before[1])
	}

	afterLocal := strings.Index(code.Text, "x=Shared") + len("x=")
	assertScopeNames(t,
		code.GetSymbolsInScope(afterLocal, SymbolMatchLocal|SymbolMatchTable|SymbolMatchFunction, IncludeFunctionKindAll),
		[]string{"Shared", "Shared", "Shared"},
	)
}

func TestKustoCodeSymbolsInScopeTreatsAmbientSymbolsAsOuterLocals(t *testing.T) {
	ambientFunction := NewScalarFunctionSymbol("AmbientFunction", ScalarTypes.String.Type())
	ambientShared := NewVariableSymbol("Shared", ScalarTypes.String.Type(), false, nil, nil)
	clientOnly := NewParameterSymbol("ClientOnly", ScalarTypes.String.Type(), "")
	globals := NewGlobalState(nil).
		WithFunctions(nil).
		WithAggregates(nil).
		WithPlugIns(nil).
		WithAmbientSymbols([]Symbol{ambientFunction, ambientShared}).
		WithClientSymbols([]Symbol{clientOnly})
	text := `let Shared=1; print x=Shared`
	code := ParseAndAnalyze(text, globals)
	position := strings.Index(text, "x=Shared") + len("x=")

	assertScopeNames(t,
		code.GetSymbolsInScope(position, SymbolMatchLocal, IncludeFunctionKindNone),
		[]string{"Shared", "Shared"},
	)
	assertScopeNames(t,
		code.GetSymbolsInScope(position, SymbolMatchFunction, IncludeFunctionKindLocalFunctions),
		[]string{"AmbientFunction"},
	)
	assertScopeNames(t,
		code.GetSymbolsInScope(position, SymbolMatchFunction, IncludeFunctionKindBuiltInFunctions),
		[]string{},
	)
	for _, symbol := range code.GetSymbolsInScope(position, SymbolMatchAny, IncludeFunctionKindAll) {
		if symbol.SymbolName() == clientOnly.SymbolName() {
			t.Fatalf("client symbol leaked into ordinary scope: %#v", symbol)
		}
	}
}

func assertScopeNames(t *testing.T, symbols []Symbol, want []string) {
	t.Helper()
	if got := scopeNames(symbols); !reflect.DeepEqual(got, want) {
		t.Fatalf("scope names = %v, want %v", got, want)
	}
}

func scopeNames(symbols []Symbol) []string {
	names := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		if symbol != nil {
			names = append(names, symbol.SymbolName())
		}
	}
	return names
}

func assertScopeColumnNames(t *testing.T, table *TableSymbol, want []string) {
	t.Helper()
	columns := table.Columns()
	names := make([]string, 0, len(columns))
	for _, column := range columns {
		names = append(names, column.SymbolName())
	}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("column names = %v, want %v", names, want)
	}
}
