package kustolang

import (
	"fmt"
	"strconv"
	"strings"
)

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/AvoidUsingHasWithIPv4Strings.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var avoidUsingHasWithIPv4StringsDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodeAvoidUsingHasWithIPv4Strings,
	DiagnosticCategoryPerformance,
	DiagnosticSeveritySuggestion,
	"Avoid using has operator with IPv4 strings.",
	"Avoid using the 'has' operator with IPv4 strings. Use the 'has_ipv4' function instead.",
)

func NewAvoidUsingHasWithIPv4StringsAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "AvoidUsingHasWithIPv4StringsAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{avoidUsingHasWithIPv4StringsDiagnostic}
		},
		Analyze:   analyzeAvoidUsingHasWithIPv4Strings,
		FixAction: avoidUsingHasWithIPv4StringsFixAction,
		FixEdits:  avoidUsingHasWithIPv4StringsFixEdits,
	})
}

func analyzeAvoidUsingHasWithIPv4Strings(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil {
		return diagnostics
	}
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		cancellationToken.ThrowIfCancellationRequested()
		switch expr := WrapGeneratedSyntaxElement(element).(type) {
		case *BinaryExpression:
			operator := expr.Operator()
			if operator == nil || !analyzerExpressionIsDbColumn(expr.Left(), code.Globals) {
				return
			}
			functionName := ""
			switch expr.Kind() {
			case SyntaxKindHasExpression, SyntaxKindHasCsExpression, SyntaxKindContainsExpression, SyntaxKindContainsCsExpression:
				if analyzerExpressionIsIPv4Constant(expr.Right()) {
					functionName = "has_ipv4"
				}
			case SyntaxKindHasPrefixExpression, SyntaxKindHasPrefixCsExpression:
				if analyzerExpressionIsIPv4PrefixConstant(expr.Right()) {
					functionName = "has_ipv4_prefix"
				}
			}
			if functionName != "" {
				diagnostics = append(diagnostics, avoidUsingHasWithIPv4StringsDiagnosticFor(operator.RawText(), functionName).
					WithLocation(elementTextStart(operator, triviaStart+expr.Left().FullWidth()), operator.Width()))
			}
		case *HasAnyExpression:
			operator := expr.Operator()
			if operator == nil || !analyzerExpressionIsDbColumn(expr.Left(), code.Globals) || !hasAnyExpressionHasOnlyIPv4Constants(expr) {
				return
			}
			diagnostics = append(diagnostics, avoidUsingHasWithIPv4StringsDiagnosticFor(operator.RawText(), "has_any_ipv4").
				WithLocation(elementTextStart(operator, triviaStart+expr.Left().FullWidth()), operator.Width()))
		}
	})
	return diagnostics
}

func avoidUsingHasWithIPv4StringsDiagnosticFor(operatorName string, ipv4FunctionName string) Diagnostic {
	return avoidUsingHasWithIPv4StringsDiagnostic.WithMessage(
		fmt.Sprintf("Avoid using the '%s' operator with IPv4 strings. Use the '%s' function instead.", operatorName, ipv4FunctionName),
	)
}

