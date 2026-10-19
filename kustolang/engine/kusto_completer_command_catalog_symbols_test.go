package kustolang

import "testing"

// Covers the command branch of KustoCompleter.GetSymbolCompletions against
// source-shaped catalog hints, including malformed required-name recovery.
func TestKustoCompleterCommandCatalogHintInventory(t *testing.T) {
	globals := commandCatalogCompletionGlobals()
	code := ParseAndAnalyze(".show version", globals)
	completer := newKustoCompleter(code, CompletionOptionsDefault, CancellationToken{})
	hints := CompletionHintTable | CompletionHintExternalTable | CompletionHintMaterializedView |
		CompletionHintEntityGroup | CompletionHintStoredQueryResult | CompletionHintGraphModel |
		CompletionHintGraphSnapshot | CompletionHintDatabaseFunction | CompletionHintDatabase |
		CompletionHintCluster | CompletionHintOption
	symbols := completer.commandSymbolsForCompletionHint(0, hints)
	for _, name := range []string{
		"Events", "Archive", "Daily", "Targets", "Saved", "SecurityGraph", "Current", "DbFn",
		"db", "OtherDB", "help.kusto.windows.net", "remote.kusto.windows.net", "query_language",
	} {
		if !containsScopeSymbolName(symbols, name) {
			t.Fatalf("combined command hint symbols are missing %q: %#v", name, scopeNames(symbols))
		}
	}
	if containsScopeSymbolName(symbols, "strlen") {
		t.Fatalf("command function hints included a built-in function: %#v", scopeNames(symbols))
	}
}

func TestKustoCompleterCommandCatalogItemsUseNameOnlyEdits(t *testing.T) {
	globals := commandCatalogCompletionGlobals()
	database := globals.Cluster.GetDatabase("OtherDB")
	databaseItems := commandSymbolCompletionItems(database)
	if len(databaseItems) != 2 || databaseItems[0].DisplayText != "OtherDB" || databaseItems[1].DisplayText != "OtherAlias" {
		t.Fatalf("database completion items = %#v", databaseItems)
	}
	if databaseItems[1].BeforeText() != "OtherAlias" {
		t.Fatalf("database alternate edit = %q", databaseItems[1].BeforeText())
	}

	clusterItems := commandSymbolCompletionItems(globals.GetCluster("remote.kusto.windows.net"))
	if len(clusterItems) != 1 || clusterItems[0].Kind != CompletionKindCluster || clusterItems[0].BeforeText() != `['remote.kusto.windows.net']` {
		t.Fatalf("cluster completion items = %#v", clusterItems)
	}

	function := globals.Database.GetFunction("DbFn")
	functionItems := commandSymbolCompletionItems(function)
	if len(functionItems) != 1 || functionItems[0].Kind != CompletionKindDatabaseFunction ||
		functionItems[0].DisplayText != "DbFn" || functionItems[0].BeforeText() != "DbFn" {
		t.Fatalf("function completion items = %#v", functionItems)
	}
	if items := commandSymbolCompletionItems(function.Hide()); len(items) != 0 {
		t.Fatalf("hidden command function completion items = %#v", items)
	}
}

func TestKustoCompleterCommandCatalogSymbolsOnMalformedCommands(t *testing.T) {
	globals := commandCatalogCompletionGlobals()
	tests := []struct {
		text       string
		want       []string
		wantAbsent []string
	}{
		{text: ".show databases (", want: []string{"db", "OtherDB", "OtherAlias"}},
		{text: ".show function ", want: []string{"DbFn"}, wantAbsent: []string{"strlen", "DbFn()"}},
		{text: ".show graph_model ", want: []string{"SecurityGraph"}},
		{text: ".show entity_group ", want: []string{"Targets"}},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			code := ParseAndAnalyze(test.text, globals)
			if len(code.GetSyntaxDiagnostics()) == 0 {
				t.Fatalf("malformed command %q produced no syntax diagnostics", test.text)
			}
			items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(test.text), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, items, test.want...)
			assertCompletionExcludes(t, items, test.wantAbsent...)
		})
	}
}

func commandCatalogCompletionGlobals() *GlobalState {
	databaseFunction := NewComputedFunctionSymbol("DbFn", "print x=1", TabularityTabular)
	current := NewDatabaseSymbol("db",
		NewTableSymbol("Events"),
		NewExternalTableSymbol("Archive"),
		NewMaterializedViewSymbol("Daily", nil, "Events | take 1"),
		NewEntityGroupSymbol("Targets"),
		NewStoredQueryResultSymbol("Saved"),
		NewGraphModelSymbolFromNames("SecurityGraph", nil, nil, []string{"Current"}),
		databaseFunction,
	)
	other := NewDatabaseSymbolWithAlternateName("OtherDB", "OtherAlias", nil, false)
	currentCluster := NewClusterSymbol("help.kusto.windows.net", current, other)
	remoteCluster := NewClusterSymbol("remote.kusto.windows.net", NewDatabaseSymbol("RemoteDB"))
	option := NewOptionSymbolWithType("query_language", "", ScalarTypes.String, nil)
	return NewGlobalState(current).
		WithClusterList([]*ClusterSymbol{currentCluster, remoteCluster}).
		WithCluster(currentCluster).
		WithDatabase(current).
		WithOptions([]*OptionSymbol{option})
}
