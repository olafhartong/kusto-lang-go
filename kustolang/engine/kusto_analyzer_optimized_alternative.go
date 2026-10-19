package kustolang

import "fmt"

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/PreferUsingOptimizedAlternative.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var preferUsingOptimizedAlternativeDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodePreferUsingOptimizedAlternative,
	DiagnosticCategoryPerformance,
	DiagnosticSeveritySuggestion,
	"Prefer using more optimized alternative when possible.",
	"",
)

func NewPreferUsingOptimizedAlternativeAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "PreferUsingOptimizedAlternative",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{preferUsingOptimizedAlternativeDiagnostic}
		},
		Analyze: analyzePreferUsingOptimizedAlternative,
	})
}

func analyzePreferUsingOptimizedAlternative(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil {
		return diagnostics
	}
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		cancellationToken.ThrowIfCancellationRequested()
		call, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression)
		if !ok {
			return
		}
		functionSymbol := referencedFunctionSymbol(call)
		if functionSymbol == nil || functionSymbol.OptimizedAlternative() == "" || call.Name() == nil {
			return
		}
		diagnostics = append(diagnostics, preferUsingOptimizedAlternativeDiagnostic.
			WithMessage(fmt.Sprintf("If possible, consider switching to '%s', which is more optimized than '%s'.", functionSymbol.OptimizedAlternative(), functionSymbol.SymbolName())).
			WithLocation(elementTextStart(call.Name(), triviaStart), call.Name().Width()))
	})
	return diagnostics
}
