package kustolang

import "testing"

func TestNativeBinderDataScopePreservesTableAndDownstreamShape(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	query := ParseQuery(`T datascope=all | project A`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	scoped := firstBoundNode[*DataScopeExpression](t, query)
	result := expressionTableResult(scoped)
	if !scoped.IsBound() || result != table || len(scoped.SemanticDiagnostics()) != 0 {
		t.Fatalf("data-scope table result = bound:%t result:%#v diagnostics:%#v", scoped.IsBound(), result, scoped.SemanticDiagnostics())
	}
	projects := boundNodes[*ProjectOperator](query)
	if len(projects) != 1 || expressionTableResult(projects[0]) == nil || len(expressionTableResult(projects[0]).Columns()) != 1 {
		t.Fatalf("data-scope downstream project = %#v", projects)
	}
}

func TestNativeBinderDataScopeRejectsGraphButPreservesGraphIdentity(t *testing.T) {
	query := ParseQuery(`G datascope=all`)
	binder := NewBinder(NewGlobalState(nil))
	binder.LocalScope.AddSymbol(NewVariableSymbol("G", testGraphSymbol(), false, nil, nil))
	binder.TryBind(query)
	scoped := firstBoundNode[*DataScopeExpression](t, query)
	diagnostics := scoped.SemanticDiagnostics()
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS111" || diagnostics[0].Message != "Tabular value expected." {
		t.Fatalf("data-scope graph diagnostics = %#v", diagnostics)
	}
	if graph, ok := scoped.ReferencedSymbol().(*GraphSymbol); !ok || graph == nil {
		t.Fatalf("data-scope graph result = %#v", scoped.ReferencedSymbol())
	}
}

func TestNativeBinderDataScopeUnresolvedOperandDoesNotCascade(t *testing.T) {
	query := ParseQuery(`Missing datascope=all | count`)
	NewBinder(NewGlobalState(nil)).TryBind(query)
	scoped := firstBoundNode[*DataScopeExpression](t, query)
	if diagnostics := scoped.SemanticDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unresolved data-scope wrapper diagnostics = %#v", diagnostics)
	}
	reference := firstBoundNode[*NameReference](t, query)
	assertSemanticDiagnostic(t, reference, "KS142")
}
