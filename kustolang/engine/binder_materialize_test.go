package kustolang

import "testing"

func TestNativeBinderMaterializePreservesTableResultAndLetBinding(t *testing.T) {
	query := ParseQuery(`let M = materialize(datatable(A:long)[1] | project A); M | project A`)
	NewBinder(NewGlobalState(nil)).TryBind(query)

	materialize := firstBoundNode[*MaterializeExpression](t, query)
	table := expressionTableResult(materialize)
	if table == nil || len(table.Columns()) != 1 || table.Columns()[0].SymbolName() != "A" {
		t.Fatalf("materialize table result = %#v", table)
	}
	if diagnostics := materialize.SemanticDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid materialize diagnostics = %#v", diagnostics)
	}
	project := boundNodes[*ProjectOperator](query)
	if len(project) != 2 || expressionTableResult(project[1]) == nil || len(expressionTableResult(project[1]).Columns()) != 1 {
		t.Fatalf("downstream materialized projection = %#v", project)
	}
}

func TestNativeBinderMaterializeMalformedExpressionDoesNotCascade(t *testing.T) {
	query := ParseQuery(`let M = materialize(; M`)
	NewBinder(NewGlobalState(nil)).TryBind(query)
	materialize := firstBoundNode[*MaterializeExpression](t, query)
	if diagnosticsContain(materialize.SemanticDiagnostics(), "KS231") {
		t.Fatalf("missing materialize expression added KS231: %#v", materialize.SemanticDiagnostics())
	}
	if !diagnosticsContain(query.ContainedDiagnostics(), "KS006") || !diagnosticsContain(query.ContainedDiagnostics(), "KS005") {
		t.Fatalf("malformed materialize parser diagnostics = %#v", query.ContainedDiagnostics())
	}
}
