package kustolang

import "testing"

func TestNativeFindSmartProjectionFilteringAndLineage(t *testing.T) {
	leftUser := NewColumnSymbol("User", ScalarTypes.String.Type())
	left := NewTableSymbol("Left", leftUser, NewColumnSymbol("LeftOnly", ScalarTypes.Long.Type()))
	rightUser := NewColumnSymbol("User", ScalarTypes.String.Type())
	right := NewTableSymbol("Right", rightUser, NewColumnSymbol("RightOnly", ScalarTypes.Bool.Type()))
	unrelated := NewTableSymbol("Other", NewColumnSymbol("OtherOnly", ScalarTypes.String.Type()))
	query := ParseQuery(`find in (Left, Right, Other) where User has "a"`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right, unrelated))).TryBind(query)

	find := firstBoundNode[*FindOperator](t, query)
	assertDiagnosticCodes(t, find)
	result := expressionTableResult(find)
	assertFindColumnNames(t, result, "source_", "User", "pack_")
	user := result.GetColumn("User")
	if len(user.OriginalColumns) != 2 || user.OriginalColumns[0] != leftUser || user.OriginalColumns[1] != rightUser {
		t.Fatalf("User lineage = %#v", user.OriginalColumns)
	}
}

func TestNativeFindExplicitProjectionPackAwaySourceAndDownstream(t *testing.T) {
	user := NewColumnSymbol("User", ScalarTypes.String.Type())
	value := NewColumnSymbol("Value", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", user, value).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	query := ParseQuery(`find withsource=Src in (T) where User has "a" project User, Value:string, pack(*) __projectAway * | project pack_`)
	NewBinder(globals).TryBind(query)
	find := firstBoundNode[*FindOperator](t, query)
	assertDiagnosticCodes(t, find)
	result := expressionTableResult(find)
	assertFindColumnNames(t, result, "Src", "pack_")
	if result.IsOpen() || result.IsSorted() || result.IsSerialized() {
		t.Fatalf("explicit find properties = open:%t sorted:%t serialized:%t", result.IsOpen(), result.IsSorted(), result.IsSerialized())
	}
	if pack := result.GetColumn("pack_"); !sameType(pack.Type, ScalarTypes.Dynamic.Type()) || len(pack.OriginalColumns) != 0 {
		t.Fatalf("explicit pack = %#v", pack)
	}
	project := firstBoundNode[*ProjectOperator](t, query)
	if expressionTableResult(project).GetColumn("pack_") != result.GetColumn("pack_") {
		t.Fatal("downstream project did not bind against find result")
	}

	typed := ParseQuery(`find in (T) where User has "a" project Value:string`)
	NewBinder(globals).TryBind(typed)
	typedFind := firstBoundNode[*FindOperator](t, typed)
	typedColumn := expressionTableResult(typedFind).GetColumn("Value")
	if typedColumn.Type != ScalarTypes.String.Type() || len(typedColumn.OriginalColumns) != 1 || typedColumn.OriginalColumns[0] != value {
		t.Fatalf("typed projection = %#v", typedColumn)
	}
}

func TestNativeFindIncludesZeroArgumentLocalViewsInImpliedTables(t *testing.T) {
	code := ParseAndAnalyze(`let V = view () { datatable(ViewOnly:string)["x"] };
find where ViewOnly == "x" project ViewOnly`, NewGlobalState(NewDatabaseSymbol("db")))
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("implied-view find diagnostics = %#v", diagnostics)
	}
	find := firstBoundNode[*FindOperator](t, code.Tree.Root)
	assertFindColumnNames(t, expressionTableResult(find), "source_", "ViewOnly")
}

func TestNativeFindDiagnosticsRestorationRollbackAndIdempotence(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	query := ParseQuery(`find in (T) where 1 project pack(*), A`)
	binder := NewBinder(globals)
	outer := NewTableSymbol("Outer", NewColumnSymbol("Marker", ScalarTypes.String.Type()))
	local := NewLocalScope(nil)
	binder.rowScope, binder.scopeKind, binder.LocalScope = outer, ScopeKindAggregate, local
	tree := newTreeBinder(binder)
	tree.visitElement(query)
	find := firstBoundNode[*FindOperator](t, query)
	assertDiagnosticCodes(t, find, "KS141", "KS185")
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local {
		t.Fatalf("find leaked binder state: row=%#v kind=%v local=%p", binder.rowScope, binder.scopeKind, binder.LocalScope)
	}
	before, _ := find.GetSemanticInfo()
	tree.visitElement(query)
	after, _ := find.GetSemanticInfo()
	if len(after.Diagnostics) != len(before.Diagnostics) {
		t.Fatalf("find rebinding changed diagnostics: before=%#v after=%#v", before.Diagnostics, after.Diagnostics)
	}

}

func assertFindColumnNames(t *testing.T, table *TableSymbol, want ...string) {
	t.Helper()
	if table == nil || len(table.Columns()) != len(want) {
		t.Fatalf("columns = %#v, want %v", table, want)
	}
	for i, name := range want {
		if table.Columns()[i].SymbolName() != name {
			t.Fatalf("column %d = %q, want %q", i, table.Columns()[i].SymbolName(), name)
		}
	}
}
