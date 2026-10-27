package kustolang

import (
	"fmt"
)

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/NullAggregationAnalyzer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var nullAggregationDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodeNullAggregation,
	DiagnosticCategoryCorrectness,
	DiagnosticSeverityWarning,
	"Avoid using operations on possible null values inside sum aggregates",
	"",
)

func NewNullAggregationAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "NullAggregationAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{nullAggregationDiagnostic}
		},
		Analyze: analyzeNullAggregation,
	})
}

func analyzeNullAggregation(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil {
		return diagnostics
	}
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		cancellationToken.ThrowIfCancellationRequested()
		summarize, ok := WrapGeneratedSyntaxElement(element).(*SummarizeOperator)
		if !ok {
			return
		}
		aggregateList := summarize.GetChild(2)
		if aggregateList == nil {
			return
		}
		aggregateTriviaStart := syntaxChildTriviaStart(summarize, triviaStart, 2)
		walkSyntaxWithTriviaStart(aggregateList, aggregateTriviaStart, func(aggregateElement SyntaxElement, aggregateElementTriviaStart int) {
			cancellationToken.ThrowIfCancellationRequested()
			call, ok := WrapGeneratedSyntaxElement(aggregateElement).(*FunctionCallExpression)
			if !ok || !isSumAggregateCall(call, code.Globals) {
				return
			}
			binary := firstFunctionArgumentBinary(call)
			if binary == nil || !isNullAggregationBinaryKind(binary.Kind()) {
				return
			}
			diagnostics = append(diagnostics, nullAggregationDiagnostic.
				WithMessage(nullAggregationMessage(call, binary)).
				WithLocation(elementTextStart(call, aggregateElementTriviaStart), call.Width()))
		})
	})
	return diagnostics
}

func isSumAggregateCall(call *FunctionCallExpression, globals *GlobalState) bool {
	if call == nil {
		return false
	}
	if symbol := referencedFunctionSymbol(call); symbol != nil {
		return symbol == Aggregates.Sum
	}
	name := NameReferenceSimpleName(call.Name())
	if name == "" {
		return false
	}
	if globals != nil && globals.GetAggregate(name) == Aggregates.Sum {
		return true
	}
	return ordinalEqualFold(name, Aggregates.Sum.SymbolName())
}

func referencedFunctionSymbol(call *FunctionCallExpression) *FunctionSymbol {
	if call == nil {
		return nil
	}
	if symbol, ok := call.ReferencedSymbol().(*FunctionSymbol); ok {
		return symbol
	}
	if call.Name() != nil {
		if symbol, ok := call.Name().ReferencedSymbol().(*FunctionSymbol); ok {
			return symbol
		}
	}
	return nil
}

func firstFunctionArgumentBinary(call *FunctionCallExpression) *BinaryExpression {
	binary, _ := functionCallArgumentAt(call, 0).(*BinaryExpression)
	return binary
}

func isNullAggregationBinaryKind(kind SyntaxKind) bool {
	return kind == SyntaxKindAddExpression || kind == SyntaxKindSubtractExpression
}

func nullAggregationMessage(call *FunctionCallExpression, binary *BinaryExpression) string {
	callText := syntaxText(call)
	functionName := syntaxText(call.Name())
	left := syntaxText(binary.Left())
	right := syntaxText(binary.Right())
	operatorText := ""
	if binary.Operator() != nil {
		operatorText = binary.Operator().RawText()
	}
	return fmt.Sprintf(
		"If any of the columns referenced in '%s' contains nulls, significant values might not be included due to scalar arithmetic null propagation rules.\nConsider rewriting it as '%s(%s) %s %s(%s)'.",
		callText,
		functionName,
		left,
		operatorText,
		functionName,
		right,
	)
}

func syntaxChildTriviaStart(parent SyntaxElement, parentTriviaStart int, childIndex int) int {
	start := parentTriviaStart
	if parent == nil {
		return start
	}
	for i := 0; i < childIndex && i < parent.ChildCount(); i++ {
		if child := parent.GetChild(i); child != nil {
			start += child.FullWidth()
		}
	}
	return start
}

func syntaxText(element SyntaxElement) string {
	if element == nil {
		return ""
	}
	return element.Text(IncludeTriviaMinimal)
}
