package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/AvoidUsingLegacyPartitionAnalyzer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var avoidUsingLegacyPartitionDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodeAvoidUsingLegacyPartition,
	DiagnosticCategoryGeneral,
	DiagnosticSeveritySuggestion,
	"It is recommended to use the `native` or `shuffle` strategy of the partition operator. The legacy mode is limited to 64 partitions and is less efficient than the `native` or `shuffle` strategies.",
	"",
)

func NewAvoidUsingLegacyPartitionAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "AvoidUsingLegacyPartitionAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{avoidUsingLegacyPartitionDiagnostic}
		},
		Analyze: analyzeAvoidUsingLegacyPartition,
	})
}

func analyzeAvoidUsingLegacyPartition(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil {
		return diagnostics
	}
	abort := false
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		if abort {
			return
		}
		cancellationToken.ThrowIfCancellationRequested()
		partition, ok := WrapGeneratedSyntaxElement(element).(*PartitionOperator)
		if !ok {
			return
		}
		if _, ok := partition.Operand().(*PartitionSubquery); !ok {
			return
		}
		usingLegacy, validStrategy := partitionUsesLegacyStrategy(partition)
		if !validStrategy {
			abort = true
			return
		}
		if usingLegacy {
			start := elementTextStart(partition, triviaStart)
			diagnostics = append(diagnostics, avoidUsingLegacyPartitionDiagnostic.WithByteLocation(start, elementEnd(partition, triviaStart)-start))
		}
	})
	return diagnostics
}

func partitionUsesLegacyStrategy(partition *PartitionOperator) (usingLegacy bool, validStrategy bool) {
	usingLegacy = true
	validStrategy = true
	if partition == nil || partition.Parameters() == nil || partition.Parameters().Count() == 0 {
		return usingLegacy, validStrategy
	}
	for _, parameter := range partition.Parameters().Elements() {
		name := NameDeclarationSimpleName(parameter.Name())
		switch name {
		case SyntaxText(SyntaxKindHintDotStrategyKeyword):
			if !ExpressionIsLiteral(parameter.Expression()) {
				return false, false
			}
			usingLegacy = ExpressionLiteralValue(parameter.Expression()) == "legacy"
		case SyntaxText(SyntaxKindHintDotShuffleKeyKeyword):
			usingLegacy = false
		}
	}
	return usingLegacy, validStrategy
}
