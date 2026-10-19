package kustolang

import "reflect"

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/AnalyzerUtilities.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func analyzerUtilityIsDbColumn(expr Expression, globals *GlobalState) bool {
	if expr == nil || globals == nil {
		return false
	}
	accessor, ok := expr.(interface{ ReferencedSymbol() Symbol })
	if !ok {
		return false
	}
	column, ok := accessor.ReferencedSymbol().(*ColumnSymbol)
	return ok && globals.GetTable(column) != nil
}

func analyzerUtilityHasDbColumnOperand(expr Expression, globals *GlobalState) bool {
	return analyzerUtilityHasMatchingOperand(expr, func(operand Expression) bool {
		return analyzerUtilityIsDbColumn(operand, globals)
	})
}

func analyzerUtilityIsShortStringConstant(expr Expression, globals *GlobalState, lengthNotShort ...int) bool {
	threshold := 3
	if len(lengthNotShort) > 0 {
		threshold = lengthNotShort[0]
	}
	value, ok := ExpressionConstantValue(expr).(string)
	return ok && len(value) < threshold
}

func analyzerUtilityHasMatchingOperand(expr Expression, fnMatches func(Expression) bool, includeListOperandValues ...bool) bool {
	return analyzerUtilityGetFirstMatchingOperand(expr, fnMatches, includeListOperandValues...) != nil
}

func analyzerUtilityHasAnyMatchingElement(list *SeparatedSyntaxListOf[Expression], fnMatches func(Expression) bool) bool {
	return analyzerUtilityGetFirstMatchingElement(list, fnMatches) != nil
}

func analyzerUtilityGetFirstMatchingOperand(expr Expression, fnMatches func(Expression) bool, includeListOperandValues ...bool) Expression {
	if expr == nil || fnMatches == nil {
		return nil
	}
	includeListValues := len(includeListOperandValues) > 0 && includeListOperandValues[0]
	switch node := expr.(type) {
	case *BinaryExpression:
		if fnMatches(node.Left()) {
			return node.Left()
		}
		if fnMatches(node.Right()) {
			return node.Right()
		}
	case *PrefixUnaryExpression:
		if fnMatches(node.Expression()) {
			return node.Expression()
		}
	case *HasAllExpression:
		if fnMatches(node.Left()) {
			return node.Left()
		}
		if includeListValues && node.Right() != nil {
			return analyzerUtilityGetFirstMatchingElement(node.Right().Expressions(), fnMatches)
		}
	case *HasAnyExpression:
		if fnMatches(node.Left()) {
			return node.Left()
		}
		if includeListValues && node.Right() != nil {
			return analyzerUtilityGetFirstMatchingElement(node.Right().Expressions(), fnMatches)
		}
	}
	return nil
}

func analyzerUtilityGetFirstMatchingElement(list *SeparatedSyntaxListOf[Expression], fnMatches func(Expression) bool) Expression {
	if list == nil || fnMatches == nil {
		return nil
	}
	for i := 0; i < list.Count(); i++ {
		element := list.At(i)
		if element != nil && fnMatches(element.Element()) {
			return element.Element()
		}
	}
	return nil
}

func analyzerUtilityAddRenameEdits(edits []TextEdit, name Name, nameTriviaStart int, newName string) []TextEdit {
	if name == nil {
		return edits
	}
	return append(edits, NewTextReplacement(elementTextStart(name, nameTriviaStart), name.ByteWidth(), newName))
}

func analyzerUtilityAddRemoveArgumentEdits(edits []TextEdit, functionCall *FunctionCallExpression, functionCallTriviaStart int, argumentIndex int) []TextEdit {
	argument, positions, ok := analyzerUtilityFunctionCallArgumentPosition(functionCall, functionCallTriviaStart, argumentIndex)
	if !ok {
		return edits
	}
	prevToken, prevTriviaStart := analyzerUtilityPreviousArgumentToken(functionCall, functionCallTriviaStart, argumentIndex)
	nextToken, nextTriviaStart := analyzerUtilityNextArgumentToken(functionCall, functionCallTriviaStart, argumentIndex)
	start := positions.ArgumentTriviaStart
	end := elementEnd(argument, positions.ArgumentTriviaStart)
	if prevToken != nil && nextToken != nil {
		switch prevToken.Kind() {
		case SyntaxKindCommaToken:
			start = elementTextStart(prevToken, prevTriviaStart)
		case SyntaxKindOpenParenToken:
			start = elementEnd(prevToken, prevTriviaStart)
		}
		switch nextToken.Kind() {
		case SyntaxKindCommaToken, SyntaxKindCloseParenToken:
			end = elementTextStart(nextToken, nextTriviaStart)
		}
	} else if prevToken != nil {
		switch prevToken.Kind() {
		case SyntaxKindCommaToken:
			start = elementTextStart(prevToken, prevTriviaStart)
		case SyntaxKindOpenParenToken:
			start = elementEnd(prevToken, prevTriviaStart)
		}
	} else if nextToken != nil {
		switch nextToken.Kind() {
		case SyntaxKindCommaToken, SyntaxKindCloseParenToken:
			end = elementTextStart(nextToken, nextTriviaStart)
		}
	}
	return append(edits, NewTextDeletion(start, end-start))
}

