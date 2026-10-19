package kustolang

import "testing"

func TestGraphSymbolBasics(t *testing.T) {
	edge := NewTableSymbol("Edges", NewColumnSymbol("Source", ScalarTypes.String.Type()))
	node := NewTableSymbol("Nodes", NewColumnSymbol("Id", ScalarTypes.String.Type()))
	graph := NewGraphSymbol("G", edge, nil)

	if graph.SymbolName() != "G" || graph.SymbolKind() != SymbolKindGraph || graph.SymbolTabularity() != TabularityOther {
		t.Fatalf("graph metadata = name:%q kind:%s tabularity:%s", graph.SymbolName(), graph.SymbolKind(), graph.SymbolTabularity())
	}
	if graph.EdgeShape != edge || graph.NodeShape != nil {
		t.Fatalf("graph shapes = edge:%#v node:%#v", graph.EdgeShape, graph.NodeShape)
	}
	if emptyEdge := NewGraphSymbol("EmptyEdge", nil, nil); emptyEdge.EdgeShape != EmptyTableSymbol {
		t.Fatalf("nil edge shape should use empty table, got %#v", emptyEdge.EdgeShape)
	}
	if EmptyGraphSymbol.EdgeShape != EmptyTableSymbol || EmptyGraphSymbol.NodeShape != EmptyTableSymbol {
		t.Fatalf("empty graph shapes = edge:%#v node:%#v", EmptyGraphSymbol.EdgeShape, EmptyGraphSymbol.NodeShape)
	}

	renamed := graph.WithName("H")
	if renamed == graph || renamed.SymbolName() != "H" || renamed.EdgeShape != edge {
		t.Fatalf("WithName = %#v", renamed)
	}
	if graph.WithEdgeShape(edge) != graph {
		t.Fatal("WithEdgeShape should preserve receiver when edge shape is unchanged")
	}
	newEdge := NewTableSymbol("Edges2", NewColumnSymbol("Target", ScalarTypes.String.Type()))
	updatedEdge := graph.WithEdgeShape(newEdge)
	if updatedEdge == graph || updatedEdge.EdgeShape != newEdge || updatedEdge.NodeShape != nil {
		t.Fatalf("WithEdgeShape changed graph = %#v", updatedEdge)
	}
	if graph.WithNodeShape(node) != graph {
		t.Fatal("WithNodeShape should preserve receiver when current node shape is nil")
	}
	graphWithNode := NewGraphSymbol("G", edge, node)
	newNode := NewTableSymbol("Nodes2", NewColumnSymbol("Name", ScalarTypes.String.Type()))
	updatedNode := graphWithNode.WithNodeShape(newNode)
	if updatedNode == graphWithNode || updatedNode.NodeShape != newNode || updatedNode.EdgeShape != edge {
		t.Fatalf("WithNodeShape changed graph = %#v", updatedNode)
	}
}

func TestGraphSymbolCombinesShapes(t *testing.T) {
	edgeA := NewTableSymbol("A",
		NewColumnSymbol("Source", ScalarTypes.String.Type()),
		NewColumnSymbol("Weight", ScalarTypes.Int.Type()),
	)
	edgeB := NewTableSymbol("B",
		NewColumnSymbol("Weight", ScalarTypes.Real.Type()),
		NewColumnSymbol("Target", ScalarTypes.String.Type()),
	)
	nodeA := NewTableSymbol("N1",
		NewColumnSymbol("NodeId", ScalarTypes.String.Type()),
		NewColumnSymbol("Label", ScalarTypes.String.Type()),
	)
	nodeB := NewTableSymbol("N2",
		NewColumnSymbol("Label", ScalarTypes.Long.Type()),
		NewColumnSymbol("TenantId", ScalarTypes.String.Type()),
	)

	graph := NewGraphSymbolFromShapes("G", []*TableSymbol{edgeA, edgeB}, []*TableSymbol{nodeA, nodeB})
	requireTableColumn(t, graph.EdgeShape, "Source", ScalarTypes.String.Type())
	requireTableColumn(t, graph.EdgeShape, "Weight", ScalarTypes.Real.Type())
	requireTableColumn(t, graph.EdgeShape, "Target", ScalarTypes.String.Type())
	requireColumnCount(t, graph.EdgeShape, 3)

	requireTableColumn(t, graph.NodeShape, "NodeId", ScalarTypes.String.Type())
	requireTableColumn(t, graph.NodeShape, "Label", ScalarTypes.Dynamic.Type())
	requireTableColumn(t, graph.NodeShape, "TenantId", ScalarTypes.String.Type())
	requireColumnCount(t, graph.NodeShape, 3)

	nodeOnly := NewGraphSymbolFromNodeShapes("G", edgeA, []*TableSymbol{nodeA, nodeB})
	if nodeOnly.EdgeShape != edgeA {
		t.Fatalf("node-shape constructor edge = %#v", nodeOnly.EdgeShape)
	}
	requireTableColumn(t, nodeOnly.NodeShape, "Label", ScalarTypes.Dynamic.Type())
}

