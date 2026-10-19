package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/AvoidStrlenWithDynamic.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var avoidStrlenWithDynamicDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodeAvoidStrlenWithDynamic,
	DiagnosticCategoryPerformance,
	DiagnosticSeveritySuggestion,
	"Avoid using the strlen() with dynamic arguments.",
	"Avoid using strlen() with dynamic arguments that may be large bags or arrays.",
)

func NewAvoidStrlenWithDynamicAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "AvoidStrlenWithDynamicAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{avoidStrlenWithDynamicDiagnostic}
		},
		Analyze: analyzeAvoidStrlenWithDynamic,
	})
}

func analyzeAvoidStrlenWithDynamic(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil {
		return diagnostics
	}
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		cancellationToken.ThrowIfCancellationRequested()
		call, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression)
		if !ok || referencedFunctionSymbol(call) != Functions.Strlen || functionCallArgumentCount(call) != 1 {
			return
		}
		if IsDynamicArrayOrBag(ExpressionResultType(functionCallArgumentAt(call, 0))) {
			diagnostics = append(diagnostics, avoidStrlenWithDynamicDiagnostic.WithByteLocation(elementTextStart(call, triviaStart), call.Width()))
		}
	})
	return diagnostics
}
