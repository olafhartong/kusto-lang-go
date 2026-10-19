package kustolang

import (
	"strings"
	"testing"
)

func TestNativeSourceReferenceFunctionsResolveConfiguredClustersAndDatabases(t *testing.T) {
	current := NewDatabaseSymbol("db", NewTableSymbol("Current", NewColumnSymbol("X", ScalarTypes.String.Type())))
	samples := NewDatabaseSymbol("Samples", NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type())))
	currentCluster := NewClusterSymbol("current.kusto.windows.net", current)
	helpCluster := NewClusterSymbol("help.kusto.windows.net", samples)
	globals := NewGlobalState(current).
		WithClusterList([]*ClusterSymbol{currentCluster, helpCluster}).
		WithCluster(currentCluster).
		WithDatabase(current)

	for _, test := range []struct {
		name  string
		query string
		want  string
	}{
		{name: "qualified short cluster", query: `cluster("HELP").database("samples").T | project A`, want: "A"},
		{name: "qualified URI cluster", query: `cluster("https://help.kusto.windows.net").database("Samples").T | project A`, want: "A"},
		{name: "current database", query: `database().Current | project X`, want: "X"},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.query)
			NewBinder(globals).TryBind(query)
			if diagnostics := GetContainedDiagnosticsPreserveDuplicates(query, DiagnosticsIncludeDefault); len(diagnostics) != 0 {
				references := boundNodes[*NameReference](query)
				symbols := make([]string, 0, len(references))
				for _, reference := range references {
					symbols = append(symbols, NameReferenceSimpleName(reference)+":"+symbolDebugType(reference.ReferencedSymbol()))
				}
				t.Fatalf("diagnostics = %#v; symbols = %v", diagnostics, symbols)
			}
			output := firstBoundNode[*ProjectOperator](t, query)
			table := semanticTable(t, output)
			if table.GetColumn(test.want) == nil {
				t.Fatalf("project result = %#v, want column %q", table, test.want)
			}
		})
	}
}

func symbolDebugType(symbol Symbol) string {
	if symbol == nil {
		return "<nil>"
	}
	return string(symbol.SymbolKind()) + "/" + symbol.SymbolName()
}

func TestNativeSourceReferenceFunctionsDiagnoseUnknownAndMalformedNames(t *testing.T) {
	database := NewDatabaseSymbol("db", NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type())))
	currentCluster := NewClusterSymbol("current.kusto.windows.net", database)
	globals := NewGlobalState(database).
		WithClusterList([]*ClusterSymbol{currentCluster}).
		WithCluster(currentCluster).
		WithDatabase(database)

	for _, test := range []struct {
		name  string
		query string
		code  string
	}{
		{name: "unknown cluster", query: `cluster("missing").database("db").T`, code: "KS207"},
		{name: "unknown database", query: `database("missing").T`, code: "KS208"},
		{name: "missing cluster argument", query: `cluster().database("db").T`, code: "KS119"},
		{name: "missing database argument", query: `database(`, code: "KS005"},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.query)
			NewBinder(globals).TryBind(query)
			diagnostics := GetContainedDiagnosticsPreserveDuplicates(query, DiagnosticsIncludeDefault)
			found := false
			for _, diagnostic := range diagnostics {
				if diagnostic.Code == test.code {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("diagnostics = %#v, want %s", diagnostics, test.code)
			}
		})
	}
}

func TestNativeFuzzyUnionBindsSourceWarningsAndRestoresState(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	for _, test := range []struct {
		name  string
		query string
		kind  string
		open  bool
	}{
		{name: "bare table", query: `union isfuzzy=true Missing, T`, kind: "table", open: true},
		{name: "table intrinsic", query: `union isfuzzy=true (table("Missing")), T`, kind: "table", open: true},
		{name: "external table", query: `union isfuzzy=true (external_table("Missing")), T`, kind: "external table"},
		{name: "materialized view", query: `union isfuzzy=true (materialized_view("Missing")), T`, kind: "materialized view"},
		{name: "stored result", query: `union isfuzzy=true (stored_query_result("Missing")), T`, kind: "stored query result"},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.query)
			binder := NewBinder(globals)
			if !binder.TryBind(query) {
				t.Fatal("fuzzy union was not bound")
			}
			diagnostics := GetContainedDiagnosticsPreserveDuplicates(query, DiagnosticsIncludeDefault)
			if len(diagnostics) != 1 || diagnostics[0].Code != "KS205" || diagnostics[0].Severity != DiagnosticSeverityWarning ||
				!strings.Contains(diagnostics[0].Message, "currently accessible "+test.kind) {
				t.Fatalf("diagnostics = %#v, want fuzzy %s warning", diagnostics, test.kind)
			}
			union := firstBoundNode[*UnionOperator](t, query)
			if result := semanticTable(t, union); result.GetColumn("A") == nil || result.IsOpen() != test.open {
				t.Fatalf("union result = %#v, want open=%t and column A", result, test.open)
			}
			ordinary := binder.ResolveName("Other", SymbolMatchTable, true)
			if len(ordinary.Diagnostics) != 1 || ordinary.Diagnostics[0].Code != "KS204" {
				t.Fatalf("fuzzy state leaked after union: %#v", ordinary.Diagnostics)
			}
		})
	}
}