func TestMergeGraphSymbols(t *testing.T) {
	leftEdge := NewTableSymbol("LeftEdges",
		NewColumnSymbol("Source", ScalarTypes.String.Type()),
		NewColumnSymbol("Weight", ScalarTypes.Int.Type()),
	)
	rightEdge := NewTableSymbol("RightEdges",
		NewColumnSymbol("Weight", ScalarTypes.Real.Type()),
		NewColumnSymbol("Target", ScalarTypes.String.Type()),
	)
	leftNode := NewTableSymbol("LeftNodes", NewColumnSymbol("Name", ScalarTypes.String.Type()))
	rightNode := NewTableSymbol("RightNodes", NewColumnSymbol("Name", ScalarTypes.Long.Type()))

	left := NewGraphSymbol("Left", leftEdge, leftNode)
	right := NewGraphSymbol("Right", rightEdge, nil)
	merged := MergeGraphSymbols(left, right)
	if merged == nil || merged.SymbolName() != "" {
		t.Fatalf("merged graph = %#v", merged)
	}
	requireTableColumn(t, merged.EdgeShape, "Source", ScalarTypes.String.Type())
	requireTableColumn(t, merged.EdgeShape, "Weight", ScalarTypes.Real.Type())
	requireTableColumn(t, merged.EdgeShape, "Target", ScalarTypes.String.Type())
	if merged.NodeShape != leftNode {
		t.Fatalf("single-sided node shape should be preserved, got %#v", merged.NodeShape)
	}

	mergedNodes := MergeGraphSymbols(left, NewGraphSymbol("Right", rightEdge, rightNode))
	requireTableColumn(t, mergedNodes.NodeShape, "Name", ScalarTypes.Dynamic.Type())
	if MergeGraphSymbols(left, nil) != left || MergeGraphSymbols(nil, right) != right || MergeGraphSymbols(nil, nil) != nil {
		t.Fatal("nil graph merge cases failed")
	}
}

func TestGraphSymbolMatches(t *testing.T) {
	graph := NewGraphSymbol("MyGraph", EmptyTableSymbol, nil)
	if !SymbolMatches(graph, SymbolMatchGraph) || !SymbolMatches(graph, SymbolMatchDefault) {
		t.Fatal("graph should match graph and default symbol matches")
	}
	if !SymbolMatchesName(graph, "mygraph", SymbolMatchGraph, true) {
		t.Fatal("case-insensitive graph match failed")
	}
	if SymbolMatchesName(graph, "mygraph", SymbolMatchGraph, false) {
		t.Fatal("case-sensitive graph match should fail")
	}
	if !SymbolMatches(graph, SymbolMatchGraph|SymbolMatchTabular) {
		t.Fatal("graph flag should match before tabularity filters, like upstream")
	}
	if SymbolMatches(graph, SymbolMatchTabular) || SymbolMatches(graph, SymbolMatchFunction) {
		t.Fatal("graph should not match unrelated filters without graph flag")
	}
}

func requireTableColumn(t *testing.T, table *TableSymbol, name string, want *TypeSymbol) {
	t.Helper()
	if table == nil {
		t.Fatalf("missing table for column %q", name)
	}
	for _, column := range table.Columns() {
		if column.SymbolName() == name {
			if !sameType(column.Type, want) {
				t.Fatalf("column %q type = %s, want %s", name, column.Type.SymbolName(), want.SymbolName())
			}
			return
		}
	}
	t.Fatalf("missing column %q in %#v", name, table.Columns())
}

func requireColumnCount(t *testing.T, table *TableSymbol, want int) {
	t.Helper()
	if table == nil {
		t.Fatalf("column count on nil table, want %d", want)
	}
	if got := len(table.Columns()); got != want {
		t.Fatalf("column count = %d, want %d (%#v)", got, want, table.Columns())
	}
}
