package kustolang

import "testing"

func TestNativeRangeBindsLocalAndUnresolvedFunctionCalls(t *testing.T) {
	local := ParseQuery(`let Next=(x:long) { x + 1 }; range X from Next(0) to Next(2) step Next(0)`)
	NewBinder(NewGlobalState(nil)).TryBind(local)
	operator := firstBoundNode[*RangeOperator](t, local)
	if column := semanticTable(t, operator).GetColumn("X"); column == nil || !sameType(column.Type, ScalarTypes.Long.Type()) {
		t.Fatalf("local-call range column = %#v, want long", column)
	}
	for _, call := range boundNodes[*FunctionCallExpression](operator) {
		if !call.IsBound() || call.ReferencedSignature() == nil || call.GetCalledFunctionBody() == nil {
			t.Fatalf("local range call = bound:%t signature:%#v body:%#v", call.IsBound(), call.ReferencedSignature(), call.GetCalledFunctionBody())
		}
	}

	unresolved := ParseQuery("range X from Missing(1) to 2 step 1")
	NewBinder(NewGlobalState(nil)).TryBind(unresolved)
	unresolvedOperator := firstBoundNode[*RangeOperator](t, unresolved)
	call := firstBoundNode[*FunctionCallExpression](t, unresolvedOperator)
	if !call.IsBound() || !diagnosticsContain(call.Name().SemanticDiagnostics(), "KS211") {
		t.Fatalf("unresolved call = bound:%t diagnostics:%#v", call.IsBound(), call.Name().SemanticDiagnostics())
	}
	if !unresolvedOperator.IsBound() || !sameType(semanticTable(t, unresolvedOperator).GetColumn("X").Type, ScalarTypes.Unknown.Type()) {
		t.Fatal("range with unresolved call did not bind to an unknown column")
	}
}

func TestNativeRangeBindsPathExpressionsWithoutRowScope(t *testing.T) {
	query := ParseQuery(`let Bag=dynamic({"A": 1}); T | range X from Bag.A to A step 1`)
	tableColumn := NewColumnSymbol("A", ScalarTypes.Long.Type())
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", tableColumn))))
	binder.TryBind(query)

	operator := firstBoundNode[*RangeOperator](t, query)
	path := firstBoundNode[*PathExpression](t, operator)
	if !path.IsBound() || !isDynamicType(ExpressionResultType(path)) {
		t.Fatalf("path = bound:%t type:%#v", path.IsBound(), ExpressionResultType(path))
	}
	rowName, ok := operator.To().(*NameReference)
	if !ok {
		t.Fatalf("range to expression = %T, want name reference", operator.To())
	}
	if rowName.ReferencedSymbol() == tableColumn || !diagnosticsContain(rowName.SemanticDiagnostics(), "KS142") {
		t.Fatalf("range row name resolved in cleared scope: symbol:%#v diagnostics:%#v", rowName.ReferencedSymbol(), rowName.SemanticDiagnostics())
	}
	assertDiagnosticCodes(t, operator, "KS175", "KS138")
}

func TestNativeRangeRestoresStateAndIsIdempotent(t *testing.T) {
	query := ParseQuery("range X from 1 + 0 to strlen('abc') step +1")
	binder := NewBinder(NewGlobalState(nil))
	outer := NewTableSymbol("Outer", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	path := NewDatabaseSymbol("Path")
	local := binder.LocalScope
	binder.rowScope = outer
	binder.pathScope = path
	binder.scopeKind = ScopeKindAggregate
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.nodeBinder.inPipe = true
	tree.bind(query)

	operator := firstBoundNode[*RangeOperator](t, query)
	firstTable := semanticTable(t, operator)
	firstColumn := firstTable.GetColumn("X")
	firstInfo, _ := operator.GetSemanticInfo()
	if binder.rowScope != outer || binder.pathScope != path || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
		!tree.inPipe || !tree.inWildcard || !tree.nodeBinder.inPipe {
		t.Fatal("range binding changed caller binder or tree state")
	}

	tree.bind(query)
	secondInfo, _ := operator.GetSemanticInfo()
	if semanticTable(t, operator) != firstTable || semanticTable(t, operator).GetColumn("X") != firstColumn || secondInfo.ResultType != firstInfo.ResultType {
		t.Fatal("repeat binding replaced range semantic identity")
	}
}

func TestNativeRangeMalformedChildrenDoNotCommitPartialBindings(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "range X from 1 to 2 + step 1").(*RangeOperator)
	binder := NewBinder(NewGlobalState(nil))
	outer := NewTableSymbol("Outer", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	binder.rowScope = outer
	binder.scopeKind = ScopeKindAggregate
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.nodeBinder.inPipe = true
	tree.bind(operator)

	if operator.IsBound() || simpleBinderSyntaxNode(operator.From()).IsBound() {
		t.Fatal("malformed range committed partial child or operator bindings")
	}
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || !tree.inPipe || !tree.nodeBinder.inPipe {
		t.Fatal("malformed range did not restore caller state")
	}

	recovered := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "range n from to 5 step").(*RangeOperator)
	NewBinder(NewGlobalState(nil)).TryBind(recovered)
	if !recovered.IsBound() {
		t.Fatal("range parser recovery with missing name expressions was not bound")
	}
}
