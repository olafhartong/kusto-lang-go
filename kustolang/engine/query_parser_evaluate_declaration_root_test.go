package kustolang

import "testing"

func TestQueryParserEvaluateIsNameAtDeclarationRoot(t *testing.T) {
	query := ParseQuery("let X = evaluate;")
	statements := query.GetChild(1).(*SyntaxList)
	let := statements.At(0).(*SeparatedElement).Element().(*LetStatement)
	reference, ok := let.Expression().(*NameReference)
	if !ok || NameReferenceSimpleName(reference) != "evaluate" {
		t.Fatalf("let expression = %#v, want evaluate name reference", let.Expression())
	}
}
