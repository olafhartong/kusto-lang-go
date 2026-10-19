package kustolang

import (
	"strings"
	"testing"
)

func TestPreferUsingOptimizedAlternativeAnalyzerReportsSemanticFunction(t *testing.T) {
	text := "T | summarize count_distinct(User)"
	code := Parse(text)
	call := firstParsedFunctionCall(t, code)
	call.SetSemanticInfo(NewSemanticInfoForSymbol(Aggregates.CountDistinct, ScalarTypes.Long.Type()))

	diagnostics := NewPreferUsingOptimizedAlternativeAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Code != KustoAnalyzerCodePreferUsingOptimizedAlternative ||
		diagnostic.Category != DiagnosticCategoryPerformance ||
		diagnostic.Severity != DiagnosticSeveritySuggestion ||
		diagnostic.Start != strings.Index(text, "count_distinct") ||
		diagnostic.Length != len("count_distinct") {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
	wantMessage := "If possible, consider switching to 'dcount', which is more optimized than 'count_distinct'."
	if diagnostic.Message != wantMessage {
		t.Fatalf("message = %q, want %q", diagnostic.Message, wantMessage)
	}
}

func TestPreferUsingOptimizedAlternativeAnalyzerSkipsUnboundAndUnoptimizedFunctions(t *testing.T) {
	unbound := Parse("T | summarize count_distinct(User)")
	if diagnostics := NewPreferUsingOptimizedAlternativeAnalyzer().Analyze(unbound, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("unbound diagnostics = %#v, want empty", diagnostics)
	}

	bound := Parse("T | summarize dcount(User)")
	call := firstParsedFunctionCall(t, bound)
	call.SetSemanticInfo(NewSemanticInfoForSymbol(Aggregates.DCount, ScalarTypes.Long.Type()))
	if diagnostics := NewPreferUsingOptimizedAlternativeAnalyzer().Analyze(bound, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("unoptimized diagnostics = %#v, want empty", diagnostics)
	}
}

func TestPreferUsingOptimizedAlternativeAnalyzerUsesNameSemanticInfo(t *testing.T) {
	code := Parse("T | summarize custom(User)")
	call := firstParsedFunctionCall(t, code)
	call.Name().SetSemanticInfo(NewSemanticInfoForSymbol(Aggregates.CountDistinctIf, ScalarTypes.Long.Type()))

	diagnostics := NewPreferUsingOptimizedAlternativeAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 || !strings.Contains(diagnostics[0].Message, "'dcountif'") {
		t.Fatalf("diagnostics = %#v, want dcountif warning", diagnostics)
	}
}

func TestPreferUsingOptimizedAlternativeAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewPreferUsingOptimizedAlternativeAnalyzer().Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodePreferUsingOptimizedAlternative ||
		diagnostics[0].Description != "Prefer using more optimized alternative when possible." ||
		diagnostics[0].Message != "Prefer using more optimized alternative when possible." {
		t.Fatalf("diagnostic metadata = %#v", diagnostics[0])
	}
}

func firstParsedFunctionCall(t *testing.T, code *KustoCode) *FunctionCallExpression {
	t.Helper()
	var call *FunctionCallExpression
	WalkSyntax(code.Syntax(), func(element SyntaxElement) {
		if call != nil {
			return
		}
		if found, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression); ok {
			call = found
		}
	})
	if call == nil {
		t.Fatalf("function call not found in %q", code.Text)
	}
	return call
}
