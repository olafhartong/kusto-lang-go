package kustolang

import (
	"strings"
	"testing"
)

func TestNullAggregationAnalyzerReportsSumWithAddOrSubtract(t *testing.T) {
	text := "T | summarize sum(Price+Tax), Total=sum(Revenue-Cost)"
	diagnostics := NewNullAggregationAnalyzer().Analyze(Parse(text), nil, CancellationToken{})

	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %#v, want two", diagnostics)
	}
	first := diagnostics[0]
	if first.Code != KustoAnalyzerCodeNullAggregation ||
		first.Category != DiagnosticCategoryCorrectness ||
		first.Severity != DiagnosticSeverityWarning ||
		first.Start != strings.Index(text, "sum(Price+Tax)") ||
		first.Length != len("sum(Price+Tax)") {
		t.Fatalf("first diagnostic = %#v", first)
	}
	wantMessage := "If any of the columns referenced in 'sum(Price+Tax)' contains nulls, significant values might not be included due to scalar arithmetic null propagation rules.\nConsider rewriting it as 'sum(Price) + sum(Tax)'."
	if first.Message != wantMessage {
		t.Fatalf("first message = %q, want %q", first.Message, wantMessage)
	}

	second := diagnostics[1]
	if second.Start != strings.Index(text, "sum(Revenue-Cost)") ||
		second.Length != len("sum(Revenue-Cost)") ||
		!strings.Contains(second.Message, "sum(Revenue) - sum(Cost)") {
		t.Fatalf("second diagnostic = %#v", second)
	}
}

func TestNullAggregationAnalyzerSkipsOtherAggregatesAndNonAdditiveArguments(t *testing.T) {
	text := "T | summarize sum(), sum(Price), sum(Price*Tax), avg(Price+Tax) by bin(Timestamp+Offset, 1h)"
	diagnostics := NewNullAggregationAnalyzer().Analyze(Parse(text), nil, CancellationToken{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want empty", diagnostics)
	}
}

func TestNullAggregationAnalyzerMatchesSemanticSumSymbol(t *testing.T) {
	code := Parse("T | summarize Alias=custom(Price+Tax)")
	var call *FunctionCallExpression
	WalkSyntax(code.Syntax(), func(element SyntaxElement) {
		if found, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression); ok {
			call = found
		}
	})
	if call == nil {
		t.Fatalf("function call not found")
	}
	call.SetSemanticInfo(NewSemanticInfoForSymbol(Aggregates.Sum, ScalarTypes.Long.Type()))

	diagnostics := NewNullAggregationAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 || diagnostics[0].Code != KustoAnalyzerCodeNullAggregation {
		t.Fatalf("diagnostics = %#v, want null aggregation warning", diagnostics)
	}
}

func TestNullAggregationAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewNullAggregationAnalyzer().Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodeNullAggregation ||
		diagnostics[0].Description != "Avoid using operations on possible null values inside sum aggregates" ||
		diagnostics[0].Message != "Avoid using operations on possible null values inside sum aggregates" {
		t.Fatalf("diagnostic metadata = %#v", diagnostics[0])
	}
}
