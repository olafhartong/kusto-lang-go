package kustolang

import (
	"fmt"
	"strconv"
)

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/AvoidUsingContainsAnalyzer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var avoidUsingContainsDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodeAvoidUsingContains,
	DiagnosticCategoryPerformance,
	DiagnosticSeveritySuggestion,
	"Avoid using contains operator",
	"Avoid using the 'contains' operator as it has a high compute price.\nUse the 'has' operator in cases when full term match is desired (see: https://aka.ms/kusto.stringterms).",
)

func NewAvoidUsingContainsAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "AvoidUsingContainsAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{avoidUsingContainsDiagnostic}
		},
		Analyze:   analyzeAvoidUsingContains,
		FixAction: avoidUsingContainsFixAction,
		FixEdits:  avoidUsingContainsFixEdits,
	})
}

func analyzeAvoidUsingContains(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil {
		return diagnostics
	}
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		cancellationToken.ThrowIfCancellationRequested()
		binary, ok := WrapGeneratedSyntaxElement(element).(*BinaryExpression)
		if !ok || binary.Operator() == nil {
			return
		}
		if hasKind := containsAnalyzerHasOperatorKind(binary.Operator().Kind()); hasKind == SyntaxKindNone {
			return
		}
		if analyzerExpressionIsDbColumn(binary.Left(), code.Globals) && !analyzerExpressionIsShortStringConstant(binary.Right(), 3) {
			diagnostics = append(diagnostics, avoidUsingContainsDiagnostic.WithByteLocation(elementTextStart(binary.Operator(), triviaStart+binary.Left().FullWidth()), binary.Operator().Width()))
		}
	})
	return diagnostics
}

func avoidUsingContainsFixAction(code *KustoCode, diagnostic Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction {
	if diagnostic.Code != KustoAnalyzerCodeAvoidUsingContains {
		return actions
	}
	token := syntaxTokenAtDiagnosticStart(code, diagnostic.Start)
	if token == nil {
		return actions
	}
	hasKind := containsAnalyzerHasOperatorKind(token.Kind())
	if hasKind == SyntaxKindNone {
		return actions
	}
	hasOperator := SyntaxText(hasKind)
	return append(actions, NewCodeActionTitle(
		fmt.Sprintf("Change to '%s'", hasOperator),
		fmt.Sprintf("Replace operator '%s' with operator '%s'", token.RawText(), hasOperator),
		[]string{strconv.Itoa(diagnostic.Start), hasOperator},
	))
}

func avoidUsingContainsFixEdits(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits {
	data := []string(nil)
	if action != nil {
		data = action.ApplyData()
	}
	if len(data) == 2 {
		if operatorStart, err := strconv.Atoi(data[0]); err == nil {
			token := syntaxTokenAtDiagnosticStart(code, operatorStart)
			if token != nil {
				return NewFixEditsList(operatorStart, NewTextReplacement(operatorStart, token.Width(), data[1]))
			}
		}
	}
	return NewFixEdits(caretPosition, nil)
}

func containsAnalyzerHasOperatorKind(tokenKind SyntaxKind) SyntaxKind {
	switch tokenKind {
	case SyntaxKindContainsKeyword:
		return SyntaxKindHasKeyword
	case SyntaxKindContainsCsKeyword, SyntaxKindContains_CsKeyword:
		return SyntaxKindHasCsKeyword
	case SyntaxKindNotContainsKeyword, SyntaxKindNotBangContainsKeyword:
		return SyntaxKindNotHasKeyword
	case SyntaxKindNotContainsCsKeyword, SyntaxKindNotBangContainsCsKeyword:
		return SyntaxKindNotHasCsKeyword
	default:
		return SyntaxKindNone
	}
}

func analyzerExpressionIsDbColumn(expr Expression, globals *GlobalState) bool {
	return analyzerUtilityIsDbColumn(expr, globals)
}

func analyzerExpressionIsShortStringConstant(expr Expression, lengthNotShort int) bool {
	return analyzerUtilityIsShortStringConstant(expr, nil, lengthNotShort)
}

func syntaxTokenAtDiagnosticStart(code *KustoCode, start int) *SyntaxToken {
	if code == nil || code.Syntax() == nil {
		return nil
	}
	var found *SyntaxToken
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		if found != nil {
			return
		}
		token, ok := element.(*SyntaxToken)
		if !ok {
			return
		}
		if elementTextStart(token, triviaStart) == start {
			found = token
		}
	})
	return found
}
