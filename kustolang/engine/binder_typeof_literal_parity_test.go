package kustolang

import (
	"strings"
	"testing"
)

func TestNativeBinderTypeOfLiteralAcceptedByLiteralFunctionParameter(t *testing.T) {
	query := `datatable(Raw:string)["age=42"] | project V=extract(@"age=(\d+)", 1, Raw, typeof(long))`
	code := ParseAndAnalyze(query, nil)
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid extract typeof diagnostics = %#v", diagnostics)
	}

	call := firstBoundNode[*FunctionCallExpression](t, code.Syntax())
	if call.ReferencedSignature() == nil || !sameType(call.RawResultType(), ScalarTypes.Long.Type()) {
		t.Fatalf("extract semantics = signature:%#v type:%s", call.ReferencedSignature(), typeName(call.RawResultType()))
	}
	typeOf := firstBoundNode[*TypeOfLiteralExpression](t, code.Syntax())
	if !ExpressionIsLiteral(typeOf) || !sameType(typeOf.RawResultType(), ScalarTypes.TypeLit.Type()) {
		t.Fatalf("typeof semantics = literal:%t type:%s", ExpressionIsLiteral(typeOf), typeName(typeOf.RawResultType()))
	}
}

func TestNativeTreeBinderStagesTypeOfLiteralInNestedSummarizeScalar(t *testing.T) {
	query := `datatable(Raw:string)["age=42"] | summarize V=take_any(extract(@"age=(\d+)", 1, Raw, typeof(long)))`
	code := ParseAndAnalyze(query, nil)
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("summarize extract typeof diagnostics = %#v", diagnostics)
	}

	for _, call := range boundNodes[*FunctionCallExpression](code.Syntax()) {
		if !call.IsBound() || call.ReferencedSignature() == nil {
			t.Fatalf("nested summarize call was not bound: %#v", call)
		}
	}
	typeOf := firstBoundNode[*TypeOfLiteralExpression](t, code.Syntax())
	if !typeOf.IsBound() || !sameType(typeOf.RawResultType(), ScalarTypes.TypeLit.Type()) {
		t.Fatalf("staged typeof semantics = bound:%t type:%s", typeOf.IsBound(), typeName(typeOf.RawResultType()))
	}
	summarize := firstBoundNode[*SummarizeOperator](t, code.Syntax())
	result := expressionTableResult(summarize)
	if column := result.GetColumn("V"); column == nil || !sameType(column.Type, ScalarTypes.Long.Type()) {
		t.Fatalf("summarize result = %#v, want V:long", result)
	}
}

func TestNativeBinderTypeOfLiteralInvalidTypeDiagnosticLocation(t *testing.T) {
	query := `datatable(Raw:string)["age=42"] | project V=extract(@"age=(\d+)", 1, Raw, typeof(NotAType))`
	code := ParseAndAnalyze(query, nil)
	start := strings.Index(query, "NotAType")
	assertDiagnosticLocations(t, code.GetDiagnostics(), diagnosticLocation{"KS160", start, len("NotAType")})

	typeOf := firstBoundNode[*TypeOfLiteralExpression](t, code.Syntax())
	if !sameType(typeOf.RawResultType(), ScalarTypes.TypeLit.Type()) {
		t.Fatalf("invalid typeof result type = %s, want type", typeName(typeOf.RawResultType()))
	}
}

func TestNativeBinderTypeOfLiteralMalformedCloseIsDiagnosed(t *testing.T) {
	query := `print V=extract(@"age=(\d+)", 1, "age=42", typeof(long`
	code := ParseAndAnalyze(query, nil)
	diagnostics := code.GetDiagnostics()
	if len(diagnostics) == 0 || !diagnosticsContain(diagnostics, "KS005") {
		t.Fatalf("malformed typeof diagnostics = %#v, want KS005", diagnostics)
	}
}
