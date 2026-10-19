package kustolang

import (
	"strings"
	"testing"
)

func TestStdevTimespanConversionAnalyzerReportsTimespanStdev(t *testing.T) {
	text := "T | summarize S=stdev(Duration)"
	code := Parse(text)
	call := parsedFunctionCallByName(t, code, "stdev")
	call.SetSemanticInfo(NewSemanticInfoForSymbol(Aggregates.Stdev, ScalarTypes.Real.Type()))
	setExpressionSemanticType(t, functionCallArgumentAt(call, 0), ScalarTypes.TimeSpan.Type())

	diagnostics := NewStdevTimespanConversionAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Code != KustoAnalyzerCodeStdevTimespanConversion ||
		diagnostic.Category != DiagnosticCategoryCorrectness ||
		diagnostic.Severity != DiagnosticSeveritySuggestion ||
		diagnostic.Start != strings.Index(text, "stdev") ||
		diagnostic.Length != len("stdev") {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestStdevTimespanConversionAnalyzerSkipsConvertedStdev(t *testing.T) {
	code := Parse("T | summarize S=totimespan(stdev(Duration))")
	outer := parsedFunctionCallByName(t, code, "totimespan")
	outer.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.ToTimespan, ScalarTypes.TimeSpan.Type()))
	inner := parsedFunctionCallByName(t, code, "stdev")
	inner.SetSemanticInfo(NewSemanticInfoForSymbol(Aggregates.Stdev, ScalarTypes.Real.Type()))
	setExpressionSemanticType(t, functionCallArgumentAt(inner, 0), ScalarTypes.TimeSpan.Type())

	diagnostics := NewStdevTimespanConversionAnalyzer().Analyze(code, nil, CancellationToken{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want empty", diagnostics)
	}
}

func TestStdevTimespanConversionAnalyzerSkipsUnboundNonTimespanAndWrongArity(t *testing.T) {
	unbound := Parse("T | summarize S=stdev(Duration)")
	if diagnostics := NewStdevTimespanConversionAnalyzer().Analyze(unbound, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("unbound diagnostics = %#v, want empty", diagnostics)
	}

	longArg := Parse("T | summarize S=stdev(Duration)")
	longCall := parsedFunctionCallByName(t, longArg, "stdev")
	longCall.SetSemanticInfo(NewSemanticInfoForSymbol(Aggregates.Stdev, ScalarTypes.Real.Type()))
	setExpressionSemanticType(t, functionCallArgumentAt(longCall, 0), ScalarTypes.Long.Type())
	if diagnostics := NewStdevTimespanConversionAnalyzer().Analyze(longArg, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("long diagnostics = %#v, want empty", diagnostics)
	}

	wrongArity := Parse("T | summarize S=stdev(Duration, Other)")
	wrongArityCall := parsedFunctionCallByName(t, wrongArity, "stdev")
	wrongArityCall.SetSemanticInfo(NewSemanticInfoForSymbol(Aggregates.Stdev, ScalarTypes.Real.Type()))
	setExpressionSemanticType(t, functionCallArgumentAt(wrongArityCall, 0), ScalarTypes.TimeSpan.Type())
	if diagnostics := NewStdevTimespanConversionAnalyzer().Analyze(wrongArity, nil, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("wrong-arity diagnostics = %#v, want empty", diagnostics)
	}
}

func TestStdevTimespanConversionAnalyzerDiagnosticsMetadata(t *testing.T) {
	diagnostics := NewStdevTimespanConversionAnalyzer().Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if diagnostics[0].Code != KustoAnalyzerCodeStdevTimespanConversion ||
		diagnostics[0].Description != "This use of stdev() results in a number. To get a timespan use totimespan() on the result." ||
		diagnostics[0].Message != "This use of stdev() results in a number. To get a timespan use totimespan() on the result." {
		t.Fatalf("diagnostic metadata = %#v", diagnostics[0])
	}
}

func parsedFunctionCallByName(t *testing.T, code *KustoCode, name string) *FunctionCallExpression {
	t.Helper()
	var call *FunctionCallExpression
	WalkSyntax(code.Syntax(), func(element SyntaxElement) {
		if call != nil {
			return
		}
		if found, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression); ok && NameReferenceSimpleName(found.Name()) == name {
			call = found
		}
	})
	if call == nil {
		t.Fatalf("function call %q not found in %q", name, code.Text)
	}
	return call
}
