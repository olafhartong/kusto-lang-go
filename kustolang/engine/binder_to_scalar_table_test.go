package kustolang

import "testing"

func TestNativeToScalarTableResultsConstantsAndIdentity(t *testing.T) {
	column := NewColumnSymbol("A", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", column).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	empty := NewTableSymbol("Empty")
	globals := NewGlobalState(NewDatabaseSymbol("db", table, empty))

	scalar := parsedQueryExpression(t, `toscalar(T)`).(*ToScalarExpression)
	NewBinder(globals).TryBind(scalar)
	if ExpressionRawResultType(scalar) != column.Type || !ExpressionIsConstant(scalar) {
		t.Fatalf("toscalar table result = %v, constant=%t", ExpressionRawResultType(scalar), ExpressionIsConstant(scalar))
	}

	scalarValue := parsedQueryExpression(t, `toscalar(1)`).(*ToScalarExpression)
	NewBinder(globals).TryBind(scalarValue)
	if ExpressionRawResultType(scalarValue) != ScalarTypes.Long.Type() || !ExpressionIsConstant(scalarValue) {
		t.Fatalf("toscalar scalar result = %v, constant=%t", ExpressionRawResultType(scalarValue), ExpressionIsConstant(scalarValue))
	}

	tabular := parsedQueryExpression(t, `totable(T)`).(*ToTableExpression)
	NewBinder(globals).TryBind(tabular)
	if got := semanticTable(t, tabular); got != table || !got.IsOpen() || !got.IsSorted() || !got.IsSerialized() {
		t.Fatalf("totable did not preserve exact table identity/state: %#v", got)
	}
	if ExpressionIsConstant(tabular) {
		t.Fatal("totable result is constant")
	}

	emptyScalar := parsedQueryExpression(t, `toscalar(Empty)`).(*ToScalarExpression)
	NewBinder(globals).TryBind(emptyScalar)
	assertDiagnosticCodes(t, emptyScalar, "KS190")
	if ExpressionRawResultType(emptyScalar) != &Error.TypeSymbol || !ExpressionIsConstant(emptyScalar) {
		t.Fatal("empty toscalar did not produce a constant error")
	}
}

func TestNativeToScalarTableErrorsKindsAndDownstreamPipes(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	badScalar := parsedQueryExpression(t, `toscalar(1)`).(*ToScalarExpression)
	database := NewDatabaseSymbol("other")
	badScalar.Expression().(*LiteralExpression).SetSemanticInfo(NewSemanticInfoForReferencedSymbol(database))
	badScalar.SetSemanticInfo(*newSimpleNodeBinder(NewBinder(globals)).bindToScalarExpression(badScalar))
	assertDiagnosticCodes(t, badScalar, "KS112")
	if !ExpressionIsConstant(badScalar) {
		t.Fatal("invalid toscalar is not constant")
	}

	badTable := parsedQueryExpression(t, `totable(1)`).(*ToTableExpression)
	NewBinder(globals).TryBind(badTable)
	assertDiagnosticCodes(t, badTable, "KS111")
	if ExpressionIsConstant(badTable) {
		t.Fatal("invalid totable is constant")
	}

	validKind := parsedQueryExpression(t, `totable kind=nooptimization (T)`).(*ToTableExpression)
	NewBinder(globals).TryBind(validKind)
	assertDiagnosticCodes(t, validKind)
	if validKind.KindParameter().Name().ReferencedSymbol() != QueryOperatorParameters.ToTableKindParameter {
		t.Fatal("kind parameter did not bind to its declaration")
	}

	invalidKind := parsedQueryExpression(t, `toscalar kind=NOOPTIMIZATION (1)`).(*ToScalarExpression)
	NewBinder(globals).TryBind(invalidKind)
	assertDiagnosticCodes(t, invalidKind, "KS140")

	query := ParseQuery(`totable(T) | project A | count`)
	NewBinder(globals).TryBind(query)
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*CountOperator](t, query)), "Count")
}

func TestNativeToScalarTablePreflightRollbackRetryAndIdempotence(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.rowScope = NewTableSymbol("Outer", NewColumnSymbol("X", ScalarTypes.String.Type()))
	binder.scopeKind = ScopeKindAggregate
	binder.LocalScope = NewLocalScope(nil)

	node := parsedQueryExpression(t, `totable(T)`).(*ToTableExpression)
	original := node.SyntaxNode.children[4]
	node.SyntaxNode.children[4] = nil
	binder.TryBind(node)
	if node.IsBound() || simpleBinderSyntaxNode(node.Expression()).IsBound() {
		t.Fatal("malformed preflight committed semantics")
	}
	if binder.rowScope.SymbolName() != "Outer" || binder.scopeKind != ScopeKindAggregate {
		t.Fatal("malformed preflight changed binder state")
	}

	node.SyntaxNode.children[4] = original
	binder.TryBind(node)
	if semanticTable(t, node) != table {
		t.Fatal("valid retry did not preserve the table")
	}
	first, _ := node.GetSemanticInfo()
	binder.TryBind(node)
	second, _ := node.GetSemanticInfo()
	if first.ReferencedSymbol() != second.ReferencedSymbol() || first.ResultType != second.ResultType {
		t.Fatal("idempotent bind replaced committed semantics")
	}
	if binder.rowScope.SymbolName() != "Outer" || binder.scopeKind != ScopeKindAggregate {
		t.Fatal("successful bind did not restore binder state")
	}
}