func TestNativeDatabaseSourceFunctionsPreserveConfiguredSymbols(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	external := NewExternalTableSymbol("External", NewColumnSymbol("E", ScalarTypes.String.Type()))
	view := NewMaterializedViewSymbol("MV", []*ColumnSymbol{NewColumnSymbol("M", ScalarTypes.Real.Type())}, "")
	group := NewEntityGroupSymbol("Group", table)
	stored := NewStoredQueryResultSymbol("Stored", NewColumnSymbol("S", ScalarTypes.DateTime.Type()))
	edge := NewTableSymbol("Edge", NewColumnSymbol("Source", ScalarTypes.String.Type()))
	node := NewTableSymbol("Node", NewColumnSymbol("NodeId", ScalarTypes.String.Type()))
	model := NewGraphModelSymbol("Model", []*Signature{NewSymbolSignature(edge)}, []*Signature{NewSymbolSignature(node)},
		[]*GraphSnapshotSymbol{NewGraphSnapshotSymbol("daily")})
	samples := NewDatabaseSymbol("Samples", table, external, view, group, stored, model)
	current := NewDatabaseSymbol("db")
	currentCluster := NewClusterSymbol("current.kusto.windows.net", current)
	helpCluster := NewClusterSymbol("help.kusto.windows.net", samples)
	globals := NewGlobalState(current).
		WithClusterList([]*ClusterSymbol{currentCluster, helpCluster}).
		WithCluster(currentCluster).
		WithDatabase(current)
	binder := NewBinder(globals)

	for _, test := range []struct {
		name  string
		query string
		call  string
		want  Symbol
	}{
		{name: "table", query: `cluster("help").database("Samples").table("T")`, call: "table", want: table},
		{name: "external", query: `cluster("help").database("Samples").external_table("External")`, call: "external_table", want: external},
		{name: "materialized view", query: `cluster("help").database("Samples").materialized_view("MV")`, call: "materialized_view", want: view},
		{name: "entity group", query: `cluster("help").database("Samples").entity_group("Group")`, call: "entity_group", want: group},
		{name: "stored result", query: `cluster("help").database("Samples").stored_query_result("Stored")`, call: "stored_query_result", want: stored},
	} {
		t.Run(test.name, func(t *testing.T) {
			expression := ParseExpression(test.query)
			binder.TryBind(expression)
			call := namedFunctionCall(t, expression, test.call)
			result, _ := call.SyntaxNode.Payload(functionCallResultSymbolPayload)
			if result != test.want {
				t.Fatalf("result symbol = %#v, want exact %#v", result, test.want)
			}
			if diagnostics := GetContainedDiagnosticsPreserveDuplicates(expression, DiagnosticsIncludeDefault); len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
		})
	}

	graphExpression := ParseExpression(`cluster("help").database("Samples").graph("Model")`)
	binder.TryBind(graphExpression)
	graphCall := namedFunctionCall(t, graphExpression, "graph")
	graphResult, _ := graphCall.SyntaxNode.Payload(functionCallResultSymbolPayload)
	graph, ok := graphResult.(*GraphSymbol)
	if !ok || graph.EdgeShape.GetColumn("Source") == nil || graph.NodeShape.GetColumn("NodeId") == nil {
		t.Fatalf("graph result = %#v", graphResult)
	}
	repeated := ParseExpression(`cluster("help").database("Samples").graph("Model")`)
	binder.TryBind(repeated)
	repeatedResult, _ := namedFunctionCall(t, repeated, "graph").SyntaxNode.Payload(functionCallResultSymbolPayload)
	if repeatedResult != graphResult {
		t.Fatal("graph model result identity was not cached")
	}

	missingSnapshot := ParseExpression(`cluster("help").database("Samples").graph("Model", "missing")`)
	binder.TryBind(missingSnapshot)
	assertSemanticDiagnostic(t, namedFunctionCall(t, missingSnapshot, "graph"), "KS261")
}

