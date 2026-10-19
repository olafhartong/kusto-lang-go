package kustolang

import (
	"strings"
	"testing"
)

func TestKustoCodeCurrentDirectiveContextChangesGlobalScopeAndCompletion(t *testing.T) {
	globals, currentTable, sampleTable := directiveContextGlobals()
	text := "#database Samples\n"
	code := ParseWithGlobals(text, globals).Analyze()
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("directive diagnostics = %#v", diagnostics)
	}

	before := code.GetSymbolsInScope(0, SymbolMatchTable, IncludeFunctionKindNone)
	assertScopeNames(t, before, []string{currentTable.SymbolName()})

	after := code.GetSymbolsInScope(len("#database Samples"), SymbolMatchTable, IncludeFunctionKindNone)
	assertScopeNames(t, after, []string{sampleTable.SymbolName()})
	assertScopeNames(t,
		code.GetSymbolsInScope(len(text), SymbolMatchFunction, IncludeFunctionKindDatabaseFunctions),
		[]string{"SampleFn"},
	)
	if got := code.GetSpeculativeReferencedSymbol(len(text), sampleTable.SymbolName(), SymbolMatchTable); got != sampleTable {
		t.Fatalf("directive-scoped speculative table = %#v, want %#v", got, sampleTable)
	}
	if got := code.GetSpeculativeReferencedSymbol(len(text), currentTable.SymbolName(), SymbolMatchTable); got != nil {
		t.Fatalf("original database table leaked after directive: %#v", got)
	}

	completer := newKustoCompleter(code, CompletionOptionsDefault, CancellationToken{})
	editStart, _ := completer.getEditRange(len(text))
	if !completer.shouldComplete(len(text)) {
		t.Fatalf("completion suppressed at %d (edit start %d)", len(text), editStart)
	}
	items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	if scope := getQueryCompletionScope(code, len(text)); !scope.root || scope.globals == nil || scope.globals.Database.GetTable(sampleTable.SymbolName()) != sampleTable {
		t.Fatalf("directive completion scope = %#v", scope)
	}
	assertCompletionIncludes(t, items, sampleTable.SymbolName())
	assertCompletionIncludes(t, items, "SampleFn()")
	assertCompletionExcludes(t, items, currentTable.SymbolName())
	assertCompletionExcludes(t, items, "CurrentFn()")
}

func TestKustoCodeCurrentConnectDirectiveChangesClusterAndDatabaseContext(t *testing.T) {
	globals, currentTable, _ := directiveContextGlobals()
	remoteTable := NewTableSymbol("RemoteTable", NewColumnSymbol("RemoteColumn", ScalarTypes.Bool.Type()))
	remoteDatabase := NewDatabaseSymbol("RemoteDB", remoteTable)
	remoteCluster := NewClusterSymbol("remote.kusto.windows.net", remoteDatabase)
	globals = globals.WithClusterList(append(globals.Clusters, remoteCluster))

	text := "#connect cluster(\"remote.kusto.windows.net\").database(\"RemoteDB\")\n"
	code := ParseWithGlobals(text, globals).Analyze()
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("connect diagnostics = %#v", diagnostics)
	}
	assertScopeNames(t,
		code.GetSymbolsInScope(len(text), SymbolMatchTable, IncludeFunctionKindNone),
		[]string{remoteTable.SymbolName()},
	)
	items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	assertCompletionIncludes(t, items, remoteTable.SymbolName())
	assertCompletionExcludes(t, items, currentTable.SymbolName())
}

func TestKustoCodeCurrentDirectiveContextSurvivesMalformedQueryRecovery(t *testing.T) {
	globals, currentTable, sampleTable := directiveContextGlobals()
	text := "#database Samples\n" + sampleTable.SymbolName() + " | project Value=("
	code := ParseWithGlobals(text, globals).Analyze()
	diagnostics := code.GetDiagnostics()
	if len(diagnostics) == 0 {
		t.Fatal("malformed project should retain syntax diagnostics")
	}
	if !hasDiagnosticCode(diagnostics, "KS006") {
		t.Fatalf("malformed project diagnostics = %#v, want KS006", diagnostics)
	}

	items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	assertCompletionIncludes(t, items, "SampleColumn")
	assertCompletionExcludes(t, items, "CurrentColumn")

	unknownText := "#database Missing\n"
	unknown := ParseWithGlobals(unknownText, globals).Analyze()
	if !hasDiagnosticCode(unknown.GetDiagnostics(), "KS208") {
		t.Fatalf("unknown database diagnostics = %#v, want KS208", unknown.GetDiagnostics())
	}
	unknownItems := NewKustoCodeServiceForCode(unknown).GetCompletionItems(len(unknownText), CompletionOptionsDefault, CancellationToken{}).Items
	assertCompletionExcludes(t, unknownItems, currentTable.SymbolName(), sampleTable.SymbolName())
}

func TestKustoCodeCurrentDirectiveContextFeedsMalformedCommandCompletion(t *testing.T) {
	globals, _, sampleTable := directiveContextGlobals()
	text := "#database Samples\n.alter table " + sampleTable.SymbolName() + " column-docstrings ("
	code := ParseWithGlobals(text, globals).Analyze()
	if diagnostics := code.GetDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("incomplete command should retain syntax diagnostics")
	}
	items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	assertCompletionIncludes(t, items, "SampleColumn")
	assertCompletionExcludes(t, items, "CurrentColumn")
}

func directiveContextGlobals() (*GlobalState, *TableSymbol, *TableSymbol) {
	currentTable := NewTableSymbol("CurrentTable", NewColumnSymbol("CurrentColumn", ScalarTypes.String.Type()))
	sampleTable := NewTableSymbol("SampleTable", NewColumnSymbol("SampleColumn", ScalarTypes.Long.Type()))
	current := NewDatabaseSymbol("CurrentDB", currentTable, NewComputedFunctionSymbol("CurrentFn", "print Current=1", TabularityTabular))
	samples := NewDatabaseSymbol("Samples", sampleTable, NewComputedFunctionSymbol("SampleFn", "print Sample=1", TabularityTabular))
	cluster := NewClusterSymbol("help.kusto.windows.net", current, samples)
	globals := NewGlobalState(current).
		WithClusterList([]*ClusterSymbol{cluster}).
		WithCluster(cluster).
		WithDatabase(current)
	return globals, currentTable, sampleTable
}

func hasDiagnosticCode(diagnostics []Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if strings.EqualFold(diagnostic.Code, code) {
			return true
		}
	}
	return false
}
