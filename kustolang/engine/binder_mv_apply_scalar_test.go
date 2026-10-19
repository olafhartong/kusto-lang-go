package kustolang

import "testing"

func TestNativeMvApplyStagesBagKeysAndBindsDownstream(t *testing.T) {
	bag := NewColumnSymbol("Bag", ScalarTypes.DynamicBag.Type())
	keep := NewColumnSymbol("Keep", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", bag, keep)
	query := ParseQuery(`T | mv-apply bag_keys(Bag) on (project Key=Column1) | project Key, Keep`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	apply := firstBoundNode[*MvApplyOperator](t, query)
	assertDiagnosticCodes(t, apply)
	call := firstBoundNode[*FunctionCallExpression](t, apply)
	if !call.IsBound() || call.ReferencedSymbol() != Functions.BagKeys {
		t.Fatalf("bag_keys binding = bound:%t symbol:%#v", call.IsBound(), call.ReferencedSymbol())
	}
	result := semanticTable(t, apply)
	assertOrderedColumns(t, result, "Bag", "Keep", "Key")
	if result.GetColumn("Key").Type != ScalarTypes.DynamicString.Type() {
		t.Fatalf("bag_keys expanded type = %v, want dynamic(string)", result.GetColumn("Key").Type)
	}
	downstream := boundNodes[*ProjectOperator](query)
	if len(downstream) != 2 || !downstream[1].IsBound() {
		t.Fatalf("downstream project binding = %#v", downstream)
	}
	assertOrderedColumns(t, semanticTable(t, downstream[1]), "Key", "Keep")
}

func TestNativeMvApplyNonDynamicDiagnosticRetainsExpressionType(t *testing.T) {
	text := NewColumnSymbol("Text", ScalarTypes.String.Type())
	table := NewTableSymbol("T", text)
	query := ParseQuery(`T | mv-apply Text on (project Text) | where isnotempty(Text)`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	apply := firstBoundNode[*MvApplyOperator](t, query)
	assertDiagnosticCodes(t, apply, "KS141")
	result := semanticTable(t, apply)
	if result.GetColumn("Text") == nil || result.GetColumn("Text").Type != ScalarTypes.String.Type() {
		t.Fatalf("non-dynamic expansion type = %#v, want string", result.GetColumn("Text"))
	}
	if !firstBoundNode[*FilterOperator](t, query).IsBound() {
		t.Fatal("downstream where did not bind against retained scalar type")
	}
}

func TestNativeMvApplySuppressesDynamicDiagnosticForErrorExpression(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Keep", ScalarTypes.Long.Type()))
	query := ParseQuery(`T | mv-apply Missing on (take 1) | count`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	apply := firstBoundNode[*MvApplyOperator](t, query)
	if !apply.IsBound() {
		t.Fatal("mv-apply with an unresolved expression did not bind")
	}
	if hasSemanticDiagnostic(apply, "KS141") {
		t.Fatalf("mv-apply diagnostics = %#v, unexpected dynamic-type diagnostic", apply.SemanticDiagnostics())
	}
	missing := firstBoundNode[*NameReference](t, apply)
	if ExpressionResultType(missing) != &Error.TypeSymbol {
		t.Fatalf("unresolved expression type = %v, want error", ExpressionResultType(missing))
	}
	if !firstBoundNode[*CountOperator](t, query).IsBound() {
		t.Fatal("downstream count did not bind after the error expression")
	}
}

func TestNativeMvApplyMalformedScalarRollsBack(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Bag", ScalarTypes.DynamicBag.Type())).WithIsOpen(true)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	recovered := ParseQuery(`T | mv-apply bag_keys(Bag on (take 1)`)
	NewBinder(globals).TryBind(recovered)
	if apply := firstBoundNode[*MvApplyOperator](t, recovered); !apply.IsBound() ||
		!diagnosticsContainMessage(apply.ContainedDiagnostics(), "KS005", "Expected: )") {
		t.Fatalf("recovered mv-apply call = bound:%t diagnostics:%#v", apply.IsBound(), apply.ContainedDiagnostics())
	}

	query := ParseQuery(`T | mv-apply bag_keys(Bag) on (take 1)`)
	apply := firstBoundNode[*MvApplyOperator](t, query)
	firstBoundNode[*FunctionCallExpression](t, apply).ArgumentList().OpenParen().text = ""
	binder := NewBinder(globals)
	outer := NewTableSymbol("Outer", NewColumnSymbol("Marker", ScalarTypes.Bool.Type()))
	binder.rowScope = outer
	binder.scopeKind = ScopeKindAggregate
	local := binder.LocalScope
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.nodeBinder.defaultColumnNameSuffix = 23
	tree.bind(query)

	if apply.IsBound() {
		t.Fatal("mv-apply with an injected invalid function call was bound")
	}
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
		!tree.inPipe || !tree.inWildcard || tree.nodeBinder.defaultColumnNameSuffix != 23 {
		t.Fatal("malformed mv-apply scalar changed binder or tree state")
	}
	if binder.openTableInferredColumns != nil || binder.openTableInferredOrder != nil {
		t.Fatal("malformed mv-apply scalar leaked inferred columns")
	}
	if expression := apply.Expressions().At(0).Element().Expression(); expressionIsBound(expression) {
		t.Fatal("malformed mv-apply scalar retained staged semantic information")
	}
}
