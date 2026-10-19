package kustolang

import (
	"fmt"
	"strconv"
)

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/AvoidUsingFormatDateTimeInPredicateAnalyzer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var avoidUsingFormatDateTimeInPredicateDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodeAvoidUsingFormatDateTimeInPredicate,
	DiagnosticCategoryPerformance,
	DiagnosticSeveritySuggestion,
	"Avoid using format_datetime() in a filter or predicate",
	"Avoid using the 'format_datetime' function in a filter or predicate.\nInstead, use specific datetime functions like startofday or bin.",
)

func NewAvoidUsingFormatDateTimeInPredicateAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "AvoidUsingFormatDateTimeInPredicateAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{avoidUsingFormatDateTimeInPredicateDiagnostic}
		},
		Analyze:   analyzeAvoidUsingFormatDateTimeInPredicate,
		FixAction: avoidUsingFormatDateTimeInPredicateFixAction,
		FixEdits:  avoidUsingFormatDateTimeInPredicateFixEdits,
	})
}

func analyzeAvoidUsingFormatDateTimeInPredicate(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil {
		return diagnostics
	}
	walkSyntaxWithTriviaStartAndAncestors(code.Syntax(), 0, func(element SyntaxElement, triviaStart int, ancestors []SyntaxElement) {
		cancellationToken.ThrowIfCancellationRequested()
		call, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression)
		if !ok || referencedFunctionSymbol(call) != Functions.FormatDatetime {
			return
		}
		if isFormatDatetimeInFilterOrPredicate(ancestors) {
			diagnostics = append(diagnostics, avoidUsingFormatDateTimeInPredicateDiagnostic.WithByteLocation(elementTextStart(call, triviaStart), call.Width()))
		}
	})
	return diagnostics
}

func isFormatDatetimeInFilterOrPredicate(ancestors []SyntaxElement) bool {
	for _, ancestor := range ancestors {
		switch node := WrapGeneratedSyntaxElement(ancestor).(type) {
		case *FilterOperator:
			return true
		case *BinaryExpression:
			if isFormatDatetimeComparisonOperator(node.Kind()) {
				return true
			}
		}
	}
	return false
}

func isFormatDatetimeComparisonOperator(kind SyntaxKind) bool {
	switch kind {
	case SyntaxKindEqualExpression,
		SyntaxKindNotEqualExpression,
		SyntaxKindEqualTildeExpression,
		SyntaxKindGreaterThanOrEqualExpression,
		SyntaxKindGreaterThanExpression,
		SyntaxKindLessThanExpression,
		SyntaxKindLessThanOrEqualExpression:
		return true
	default:
		return false
	}
}

