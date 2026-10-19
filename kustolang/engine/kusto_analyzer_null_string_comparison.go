package kustolang

import (
	"fmt"
	"strconv"
)

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/AvoidUsingNullStringComparisonAnalyzer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var avoidUsingIsNullStringComparisonDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodeAvoidUsingIsNullStringComparision,
	DiagnosticCategoryCorrectness,
	DiagnosticSeverityWarning,
	"Avoid using isnull on string arguments",
	"Avoid using isnull on string arguments, use isempty() instead",
)

var avoidUsingIsNotNullStringComparisonDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodeAvoidUsingIsNotNullStringComparison,
	DiagnosticCategoryCorrectness,
	DiagnosticSeverityWarning,
	"Avoid using isnotnull on string arguments",
	"Avoid using isnotnull on string arguments, use isnotempty() instead",
)

func NewAvoidUsingNullStringComparisonAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "AvoidUsingNullStringComparisonAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{
				avoidUsingIsNullStringComparisonDiagnostic,
				avoidUsingIsNotNullStringComparisonDiagnostic,
			}
		},
		Analyze:   analyzeAvoidUsingNullStringComparison,
		FixAction: avoidUsingNullStringComparisonFixAction,
		FixEdits:  avoidUsingNullStringComparisonFixEdits,
	})
}

func analyzeAvoidUsingNullStringComparison(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
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
		if functionSymbol != Functions.IsNull && functionSymbol != Functions.IsNotNull {
			return
		}
		if !sameType(ExpressionResultType(functionCallArgumentAt(call, 0)), ScalarTypes.String.Type()) {
			return
		}
		diagnostic := avoidUsingIsNotNullStringComparisonDiagnostic
		if functionSymbol == Functions.IsNull {
			diagnostic = avoidUsingIsNullStringComparisonDiagnostic
		}
		diagnostics = append(diagnostics, diagnostic.WithByteLocation(elementTextStart(call, triviaStart), call.Width()))
	})
	return diagnostics
}

func avoidUsingNullStringComparisonFixAction(code *KustoCode, diagnostic Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction {
	call, callTriviaStart := functionCallAtDiagnosticLocation(code, diagnostic)
	if call == nil || call.Name() == nil {
		return actions
	}

	functionName := NameReferenceSimpleName(call.Name())
	newFunctionName := ""
	switch functionName {
	case Functions.IsNull.SymbolName():
		newFunctionName = Functions.IsEmpty.SymbolName()
	case Functions.IsNotNull.SymbolName():
		newFunctionName = Functions.IsNotEmpty.SymbolName()
	}
	if newFunctionName == "" {
		return actions
	}

	nameStart := elementTextStart(call.Name(), callTriviaStart)
	return append(actions, NewCodeActionTitle(
		fmt.Sprintf("Change to '%s'", newFunctionName),
		fmt.Sprintf("Replace call to function '%s' with function '%s'", functionName, newFunctionName),
		[]string{strconv.Itoa(nameStart), strconv.Itoa(call.Name().Width()), newFunctionName},
	))
}

func avoidUsingNullStringComparisonFixEdits(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits {
	data := []string(nil)
	if action != nil {
		data = action.ApplyData()
	}
	if len(data) == 3 {
		if functionNameStart, err := strconv.Atoi(data[0]); err == nil {
			if functionNameLength, err := strconv.Atoi(data[1]); err == nil {
				return NewFixEditsList(functionNameStart, NewTextReplacement(functionNameStart, functionNameLength, data[2]))
			}
		}
	}
	return NewFixEdits(caretPosition, nil)
}

func functionCallAtDiagnosticLocation(code *KustoCode, diagnostic Diagnostic) (*FunctionCallExpression, int) {
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
		if elementTextStart(call, triviaStart) == diagnostic.Start && call.Width() == diagnostic.Length {
			found = call
			foundTriviaStart = triviaStart
		}
	})
	return found, foundTriviaStart
}
