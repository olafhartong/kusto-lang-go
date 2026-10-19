package kustolang

import "testing"

func TestNativeGraphWherePreservesShapesAndChecksExactBool(t *testing.T) {
	graph := testGraphSymbol()
	query := ParseQuery(`G | graph-where-nodes true | graph-where-edges false`)
	binder := NewBinder(DefaultGlobalState)
	binder.LocalScope.AddSymbol(NewVariableSymbol("G", graph, false, nil, nil))
	binder.TryBind(query)

	nodes := firstBoundNode[*GraphWhereNodesOperator](t, query)
	assertDiagnosticCodes(t, nodes)
	nodesResult := graphOperatorResult(t, nodes)
	if nodesResult == graph || nodesResult.EdgeShape != graph.EdgeShape || nodesResult.NodeShape != graph.NodeShape {
		t.Fatalf("graph-where-nodes result = %#v, want copied graph with preserved shapes", nodesResult)
	}
	edges := firstBoundNode[*GraphWhereEdgesOperator](t, query)
	assertDiagnosticCodes(t, edges)
	edgesResult := graphOperatorResult(t, edges)
	if edgesResult.EdgeShape != graph.EdgeShape || edgesResult.NodeShape != graph.NodeShape {
		t.Fatalf("graph-where-edges did not preserve graph shapes: %#v", edgesResult)
	}
}

func TestNativeGraphMarkComponentsAddsDefaultAndNamedColumns(t *testing.T) {
	for _, test := range []struct {
		query string
		want  string
	}{
		{`G | graph-mark-components`, "ComponentId"},
		{`G | graph-mark-components kind=strong with_component_id=Community`, "Community"},
	} {
		graph := testGraphSymbol()
		query := ParseQuery(test.query)
		binder := NewBinder(DefaultGlobalState)
		binder.LocalScope.AddSymbol(NewVariableSymbol("G", graph, false, nil, nil))
		binder.TryBind(query)
		mark := firstBoundNode[*GraphMarkComponentsOperator](t, query)
		assertDiagnosticCodes(t, mark)
		result := graphOperatorResult(t, mark)
		column := result.NodeShape.GetColumn(test.want)
		if column == nil || column.Type != ScalarTypes.Long.Type() {
			t.Fatalf("%s column = %#v, want long", test.want, column)
		}
		if result.EdgeShape != graph.EdgeShape || result.NodeShape == graph.NodeShape {
			t.Fatalf("mark-components shape mutation = %#v", result)
		}
	}

	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	query := ParseQuery(`T | graph-mark-components with_component_id=Component`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
	mark := firstBoundNode[*GraphMarkComponentsOperator](t, query)
	assertDiagnosticCodes(t, mark)
	result := graphOperatorResult(t, mark)
	if result.EdgeShape != table || result.NodeShape.GetColumn("Component") == nil {
		t.Fatalf("table-input component result = %#v, want T edges and Component node column", result)
	}
}

func TestNativeGraphToTableSingleOutputsHashColumns(t *testing.T) {
	for _, test := range []struct {
		query   string
		columns []string
	}{
		{`G | graph-to-table nodes with_node_id=NodeHash`, []string{"NodeHash", "Id", "Active"}},
		{`G | graph-to-table edges with_source_id=SourceHash with_target_id=TargetHash`, []string{"SourceHash", "TargetHash", "Source", "Target", "Weight"}},
	} {
		query := ParseQuery(test.query)
		binder := NewBinder(DefaultGlobalState)
		binder.LocalScope.AddSymbol(NewVariableSymbol("G", testGraphSymbol(), false, nil, nil))
		binder.TryBind(query)
		operator := firstBoundNode[*GraphToTableOperator](t, query)
		assertDiagnosticCodes(t, operator)
		assertOrderedColumns(t, graphOperatorTableResult(t, operator), test.columns...)
	}
}

func TestNativeGraphToTableClosesOpenShape(t *testing.T) {
	nodes := NewTableSymbol("", NewColumnSymbol("NodeId", ScalarTypes.Unknown.Type())).WithIsOpen(true)
	graph := NewGraphSymbol("", EmptyTableSymbol, nodes)
	query := ParseQuery(`G | graph-to-table nodes`)
	binder := NewBinder(DefaultGlobalState)
	binder.LocalScope.AddSymbol(NewVariableSymbol("G", graph, false, nil, nil))
	binder.TryBind(query)

	result := graphOperatorTableResult(t, firstBoundNode[*GraphToTableOperator](t, query))
	if result.IsOpen() || result.GetColumn("NodeId") == nil {
		t.Fatalf("graph-to-table result = open:%t columns:%#v, want closed with NodeId", result.IsOpen(), result.Columns())
	}
}

func TestNativeGraphToTableTwoOutputsPublishesAliases(t *testing.T) {
	query := ParseQuery(`G | graph-to-table nodes as N with_node_id=NodeHash, edges as E with_source_id=SourceHash with_target_id=TargetHash`)
	binder := NewBinder(DefaultGlobalState)
	binder.LocalScope.AddSymbol(NewVariableSymbol("G", testGraphSymbol(), false, nil, nil))
	binder.TryBind(query)

	operator := firstBoundNode[*GraphToTableOperator](t, query)
	assertDiagnosticCodes(t, operator)
	info, ok := operator.GetSemanticInfo()
	if !ok {
		t.Fatal("graph-to-table was not bound")
	}
	group, ok := info.ReferencedSymbol().(*GroupSymbol)
	if !ok || len(group.Members()) != 2 {
		t.Fatalf("graph-to-table result = %#v, want two-table group", info.ReferencedSymbol())
	}
	nodes, _ := group.Members()[0].(*TableSymbol)
	edges, _ := group.Members()[1].(*TableSymbol)
	if nodes == nil || nodes.SymbolName() != "N" || edges == nil || edges.SymbolName() != "E" {
		t.Fatalf("group members = %#v, want N and E tables", group.Members())
	}
	assertOrderedColumns(t, nodes, "NodeHash", "Id", "Active")
	assertOrderedColumns(t, edges, "SourceHash", "TargetHash", "Source", "Target", "Weight")
	if !binder.LocalScope.ContainsSymbol("N") || !binder.LocalScope.ContainsSymbol("E") {
		t.Fatal("graph-to-table aliases were not added to local scope")
	}
	clauses := operator.OutputClause().Elements()
	for i, name := range []string{"N", "E"} {
		declaration := clauses[i].Element().AsClause().Name()
		if table, ok := declaration.ReferencedSymbol().(*TableSymbol); !ok || table.SymbolName() != name {
			t.Fatalf("alias declaration %s symbol = %#v", name, declaration.ReferencedSymbol())
		}
	}
}

func TestNativeGraphOutputDiagnosticsAndMalformedSyntax(t *testing.T) {
	standalone := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, `graph-where-nodes true`).(*GraphWhereNodesOperator)
	NewBinder(nil).TryBind(standalone)
	assertDiagnosticCodes(t, standalone, "KS174")

	tests := []struct {
		query string
		codes []string
	}{
		{`T | graph-where-edges A > 0`, []string{}},
		{`G | graph-where-nodes 1`, []string{"KS141"}},
		{`G | graph-mark-components kind=sideways`, []string{"KS005"}},
		{`G | graph-to-table nodes, nodes`, []string{"KS229"}},
		{`G | graph-to-table nodes, edges`, []string{"KS273", "KS273"}},
		{`G | graph-to-table`, []string{}},
		{`G | graph-where-nodes`, []string{}},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			query := ParseQuery(test.query)
			binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type())))))
			binder.LocalScope.AddSymbol(NewVariableSymbol("G", testGraphSymbol(), false, nil, nil))
			binder.TryBind(query)
			var node SyntaxElement
			if candidate := firstOptionalBoundNode[*GraphWhereNodesOperator](query); candidate != nil {
				node = candidate
			} else if candidate := firstOptionalBoundNode[*GraphWhereEdgesOperator](query); candidate != nil {
				node = candidate
			} else if candidate := firstOptionalBoundNode[*GraphMarkComponentsOperator](query); candidate != nil {
				node = candidate
			} else if candidate := firstOptionalBoundNode[*GraphToTableOperator](query); candidate != nil {
				node = candidate
			}
			if node == nil {
				t.Fatalf("%q: graph operator not parsed", test.query)
			}
			diagnosticNode, ok := node.(interface{ SemanticDiagnostics() []Diagnostic })
			if !ok {
				t.Fatalf("%q: %T has no semantic diagnostics", test.query, node)
			}
			assertDiagnosticCodes(t, diagnosticNode, test.codes...)
		})
	}
}

