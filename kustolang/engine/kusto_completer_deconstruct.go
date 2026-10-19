package kustolang

import "strings"

// Ported from GetDeconstructCompletions and GetUnpipedExpression in
// src/Kusto.Language/Editor/Kusto/KustoCompleter.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func (c *kustoCompleter) getDeconstructCompletions(position int, builder *kustoCompletionBuilder) completionMode {
	parenthesized, call, tuple := c.deconstructCompletionContext(position)
	if parenthesized == nil || call == nil || tuple == nil {
		return completionModeCombined
	}
	columns := tuple.Columns()
	displayText := "(" + columns[0].SymbolName() + ", ...) = "
	markers := make([]string, 0, len(columns))
	for _, column := range columns {
		markers = append(markers, "[["+column.SymbolName()+"]]")
	}
	item := NewCompletionItem(CompletionKindExample, displayText,
		WithCompletionMatchText(NameReferenceSimpleName(call.Name())),
		WithCompletionPriority(CompletionPriorityTop),
		WithCompletionApplyTextMarkers(strings.Join(markers, ", ")+") = "))
	builder.add(item)
	return completionModeCombined
}

func (c *kustoCompleter) deconstructCompletionContext(position int) (*ParenthesizedExpression, *FunctionCallExpression, *TupleSymbol) {
	if c == nil || c.code == nil || c.code.Syntax() == nil || position < 0 || position > len(c.code.Text) {
		return nil, nil, nil
	}
	var selected *ParenthesizedExpression
	var selectedCall *FunctionCallExpression
	var selectedTuple *TupleSymbol
	selectedStart := -1
	walkSyntaxWithTriviaStartAndAncestors(c.code.Syntax(), 0, func(element SyntaxElement, _ int, _ []SyntaxElement) {
		parenthesized, ok := WrapGeneratedSyntaxElement(element).(*ParenthesizedExpression)
		if !ok || parenthesized == nil || parenthesized.OpenParen() == nil || parenthesized.Expression() == nil || parenthesized.CloseParen() == nil {
			return
		}
		open := parenthesized.OpenParen()
		if open.Width() == 0 || parenthesized.Expression().Width() == 0 || parenthesized.CloseParen().Width() != 0 || position != open.End() {
			return
		}
		call, ok := completionUnpipedExpression(parenthesized.Expression()).(*FunctionCallExpression)
		if !ok || call == nil || call.Name() == nil {
			return
		}
		switch WrapGeneratedSyntaxElement(call.Parent()).(type) {
		case NamedExpression, *LetStatement:
			return
		}
		tuple, ok := TupleForType(ExpressionRawResultType(call))
		if !ok || tuple == nil || len(tuple.Columns()) <= 1 || open.TextStart() < selectedStart {
			return
		}
		selected = parenthesized
		selectedCall = call
		selectedTuple = tuple
		selectedStart = open.TextStart()
	})
	return selected, selectedCall, selectedTuple
}

func completionUnpipedExpression(expression Expression) Expression {
	for {
		pipe, ok := expression.(*PipeExpression)
		if !ok || pipe == nil || pipe.Expression() == nil {
			return expression
		}
		expression = pipe.Expression()
	}
}
