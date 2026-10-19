package kustolang

import "testing"

func TestNativeFacetResultShapeIdentityAndState(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.Long.Type())
	d := NewColumnSymbol("D", ScalarTypes.Dynamic.Type())
	table := NewTableSymbol("T", a, d).WithIsOpen(true).WithIsSerialized(true).WithIsSorted(true)
	query := ParseQuery(`T | facet by A, D with (project A)`)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.TryBind(query)

	facet := firstBoundNode[*FacetOperator](t, query)
	members := facetResultGroup(t, facet).Members()
	if len(members) != 3 {
		t.Fatalf("facet members = %d, want with result plus two facets", len(members))
	}
	withResult := members[0].(*TableSymbol)
	first := members[1].(*TableSymbol)
	second := members[2].(*TableSymbol)
	if withResult.SymbolName() != "" || first.SymbolName() != "Facet" || second.SymbolName() != "Facet_2" {
		t.Fatalf("facet member names = %q, %q, %q", withResult.SymbolName(), first.SymbolName(), second.SymbolName())
	}
	if withResult.GetColumn("A") != a || first.GetColumn("A") != a || second.GetColumn("D") != d {
		t.Fatal("facet did not preserve source column identity")
	}
	withExpression := facetWithExpression(facet.WithClause())
	if withResult != expressionTableResult(withExpression) {
		t.Fatal("facet cloned the with-clause result table")
	}
	assertOrderedColumns(t, first, "A", "count_A")
	assertOrderedColumns(t, second, "D", "count_D")
	if !sameType(first.GetColumn("count_A").Type, ScalarTypes.Long.Type()) ||
		!sameType(second.GetColumn("count_D").Type, ScalarTypes.Long.Type()) {
		t.Fatal("facet count columns are not long")
	}
	expressions := facet.Expressions()
	for index, test := range []struct {
		count    *ColumnSymbol
		original *ColumnSymbol
	}{
		{count: first.GetColumn("count_A"), original: a},
		{count: second.GetColumn("count_D"), original: d},
	} {
		if len(test.count.OriginalColumns) != 1 || test.count.OriginalColumns[0] != test.original ||
			test.count.Source != expressions.At(index).Element() {
			t.Fatalf("facet count %d lineage = originals:%#v source:%#v", index, test.count.OriginalColumns, test.count.Source)
		}
	}
	for i, result := range []*TableSymbol{first, second} {
		if !result.IsOpen() || !result.IsSerialized() || result.IsSorted() {
			t.Fatalf("facet %d state = open:%t serialized:%t sorted:%t", i, result.IsOpen(), result.IsSerialized(), result.IsSorted())
		}
	}
	if !withResult.IsOpen() || !withResult.IsSerialized() || !withResult.IsSorted() {
		t.Fatalf("with result state = open:%t serialized:%t sorted:%t", withResult.IsOpen(), withResult.IsSerialized(), withResult.IsSorted())
	}
	assertDiagnosticCodes(t, facet, "KS116")
}

func TestNativeFacetNonColumnResultsRetainExpressionSource(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	query := ParseQuery(`let x = 1; T | facet by x`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	facet := firstBoundNode[*FacetOperator](t, query)
	assertDiagnosticCodes(t, facet, "KS109")
	expression := facet.Expressions().At(0).Element()
	result := facetResultGroup(t, facet).Members()[0].(*TableSymbol)
	value, count := result.GetColumn("x"), result.GetColumn("count_x")
	if value == nil || count == nil || value.Source != expression || count.Source != expression ||
		len(value.OriginalColumns) != 0 || len(count.OriginalColumns) != 0 ||
		!sameType(value.Type, ScalarTypes.Long.Type()) || !sameType(count.Type, ScalarTypes.Long.Type()) {
		t.Fatalf("non-column facet lineage = value:%#v count:%#v", value, count)
	}
}

func TestNativeFacetDynamicNonColumnShortCircuitsTypeDiagnostic(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	query := ParseQuery(`let d = dynamic({"x": 1}); T | facet by d`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	facet := firstBoundNode[*FacetOperator](t, query)
	assertDiagnosticCodes(t, facet, "KS109")
	result := facetResultGroup(t, facet).Members()[0].(*TableSymbol)
	if value := result.GetColumn("d"); value == nil || !isDynamicType(value.Type) {
		t.Fatalf("dynamic non-column facet value = %#v", value)
	}
}

func TestNativeFacetDiagnosticsRecoveryAndAllowlist(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	right := NewTableSymbol("Other", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	query := ParseQuery(`T | facet by Missing, with join Other on A`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table, right))).TryBind(query)

	facet := firstBoundNode[*FacetOperator](t, query)
	assertDiagnosticCodes(t, facet, "KS177")
	members := facetResultGroup(t, facet).Members()
	if len(members) < 2 {
		t.Fatalf("recovered facet members = %d, want recovered keys retained", len(members))
	}
	missing := members[len(members)-1].(*TableSymbol)
	assertOrderedColumns(t, missing, "", "count_")
	if !sameType(missing.GetColumn("").Type, ScalarTypes.Unknown.Type()) {
		t.Fatalf("missing facet type = %#v, want unknown", missing.GetColumn("").Type)
	}
	missingExpression := facet.Expressions().At(facet.Expressions().Count() - 1).Element()
	if missing.GetColumn("").Source != missingExpression || missing.GetColumn("count_").Source != missingExpression ||
		len(missing.GetColumn("").OriginalColumns) != 0 || len(missing.GetColumn("count_").OriginalColumns) != 0 {
		t.Fatalf("missing facet lineage = value:%#v count:%#v", missing.GetColumn(""), missing.GetColumn("count_"))
	}

	parenthesized := ParseQuery(`T | facet by A with (join Other on A)`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table, right))).TryBind(parenthesized)
	parenthesizedFacet := firstBoundNode[*FacetOperator](t, parenthesized)
	assertDiagnosticCodes(t, parenthesizedFacet)
	parenthesizedMembers := facetResultGroup(t, parenthesizedFacet).Members()
	with := facetWithExpression(parenthesizedFacet.WithClause())
	if len(parenthesizedMembers) != 2 || parenthesizedMembers[0] != expressionTableResult(with) {
		t.Fatal("parenthesized facet join did not preserve its exact with result")
	}

	standalone := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, `facet by A`).(*FacetOperator)
	NewBinder(nil).TryBind(standalone)
	assertDiagnosticCodes(t, standalone, "KS174")
}

