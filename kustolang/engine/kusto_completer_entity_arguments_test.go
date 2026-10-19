package kustolang

import "testing"

func TestKustoCompleterOffersEntityExamplesFromReturnTypeKinds(t *testing.T) {
	globals := completionEntityTestGlobals()
	tests := []struct {
		text       string
		example    string
		notExample string
	}{
		{text: "table(", example: "'Events'", notExample: "'Archive'"},
		{text: "external_table(", example: "'Archive'", notExample: "'Events'"},
		{text: "materialized_view(", example: "'LatestEvents'", notExample: "'Events'"},
		{text: "entity_group(", example: "'Regional'", notExample: "'Events'"},
		{text: "stored_query_result(", example: "'SavedResult'", notExample: "'Events'"},
		{text: "graph(", example: "'SecurityGraph'", notExample: "'Events'"},
		{text: "graph('SecurityGraph', ", example: "'Daily'", notExample: "'SecurityGraph'"},
		{text: "cluster(", example: "'help.kusto.windows.net'"},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			service := NewKustoCodeService(test.text, globals)
			if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
				t.Fatalf("incomplete entity call %q produced no diagnostics", test.text)
			}
			items := service.GetCompletionItems(len(test.text), CompletionOptionsDefault, CancellationToken{}).Items
			requireCompletionKind(t, items, test.example, CompletionKindExample)
			if test.notExample != "" {
				requireNoCompletion(t, items, test.notExample)
			}
		})
	}
}

func TestKustoCompleterOffersDatabaseAliasesAndLocalStringArguments(t *testing.T) {
	globals := completionEntityTestGlobals()
	databaseText := "cluster('help.kusto.windows.net').database("
	databaseService := NewKustoCodeService(databaseText, globals)
	if diagnostics := databaseService.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
		t.Fatalf("incomplete qualified database call %q produced no diagnostics", databaseText)
	}
	databaseItems := databaseService.GetCompletionItems(len(databaseText), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, databaseItems, "'Samples'", CompletionKindExample)
	requireCompletionKind(t, databaseItems, "'SamplesAlias'", CompletionKindExample)
	requireNoCompletion(t, databaseItems, "'Current'")

	tableText := "cluster('help.kusto.windows.net').database('Samples').table("
	tableService := NewKustoCodeService(tableText, globals)
	if diagnostics := tableService.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
		t.Fatalf("incomplete qualified table call %q produced no diagnostics", tableText)
	}
	tableItems := tableService.GetCompletionItems(len(tableText), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, tableItems, "'StormEvents'", CompletionKindExample)
	requireNoCompletion(t, tableItems, "'Events'")

	localText := "let TableName='Events'; table("
	localItems := NewKustoCodeService(localText, globals).GetCompletionItems(len(localText), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, localItems, "TableName", CompletionKindVariable)
	requireCompletionKind(t, localItems, "'Events'", CompletionKindExample)

	parameterText := "let F=(EntityName:string) { table("
	parameterItems := NewKustoCodeService(parameterText, globals).GetCompletionItems(len(parameterText), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, parameterItems, "EntityName", CompletionKindParameter)
}

func TestKustoCompleterUnknownGraphModelHasIsolatedEmptySnapshotContext(t *testing.T) {
	globals := completionEntityTestGlobals()
	text := "graph('MissingModel', "
	service := NewKustoCodeService(text, globals)
	if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
		t.Fatalf("unknown graph model query %q produced no diagnostics", text)
	}
	items := service.GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	requireNoCompletion(t, items, "'Daily'")
	requireNoCompletion(t, items, "'SecurityGraph'")
}

func completionEntityTestGlobals() *GlobalState {
	events := NewTableSymbol("Events", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	external := NewExternalTableSymbol("Archive", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	view := NewMaterializedViewSymbol("LatestEvents", []*ColumnSymbol{NewColumnSymbol("Account", ScalarTypes.String.Type())}, "Events")
	group := NewEntityGroupSymbol("Regional", events)
	stored := NewStoredQueryResultSymbol("SavedResult", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	graph := NewGraphModelSymbolFromNames("SecurityGraph", nil, nil, []string{"Daily", "Weekly"})
	current := NewDatabaseSymbol("Current", events, external, view, group, stored, graph)
	currentCluster := NewClusterSymbol("current.kusto.windows.net", current)

	remoteTable := NewTableSymbol("StormEvents", NewColumnSymbol("State", ScalarTypes.String.Type()))
	remoteDatabase := NewDatabaseSymbolWithAlternateName("Samples", "SamplesAlias", []Symbol{remoteTable}, false)
	remoteCluster := NewClusterSymbol("help.kusto.windows.net", remoteDatabase)
	return NewGlobalState(current).
		WithClusterList([]*ClusterSymbol{currentCluster, remoteCluster}).
		WithCluster(currentCluster).
		WithDatabase(current)
}