func avoidUsingFormatDateTimeInPredicateFixAction(code *KustoCode, diagnostic Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction {
	if diagnostic.Code != KustoAnalyzerCodeAvoidUsingFormatDateTimeInPredicate {
		return actions
	}
	call, _ := functionCallAtDiagnosticLocation(code, diagnostic)
	if call == nil || functionCallArgumentCount(call) != 2 {
		return actions
	}
	format, ok := ExpressionConstantValue(functionCallArgumentAt(call, 1)).(string)
	if !ok {
		return actions
	}
	alternative := formatDatetimeAlternativeFunctionOrPart(format)
	if alternative == "" {
		return actions
	}
	replacement := alternative
	if formatDatetimeAlternativeIsDateTimePart(alternative) {
		replacement = fmt.Sprintf("datetime_part('%s',)", alternative)
	}
	return append(actions, NewCodeAction(
		"convert_format_datetime",
		fmt.Sprintf("Change to '%s'", replacement),
		fmt.Sprintf("Change use of function 'format_datetime' to function '%s'", replacement),
		[]string{strconv.Itoa(diagnostic.Start), strconv.Itoa(diagnostic.Length), alternative},
	))
}

func avoidUsingFormatDateTimeInPredicateFixEdits(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits {
	data := []string(nil)
	if action != nil {
		data = action.ApplyData()
	}
	if len(data) != 3 {
		return NewFixEdits(caretPosition, nil)
	}
	start, err := strconv.Atoi(data[0])
	if err != nil {
		return NewFixEdits(caretPosition, nil)
	}
	length, err := strconv.Atoi(data[1])
	if err != nil {
		return NewFixEdits(caretPosition, nil)
	}
	alternative := data[2]
	call, callTriviaStart := functionCallAtLocation(code, start, length)
	if call == nil || call.Name() == nil || functionCallArgumentCount(call) != 2 {
		return NewFixEdits(caretPosition, nil)
	}

	edits := []TextEdit{}
	if formatDatetimeAlternativeIsDateTimePart(alternative) {
		edits = analyzerUtilityAddRenameEdits(edits, call.Name().Name(), callTriviaStart, Functions.DatetimePart.SymbolName())
		edits = analyzerUtilityAddInsertArgumentEdits(edits, call, callTriviaStart, 0, fmt.Sprintf("'%s'", alternative))
	} else {
		edits = analyzerUtilityAddRenameEdits(edits, call.Name().Name(), callTriviaStart, alternative)
	}
	edits = analyzerUtilityAddRemoveArgumentEdits(edits, call, callTriviaStart, 1)

	if outer, outerTriviaStart := formatDatetimeOuterCastFunctionCall(code, call); outer != nil {
		edits = analyzerUtilityAddRemoveOuterFunctionCallEdits(edits, outer, outerTriviaStart, call, callTriviaStart)
	}
	return NewFixEdits(elementTextStart(call.Name(), callTriviaStart), edits)
}

func formatDatetimeAlternativeFunctionOrPart(format string) string {
	switch format {
	case "d", "dd":
		return Functions.DayOfMonth.SymbolName()
	case "fff", "FFF":
		return "millisecond"
	case "ffffff", "FFFFFF":
		return "microsecond"
	case "H", "HH":
		return Functions.HourOfDay.SymbolName()
	case "m", "mm":
		return "minute"
	case "M", "MM":
		return Functions.MonthOfYear.SymbolName()
	case "s", "ss":
		return "second"
	case "y", "yy", "yyyy":
		return Functions.GetYear.SymbolName()
	default:
		return ""
	}
}

func formatDatetimeAlternativeIsDateTimePart(value string) bool {
	for _, part := range KustoFacts.DateTimeParts {
		if part == value {
			return true
		}
	}
	return false
}

func functionCallAtLocation(code *KustoCode, start int, length int) (*FunctionCallExpression, int) {
	if code == nil || code.Syntax() == nil {
		return nil, 0
	}
	var found *FunctionCallExpression
	foundTriviaStart := 0
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		if found != nil {
			return
		}
		call, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression)
		if !ok {
			return
		}
		if elementTextStart(call, triviaStart) == start && call.Width() == length {
			found = call
			foundTriviaStart = triviaStart
		}
	})
	return found, foundTriviaStart
}

func formatDatetimeOuterCastFunctionCall(code *KustoCode, inner *FunctionCallExpression) (*FunctionCallExpression, int) {
	if code == nil || code.Syntax() == nil || inner == nil {
		return nil, 0
	}
	var found *FunctionCallExpression
	foundTriviaStart := 0
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		if found != nil {
			return
		}
		call, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression)
		if !ok || call == inner || call.Name() == nil || !formatDatetimeIsOuterCastName(call.Name()) {
			return
		}
		for i := 0; i < functionCallArgumentCount(call); i++ {
			if sameSyntaxExpression(functionCallArgumentAt(call, i), inner) {
				found = call
				foundTriviaStart = triviaStart
				return
			}
		}
	})
	return found, foundTriviaStart
}

func formatDatetimeIsOuterCastName(name *NameReference) bool {
	switch NameReferenceSimpleName(name) {
	case Functions.ToLong.SymbolName(), Functions.ToInt.SymbolName():
		return true
	default:
		return false
	}
}
