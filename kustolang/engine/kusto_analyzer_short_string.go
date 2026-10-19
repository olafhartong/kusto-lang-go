package kustolang

import "unicode/utf16"

// Ported from the the upstream compatibility implementation implementation in
// src/Kusto.Language/Editor/Kusto/KustoAnalyzers.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var avoidUsingShortStringComparisonDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodeAvoidUsingShortStringComparision,
	DiagnosticCategoryPerformance,
	DiagnosticSeveritySuggestion,
	"Avoid short string comparisons",
	"Avoid using short strings (less than 4 characters) for string comparison operations (see: https://aka.ms/kusto.stringterms).",
)

func NewAvoidUsingShortStringComparisionAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "AvoidUsingShortStringComparisionAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{avoidUsingShortStringComparisonDiagnostic}
		},
		Analyze: analyzeAvoidUsingShortStringComparison,
	})
}

func analyzeAvoidUsingShortStringComparison(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil {
		return diagnostics
	}
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		cancellationToken.ThrowIfCancellationRequested()
		expression, ok := WrapGeneratedSyntaxElement(element).(Expression)
		if !ok || !shortStringAnalyzerIsRelevantComparison(expression) || !analyzerUtilityHasDbColumnOperand(expression, code.Globals) {
			return
		}
		location := analyzerUtilityGetFirstMatchingOperand(expression, shortStringAnalyzerIsShortStringConstant, true)
		if location != nil {
			locationStart := shortStringAnalyzerLocationStart(expression, triviaStart, location)
			diagnostics = append(diagnostics, avoidUsingShortStringComparisonDiagnostic.WithByteLocation(locationStart, location.Width()))
		}
	})
	return diagnostics
}

func shortStringAnalyzerLocationStart(root Expression, rootTriviaStart int, location Expression) int {
	start := rootTriviaStart
	walkSyntaxWithTriviaStart(root, rootTriviaStart, func(element SyntaxElement, triviaStart int) {
		candidate, ok := WrapGeneratedSyntaxElement(element).(Expression)
		if ok && sameSyntaxExpression(candidate, location) {
			start = elementTextStart(candidate, triviaStart)
		}
	})
	return start
}

func shortStringAnalyzerIsRelevantComparison(expression Expression) bool {
	switch expression.Kind() {
	case SyntaxKindEqualExpression,
		SyntaxKindNotEqualExpression,
		SyntaxKindHasExpression,
		SyntaxKindNotHasExpression,
		SyntaxKindHasCsExpression,
		SyntaxKindNotHasCsExpression,
		SyntaxKindStartsWithExpression,
		SyntaxKindNotStartsWithExpression,
		SyntaxKindHasAnyExpression,
		SyntaxKindHasAllExpression:
		return true
	default:
		return false
	}
}

func shortStringAnalyzerIsShortStringConstant(expression Expression) bool {
	value, ok := ExpressionConstantValue(expression).(string)
	return ok && value != "" && len(utf16.Encode([]rune(value))) < 4
}
