package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/StdevTimespanConversionAnalyzer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var stdevTimespanConversionDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodeStdevTimespanConversion,
	DiagnosticCategoryCorrectness,
	DiagnosticSeveritySuggestion,
	"This use of stdev() results in a number. To get a timespan use totimespan() on the result.",
	"",
)

func NewStdevTimespanConversionAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "StdevTimespanConversionAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{stdevTimespanConversionDiagnostic}
		},
		Analyze: analyzeStdevTimespanConversion,
	})
}

func analyzeStdevTimespanConversion(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil {
		return diagnostics
	}
	walkSyntaxWithTriviaStartAndAncestors(code.Syntax(), 0, func(element SyntaxElement, triviaStart int, ancestors []SyntaxElement) {
		cancellationToken.ThrowIfCancellationRequested()
		call, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression)
		if !ok || referencedFunctionSymbol(call) != Aggregates.Stdev || functionCallArgumentCount(call) != 1 {
			return
		}
		if !sameType(ExpressionResultType(functionCallArgumentAt(call, 0)), ScalarTypes.TimeSpan.Type()) {
			return
		}
		if isDirectArgumentOfFunction(ancestors, Functions.ToTimespan) {
			return
		}
		if call.Name() != nil {
			diagnostics = append(diagnostics, stdevTimespanConversionDiagnostic.WithByteLocation(elementTextStart(call.Name(), triviaStart), call.Name().Width()))
		}
	})
	return diagnostics
}

func isDirectArgumentOfFunction(ancestors []SyntaxElement, symbol *FunctionSymbol) bool {
	if symbol == nil || len(ancestors) < 4 {
		return false
	}
	if _, ok := ancestors[len(ancestors)-1].(*SeparatedElement); !ok {
		return false
	}
	if _, ok := ancestors[len(ancestors)-2].(*SyntaxList); !ok {
		return false
	}
	if _, ok := WrapGeneratedSyntaxElement(ancestors[len(ancestors)-3]).(*ExpressionList); !ok {
		return false
	}
	parentCall, ok := WrapGeneratedSyntaxElement(ancestors[len(ancestors)-4]).(*FunctionCallExpression)
	return ok && referencedFunctionSymbol(parentCall) == symbol
}