func analyzerUtilityAddInsertArgumentEdits(edits []TextEdit, functionCall *FunctionCallExpression, functionCallTriviaStart int, argumentIndex int, newArgumentText string) []TextEdit {
	if functionCall == nil || functionCall.ArgumentList() == nil || functionCall.ArgumentList().Expressions() == nil {
		return edits
	}
	argumentList := functionCall.ArgumentList()
	argumentListTriviaStart := syntaxChildTriviaStart(functionCall, functionCallTriviaStart, 1)
	closeParenTriviaStart := syntaxChildTriviaStart(argumentList, argumentListTriviaStart, 2)
	expressions := argumentList.Expressions()
	if expressions.Count() == 0 {
		edits = append(edits, NewTextInsertion(closeParenTriviaStart, newArgumentText))
	}
	if argumentIndex >= expressions.Count() {
		edits = append(edits, NewTextInsertion(closeParenTriviaStart, ", "+newArgumentText))
	} else if argumentIndex >= 0 {
		if argument, positions, ok := analyzerUtilityFunctionCallArgumentPosition(functionCall, functionCallTriviaStart, argumentIndex); ok {
			edits = append(edits, NewTextInsertion(elementTextStart(argument, positions.ArgumentTriviaStart), newArgumentText+", "))
		}
	}
	return edits
}

func analyzerUtilityAddAddArgumentEdits(edits []TextEdit, functionCall *FunctionCallExpression, functionCallTriviaStart int, newArgumentText string) []TextEdit {
	count := functionCallArgumentCount(functionCall)
	return analyzerUtilityAddInsertArgumentEdits(edits, functionCall, functionCallTriviaStart, count, newArgumentText)
}

func analyzerUtilityGetFunctionCall(root SyntaxElement, argument Expression) *FunctionCallExpression {
	if root == nil || argument == nil {
		return nil
	}
	var found *FunctionCallExpression
	WalkSyntax(root, func(element SyntaxElement) {
		if found != nil {
			return
		}
		call, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression)
		if !ok || call.ArgumentList() == nil || call.ArgumentList().Expressions() == nil {
			return
		}
		expressions := call.ArgumentList().Expressions()
		for i := 0; i < expressions.Count(); i++ {
			item := expressions.At(i)
			if item != nil && sameSyntaxExpression(item.Element(), argument) {
				found = call
				return
			}
		}
	})
	return found
}

func analyzerUtilityAddRemoveOuterFunctionCallEdits(edits []TextEdit, outer *FunctionCallExpression, outerTriviaStart int, argument Expression, argumentTriviaStart int) []TextEdit {
	if outer == nil || argument == nil {
		return edits
	}
	outerStart := elementTextStart(outer, outerTriviaStart)
	argumentEnd := elementEnd(argument, argumentTriviaStart)
	edits = append(edits,
		NewTextDeletion(outerStart, argumentTriviaStart-outerStart),
		NewTextDeletion(argumentEnd, elementEnd(outer, outerTriviaStart)-argumentEnd),
	)
	return edits
}

type analyzerUtilityArgumentPosition struct {
	SeparatedTriviaStart int
	ArgumentTriviaStart  int
}

func analyzerUtilityFunctionCallArgumentPosition(functionCall *FunctionCallExpression, functionCallTriviaStart int, argumentIndex int) (Expression, analyzerUtilityArgumentPosition, bool) {
	if functionCall == nil || argumentIndex < 0 || functionCall.ArgumentList() == nil || functionCall.ArgumentList().Expressions() == nil {
		return nil, analyzerUtilityArgumentPosition{}, false
	}
	expressions := functionCall.ArgumentList().Expressions()
	if argumentIndex >= expressions.Count() {
		return nil, analyzerUtilityArgumentPosition{}, false
	}
	separated := expressions.At(argumentIndex)
	if separated == nil || separated.Element() == nil {
		return nil, analyzerUtilityArgumentPosition{}, false
	}
	separatedTriviaStart, ok := analyzerUtilitySeparatedArgumentTriviaStart(functionCall, functionCallTriviaStart, argumentIndex)
	if !ok {
		return nil, analyzerUtilityArgumentPosition{}, false
	}
	return separated.Element(), analyzerUtilityArgumentPosition{
		SeparatedTriviaStart: separatedTriviaStart,
		ArgumentTriviaStart:  separatedTriviaStart,
	}, true
}