func TestNativeEntityGroupSourceFunctionBindsDefinitionInDeclarationContext(t *testing.T) {
	ownerTable := NewTableSymbol("T", NewColumnSymbol("Owner", ScalarTypes.String.Type()))
	auxTable := NewTableSymbol("U", NewColumnSymbol("Aux", ScalarTypes.Long.Type()))
	group := NewEntityGroupSymbolFromDefinition("Targets", `entity_group [T, database("Aux").U]`, "")
	owner := NewDatabaseSymbol("Owner", ownerTable, group)
	aux := NewDatabaseSymbol("Aux", auxTable)
	remoteCluster := NewClusterSymbol("remote.kusto.windows.net", owner, aux)

	callerTable := NewTableSymbol("T", NewColumnSymbol("Caller", ScalarTypes.Bool.Type()))
	caller := NewDatabaseSymbol("Caller", callerTable)
	callerCluster := NewClusterSymbol("caller.kusto.windows.net", caller)
	globals := NewGlobalState(caller).
		WithClusterList([]*ClusterSymbol{callerCluster, remoteCluster}).
		WithCluster(callerCluster).
		WithDatabase(caller)

	expression := ParseExpression(`cluster("remote").database("Owner").entity_group("Targets")`)
	NewBinder(globals).TryBind(expression)
	call := namedFunctionCall(t, expression, "entity_group")
	result, _ := call.SyntaxNode.Payload(functionCallResultSymbolPayload)
	computed, ok := result.(*EntityGroupSymbol)
	if !ok {
		t.Fatalf("computed entity group = %T, want *EntityGroupSymbol", result)
	}
	members := computed.Members()
	if len(members) != 2 || members[0] != ownerTable || members[1] != auxTable {
		t.Fatalf("computed members = %#v, want declaration-context tables %#v, %#v", members, ownerTable, auxTable)
	}
	if computed == group {
		t.Fatal("definition-backed lookup returned the declaration instead of its computed body result")
	}
	if call.GetCalledFunctionBody() == nil {
		t.Fatal("definition-backed lookup did not retain its bound expansion")
	}
	if diagnostics := GetContainedDiagnosticsPreserveDuplicates(expression, DiagnosticsIncludeDefault|DiagnosticsIncludeExpansion); len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestNativeEntityGroupSourceFunctionPreservesMemberIdentityAndBlocksRecursion(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	members := NewEntityGroupSymbol("Members", table)
	loop := NewEntityGroupSymbolFromDefinition("Loop", `entity_group("Loop")`, "")
	database := NewDatabaseSymbol("db", table, members, loop)
	binder := NewBinder(NewGlobalState(database))

	memberExpression := ParseExpression(`entity_group("Members")`)
	binder.TryBind(memberExpression)
	memberResult, _ := namedFunctionCall(t, memberExpression, "entity_group").SyntaxNode.Payload(functionCallResultSymbolPayload)
	if memberResult != members {
		t.Fatalf("member-backed result = %#v, want exact %#v", memberResult, members)
	}
	memberCall := namedFunctionCall(t, memberExpression, "entity_group")
	if memberCall.ReferencedSymbol() != members || memberCall.Name().ReferencedSymbol() != Functions.EntityGroup {
		t.Fatalf("member-backed semantics = call:%#v name:%#v", memberCall.ReferencedSymbol(), memberCall.Name().ReferencedSymbol())
	}

	macroQuery := ParseQuery(`macro-expand entity_group("Members") as Scope (Scope)`)
	binder.TryBind(macroQuery)
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*MacroExpandOperator](t, macroQuery)), "A")

	loopExpression := ParseExpression(`entity_group("Loop")`)
	binder.TryBind(loopExpression)
	loopCall := namedFunctionCall(t, loopExpression, "entity_group")
	loopResult, _ := loopCall.SyntaxNode.Payload(functionCallResultSymbolPayload)
	computed, ok := loopResult.(*EntityGroupSymbol)
	if !ok || len(computed.Members()) != 0 {
		t.Fatalf("recursive result = %#v, want empty recovery entity group", loopResult)
	}
}

func TestNativeEntityGroupDefinitionDiagnosticsAreReportedAsExpansionErrors(t *testing.T) {
	group := NewEntityGroupSymbolFromDefinition("Broken", `entity_group [Missing]`, "")
	database := NewDatabaseSymbol("db", group)
	expression := ParseExpression(`entity_group("Broken")`)
	NewBinder(NewGlobalState(database)).TryBind(expression)

	diagnostics := GetContainedDiagnosticsPreserveDuplicates(expression, DiagnosticsIncludeDefault|DiagnosticsIncludeExpansion)
	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == "KS200" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("diagnostics = %#v, want KS200 expansion error", diagnostics)
	}
}

