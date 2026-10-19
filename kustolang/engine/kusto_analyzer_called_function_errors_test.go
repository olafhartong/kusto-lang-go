package kustolang

import (
	"strings"
	"testing"
)

func TestCalledFunctionHasErrorsAnalyzerReportsFunctionCallAndSameQueryErrors(t *testing.T) {
	text := "let f = () { bad };\nT | extend Value=f()"
	code := Parse(text)
	call := firstParsedFunctionCall(t, code)
	body := firstParsedFunctionBody(t, code)
	signature := NewDeclarationSignature(body)
	function := NewFunctionSymbol("f", signature)
	callInfo := analyzerCalledFunctionInfoWithError("KS999", "called body is bad")
	call.SetSemanticInfo(NewSemanticInfoForType(ScalarTypes.Long.Type()).WithCalledFunctionInfo(callInfo))
	call.Name().SetSemanticInfo(NewSemanticInfoForSignature(signature, &function.TypeSymbol))

	diagnostics := NewCalledFunctionHasErrorsAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %#v, want propagated error and warning", diagnostics)
	}
	if diagnostics[0].Code != "KS999" || diagnostics[0].Severity != DiagnosticSeverityError {
		t.Fatalf("propagated diagnostic = %#v", diagnostics[0])
	}
	warning := diagnostics[1]
	if warning.Code != KustoAnalyzerCodeCalledFunctionHasErrors ||
		warning.Category != DiagnosticCategoryGeneral ||
		warning.Severity != DiagnosticSeverityWarning ||
		warning.Start != strings.LastIndex(text, "f()") ||
		warning.Length != len("f") ||
		warning.Message != "The function 'f' has errors in its definition" {
		t.Fatalf("warning = %#v", warning)
	}
}

func TestCalledFunctionHasErrorsAnalyzerReportsExternalFunctionCallWithoutPropagatingErrors(t *testing.T) {
	text := "T | extend Value=f()"
	code := Parse(text)
	call := firstParsedFunctionCall(t, code)
	signature := NewSignature(ScalarTypes.Long.Type())
	function := NewFunctionSymbol("f", signature)
	call.SetSemanticInfo(NewSemanticInfoForType(ScalarTypes.Long.Type()).WithCalledFunctionInfo(analyzerCalledFunctionInfoWithError("KS999", "external bad")))
	call.Name().SetSemanticInfo(NewSemanticInfoForSignature(signature, &function.TypeSymbol))

	diagnostics := NewCalledFunctionHasErrorsAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one warning", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodeCalledFunctionHasErrors ||
		diagnostics[0].Start != strings.Index(text, "f()") ||
		diagnostics[0].Message != "The function 'f' has errors in its definition" {
		t.Fatalf("warning = %#v", diagnostics[0])
	}
}

func TestCalledFunctionHasErrorsAnalyzerReportsFunctionValuedNameReference(t *testing.T) {
	text := "T | project f"
	code := Parse(text)
	name := parsedNameReferenceBySimpleName(t, code, "f")
	signature := NewSignature(ScalarTypes.Long.Type())
	function := NewFunctionSymbol("f", signature)
	name.SetSemanticInfo(NewSemanticInfoForSignature(signature, &function.TypeSymbol).WithCalledFunctionInfo(analyzerCalledFunctionInfoWithError("KS999", "bad reference")))

	diagnostics := NewCalledFunctionHasErrorsAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one warning", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodeCalledFunctionHasErrors ||
		diagnostics[0].Start != strings.LastIndex(text, "f") ||
		diagnostics[0].Length != len("f") {
		t.Fatalf("diagnostic = %#v", diagnostics[0])
	}
}

func TestCalledFunctionHasErrorsAnalyzerSkipsFunctionCallNameReferenceBranch(t *testing.T) {
	text := "T | extend Value=f()"
	code := Parse(text)
	call := firstParsedFunctionCall(t, code)
	signature := NewSignature(ScalarTypes.Long.Type())
	function := NewFunctionSymbol("f", signature)
	call.Name().SetSemanticInfo(NewSemanticInfoForSignature(signature, &function.TypeSymbol).WithCalledFunctionInfo(analyzerCalledFunctionInfoWithError("KS999", "name branch bad")))

	diagnostics := NewCalledFunctionHasErrorsAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want empty", diagnostics)
	}
}

func TestCalledFunctionHasErrorsAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewCalledFunctionHasErrorsAnalyzer().Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodeCalledFunctionHasErrors ||
		diagnostics[0].Category != DiagnosticCategoryGeneral ||
		diagnostics[0].Severity != DiagnosticSeverityWarning ||
		diagnostics[0].Description != "A called function has errors in its definition" ||
		diagnostics[0].Message != "A called function has errors in its definition" {
		t.Fatalf("diagnostic metadata = %#v", diagnostics[0])
	}
}

func analyzerCalledFunctionInfoWithError(code string, message string) *FunctionCallInfo {
	return NewFunctionCallInfo(
		NewFunctionCallExpansion(NewSyntaxNode(SyntaxKindNameReference, nil, NewDiagnostic(code, message)), nil, 0),
		FunctionBodyFactsDefault,
	)
}

func firstParsedFunctionBody(t *testing.T, code *KustoCode) *FunctionBody {
	t.Helper()
	var body *FunctionBody
	WalkSyntax(code.Syntax(), func(element SyntaxElement) {
		if body != nil {
			return
		}
		if found, ok := WrapGeneratedSyntaxElement(element).(*FunctionBody); ok {
			body = found
		}
	})
	if body == nil {
		t.Fatalf("function body not found in %q", code.Text)
	}
	return body
}

func parsedNameReferenceBySimpleName(t *testing.T, code *KustoCode, name string) *NameReference {
	t.Helper()
	var reference *NameReference
	WalkSyntax(code.Syntax(), func(element SyntaxElement) {
		if reference != nil {
			return
		}
		if found, ok := WrapGeneratedSyntaxElement(element).(*NameReference); ok && NameReferenceSimpleName(found) == name {
			reference = found
		}
	})
	if reference == nil {
		t.Fatalf("name reference %q not found in %q", name, code.Text)
	}
	return reference
}
