package kustolang

import "testing"

func TestNativeBinderMaterializedViewCombineDeltaScopeResultAndDownstream(t *testing.T) {
	key := NewColumnSymbol("Key", ScalarTypes.String.Type())
	value := NewColumnSymbol("Value", ScalarTypes.Long.Type())
	base := NewTableSymbol("Base", NewColumnSymbol("BaseOnly", ScalarTypes.Bool.Type()))
	delta := NewTableSymbol("Delta", key, value)
	query := ParseQuery(`materialized-view-combine("v") base(Base) delta(Delta) aggregations(summarize Total=sum(Value) by Key) | project Key, Total`)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", base, delta)))
	binder.TryBind(query)

	combine := firstBoundNode[*MaterializedViewCombineExpression](t, query)
	summarize := firstBoundNode[*SummarizeOperator](t, query)
	combineTable := semanticTable(t, combine)
	if combineTable != semanticTable(t, summarize) {
		t.Fatal("combine result did not preserve the aggregation table")
	}
	assertOrderedColumns(t, combineTable, "Key", "Total")
	if firstNamedReference(t, summarize, "Value").ReferencedSymbol() != value ||
		firstNamedReference(t, summarize, "Key").ReferencedSymbol() != key {
		t.Fatal("aggregation expressions were not bound against the delta row scope")
	}
	project := firstBoundNode[*ProjectOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, project), "Key", "Total")
	if project.Expressions().At(0).Element().(*NameReference).ReferencedSymbol() != combineTable.GetColumn("Key") {
		t.Fatal("downstream operator did not receive the combine result as row scope")
	}
}

func TestNativeBinderMaterializedViewCombineNameRequiresExactStringLiteral(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	tests := []struct {
		name string
		text string
		want []string
	}{
		{name: "literal", text: `materialized-view-combine("v") base(T) delta(T) aggregations(summarize count())`},
		{name: "wrong exact type", text: `materialized-view-combine(1) base(T) delta(T) aggregations(summarize count())`, want: []string{"KS141"}},
		{name: "string constant is not literal", text: `let V = "v"; materialized-view-combine(V) base(T) delta(T) aggregations(summarize count())`, want: []string{"KS131"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseQuery(tt.text)
			NewBinder(globals).TryBind(query)
			clause := firstBoundNode[*MaterializedViewCombineNameClause](t, query)
			assertDiagnosticCodes(t, clause, tt.want...)
			if tt.name == "wrong exact type" && hasSemanticDiagnostic(clause, "KS131") {
				t.Fatal("wrongly typed name also produced a literal diagnostic")
			}
		})
	}
}

func TestNativeBinderMaterializedViewCombineMalformedRecoveryAndScopeRestoration(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	ambient := NewTableSymbol("Ambient", NewColumnSymbol("OuterOnly", ScalarTypes.Bool.Type()))
	binder.rowScope = ambient
	markerScope := binder.LocalScope
	tree := newTreeBinder(binder)
	tree.nodeBinder.inPipe = true

	query := ParseQuery(`materialized-view-combine("v") base(T) delta()`)
	combine := firstBoundNode[*MaterializedViewCombineExpression](t, query)
	tree.bind(combine)
	if !combine.IsBound() || ExpressionResultType(combine) != &Error.TypeSymbol {
		t.Fatalf("malformed combine = bound:%t type:%v, want bound error recovery", combine.IsBound(), ExpressionResultType(combine))
	}
	if len(combine.ContainedDiagnostics()) == 0 {
		t.Fatal("malformed combine lost parser recovery diagnostics")
	}
	if binder.rowScope != ambient || binder.LocalScope != markerScope || !tree.nodeBinder.inPipe {
		t.Fatal("malformed combine leaked binder traversal state")
	}
}

func TestNativeBinderMaterializedViewCombineRestoresOuterScopeAndRebindsIdempotently(t *testing.T) {
	deltaValue := NewColumnSymbol("DeltaValue", ScalarTypes.Long.Type())
	base := NewTableSymbol("Base", NewColumnSymbol("BaseValue", ScalarTypes.String.Type()))
	delta := NewTableSymbol("Delta", deltaValue)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", base, delta)))
	ambient := NewTableSymbol("Ambient", NewColumnSymbol("OuterOnly", ScalarTypes.Bool.Type()))
	binder.rowScope = ambient
	tree := newTreeBinder(binder)
	tree.nodeBinder.inPipe = true
	query := ParseQuery(`materialized-view-combine("v") base(Base) delta(Delta) aggregations(summarize Total=sum(DeltaValue))`)
	combine := firstBoundNode[*MaterializedViewCombineExpression](t, query)

	tree.bind(combine)
	firstInfo, ok := combine.GetSemanticInfo()
	if !ok {
		t.Fatal("combine semantic info missing")
	}
	firstResult := firstInfo.ReferencedSymbol()
	if binder.rowScope != ambient || !tree.nodeBinder.inPipe {
		t.Fatal("valid combine leaked outer binder state")
	}

	poison := NewTableSymbol("Poison", NewColumnSymbol("DeltaValue", ScalarTypes.String.Type()))
	binder.rowScope = poison
	tree.bind(combine)
	secondInfo, ok := combine.GetSemanticInfo()
	if !ok || secondInfo.ReferencedSymbol() != firstResult || len(secondInfo.Diagnostics) != len(firstInfo.Diagnostics) {
		t.Fatalf("rebinding changed combine semantics: before=%#v after=%#v", firstInfo, secondInfo)
	}
	if binder.rowScope != poison || !tree.nodeBinder.inPipe {
		t.Fatal("idempotent rebinding changed ambient state")
	}
}