func avoidUsingHasWithIPv4StringsFixAction(code *KustoCode, diagnostic Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction {
	if diagnostic.Code != KustoAnalyzerCodeAvoidUsingHasWithIPv4Strings {
		return actions
	}
	token := syntaxTokenAtDiagnosticStart(code, diagnostic.Start)
	if token == nil {
		return actions
	}
	ipv4Name := ipv4FunctionNameForTokenKind(token.Kind())
	return append(actions, NewCodeActionTitle(
		fmt.Sprintf("Change to '%s'", ipv4Name),
		fmt.Sprintf("Replace use of operator '%s' with a call to function '%s'", token.RawText(), ipv4Name),
		[]string{strconv.Itoa(diagnostic.Start), ipv4Name},
	))
}

func avoidUsingHasWithIPv4StringsFixEdits(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits {
	data := []string(nil)
	if action != nil {
		data = action.ApplyData()
	}
	if len(data) != 2 {
		return NewFixEdits(caretPosition, nil)
	}
	tokenStart, err := strconv.Atoi(data[0])
	if err != nil {
		return NewFixEdits(caretPosition, nil)
	}
	binary, binaryTriviaStart, hasAny, hasAnyTriviaStart := ipv4ExpressionAtOperatorStart(code, tokenStart)
	switch {
	case hasAny != nil:
		return ipv4HasAnyFixEdits(hasAny, hasAnyTriviaStart, data[1])
	case binary != nil:
		return ipv4BinaryFixEdits(binary, binaryTriviaStart, data[1])
	default:
		return NewFixEdits(caretPosition, nil)
	}
}

func ipv4BinaryFixEdits(binary *BinaryExpression, triviaStart int, functionName string) FixEdits {
	if binary == nil {
		return NewFixEdits(0, nil)
	}
	operator := binary.Operator()
	right := binary.Right()
	if operator == nil || right == nil {
		return NewFixEdits(0, nil)
	}
	nodeStart := elementTextStart(binary, triviaStart)
	nodeEnd := elementEnd(binary, triviaStart)
	operatorTriviaStart := triviaStart + binary.Left().FullWidth()
	operatorEnd := operatorTriviaStart + operator.FullWidth()
	rightStart := elementTextStart(right, operatorEnd)
	return NewFixEditsList(nodeStart,
		NewTextInsertion(nodeStart, functionName+"("),
		NewTextReplacement(operatorTriviaStart, rightStart-operatorTriviaStart, ", "),
		NewTextInsertion(nodeEnd, ")"),
	)
}

func ipv4HasAnyFixEdits(hasAny *HasAnyExpression, triviaStart int, functionName string) FixEdits {
	if hasAny == nil || hasAny.Operator() == nil || hasAny.Right() == nil || hasAny.Right().OpenParen() == nil {
		return NewFixEdits(0, nil)
	}
	nodeStart := elementTextStart(hasAny, triviaStart)
	operatorTriviaStart := triviaStart + hasAny.Left().FullWidth()
	operatorEnd := operatorTriviaStart + hasAny.Operator().FullWidth()
	openParen := hasAny.Right().OpenParen()
	openParenStart := elementTextStart(openParen, operatorEnd)
	openParenEnd := openParenStart + openParen.Width()
	return NewFixEditsList(nodeStart,
		NewTextInsertion(nodeStart, functionName+"("),
		NewTextReplacement(operatorTriviaStart, openParenEnd-operatorTriviaStart, ", "),
	)
}

func ipv4ExpressionAtOperatorStart(code *KustoCode, operatorStart int) (*BinaryExpression, int, *HasAnyExpression, int) {
	if code == nil || code.Syntax() == nil {
		return nil, 0, nil, 0
	}
	var foundBinary *BinaryExpression
	foundBinaryTriviaStart := 0
	var foundHasAny *HasAnyExpression
	foundHasAnyTriviaStart := 0
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		if foundBinary != nil || foundHasAny != nil {
			return
		}
		switch expr := WrapGeneratedSyntaxElement(element).(type) {
		case *BinaryExpression:
			if expr.Operator() != nil && elementTextStart(expr.Operator(), triviaStart+expr.Left().FullWidth()) == operatorStart {
				foundBinary = expr
				foundBinaryTriviaStart = triviaStart
			}
		case *HasAnyExpression:
			if expr.Operator() != nil && elementTextStart(expr.Operator(), triviaStart+expr.Left().FullWidth()) == operatorStart {
				foundHasAny = expr
				foundHasAnyTriviaStart = triviaStart
			}
		}
	})
	return foundBinary, foundBinaryTriviaStart, foundHasAny, foundHasAnyTriviaStart
}

func ipv4FunctionNameForTokenKind(kind SyntaxKind) string {
	switch kind {
	case SyntaxKindHasPrefixKeyword, SyntaxKindHasPrefixCsKeyword:
		return "has_ipv4_prefix"
	case SyntaxKindHasAnyKeyword:
		return "has_any_ipv4"
	default:
		return "has_ipv4"
	}
}

func hasAnyExpressionHasOnlyIPv4Constants(expr *HasAnyExpression) bool {
	if expr == nil || expr.Right() == nil || expr.Right().Expressions() == nil || expr.Right().Expressions().Count() == 0 {
		return false
	}
	expressions := expr.Right().Expressions()
	for i := 0; i < expressions.Count(); i++ {
		element := expressions.At(i)
		if element == nil || !analyzerExpressionIsIPv4Constant(element.Element()) {
			return false
		}
	}
	return true
}

func analyzerExpressionIsIPv4Constant(expr Expression) bool {
	value, ok := ExpressionConstantValue(expr).(string)
	return ok && analyzerStringIsIPv4(value)
}

func analyzerExpressionIsIPv4PrefixConstant(expr Expression) bool {
	value, ok := ExpressionConstantValue(expr).(string)
	return ok && analyzerStringIsIPv4Prefix(value)
}

func analyzerStringIsIPv4(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if !analyzerStringIsIPv4Number(part) {
			return false
		}
	}
	return true
}

func analyzerStringIsIPv4Prefix(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) < 2 || len(parts) > 4 || parts[len(parts)-1] != "" {
		return false
	}
	for _, part := range parts[:len(parts)-1] {
		if !analyzerStringIsIPv4Number(part) {
			return false
		}
	}
	return true
}

func analyzerStringIsIPv4Number(value string) bool {
	if len(value) == 0 || len(value) > 3 {
		return false
	}
	number := 0
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
		number = number*10 + int(ch-'0')
	}
	if len(value) == 3 {
		return number >= 100 && number <= 255
	}
	return number <= 99
}
