package kustolang

import "testing"

func TestNativeTreeBinderStandaloneNameKeepsScalarContext(t *testing.T) {
	const text = "MissingScalar"
	query := ParseQuery(text)
	NewBinder(nil).TryBind(query)

	reference, ok := queryExpressionAt(t, query, 0, true).(*NameReference)
	if !ok {
		t.Fatalf("query expression = %T, want *NameReference", queryExpressionAt(t, query, 0, true))
	}
	diagnostics := reference.SemanticDiagnostics()
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS142" {
		t.Fatalf("scalar semantic diagnostics = %#v, want only KS142", diagnostics)
	}
	if diagnostics[0].LocationText() != text {
		t.Fatalf("scalar diagnostic location text = %q, want %q", diagnostics[0].LocationText(), text)
	}
}

func TestNativeTreeBinderRecoveredNameAfterPreambleKeepsScalarContext(t *testing.T) {
	const text = "let Scoped = ZScope DATASCOPE=hotcache; Scoped | project User"
	query := ParseQuery(text)
	NewBinder(nil).TryBind(query)

	var recovered *NameReference
	WalkSyntax(query, func(element SyntaxElement) {
		reference, ok := WrapGeneratedSyntaxElement(element).(*NameReference)
		if ok && NameReferenceSimpleName(reference) == "DATASCOPE" {
			recovered = reference
		}
	})
	if recovered == nil {
		t.Fatal("recovered DATASCOPE name not found")
	}
	if _, ok := recovered.SyntaxNode.Payload(recoveredTabularStatementSourcePayload); ok {
		t.Fatal("recovered name after preamble was marked as a tabular source")
	}
	diagnostics := recovered.SemanticDiagnostics()
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS142" {
		t.Fatalf("recovered name diagnostics = %#v, want only KS142", diagnostics)
	}
}
