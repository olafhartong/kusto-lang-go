package kustolang

import "testing"

func TestNativeMvApplyTwoPhaseSchemaIdentityAndItemIndex(t *testing.T) {
	items := NewColumnSymbol("Items", ScalarTypes.Dynamic.Type())
	keep := NewColumnSymbol("Keep", ScalarTypes.String.Type())
	table := NewTableSymbol("T", items, keep).WithIsOpen(true).WithIsSerialized(true).WithIsSorted(true)
	query := ParseQuery(`T | mv-apply with_itemindex=Idx Expanded=Items to typeof(string) on (project Expanded, Idx)`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	apply := firstBoundNode[*MvApplyOperator](t, query)
	assertDiagnosticCodes(t, apply)
	result := semanticTable(t, apply)
	assertOrderedColumns(t, result, "Items", "Keep", "Expanded", "Idx")
	if result.GetColumn("Keep") != keep {
		t.Fatal("mv-apply did not preserve untouched input column identity")
	}
	if result.GetColumn("Expanded").Type != ScalarTypes.String.Type() || result.GetColumn("Idx").Type != ScalarTypes.Long.Type() {
		t.Fatalf("mv-apply output types = Expanded:%v Idx:%v", result.GetColumn("Expanded").Type, result.GetColumn("Idx").Type)
	}
	if !result.IsOpen() || !result.IsSerialized() || !result.IsSorted() {
		t.Fatalf("mv-apply table state = open:%t serialized:%t sorted:%t", result.IsOpen(), result.IsSerialized(), result.IsSorted())
	}
	if result.GetColumn("Items") != items {
		t.Fatal("aliased mv-apply did not retain the original source column identity")
	}
	indexDeclaration := apply.Parameters().At(0).Expression().(*NameDeclaration)
	if indexDeclaration.ReferencedSymbol() != result.GetColumn("Idx") {
		t.Fatal("with_itemindex declaration does not reference the projected index column")
	}
}

func TestNativeMvApplySubqueryUsesIntermediateExpandedScope(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("Items", ScalarTypes.Dynamic.Type()),
		NewColumnSymbol("Other", ScalarTypes.Long.Type()))
	query := ParseQuery(`T | mv-apply Value=Items to typeof(long) on (where Value > 0 | summarize Total=sum(Value))`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	apply := firstBoundNode[*MvApplyOperator](t, query)
	assertDiagnosticCodes(t, apply)
	result := semanticTable(t, apply)
	assertOrderedColumns(t, result, "Items", "Other", "Total")
	if result.GetColumn("Total").Type != ScalarTypes.Long.Type() {
		t.Fatalf("subquery aggregate type = %v, want long", result.GetColumn("Total").Type)
	}
}

func TestNativeMvApplyDiagnosticsAndStateRestoration(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("Text", ScalarTypes.String.Type()),
		NewColumnSymbol("Limit", ScalarTypes.Real.Type()))
	query := ParseQuery(`T | mv-apply Text limit Limit on (take 1)`)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	outer := NewTableSymbol("Outer", NewColumnSymbol("Marker", ScalarTypes.Bool.Type()))
	binder.rowScope = outer
	binder.scopeKind = ScopeKindAggregate
	tree := newTreeBinder(binder)
	tree.visitElement(query)

	apply := firstBoundNode[*MvApplyOperator](t, query)
	info, _ := apply.GetSemanticInfo()
	if len(info.Diagnostics) != 2 {
		t.Fatalf("mv-apply diagnostics = %#v, want dynamic and integer diagnostics", info.Diagnostics)
	}
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate {
		t.Fatal("mv-apply leaked binder state")
	}
	before, _ := apply.GetSemanticInfo()
	tree.visitElement(query)
	after, _ := apply.GetSemanticInfo()
	if before.ReferencedSymbol() != after.ReferencedSymbol() || len(before.Diagnostics) != len(after.Diagnostics) {
		t.Fatal("mv-apply rebinding changed semantic information")
	}
}
