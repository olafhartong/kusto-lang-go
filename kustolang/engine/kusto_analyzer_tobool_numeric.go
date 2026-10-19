package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/AvoidUsingToBoolOnNumericsAnalyzer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var avoidUsingToBoolOnNumericsDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodeAvoidUsingToBoolOnNumerics,
	DiagnosticCategoryCorrectness,
	DiagnosticSeverityWarning,
	"Avoid using tobool on numeric arguments",
	"Avoid using tobool on numeric arguments, use comparison operators instead",
)

func NewAvoidUsingToBoolOnNumericsAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "AvoidUsingToBoolOnNumericsAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{avoidUsingToBoolOnNumericsDiagnostic}
		},
		Analyze: analyzeAvoidUsingToBoolOnNumerics,
	})
}

func analyzeAvoidUsingToBoolOnNumerics(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil {
		return diagnostics
	}
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		cancellationToken.ThrowIfCancellationRequested()
		call, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression)
		if !ok || functionCallArgumentCount(call) == 0 {
			return
		}
		functionSymbol := referencedFunctionSymbol(call)
		if functionSymbol != Functions.ToBool && functionSymbol != Functions.ToBoolean {
			return
		}
		if isToBoolNumericArgumentType(ExpressionResultType(functionCallArgumentAt(call, 0))) {
			diagnostics = append(diagnostics, avoidUsingToBoolOnNumericsDiagnostic.WithByteLocation(elementTextStart(call, triviaStart), call.Width()))
		}
	})
	return diagnostics
}

func isToBoolNumericArgumentType(typ *TypeSymbol) bool {
	return sameType(typ, ScalarTypes.DateTime.Type()) ||
		sameType(typ, ScalarTypes.Int.Type()) ||
		sameType(typ, ScalarTypes.Decimal.Type()) ||
		sameType(typ, ScalarTypes.Guid.Type()) ||
		sameType(typ, ScalarTypes.Long.Type()) ||
		sameType(typ, ScalarTypes.Real.Type()) ||
		sameType(typ, ScalarTypes.TimeSpan.Type())
}
