package kustolang

import "testing"

func TestNativeForkBranchesUseIndependentIncomingRowScope(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.Long.Type())
	b := NewColumnSymbol("B", ScalarTypes.String.Type())
	table := NewTableSymbol("T", a, b).WithIsOpen(true).WithIsSerialized(true).WithIsSorted(true)
	query := ParseQuery(`T | fork (project A) (project B)`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	fork := firstBoundNode[*ForkOperator](t, query)
	group := forkResultGroup(t, fork)
	members := group.Members()
	if len(members) != 2 {
		t.Fatalf("fork members = %d, want 2", len(members))
	}
	first := members[0].(*TableSymbol)
	second := members[1].(*TableSymbol)
	if first.SymbolName() != "Results" || second.SymbolName() != "Results_2" {
		t.Fatalf("fork member names = %q, %q", first.SymbolName(), second.SymbolName())
	}
	assertOrderedColumns(t, first, "A")
	assertOrderedColumns(t, second, "B")
	if first.GetColumn("A") != a || second.GetColumn("B") != b {
		t.Fatal("fork result did not preserve branch column identity")
	}
	for i, result := range []*TableSymbol{first, second} {
		if !result.IsOpen() || !result.IsSerialized() || !result.IsSorted() {
			t.Fatalf("branch %d state = open:%t serialized:%t sorted:%t", i, result.IsOpen(), result.IsSerialized(), result.IsSorted())
		}
	}
	assertDiagnosticCodes(t, fork)
}

func TestNativeForkIncludesDeclaredAndInferredColumns(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type())).WithIsOpen(true)
	query := ParseQuery(`T | fork (where Future > 0) Named=(where OtherFuture == "x")`)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.TryBind(query)

	members := forkResultGroup(t, firstBoundNode[*ForkOperator](t, query)).Members()
	if len(members) != 2 {
		t.Fatalf("fork members = %d, want 2", len(members))
	}
	first := members[0].(*TableSymbol)
	second := members[1].(*TableSymbol)
	assertOrderedColumns(t, first, "A", "Future")
	assertOrderedColumns(t, second, "A", "Future", "OtherFuture")
	if first.GetColumn("Future") != second.GetColumn("Future") {
		t.Fatal("fork did not preserve stable inferred-column identity")
	}
	branch := firstBoundNode[*ForkExpression](t, query)
	if branch.NameEquals() != nil {
		t.Fatal("first branch unexpectedly named")
	}
	if members[1].SymbolName() != "Results_2" {
		t.Fatalf("named branch leaked into group result as %q", members[1].SymbolName())
	}
}

func TestNativeForkDiagnosticsRestorationAndIdempotence(t *testing.T) {
	left := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	right := NewTableSymbol("Other", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	query := ParseQuery(`T | fork (join Other on A) (project A)`)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right)))
	outer := NewTableSymbol("Outer", NewColumnSymbol("Marker", ScalarTypes.String.Type()))
	local := NewLocalScope(nil)
	binder.rowScope, binder.scopeKind, binder.LocalScope = outer, ScopeKindAggregate, local
	tree := newTreeBinder(binder)
	tree.visitElement(query)

	fork := firstBoundNode[*ForkOperator](t, query)
	assertDiagnosticCodes(t, fork, "KS177")
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local {
		t.Fatal("fork leaked binder state")
	}
	before, _ := fork.GetSemanticInfo()
	beforeGroup := before.ReferencedSymbol()
	tree.visitElement(query)
	after, _ := fork.GetSemanticInfo()
	if after.ReferencedSymbol() != beforeGroup || len(after.Diagnostics) != len(before.Diagnostics) {
		t.Fatalf("fork rebinding changed semantic info: before=%#v after=%#v", before, after)
	}

	standalone := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, `fork (count)`).(*ForkOperator)
	NewBinder(nil).TryBind(standalone)
	assertDiagnosticCodes(t, standalone, "KS174")
}

func TestNativeForkMalformedBranchRetainsRecoveryAndSemanticResult(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.Long.Type()),
		NewColumnSymbol("B", ScalarTypes.String.Type()))
	query := ParseQuery(`T | fork (where A > 0) ByB=(project B`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	fork := firstBoundNode[*ForkOperator](t, query)
	if !diagnosticsContainMessage(fork.ContainedDiagnostics(), "KS005", "Expected: )") {
		t.Fatalf("missing close-parenthesis recovery = %#v", fork.ContainedDiagnostics())
	}
	members := forkResultGroup(t, fork).Members()
	if len(members) != 2 {
		t.Fatalf("malformed fork members = %d, want 2", len(members))
	}
	assertOrderedColumns(t, members[0].(*TableSymbol), "A", "B")
	assertOrderedColumns(t, members[1].(*TableSymbol), "B")

	empty := ParseQuery(`T | fork ()`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(empty)
	emptyFork := firstBoundNode[*ForkOperator](t, empty)
	if !diagnosticsContainMessage(emptyFork.ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing expression recovery = %#v", emptyFork.ContainedDiagnostics())
	}
	if got := len(forkResultGroup(t, emptyFork).Members()); got != 0 {
		t.Fatalf("missing-expression fork members = %d, want 0", got)
	}
}

func forkResultGroup(t *testing.T, fork *ForkOperator) *GroupSymbol {
	t.Helper()
	info, ok := fork.GetSemanticInfo()
	if !ok {
		t.Fatal("fork is unbound")
	}
	group, ok := info.ReferencedSymbol().(*GroupSymbol)
	if !ok {
		t.Fatalf("fork result = %T, want *GroupSymbol", info.ReferencedSymbol())
	}
	return group
}
