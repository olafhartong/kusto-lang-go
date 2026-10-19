package kustolang

import "testing"

func TestKustoCompleterQueryRootTablesAndTabularFunctions(t *testing.T) {
	table := NewTableSymbol("Events", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	hiddenTable := NewTableSymbol("__Internal", NewColumnSymbol("Secret", ScalarTypes.String.Type()))
	databaseFunction := NewComputedFunctionSymbol("DbRows", "Events", TabularityTabular)
	parameterizedFunction := NewComputedFunctionSymbol("RowsFor", "Events", TabularityTabular, NewParameter("value", ScalarTypes.String.Type()))
	scalarDatabaseFunction := NewScalarFunctionSymbol("DbScalar", ScalarTypes.Long.Type())
	database := NewDatabaseSymbol("db", table, hiddenTable, databaseFunction, parameterizedFunction, scalarDatabaseFunction)
	tabularBuiltIn := NewComputedFunctionSymbol("Rows", "Events", TabularityTabular)
	scalarBuiltIn := NewScalarFunctionSymbol("ScalarOnly", ScalarTypes.Long.Type())
	hiddenFunction := NewComputedFunctionSymbol("HiddenRows", "Events", TabularityTabular).Hide()
	globals := NewGlobalState(database).WithFunctions([]*FunctionSymbol{tabularBuiltIn, scalarBuiltIn, hiddenFunction})

	items := NewKustoCodeService("", globals).GetCompletionItems(0, CompletionOptionsDefault, CancellationToken{}).Items
	tableItem := requireCompletionKind(t, items, "Events", CompletionKindTable)
	if tableItem.BeforeText() != "Events\n| " || tableItem.AfterText() != "" {
		t.Fatalf("root table insertion = before:%q after:%q", tableItem.BeforeText(), tableItem.AfterText())
	}
	databaseItem := requireCompletionKind(t, items, CompletionDisplayText(databaseFunction), CompletionKindDatabaseFunction)
	if databaseItem.BeforeText() != "DbRows\n| " || databaseItem.AfterText() != "" {
		t.Fatalf("zero-argument database function insertion = before:%q after:%q", databaseItem.BeforeText(), databaseItem.AfterText())
	}
	builtInItem := requireCompletionKind(t, items, CompletionDisplayText(tabularBuiltIn), CompletionKindBuiltInFunction)
	if builtInItem.BeforeText() != "Rows()\n| " || builtInItem.AfterText() != "" {
		t.Fatalf("zero-argument built-in insertion = before:%q after:%q", builtInItem.BeforeText(), builtInItem.AfterText())
	}
	parameterizedItem := requireCompletionKind(t, items, CompletionDisplayText(parameterizedFunction), CompletionKindDatabaseFunction)
	if parameterizedItem.BeforeText() != "RowsFor(" || parameterizedItem.AfterText() != ")" || !parameterizedItem.HasCaret() {
		t.Fatalf("parameterized function insertion = before:%q after:%q caret:%t", parameterizedItem.BeforeText(), parameterizedItem.AfterText(), parameterizedItem.HasCaret())
	}
	requireNoCompletion(t, items, "__Internal")
	requireNoCompletion(t, items, CompletionDisplayText(hiddenFunction))
	requireNoCompletion(t, items, CompletionDisplayText(scalarDatabaseFunction))
	requireNoCompletion(t, items, CompletionDisplayText(scalarBuiltIn))
}

func TestKustoCompleterInjectsSingleDatabaseFunctionParameter(t *testing.T) {
	function := NewComputedFunctionSymbol("RowsFor", "Events", TabularityTabular, NewParameter("value", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", function))
	options := CompletionOptionsDefault.WithEnableParameterInjection(true)

	items := NewKustoCodeService("", globals).GetCompletionItems(0, options, CancellationToken{}).Items
	item := requireCompletionKind(t, items, CompletionDisplayText(function), CompletionKindDatabaseFunction)
	if item.BeforeText() != "RowsFor({parameter})" || item.AfterText() != "" || item.HasCaret() {
		t.Fatalf("injected function insertion = before:%q after:%q caret:%t apply:%#v", item.BeforeText(), item.AfterText(), item.HasCaret(), item.ApplyTexts)
	}
}

func TestKustoCompleterDoesNotInjectSingleDatabaseFunctionParameterInInvokeContext(t *testing.T) {
	events := NewTableSymbol("Events", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	input := NewTableSymbol("Input", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	function := NewComputedFunctionSymbol("Transform", "source", TabularityTabular, NewParameterSymbolType("source", input))
	globals := NewGlobalState(NewDatabaseSymbol("db", events, function))
	options := CompletionOptionsDefault.WithEnableParameterInjection(true)
	text := "Events | invoke "

	items := NewKustoCodeService(text, globals).GetCompletionItems(len(text), options, CancellationToken{}).Items
	item := requireCompletionKind(t, items, CompletionDisplayText(function), CompletionKindDatabaseFunction)
	if item.BeforeText() != "Transform(" || item.AfterText() != ")" || !item.HasCaret() {
		t.Fatalf("invoke function insertion = before:%q after:%q caret:%t apply:%#v", item.BeforeText(), item.AfterText(), item.HasCaret(), item.ApplyTexts)
	}
}

func TestKustoCompleterTreatsPositionAfterLetAsQueryStart(t *testing.T) {
	table := NewTableSymbol("Events", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	function := NewComputedFunctionSymbol("DbRows", "Events", TabularityTabular)
	globals := NewGlobalState(NewDatabaseSymbol("db", table, function))
	text := "let Threshold=10; "

	items := NewKustoCodeService(text, globals).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	tableItem := requireCompletionKind(t, items, "Events", CompletionKindTable)
	if tableItem.BeforeText() != "Events\n| " {
		t.Fatalf("post-let table insertion = %q", tableItem.BeforeText())
	}
	functionItem := requireCompletionKind(t, items, CompletionDisplayText(function), CompletionKindDatabaseFunction)
	if functionItem.BeforeText() != "DbRows\n| " {
		t.Fatalf("post-let function insertion = %q", functionItem.BeforeText())
	}
	requireCompletionKind(t, items, "Threshold", CompletionKindVariable)
}

func TestKustoCompleterUsesTableKindForTableValuedLetVariables(t *testing.T) {
	table := NewTableSymbol("Events", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	for _, test := range []struct {
		name       string
		text       string
		beforeText string
	}{
		{name: "query start", text: "let Rows=Events; let Threshold=1; ", beforeText: "Rows\n| "},
		{name: "tabular operand", text: "let Rows=Events; union ", beforeText: "Rows"},
	} {
		t.Run(test.name, func(t *testing.T) {
			items := NewKustoCodeService(test.text, globals).GetCompletionItems(len(test.text), CompletionOptionsDefault, CancellationToken{}).Items
			rows := requireCompletionKind(t, items, "Rows", CompletionKindTable)
			if rows.BeforeText() != test.beforeText {
				t.Fatalf("table-valued let insertion = %q, want %q", rows.BeforeText(), test.beforeText)
			}
			if test.name == "query start" {
				requireCompletionKind(t, items, "Threshold", CompletionKindVariable)
			}
		})
	}
}

func TestKustoCompleterQueryDatabaseSymbolsIncludeAlternateNames(t *testing.T) {
	current := NewDatabaseSymbol("Current")
	other := NewDatabaseSymbolWithAlternateName("OtherDB", "Other Alias", nil, false)
	cluster := NewClusterSymbol("help.kusto.windows.net", current, other)
	globals := NewGlobalState(current).
		WithClusterList([]*ClusterSymbol{cluster}).
		WithCluster(cluster).
		WithDatabase(current)
	text := "alias database Local = "

	items := NewKustoCodeService(text, globals).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	primary := requireCompletionKind(t, items, "OtherDB", CompletionKindDatabase)
	if primary.BeforeText() != "OtherDB" {
		t.Fatalf("primary database insertion = %q", primary.BeforeText())
	}
	alternate := requireCompletionKind(t, items, "Other Alias", CompletionKindDatabase)
	if alternate.BeforeText() != "['Other Alias']" || alternate.MatchText != "Other Alias" {
		t.Fatalf("alternate database completion = before:%q match:%q", alternate.BeforeText(), alternate.MatchText)
	}
}

func TestKustoCompleterUsesUpstreamTableFamilyQueryInsertions(t *testing.T) {
	regular := NewTableSymbol("Events Log", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	external := NewExternalTableSymbol("Archive Logs", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	view := NewMaterializedViewSymbol("Daily View", []*ColumnSymbol{NewColumnSymbol("Account", ScalarTypes.String.Type())}, "Events | summarize count()")
	stored := NewStoredQueryResultSymbol("Saved Result", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", regular, external, view, stored))

	items := NewKustoCodeService("", globals).GetCompletionItems(0, CompletionOptionsDefault, CancellationToken{}).Items
	tests := []struct {
		symbol   Symbol
		kind     CompletionKind
		before   string
		priority CompletionPriority
	}{
		{regular, CompletionKindTable, "['Events Log']\n| ", CompletionPriorityNormal},
		{external, CompletionKindTable, "external_table('Archive Logs')\n| ", CompletionPriorityLow},
		{view, CompletionKindMaterialiedView, "['Daily View']\n| ", CompletionPriorityNormal},
	}
	for _, test := range tests {
		item := requireCompletionKind(t, items, CompletionDisplayText(test.symbol), test.kind)
		if item.BeforeText() != test.before || item.AfterText() != "" || item.MatchText != test.symbol.SymbolName() || item.priority != test.priority {
			t.Fatalf("%T completion = before:%q after:%q match:%q priority:%v", test.symbol, item.BeforeText(), item.AfterText(), item.MatchText, item.priority)
		}
	}
	completer := newKustoCompleter(ParseAndAnalyze("", globals), CompletionOptionsDefault, CancellationToken{})
	storedItem, ok := completer.querySymbolCompletionItem(stored, CompletionKindStoredQueryResult, true)
	if !ok || storedItem.BeforeText() != "['Saved Result']\n| " || storedItem.MatchText != "Saved Result" || storedItem.priority != CompletionPriorityNormal {
		t.Fatalf("stored-result completion = ok:%t before:%q match:%q priority:%v", ok, storedItem.BeforeText(), storedItem.MatchText, storedItem.priority)
	}
}

func TestKustoCompleterTableFamilyInsertionsInMalformedUnionOperand(t *testing.T) {
	regular := NewTableSymbol("Events", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	external := NewExternalTableSymbol("Archive", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	view := NewMaterializedViewSymbol("Daily", []*ColumnSymbol{NewColumnSymbol("Account", ScalarTypes.String.Type())}, "Events")
	stored := NewStoredQueryResultSymbol("Saved", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", regular, external, view, stored))
	text := "union Events, "
	service := NewKustoCodeService(text, globals)
	if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
		t.Fatalf("malformed union query %q produced no diagnostics", text)
	}

	items := service.GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	tests := []struct {
		symbol Symbol
		kind   CompletionKind
		before string
	}{
		{regular, CompletionKindTable, "Events"},
		{external, CompletionKindTable, "external_table('Archive')"},
		{view, CompletionKindMaterialiedView, "Daily"},
	}
	for _, test := range tests {
		item := requireCompletionKind(t, items, CompletionDisplayText(test.symbol), test.kind)
		if item.BeforeText() != test.before || item.MatchText != test.symbol.SymbolName() {
			t.Fatalf("%T malformed-union completion = before:%q match:%q", test.symbol, item.BeforeText(), item.MatchText)
		}
	}
	completer := newKustoCompleter(ParseAndAnalyze(text, globals), CompletionOptionsDefault, CancellationToken{})
	storedItem, ok := completer.querySymbolCompletionItem(stored, CompletionKindStoredQueryResult, false)
	if !ok || storedItem.BeforeText() != "Saved" || storedItem.MatchText != "Saved" {
		t.Fatalf("stored-result malformed-union insertion = ok:%t before:%q match:%q", ok, storedItem.BeforeText(), storedItem.MatchText)
	}
}

func TestKustoCompleterQueryPipelineColumnsAndScalarFunctions(t *testing.T) {
	table := NewTableSymbol("Events",
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("Score", ScalarTypes.Long.Type()),
		NewColumnSymbol("__HiddenColumn", ScalarTypes.Long.Type()))
	databaseScalar := NewScalarFunctionSymbol("DbScalar", ScalarTypes.Long.Type())
	databaseRows := NewComputedFunctionSymbol("DbRows", "Events", TabularityTabular)
	database := NewDatabaseSymbol("db", table, databaseScalar, databaseRows)
	builtInScalar := NewScalarFunctionSymbol("ScalarFn", ScalarTypes.Long.Type())
	builtInRows := NewComputedFunctionSymbol("Rows", "Events", TabularityTabular)
	globals := NewGlobalState(database).WithFunctions([]*FunctionSymbol{builtInScalar, builtInRows})
	text := "Events | where "

	items := NewKustoCodeService(text, globals).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, items, "Account", CompletionKindColumn)
	requireCompletionKind(t, items, "Score", CompletionKindColumn)
	requireCompletionKind(t, items, CompletionDisplayText(databaseScalar), CompletionKindDatabaseFunction)
	requireCompletionKind(t, items, CompletionDisplayText(builtInScalar), CompletionKindBuiltInFunction)
	requireNoCompletion(t, items, "__HiddenColumn")
	requireNoCompletion(t, items, CompletionDisplayText(databaseRows))
	requireNoCompletion(t, items, CompletionDisplayText(builtInRows))
	requireNoCompletion(t, items, "Events")
}

func TestKustoCompleterAggregateFallbackUsesAggregateKindAndRank(t *testing.T) {
	aggregate := NewScalarFunctionSymbol("custom_aggregate", ScalarTypes.Long.Type())
	globals := NewGlobalState(NewDatabaseSymbol("db")).WithAggregates([]*FunctionSymbol{aggregate})
	completer := newKustoCompleter(ParseAndAnalyze("", globals), CompletionOptionsDefault, CancellationToken{})
	builder := newKustoCompletionBuilder()

	completer.addApplicableQueryFunctions(builder, queryCompletionScope{
		scopeKind: ScopeKindAggregate,
		globals:   globals,
	}, map[string]struct{}{})

	item := requireCompletionKind(t, builder.items(), CompletionDisplayText(aggregate), CompletionKindAggregateFunction)
	if rank := completer.getOrderingRank(item); rank != CompletionRankAggregate {
		t.Fatalf("aggregate fallback rank = %v, want %v", rank, CompletionRankAggregate)
	}
}

func TestKustoCompleterIncompleteSummarizeAggregateKindAndRank(t *testing.T) {
	table := NewTableSymbol("Events", NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	aggregate := NewScalarFunctionSymbol("custom_aggregate", ScalarTypes.Long.Type())
	globals := NewGlobalState(NewDatabaseSymbol("db", table)).WithAggregates([]*FunctionSymbol{aggregate})
	text := "Events | summarize Total = "
	service := NewKustoCodeService(text, globals)
	if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
		t.Fatalf("incomplete summarize query %q produced no diagnostics", text)
	}

	items := service.GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	item := requireCompletionKind(t, items, CompletionDisplayText(aggregate), CompletionKindAggregateFunction)
	completer := newKustoCompleter(ParseAndAnalyze(text, globals), CompletionOptionsDefault, CancellationToken{})
	if rank := completer.getOrderingRank(item); rank != CompletionRankAggregate {
		t.Fatalf("incomplete summarize aggregate rank = %v, want %v", rank, CompletionRankAggregate)
	}
}

func TestKustoCompleterKeepsSemanticSymbolsForMalformedExpression(t *testing.T) {
	table := NewTableSymbol("Events",
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("Score", ScalarTypes.Long.Type()))
	function := NewScalarFunctionSymbol("ScalarFn", ScalarTypes.Long.Type())
	globals := NewGlobalState(NewDatabaseSymbol("db", table)).WithFunctions([]*FunctionSymbol{function})
	text := "Events | where ("
	service := NewKustoCodeService(text, globals)

	diagnostics := service.GetDiagnostics(true, CancellationToken{})
	if len(diagnostics) == 0 {
		t.Fatalf("malformed completion query %q produced no diagnostics", text)
	}
	items := service.GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, items, "Account", CompletionKindColumn)
	requireCompletionKind(t, items, "Score", CompletionKindColumn)
	requireCompletionKind(t, items, CompletionDisplayText(function), CompletionKindBuiltInFunction)
}

func TestKustoCompleterAddsSpaceToNonBooleanSymbolsInMalformedBooleanContexts(t *testing.T) {
	table := NewTableSymbol("Events",
		NewColumnSymbol("Enabled", ScalarTypes.Bool.Type()),
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("Score", ScalarTypes.Long.Type()))
	function := NewScalarFunctionSymbol("ScalarFn", ScalarTypes.Long.Type())
	globals := NewGlobalState(NewDatabaseSymbol("db", table)).WithFunctions([]*FunctionSymbol{function})

	for _, text := range []string{
		"Events | where ",
		"Events | where (",
		"Events | where Enabled and (",
	} {
		service := NewKustoCodeService(text, globals)
		if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
			t.Fatalf("malformed boolean query %q produced no diagnostics", text)
		}
		items := service.GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
		if item := requireCompletionKind(t, items, "Enabled", CompletionKindColumn); item.BeforeText() != "Enabled" {
			t.Fatalf("boolean completion for %q = %q", text, item.BeforeText())
		}
		for _, name := range []string{"Account", "Score"} {
			if item := requireCompletionKind(t, items, name, CompletionKindColumn); item.BeforeText() != name+" " || !item.Retrigger {
				t.Fatalf("non-boolean completion %s for %q = before:%q retrigger:%t", name, text, item.BeforeText(), item.Retrigger)
			}
		}
		if item := requireCompletionKind(t, items, CompletionDisplayText(function), CompletionKindBuiltInFunction); item.BeforeText() != "ScalarFn() " || !item.Retrigger {
			t.Fatalf("function completion for %q = before:%q retrigger:%t", text, item.BeforeText(), item.Retrigger)
		}
	}
}

func TestKustoCompleterFiltersInvokeFunctionsByIncomingSchema(t *testing.T) {
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	events := NewTableSymbol("Events", account, NewColumnSymbol("Score", ScalarTypes.Long.Type()))
	compatibleInput := NewTableSymbol("CompatibleInput", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	incompatibleInput := NewTableSymbol("IncompatibleInput", NewColumnSymbol("Account", ScalarTypes.Long.Type()))
	compatible := NewComputedFunctionSymbol("Compatible", "source", TabularityTabular, NewParameterSymbolType("source", compatibleInput))
	incompatible := NewComputedFunctionSymbol("Incompatible", "source", TabularityTabular, NewParameterSymbolType("source", incompatibleInput))
	anyTable := NewComputedFunctionSymbol("AnyTable", "source", TabularityTabular, NewParameterKind("source", ParameterTypeTabular))
	zeroArguments := NewComputedFunctionSymbol("ZeroArguments", "Events", TabularityTabular)
	scalar := NewScalarFunctionSymbol("ScalarOnly", ScalarTypes.Long.Type(), NewParameter("value", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", events, compatible, incompatible, anyTable, zeroArguments, scalar))
	text := "let Threshold=1; Events | invoke "
	service := NewKustoCodeService(text, globals)

	if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
		t.Fatalf("incomplete invoke query %q produced no diagnostics", text)
	}
	items := service.GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, items, CompletionDisplayText(Functions.Cluster), CompletionKindBuiltInFunction)
	requireCompletionKind(t, items, CompletionDisplayText(Functions.Database), CompletionKindBuiltInFunction)
	requireCompletionKind(t, items, CompletionDisplayText(compatible), CompletionKindDatabaseFunction)
	requireCompletionKind(t, items, CompletionDisplayText(anyTable), CompletionKindDatabaseFunction)
	requireNoCompletion(t, items, CompletionDisplayText(incompatible))
	requireNoCompletion(t, items, CompletionDisplayText(zeroArguments))
	requireNoCompletion(t, items, CompletionDisplayText(scalar))
}

func TestQueryInvokeFunctionAppliesCatalogExceptionsBeforeRowScopeFiltering(t *testing.T) {
	closed := NewTableSymbol("Closed", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	open := closed.WithIsOpen(true)
	other := NewScalarFunctionSymbol("Other", ScalarTypes.Long.Type(), NewParameter("value", ScalarTypes.Long.Type()))

	for _, test := range []struct {
		name     string
		function *FunctionSymbol
		rowScope *TableSymbol
		want     bool
	}{
		{name: "cluster with closed schema", function: Functions.Cluster, rowScope: closed, want: true},
		{name: "database with closed schema", function: Functions.Database, rowScope: closed, want: true},
		{name: "cluster without row scope", function: Functions.Cluster, want: true},
		{name: "database without row scope", function: Functions.Database, want: true},
		{name: "other with closed schema", function: other, rowScope: closed, want: false},
		{name: "other with open schema", function: other, rowScope: open, want: true},
		{name: "other without row scope", function: other, want: false},
		{name: "nil function", rowScope: closed, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := queryInvokeFunctionApplies(test.function, test.rowScope); got != test.want {
				t.Fatalf("queryInvokeFunctionApplies() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestKustoCompleterFiltersLocalInvokeFunctionsAfterLetStatements(t *testing.T) {
	events := NewTableSymbol("Events", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", events))
	text := "let CompatibleLocal=(source:(Account:string)) { source }; let IncompatibleLocal=(source:(Missing:string)) { source }; Events | invoke "
	service := NewKustoCodeService(text, globals)

	if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) == 0 {
		t.Fatalf("incomplete local invoke query %q produced no diagnostics", text)
	}
	items := service.GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionMatchKind(t, items, "CompatibleLocal", CompletionKindLocalFunction)
	requireNoCompletionMatch(t, items, "IncompatibleLocal")
}

func TestKustoCompleterQuerySymbolOptions(t *testing.T) {
	table := NewTableSymbol("Events", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	databaseFunction := NewComputedFunctionSymbol("DbRows", "Events", TabularityTabular)
	builtInFunction := NewComputedFunctionSymbol("Rows", "Events", TabularityTabular)
	globals := NewGlobalState(NewDatabaseSymbol("db", table, databaseFunction)).WithFunctions([]*FunctionSymbol{builtInFunction})

	withoutSymbols := CompletionOptionsDefault.WithIncludeSymbols(false)
	items := NewKustoCodeService("", globals).GetCompletionItems(0, withoutSymbols, CancellationToken{}).Items
	requireNoCompletion(t, items, "Events")
	requireNoCompletion(t, items, CompletionDisplayText(databaseFunction))

	builtInsOnly := CompletionOptionsDefault.WithIncludeFunctions(IncludeFunctionKindBuiltInFunctions)
	items = NewKustoCodeService("", globals).GetCompletionItems(0, builtInsOnly, CancellationToken{}).Items
	requireCompletionKind(t, items, "Events", CompletionKindTable)
	requireCompletionKind(t, items, CompletionDisplayText(builtInFunction), CompletionKindBuiltInFunction)
	requireNoCompletion(t, items, CompletionDisplayText(databaseFunction))

	withoutFunctions := CompletionOptionsDefault.WithIncludeFunctions(IncludeFunctionKindNone)
	items = NewKustoCodeService("", globals).GetCompletionItems(0, withoutFunctions, CancellationToken{}).Items
	requireCompletionKind(t, items, "Events", CompletionKindTable)
	requireNoCompletion(t, items, CompletionDisplayText(builtInFunction))
	requireNoCompletion(t, items, CompletionDisplayText(databaseFunction))
}

func TestKustoCompleterQuerySymbolsDeduplicateByBinderOrder(t *testing.T) {
	databaseDuplicate := NewComputedFunctionSymbol("Rows", "Events", TabularityTabular)
	builtInDuplicate := NewComputedFunctionSymbol("Rows", "Other", TabularityTabular)
	sameNameTable := NewTableSymbol("Rows", NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", sameNameTable, databaseDuplicate)).WithFunctions([]*FunctionSymbol{builtInDuplicate, builtInDuplicate})

	items := NewKustoCodeService("", globals).GetCompletionItems(0, CompletionOptionsDefault, CancellationToken{}).Items
	functionCount := 0
	tableCount := 0
	for _, item := range items {
		if item.MatchText == "Rows" {
			switch item.Kind {
			case CompletionKindBuiltInFunction:
				functionCount++
			case CompletionKindTable:
				tableCount++
			default:
				t.Fatalf("Rows completion kind = %v, want built-in function precedence from binder order", item.Kind)
			}
		}
	}
	if functionCount != 1 || tableCount != 1 {
		t.Fatalf("Rows completion counts = function:%d table:%d, want one of each; items=%#v", functionCount, tableCount, completionDisplayTexts(items))
	}
}

func TestKustoCompleterIncludesOnlyVisibleLetVariablesAndFunctions(t *testing.T) {
	text := `let Threshold=10; let Add=(x:long) { x + Threshold }; print Result=`
	items := NewKustoCodeService(text).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, items, "Threshold", CompletionKindVariable)
	requireCompletionKind(t, items, CompletionDisplayText(firstBoundNode[*FunctionDeclaration](t, ParseAndAnalyze(text).Syntax()).ReferencedSymbol()), CompletionKindLocalFunction)

	before := `print Result=; let Later=1`
	items = NewKustoCodeService(before).GetCompletionItems(len("print Result="), CompletionOptionsDefault, CancellationToken{}).Items
	requireNoCompletion(t, items, "Later")

	withoutLocals := CompletionOptionsDefault.WithIncludeFunctions(IncludeFunctionKindBuiltInFunctions | IncludeFunctionKindDatabaseFunctions)
	items = NewKustoCodeService(text).GetCompletionItems(len(text), withoutLocals, CancellationToken{}).Items
	requireNoCompletion(t, items, "Add")
}

func TestKustoCompleterUsesLexicalFunctionScope(t *testing.T) {
	text := `let Outer=1; let F=(Param:string) { let Inner=2; print Result= }; print Outside=`
	inside := len(`let Outer=1; let F=(Param:string) { let Inner=2; print Result=`)
	items := NewKustoCodeService(text).GetCompletionItems(inside, CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, items, "Outer", CompletionKindVariable)
	requireCompletionKind(t, items, "Param", CompletionKindParameter)
	requireCompletionKind(t, items, "Inner", CompletionKindVariable)
	requireNoCompletion(t, items, "F")

	items = NewKustoCodeService(text).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, items, "Outer", CompletionKindVariable)
	requireCompletionKind(t, items, CompletionDisplayText(firstBoundNode[*FunctionDeclaration](t, ParseAndAnalyze(text).Syntax()).ReferencedSymbol()), CompletionKindLocalFunction)
	requireNoCompletion(t, items, "Param")
	requireNoCompletion(t, items, "Inner")
}

func requireCompletionKind(t *testing.T, items []CompletionItem, displayText string, kind CompletionKind) CompletionItem {
	t.Helper()
	item, ok := completionItemByDisplayText(items, displayText)
	if !ok {
		t.Fatalf("completion items missing %q: %#v", displayText, completionDisplayTexts(items))
	}
	if item.Kind != kind {
		t.Fatalf("completion %q kind = %v, want %v", displayText, item.Kind, kind)
	}
	return item
}

func requireNoCompletion(t *testing.T, items []CompletionItem, displayText string) {
	t.Helper()
	if _, ok := completionItemByDisplayText(items, displayText); ok {
		t.Fatalf("completion items unexpectedly include %q: %#v", displayText, completionDisplayTexts(items))
	}
}

func requireCompletionMatchKind(t *testing.T, items []CompletionItem, matchText string, kind CompletionKind) CompletionItem {
	t.Helper()
	for _, item := range items {
		if item.MatchText == matchText {
			if item.Kind != kind {
				t.Fatalf("completion match %q kind = %v, want %v", matchText, item.Kind, kind)
			}
			return item
		}
	}
	t.Fatalf("completion items missing match %q: %#v", matchText, completionDisplayTexts(items))
	return CompletionItem{}
}

func requireNoCompletionMatch(t *testing.T, items []CompletionItem, matchText string) {
	t.Helper()
	for _, item := range items {
		if item.MatchText == matchText {
			t.Fatalf("completion items unexpectedly include match %q: %#v", matchText, completionDisplayTexts(items))
		}
	}
}