func analyzerUtilitySeparatedArgumentTriviaStart(functionCall *FunctionCallExpression, functionCallTriviaStart int, argumentIndex int) (int, bool) {
	if functionCall == nil || functionCall.ArgumentList() == nil || functionCall.ArgumentList().Expressions() == nil || argumentIndex < 0 {
		return 0, false
	}
	argumentList := functionCall.ArgumentList()
	expressions := argumentList.Expressions()
	if argumentIndex >= expressions.Count() {
		return 0, false
	}
	argumentListTriviaStart := syntaxChildTriviaStart(functionCall, functionCallTriviaStart, 1)
	expressionsTriviaStart := syntaxChildTriviaStart(argumentList, argumentListTriviaStart, 1)
	start := expressionsTriviaStart
	for i := 0; i < argumentIndex; i++ {
		if item := expressions.At(i); item != nil {
			start += item.ByteFullWidth()
		}
	}
	return start, true
}

func analyzerUtilityPreviousArgumentToken(functionCall *FunctionCallExpression, functionCallTriviaStart int, argumentIndex int) (*SyntaxToken, int) {
	if functionCall == nil || functionCall.ArgumentList() == nil {
		return nil, 0
	}
	argumentList := functionCall.ArgumentList()
	argumentListTriviaStart := syntaxChildTriviaStart(functionCall, functionCallTriviaStart, 1)
	if argumentIndex == 0 {
		return argumentList.OpenParen(), syntaxChildTriviaStart(argumentList, argumentListTriviaStart, 0)
	}
	expressions := argumentList.Expressions()
	if expressions == nil || argumentIndex-1 < 0 || argumentIndex-1 >= expressions.Count() {
		return nil, 0
	}
	previous := expressions.At(argumentIndex - 1)
	if previous == nil || previous.Separator() == nil {
		return nil, 0
	}
	previousStart, ok := analyzerUtilitySeparatedArgumentTriviaStart(functionCall, functionCallTriviaStart, argumentIndex-1)
	if !ok {
		return nil, 0
	}
	return previous.Separator(), previousStart + previous.Element().ByteFullWidth()
}

func analyzerUtilityNextArgumentToken(functionCall *FunctionCallExpression, functionCallTriviaStart int, argumentIndex int) (*SyntaxToken, int) {
	if functionCall == nil || functionCall.ArgumentList() == nil || functionCall.ArgumentList().Expressions() == nil {
		return nil, 0
	}
	argumentList := functionCall.ArgumentList()
	expressions := argumentList.Expressions()
	if argumentIndex < 0 || argumentIndex >= expressions.Count() {
		return nil, 0
	}
	current := expressions.At(argumentIndex)
	if current != nil && current.Separator() != nil {
		currentStart, ok := analyzerUtilitySeparatedArgumentTriviaStart(functionCall, functionCallTriviaStart, argumentIndex)
		if !ok {
			return nil, 0
		}
		if current.Element() == nil {
			return current.Separator(), currentStart
		}
		return current.Separator(), currentStart + current.Element().ByteFullWidth()
	}
	argumentListTriviaStart := syntaxChildTriviaStart(functionCall, functionCallTriviaStart, 1)
	return argumentList.CloseParen(), syntaxChildTriviaStart(argumentList, argumentListTriviaStart, 2)
}

func sameSyntaxExpression(left Expression, right Expression) bool {
	if left == nil || right == nil {
		return false
	}
	leftNode := syntaxNodePointer(left)
	rightNode := syntaxNodePointer(right)
	return leftNode != nil && leftNode == rightNode
}

func syntaxNodePointer(element SyntaxElement) *SyntaxNode {
	switch node := element.(type) {
	case nil:
		return nil
	case *SyntaxNode:
		return node
	}
	value := reflect.ValueOf(element)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return nil
	}
	structValue := value.Elem()
	if structValue.Kind() != reflect.Struct {
		return nil
	}
	field := structValue.FieldByName("SyntaxNode")
	if !field.IsValid() || field.IsNil() {
		return nil
	}
	node, _ := field.Interface().(*SyntaxNode)
	return node
}