func TestNativeGraphOutputRestoresStateAndIsIdempotent(t *testing.T) {
	query := ParseQuery(`G | graph-to-table nodes as N, edges as E`)
	binder := NewBinder(DefaultGlobalState)
	binder.LocalScope.AddSymbol(NewVariableSymbol("G", testGraphSymbol(), false, nil, nil))
	outer := NewTableSymbol("Outer", NewColumnSymbol("Marker", ScalarTypes.String.Type()))
	local := binder.LocalScope
	binder.rowScope, binder.scopeKind = outer, ScopeKindAggregate
	tree := newTreeBinder(binder)
	tree.visitElement(query)
	operator := firstBoundNode[*GraphToTableOperator](t, query)
	before, _ := operator.GetSemanticInfo()
	tree.visitElement(query)
	after, _ := operator.GetSemanticInfo()
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local {
		t.Fatal("graph output binder leaked row, scope kind, or local scope")
	}
	if before.ReferencedSymbol() != after.ReferencedSymbol() || len(before.Diagnostics) != len(after.Diagnostics) {
		t.Fatalf("rebinding changed semantic info: before=%#v after=%#v", before, after)
	}
}

func TestNativeGraphToTableTrailingCommaCurrentReferenceParity(t *testing.T) {
	const query = `datatable(Source:long,Target:long,Weight:long)[1,2,3] | make-graph Source --> Target | graph-to-table edges with_source_id=SrcId with_target_id=DstId ,`
	code := ParseAndAnalyze(query)
	assertDiagnosticLocations(t, code.GetDiagnostics(),
		diagnosticLocation{"KS229", 102, 49},
		diagnosticLocation{"KS230", 151, 0},
	)
	if result := terminalResultTable(code); result != nil {
		t.Fatalf("trailing-comma result = %#v, want no table", result)
	}
}

func graphOperatorResult(t *testing.T, expression Expression) *GraphSymbol {
	t.Helper()
	info, ok := expression.(interface {
		GetSemanticInfo() (SemanticInfo, bool)
	}).GetSemanticInfo()
	if !ok {
		t.Fatal("graph operator is unbound")
	}
	graph, ok := info.ReferencedSymbol().(*GraphSymbol)
	if !ok {
		t.Fatalf("graph operator result = %T, want graph", info.ReferencedSymbol())
	}
	return graph
}

func firstOptionalBoundNode[T SyntaxElement](root SyntaxElement) T {
	var zero T
	for _, node := range boundNodes[T](root) {
		return node
	}
	return zero
}