func TestNativeFacetDownstreamDoesNotUseGroupMembersAsRowScope(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	projectQuery := ParseQuery(`T | facet by A | project A`)
	NewBinder(globals).TryBind(projectQuery)
	assertDiagnosticLocations(t, GetContainedDiagnostics(projectQuery, DiagnosticsIncludeDefault),
		diagnosticLocation{"KS142", 25, 1},
	)
	project := semanticTable(t, firstBoundNode[*ProjectOperator](t, projectQuery))
	if column := project.GetColumn("A"); column == nil || !sameType(column.Type, ScalarTypes.Unknown.Type()) {
		t.Fatalf("downstream project result = %#v, want A:unknown", project)
	}

	for _, test := range []struct {
		query  string
		column string
	}{
		{query: `T | facet by A | count`, column: "Count"},
		{query: `T | facet by A | getschema`, column: "ColumnName"},
	} {
		query := ParseQuery(test.query)
		NewBinder(globals).TryBind(query)
		assertDiagnosticLocations(t, GetContainedDiagnostics(query, DiagnosticsIncludeDefault))
		rootPipe := firstBoundNode[*PipeExpression](t, query)
		result, ok := TableSymbolForType(ExpressionResultType(rootPipe))
		if !ok || result.GetColumn(test.column) == nil {
			t.Fatalf("downstream result for %q = %#v", test.query, ExpressionResultType(rootPipe))
		}
	}
}

func TestNativeFacetRestoresStateIsIdempotentAndKeepsGroupDownstream(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	query := ParseQuery(`T | facet by A | project A`)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	outer := NewTableSymbol("Outer", NewColumnSymbol("Marker", ScalarTypes.String.Type()))
	local := NewLocalScope(nil)
	binder.rowScope, binder.scopeKind, binder.LocalScope = outer, ScopeKindAggregate, local
	tree := newTreeBinder(binder)
	tree.visitElement(query)

	facet := firstBoundNode[*FacetOperator](t, query)
	before, _ := facet.GetSemanticInfo()
	group := before.ReferencedSymbol()
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local {
		t.Fatal("facet leaked binder state")
	}
	tree.visitElement(query)
	after, _ := facet.GetSemanticInfo()
	if after.ReferencedSymbol() != group || len(after.Diagnostics) != len(before.Diagnostics) {
		t.Fatalf("facet rebinding changed semantic info: before=%#v after=%#v", before, after)
	}

	inner := pipeWithOperator[*FacetOperator](query)
	if inner == nil {
		t.Fatal("facet pipe not found")
	}
	innerInfo, ok := inner.GetSemanticInfo()
	if !ok || innerInfo.ReferencedSymbol() != group {
		t.Fatalf("facet pipe result = %#v, want group", innerInfo)
	}
}

func pipeWithOperator[T QueryOperator](root SyntaxElement) *PipeExpression {
	if root == nil {
		return nil
	}
	if node := simpleBinderSyntaxNode(root); node != nil {
		if pipe, ok := WrapGeneratedSyntaxNode(node).(*PipeExpression); ok {
			if _, ok := pipe.Operator().(T); ok {
				return pipe
			}
		}
	}
	for i := 0; i < root.ChildCount(); i++ {
		if found := pipeWithOperator[T](root.GetChild(i)); found != nil {
			return found
		}
	}
	return nil
}

func facetResultGroup(t *testing.T, facet *FacetOperator) *GroupSymbol {
	t.Helper()
	info, ok := facet.GetSemanticInfo()
	if !ok {
		t.Fatalf("facet is unbound: syntax=%#v with=%T diagnostics=%#v", facet.Expressions(), facet.WithClause(), facet.ContainedDiagnostics())
	}
	group, ok := info.ReferencedSymbol().(*GroupSymbol)
	if !ok {
		t.Fatalf("facet result = %T, want *GroupSymbol", info.ReferencedSymbol())
	}
	return group
}
