package kustolang

import "testing"

func TestNativeGraphMatchBindsAliasesWhereAndProjection(t *testing.T) {
	graph := testGraphSymbol()
	query := ParseQuery(`G | graph-match (a)-[e*1..3]->(b) where a.Active project Src=a.Id, Weights=e.Weight`)
	binder := NewBinder(DefaultGlobalState)
	binder.LocalScope.AddSymbol(NewVariableSymbol("G", graph, false, nil, nil))
	binder.TryBind(query)

	match := firstBoundNode[*GraphMatchOperator](t, query)
	assertDiagnosticCodes(t, match)
	result := graphOperatorTableResult(t, match)
	assertOrderedColumns(t, result, "Src", "Weights")
	if result.GetColumn("Src").Type != ScalarTypes.Long.Type() {
		t.Fatalf("Src type = %v, want long", result.GetColumn("Src").Type)
	}
	if result.GetColumn("Weights").Type != ScalarTypes.DynamicArrayOfLong.Type() {
		t.Fatalf("Weights type = %v, want dynamic(long array)", result.GetColumn("Weights").Type)
	}

	aliases := map[string]*VariableSymbol{}
	for _, declaration := range boundNodes[*NameDeclaration](query) {
		if variable, ok := declaration.ReferencedSymbol().(*VariableSymbol); ok {
			aliases[variable.SymbolName()] = variable
		}
	}
	for _, name := range []string{"a", "b", "e"} {
		if aliases[name] == nil {
			t.Fatalf("pattern alias %q was not declared", name)
		}
		if _, ok := aliases[name].Type.(*TupleSymbol); !ok {
			t.Fatalf("pattern alias %q type = %T, want tuple", name, aliases[name].Type)
		}
	}
}

func TestNativeGraphProjectionUsesQualifiedFallbackNameCurrentReferenceParity(t *testing.T) {
	const query = `let G=datatable(Source:long,Target:long,Weight:long)[1,2,3] | make-graph Source --> Target; G | graph-match (a)-[e]->(b) where e.Weight > 0 project Srca.Source, Weight=e.Weight`
	code := ParseAndAnalyze(query)
	assertDiagnosticLocations(t, code.GetDiagnostics(), diagnosticLocation{"KS142", 148, 4})
	result := terminalResultTable(code)
	assertOrderedColumns(t, result, "Srca_Source", "Weight")
	if result.Columns()[0].Type != ScalarTypes.Unknown.Type() || result.Columns()[1].Type != ScalarTypes.Long.Type() {
		t.Fatalf("graph projection types = %v, %v", result.Columns()[0].Type, result.Columns()[1].Type)
	}
}

func TestNativeGraphShortestPathsPreservesGraphWithoutProject(t *testing.T) {
	graph := testGraphSymbol()
	query := ParseQuery(`G | graph-shortest-paths (a)-[e]->(b) where e.Weight > 0`)
	binder := NewBinder(DefaultGlobalState)
	binder.LocalScope.AddSymbol(NewVariableSymbol("G", graph, false, nil, nil))
	binder.TryBind(query)

	shortest := firstBoundNode[*GraphShortestPathsOperator](t, query)
	assertDiagnosticCodes(t, shortest)
	info, ok := shortest.GetSemanticInfo()
	if !ok || info.ReferencedSymbol() != graph {
		t.Fatalf("shortest-paths result = %#v, want original graph", info.ReferencedSymbol())
	}
}

func TestNativeGraphPatternMalformedDiagnosticsAndStateRestoration(t *testing.T) {
	graph := testGraphSymbol()
	query := ParseQuery(`G | graph-match (a)-[e]-> where a.Id project Value=e.Weight`)
	binder := NewBinder(DefaultGlobalState)
	binder.LocalScope.AddSymbol(NewVariableSymbol("G", graph, false, nil, nil))
	outer := NewTableSymbol("Outer", NewColumnSymbol("Marker", ScalarTypes.String.Type()))
	local := binder.LocalScope
	binder.rowScope, binder.scopeKind = outer, ScopeKindAggregate
	tree := newTreeBinder(binder)
	tree.visitElement(query)

	match := firstBoundNode[*GraphMatchOperator](t, query)
	if !hasSemanticDiagnostic(match, "KS238") || !hasSemanticDiagnostic(match, "KS141") {
		t.Fatalf("malformed graph diagnostics = %#v, want KS238 and KS141", match.SemanticDiagnostics())
	}
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local {
		t.Fatal("graph binder leaked row, scope kind, or local scope")
	}
	before, _ := match.GetSemanticInfo()
	tree.visitElement(query)
	after, _ := match.GetSemanticInfo()
	if before.ReferencedSymbol() != after.ReferencedSymbol() || len(before.Diagnostics) != len(after.Diagnostics) {
		t.Fatalf("graph rebinding changed semantic info: before=%#v after=%#v", before, after)
	}
}

func TestNativeGraphMatchRejectsTabularInput(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Source", ScalarTypes.String.Type()))
	query := ParseQuery(`T | graph-match (a)-->(b)`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
	assertDiagnosticCodes(t, firstBoundNode[*GraphMatchOperator](t, query), "KS219")
}

func testGraphSymbol() *GraphSymbol {
	edges := NewTableSymbol("Edges",
		NewColumnSymbol("Source", ScalarTypes.String.Type()),
		NewColumnSymbol("Target", ScalarTypes.String.Type()),
		NewColumnSymbol("Weight", ScalarTypes.Long.Type()))
	nodes := NewTableSymbol("Nodes",
		NewColumnSymbol("Id", ScalarTypes.Long.Type()),
		NewColumnSymbol("Active", ScalarTypes.Bool.Type()))
	return NewGraphSymbol("G", edges, nodes)
}

func graphOperatorTableResult(t *testing.T, expression Expression) *TableSymbol {
	t.Helper()
	info, ok := expression.(interface {
		GetSemanticInfo() (SemanticInfo, bool)
	}).GetSemanticInfo()
	if !ok {
		t.Fatal("graph operator is unbound")
	}
	table, ok := info.ReferencedSymbol().(*TableSymbol)
	if !ok {
		t.Fatalf("graph operator result = %T, want table", info.ReferencedSymbol())
	}
	return table
}
