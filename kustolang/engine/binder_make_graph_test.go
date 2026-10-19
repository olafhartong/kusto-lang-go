package kustolang

import "testing"

func TestNativeMakeGraphBindsEdgesAndExplicitNodeTables(t *testing.T) {
	edges := NewTableSymbol("Edges",
		NewColumnSymbol("Source", ScalarTypes.String.Type()),
		NewColumnSymbol("Target", ScalarTypes.String.Type()),
		NewColumnSymbol("Weight", ScalarTypes.Long.Type()))
	nodes := NewTableSymbol("Nodes",
		NewColumnSymbol("NodeId", ScalarTypes.String.Type()),
		NewColumnSymbol("Label", ScalarTypes.String.Type()))
	query := ParseQuery(`Edges | make-graph hint.strategy=shuffle Source --> Target with Nodes on NodeId`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", edges, nodes))).TryBind(query)

	makeGraph := firstBoundNode[*MakeGraphOperator](t, query)
	assertDiagnosticCodes(t, makeGraph)
	graph := makeGraphResult(t, makeGraph)
	if graph.EdgeShape != edges {
		t.Fatalf("edge shape = %#v, want Edges", graph.EdgeShape)
	}
	assertOrderedColumns(t, graph.NodeShape, "NodeId", "Label")
	clause := makeGraph.WithClause().(*MakeGraphWithTablesAndKeysClause).TablesAndKeys().At(0).Element()
	if column, ok := clause.Column().ReferencedSymbol().(*ColumnSymbol); !ok || column.SymbolName() != "NodeId" {
		t.Fatalf("node key = %#v, want Nodes.NodeId", clause.Column().ReferencedSymbol())
	}
}

func TestNativeMakeGraphImplicitIDAndPartitionResultOverride(t *testing.T) {
	edges := NewTableSymbol("Edges",
		NewColumnSymbol("Source", ScalarTypes.Long.Type()),
		NewColumnSymbol("Target", ScalarTypes.Long.Type()),
		NewColumnSymbol("Tenant", ScalarTypes.String.Type()))
	query := ParseQuery(`Edges | make-graph Source --> Target with_node_id=NodeId partitioned-by Tenant (where Tenant != "" | project Tenant)`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", edges))).TryBind(query)

	makeGraph := firstBoundNode[*MakeGraphOperator](t, query)
	assertDiagnosticCodes(t, makeGraph)
	result := expressionTableResult(makeGraph)
	if result == nil {
		t.Fatalf("partitioned make-graph result = %T, want table", makeGraph.ReferencedSymbol())
	}
	assertOrderedColumns(t, result, "Tenant")
	implicit := makeGraph.WithClause().(*MakeGraphWithImplicitIdClause)
	if column, ok := implicit.Name().ReferencedSymbol().(*ColumnSymbol); !ok || column.Type != ScalarTypes.Long.Type() {
		t.Fatalf("implicit id = %#v, want long column", implicit.Name().ReferencedSymbol())
	}
	clause := makeGraph.PartitionedByClause()
	clauseInfo, ok := clause.GetSemanticInfo()
	if !ok {
		t.Fatal("partitioned-by clause is unbound")
	}
	clauseTable, ok := clauseInfo.ReferencedSymbol().(*TableSymbol)
	if !ok {
		t.Fatalf("partitioned-by clause result = %T, want table", clauseInfo.ReferencedSymbol())
	}
	assertOrderedColumns(t, clauseTable, "Tenant")
	if clauseTable.IsSorted() {
		t.Fatal("partitioned-by clause result retained sorted state")
	}
}

func TestNativeMakeGraphPartitionedByClauseMetadataAndDiagnostics(t *testing.T) {
	edges := NewTableSymbol("Edges",
		NewColumnSymbol("Source", ScalarTypes.String.Type()),
		NewColumnSymbol("Target", ScalarTypes.String.Type()),
		NewColumnSymbol("Tenant", ScalarTypes.String.Type())).
		WithIsOpen(true).WithIsSerialized(true).WithIsSorted(true)
	query := ParseQuery(`Edges | make-graph Source --> Target partitioned-by Edges (print Tenant="x")`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", edges))).TryBind(query)

	makeGraph := firstBoundNode[*MakeGraphOperator](t, query)
	clause := makeGraph.PartitionedByClause()
	info, ok := clause.GetSemanticInfo()
	if !ok {
		t.Fatal("partitioned-by clause is unbound")
	}
	if !diagnosticsContain(info.Diagnostics, "KS109") {
		t.Fatalf("partitioned-by diagnostics = %#v, want column-expected diagnostic", info.Diagnostics)
	}
	if !diagnosticsContain(info.Diagnostics, "KS177") {
		t.Fatalf("partitioned-by diagnostics = %#v, want disallowed print diagnostic", info.Diagnostics)
	}
	table, ok := info.ReferencedSymbol().(*TableSymbol)
	if !ok {
		t.Fatalf("partitioned-by result = %T, want table", info.ReferencedSymbol())
	}
	assertOrderedColumns(t, table, "Tenant")
	if !table.IsOpen() || !table.IsSerialized() || table.IsSorted() {
		t.Fatalf("partitioned-by properties = open:%t serialized:%t sorted:%t", table.IsOpen(), table.IsSerialized(), table.IsSorted())
	}
}

func TestNativeMakeGraphRejectsDynamicIDsAndMalformedInputRestoresState(t *testing.T) {
	edges := NewTableSymbol("Edges",
		NewColumnSymbol("Source", ScalarTypes.Dynamic.Type()),
		NewColumnSymbol("Target", ScalarTypes.String.Type()))
	nodes := NewTableSymbol("Nodes", NewColumnSymbol("NodeId", ScalarTypes.Dynamic.Type()))
	query := ParseQuery(`Edges | make-graph Source --> Missing with Nodes on NodeId`)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", edges, nodes)))
	outer := NewTableSymbol("Outer", NewColumnSymbol("Marker", ScalarTypes.String.Type()))
	local := binder.LocalScope
	binder.rowScope, binder.scopeKind = outer, ScopeKindAggregate
	tree := newTreeBinder(binder)
	tree.visitElement(query)

	makeGraph := firstBoundNode[*MakeGraphOperator](t, query)
	if !hasSemanticDiagnostic(makeGraph, "KS240") || !hasSemanticDiagnostic(makeGraph.TargetColumn(), "KS142") {
		t.Fatalf("make-graph diagnostics = %#v target diagnostics = %#v, want KS240 and unresolved-column diagnostic",
			makeGraph.SemanticDiagnostics(), makeGraph.TargetColumn().SemanticDiagnostics())
	}
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local {
		t.Fatal("make-graph binder leaked row, scope kind, or local scope")
	}
	before, _ := makeGraph.GetSemanticInfo()
	tree.visitElement(query)
	after, _ := makeGraph.GetSemanticInfo()
	if before.ReferencedSymbol() != after.ReferencedSymbol() || len(before.Diagnostics) != len(after.Diagnostics) {
		t.Fatalf("make-graph rebinding changed semantic info: before=%#v after=%#v", before, after)
	}
}

func TestNativeDetachedMakeGraphUsesStandaloneScopeAndFeedsFollowingGraphOperator(t *testing.T) {
	const query = `datatable(Source:long,Target:long,Weight:long)[1,2,3]  make-graph Source --> Target | graph-to-table edges with_source_id=SrcId with_target_id=DstId`
	code := ParseAndAnalyze(query)
	assertDiagnosticLocations(t, code.GetDiagnostics(),
		diagnosticLocation{"KS005", 53, 0},
		diagnosticLocation{"KS174", 55, 10},
		diagnosticLocation{"KS240", 66, 6},
		diagnosticLocation{"KS240", 77, 6},
		diagnosticLocation{"KS142", 66, 6},
		diagnosticLocation{"KS142", 77, 6},
	)
	assertOrderedColumns(t, terminalResultTable(code), "SrcId", "DstId")
}

func makeGraphResult(t *testing.T, expression Expression) *GraphSymbol {
	t.Helper()
	info := boundSemanticInfoForExpression(expression)
	if info == nil {
		t.Fatal("make-graph is unbound")
	}
	graph, ok := info.ReferencedSymbol().(*GraphSymbol)
	if !ok {
		t.Fatalf("make-graph result = %T, want graph", info.ReferencedSymbol())
	}
	return graph
}
