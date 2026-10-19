package kustolang

import "testing"

func TestNativeBinderTypedColumnReferenceMissingColumnKeepsErrorType(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("User", ScalarTypes.String.Type()))
	query := ParseQuery(`find in (T) where User has "a" project Missing:string`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	typed := firstBoundNode[*TypedColumnReference](t, query)
	if ExpressionResultType(typed) != &Error.TypeSymbol || len(typed.SemanticDiagnostics()) != 0 {
		t.Fatalf("typed missing reference = type:%#v diagnostics:%#v", ExpressionResultType(typed), typed.SemanticDiagnostics())
	}
	if column := typed.Column(); !diagnosticsContain(column.SemanticDiagnostics(), "KS142") {
		t.Fatalf("missing column diagnostics = %#v", column.SemanticDiagnostics())
	}
}

func TestNativeBinderTypedColumnReferenceInvalidAnnotationDoesNotChangeSourceType(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("User", ScalarTypes.String.Type()),
		NewColumnSymbol("Value", ScalarTypes.Long.Type()),
	)
	query := ParseQuery(`find in (T) where User has "a" project Value:notatype`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	typed := firstBoundNode[*TypedColumnReference](t, query)
	if ExpressionResultType(typed) != ScalarTypes.Long.Type() || len(typed.SemanticDiagnostics()) != 0 {
		t.Fatalf("typed invalid annotation = type:%#v diagnostics:%#v", ExpressionResultType(typed), typed.SemanticDiagnostics())
	}
	typeNode := simpleBinderSyntaxNode(typed.Type())
	if typeNode == nil || !diagnosticsContain(typeNode.ContainedDiagnostics(), "KS160") {
		t.Fatalf("invalid annotation diagnostics = %#v", typeNode)
	}
	if annotationType := typeFromBoundExpression(typed.Type()); annotationType != &Error.TypeSymbol {
		t.Fatalf("bound annotation type = %#v, want error", annotationType)
	}
	find := firstBoundNode[*FindOperator](t, query)
	result := expressionTableResult(find)
	if result == nil || result.GetColumn("Value") == nil || result.GetColumn("Value").Type != ScalarTypes.Unknown.Type() {
		t.Fatalf("find invalid annotation output = %#v", result)
	}
}