func TestNativeSourceReferenceRecoverySymbolsAreStable(t *testing.T) {
	database := NewDatabaseSymbol("db")
	cluster := NewClusterSymbol("current.kusto.windows.net", database)
	globals := NewGlobalState(database).WithClusterList([]*ClusterSymbol{cluster}).WithCluster(cluster).WithDatabase(database)
	binder := NewBinder(globals)

	firstCluster := ParseExpression(`cluster("missing")`)
	secondCluster := ParseExpression(`cluster("missing")`)
	binder.TryBind(firstCluster)
	binder.TryBind(secondCluster)
	firstClusterResult, _ := namedFunctionCall(t, firstCluster, "cluster").SyntaxNode.Payload(functionCallResultSymbolPayload)
	secondClusterResult, _ := namedFunctionCall(t, secondCluster, "cluster").SyntaxNode.Payload(functionCallResultSymbolPayload)
	if firstClusterResult != secondClusterResult {
		t.Fatal("unknown cluster recovery identity was not cached")
	}

	firstDatabase := ParseExpression(`database("missing")`)
	secondDatabase := ParseExpression(`database("missing")`)
	binder.TryBind(firstDatabase)
	binder.TryBind(secondDatabase)
	firstDatabaseResult, _ := namedFunctionCall(t, firstDatabase, "database").SyntaxNode.Payload(functionCallResultSymbolPayload)
	secondDatabaseResult, _ := namedFunctionCall(t, secondDatabase, "database").SyntaxNode.Payload(functionCallResultSymbolPayload)
	if firstDatabaseResult != secondDatabaseResult {
		t.Fatal("unknown database recovery identity was not cached")
	}

	firstTable := ParseExpression(`database("db").table("Missing")`)
	secondTable := ParseExpression(`database("db").table("Missing")`)
	binder.TryBind(firstTable)
	binder.TryBind(secondTable)
	firstTableResult, _ := namedFunctionCall(t, firstTable, "table").SyntaxNode.Payload(functionCallResultSymbolPayload)
	secondTableResult, _ := namedFunctionCall(t, secondTable, "table").SyntaxNode.Payload(functionCallResultSymbolPayload)
	if firstTableResult != secondTableResult {
		t.Fatal("unknown table recovery identity was not cached")
	}
	assertSemanticDiagnostic(t, namedFunctionCall(t, firstTable, "table"), "KS204")
}

func TestNativeDatabaseSourceFunctionsUseIntrinsicDiagnostics(t *testing.T) {
	database := NewDatabaseSymbol("db")
	binder := NewBinder(NewGlobalState(database))
	for _, test := range []struct {
		query string
		call  string
		code  string
	}{
		{query: `database("db").table("Missing")`, call: "table", code: "KS204"},
		{query: `database("db").external_table("Missing")`, call: "external_table", code: "KS209"},
		{query: `database("db").materialized_view("Missing")`, call: "materialized_view", code: "KS210"},
		{query: `database("db").entity_group("Missing")`, call: "entity_group", code: "KS247"},
		{query: `database("db").stored_query_result("Missing")`, call: "stored_query_result", code: "KS248"},
		{query: `database("db").graph("Missing")`, call: "graph", code: "KS260"},
	} {
		expression := ParseExpression(test.query)
		binder.TryBind(expression)
		assertSemanticDiagnostic(t, namedFunctionCall(t, expression, test.call), test.code)
	}
}

func TestNativeTableFunctionPrefersLocalTablesAndZeroArgumentViews(t *testing.T) {
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db",
		NewTableSymbol("Local", NewColumnSymbol("DatabaseValue", ScalarTypes.String.Type())),
		NewTableSymbol("View", NewColumnSymbol("DatabaseViewValue", ScalarTypes.String.Type())),
	)))
	localTable := NewTableSymbol("LocalShape", NewColumnSymbol("LocalValue", ScalarTypes.Long.Type()))
	viewTable := NewTableSymbol("ViewShape", NewColumnSymbol("ViewValue", ScalarTypes.Bool.Type()))
	binder.LocalScope.AddSymbol(NewVariableSymbol("Local", localTable, false, nil, nil))
	binder.LocalScope.AddSymbol(NewFunctionSymbol("View", NewSymbolSignature(viewTable)))

	for _, test := range []struct {
		name string
		want Symbol
	}{
		{name: "Local", want: localTable},
		{name: "View", want: viewTable},
	} {
		expression := ParseExpression(`table("` + test.name + `")`)
		binder.TryBind(expression)
		result, _ := namedFunctionCall(t, expression, "table").SyntaxNode.Payload(functionCallResultSymbolPayload)
		if result != test.want {
			t.Fatalf("table(%q) = %#v, want exact local %#v", test.name, result, test.want)
		}
	}
}

func namedFunctionCall(t *testing.T, root SyntaxElement, name string) *FunctionCallExpression {
	t.Helper()
	for _, call := range boundNodes[*FunctionCallExpression](root) {
		if NameReferenceSimpleName(call.Name()) == name {
			return call
		}
	}
	t.Fatalf("function call %q not found", name)
	return nil
}
