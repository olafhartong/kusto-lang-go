package kustolang

// Ported from KustoCompleter.GetCompletionHint, GetChildHint, GetChildIndex,
// HasEmptyChild, GetNearestAncestorWithEmptyChild and TryGetCompletionContext
// in src/Kusto.Language/Editor/Kusto/KustoCompleter.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type kustoCompletionContext struct {
	node       SyntaxElement
	childIndex int
}

func (c *kustoCompleter) getCompletionHint(position int) CompletionHint {
	context, ok := c.tryGetCompletionContext(position)
	if !ok {
		return CompletionHintNone
	}
	return c.getCompletionHintForContext(position, context.node, context.childIndex, CompletionHintQuery)
}

func (c *kustoCompleter) getCompletionHintForContext(position int, contextNode SyntaxElement, childIndex int, defaultHint CompletionHint) CompletionHint {
	if contextNode == nil || childIndex < 0 || childIndex >= contextNode.ChildCount() {
		return defaultHint
	}
	hint := c.getChildHint(contextNode, childIndex, defaultHint)
	if _, separated := WrapGeneratedSyntaxElement(contextNode).(*SeparatedElement); separated && childIndex == 1 {
		return hint
	}
	if completionChildIsMissing(contextNode, childIndex) {
		return hint
	}
	_, isList := contextNode.(*SyntaxList)
	if completionChildIsEmpty(contextNode, childIndex) || completionChildIsOnNewLine(contextNode, childIndex) || isList {
		for contextNode != nil && c.completionIsWithinLineConstraint(contextNode, position) {
			for index := childIndex + 1; index < contextNode.ChildCount(); index++ {
				if !completionChildIsMissingOrEmpty(contextNode, index) && !completionChildIsOnNewLine(contextNode, index) {
					return hint
				}
				hint |= c.getChildHint(contextNode, index, defaultHint)
				if completionChildIsMissing(contextNode, index) {
					return hint
				}
			}

			parent := contextNode.Parent()
			indexInParent := 0
			if parent != nil {
				indexInParent = contextNode.IndexInParent()
			}
			if parent != nil && completionHasLineConstraint(contextNode) && !completionHasLineConstraint(parent) {
				return hint
			}
			contextNode = parent
			childIndex = indexInParent
		}
	}
	return hint
}

func (c *kustoCompleter) getChildHint(contextNode SyntaxElement, childIndex int, defaultHint CompletionHint) CompletionHint {
	if hint, ok := c.completionParameterHint(contextNode, childIndex); ok {
		return hint
	}
	for contextNode != nil {
		hint := contextNode.CompletionHint(childIndex)
		if hint != CompletionHintInherit {
			return hint
		}
		parent := contextNode.Parent()
		if parent == nil {
			break
		}
		childIndex = contextNode.IndexInParent()
		contextNode = parent
	}
	return defaultHint
}

func (c *kustoCompleter) completionParameterHint(contextNode SyntaxElement, childIndex int) (CompletionHint, bool) {
	context, ok := c.functionOrOperatorArgumentCompletionContext(contextNode, childIndex)
	if !ok {
		return CompletionHintNone, false
	}
	hint := CompletionHintNone
	for _, candidate := range functionPossibleParameters(context, c.code.Globals) {
		if candidate.signature == nil || candidate.signature.IsHidden {
			continue
		}
		hint |= completionParameterHint(candidate.signature, candidate.parameter, map[*Parameter]struct{}{})
	}
	return hint, hint != CompletionHintNone
}

func (c *kustoCompleter) functionOrOperatorArgumentCompletionContext(contextNode SyntaxElement, childIndex int) (functionArgumentCompletionContext, bool) {
	if c == nil || c.code == nil || contextNode == nil {
		return functionArgumentCompletionContext{}, false
	}

	node := contextNode
	if named, ok := WrapGeneratedSyntaxElement(node).(NamedExpression); ok {
		parameterName := ""
		if simple, ok := named.(*SimpleNamedExpression); ok {
			parameterName = NameDeclarationSimpleName(simple.Name())
		}
		parent := node.Parent()
		if parent == nil {
			return functionArgumentCompletionContext{}, false
		}
		childIndex = node.IndexInParent()
		node = parent
		if context, ok := c.functionOrOperatorArgumentCompletionContext(node, childIndex); ok {
			context.parameterName = parameterName
			return context, true
		}
		return functionArgumentCompletionContext{}, false
	}

	for completionContainsOnlyOneToken(node) {
		parent := node.Parent()
		if parent == nil {
			break
		}
		if _, isList := parent.(*SyntaxList); isList {
			break
		}
		childIndex = node.IndexInParent()
		node = parent
	}

	if separated, ok := WrapGeneratedSyntaxElement(node).(*SeparatedElement); ok && separated != nil && childIndex == 0 {
		parent := node.Parent()
		if parent == nil {
			return functionArgumentCompletionContext{}, false
		}
		childIndex = node.IndexInParent()
		node = parent
	}

	if list, ok := node.(*SyntaxList); ok {
		expressionList, _ := WrapGeneratedSyntaxElement(list.Parent()).(*ExpressionList)
		if expressionList != nil {
			if call, ok := WrapGeneratedSyntaxElement(expressionList.Parent()).(*FunctionCallExpression); ok {
				return c.functionCallArgumentCompletionContext(call, childIndex)
			}
		}
	}
	if expressionList, ok := WrapGeneratedSyntaxElement(node).(*ExpressionList); ok {
		if call, ok := WrapGeneratedSyntaxElement(expressionList.Parent()).(*FunctionCallExpression); ok {
			return c.functionCallArgumentCompletionContext(call, 0)
		}
	}
	if binary, ok := WrapGeneratedSyntaxElement(node).(*BinaryExpression); ok {
		operator, _ := binary.ReferencedSymbol().(*OperatorSymbol)
		if operator == nil {
			operator = completionOperatorForSyntaxKind(binary.Kind())
		}
		if operator != nil && len(operator.Signatures()) > 0 {
			argumentIndex := 1
			if childIndex == 0 {
				argumentIndex = 0
			}
			return functionArgumentCompletionContext{
				operator:      operator,
				signatures:    operator.Signatures(),
				arguments:     []Expression{binary.Left(), binary.Right()},
				argumentIndex: argumentIndex,
			}, true
		}
	}
	return functionArgumentCompletionContext{}, false
}

func (c *kustoCompleter) functionCallArgumentCompletionContext(call *FunctionCallExpression, argumentIndex int) (functionArgumentCompletionContext, bool) {
	if call == nil {
		return functionArgumentCompletionContext{}, false
	}
	function, _ := call.ReferencedSymbol().(*FunctionSymbol)
	if function == nil && call.Name() != nil {
		function, _ = call.Name().ReferencedSymbol().(*FunctionSymbol)
	}
	if function == nil {
		function = c.resolveCompletionFunction(call)
	}
	if function == nil || len(function.Signatures()) == 0 {
		return functionArgumentCompletionContext{}, false
	}
	arguments := functionCallCompletionArguments(call)
	if completionIsInvokeFunctionContext(call.Parent()) {
		arguments = append([]Expression{nil}, arguments...)
		argumentIndex++
	}
	return functionArgumentCompletionContext{
		call:          call,
		function:      function,
		signatures:    function.Signatures(),
		arguments:     arguments,
		argumentIndex: argumentIndex,
	}, true
}

func completionContainsOnlyOneToken(node SyntaxElement) bool {
	for node != nil && node.ChildCount() == 1 {
		child := node.GetChild(0)
		if _, ok := child.(*SyntaxToken); ok {
			return true
		}
		node = child
	}
	return false
}

func completionIsInvokeFunctionContext(node SyntaxElement) bool {
	var operator QueryOperator
	var functionCall *FunctionCallExpression
	for current := node; current != nil && (operator == nil || functionCall == nil); current = current.Parent() {
		wrapped := WrapGeneratedSyntaxElement(current)
		if operator == nil {
			operator, _ = wrapped.(QueryOperator)
		}
		if functionCall == nil {
			functionCall, _ = wrapped.(*FunctionCallExpression)
		}
	}
	invoke, ok := operator.(*InvokeOperator)
	return ok && invoke != nil && (functionCall == nil || functionCall.TextStart() < invoke.TextStart())
}

func completionChildIsMissingOrEmpty(node SyntaxElement, index int) bool {
	return completionChildIsMissing(node, index) || completionChildIsEmpty(node, index)
}

func completionChildIsMissing(node SyntaxElement, index int) bool {
	if node == nil || index < 0 || index >= node.ChildCount() {
		return false
	}
	return completionElementIsMissing(node.GetChild(index))
}

func completionElementIsMissing(element SyntaxElement) bool {
	if element == nil || element.Width() != 0 {
		return false
	}
	if token, ok := element.(*SyntaxToken); ok {
		return token.Kind() != SyntaxKindEndOfTextToken
	}
	return SyntaxIsMissing(element)
}

func completionChildIsEmpty(node SyntaxElement, index int) bool {
	if node == nil || index < 0 || index >= node.ChildCount() {
		return false
	}
	child := node.GetChild(index)
	return (child == nil || child.Width() == 0) && !completionElementIsMissing(child)
}

func completionChildIsOnNewLine(node SyntaxElement, index int) bool {
	if node == nil || index < 0 || index >= node.ChildCount() {
		return false
	}
	child := node.GetChild(index)
	if child == nil {
		return false
	}
	token := child.GetFirstToken()
	if token == nil {
		return false
	}
	lineBreak := NextLineBreakStart(token.Trivia(), 0)
	return lineBreak >= 0 && IsWhitespaceOnly(token.Trivia(), 0, lineBreak)
}

func completionLineConstraint(element SyntaxElement) *Directive {
	return GetFirstAncestorOrSelf[*Directive](element)
}

func completionHasLineConstraint(element SyntaxElement) bool {
	return completionLineConstraint(element) != nil
}

func (c *kustoCompleter) completionIsWithinLineConstraint(element SyntaxElement, position int) bool {
	constraint := completionLineConstraint(element)
	if constraint == nil {
		return true
	}
	first, last := constraint.GetFirstToken(), constraint.GetLastToken()
	if first == nil || last == nil || c == nil || c.code == nil {
		return false
	}
	firstLine, _, firstOK := c.code.TryGetLineAndOffset(first.TextStart())
	lastPosition := last.End() - 1
	if lastPosition < last.TextStart() {
		lastPosition = last.TextStart()
	}
	lastLine, _, lastOK := c.code.TryGetLineAndOffset(lastPosition)
	positionLine, _, positionOK := c.code.TryGetLineAndOffset(position)
	return firstOK && lastOK && positionOK && positionLine >= firstLine && positionLine <= lastLine
}

func completionChildIndex(node SyntaxElement, position int) int {
	if node == nil {
		return -1
	}
	firstChildBeyond := node.ChildCount()
	lastWasNullOrEmpty := false
findBeyond:
	for index := 0; index < node.ChildCount(); index++ {
		child := node.GetChild(index)
		if child != nil {
			switch {
			case position > child.TextStart() && position < child.End() && child.Width() > 0:
				return index
			case position < child.TextStart():
				firstChildBeyond = index
				break findBeyond
			case position == child.TextStart():
				if child.Width() == 0 {
					firstChildBeyond = index + 1
					break findBeyond
				}
				if lastWasNullOrEmpty {
					firstChildBeyond = index
					break findBeyond
				}
			default:
				lastWasNullOrEmpty = child.Width() == 0
			}
		} else {
			lastWasNullOrEmpty = true
		}
	}

	bestEmptyChild := firstChildBeyond - 1
	for index := firstChildBeyond - 1; index >= 0; index-- {
		child := node.GetChild(index)
		if child != nil && child.Width() > 0 {
			break
		}
		bestEmptyChild = index
	}
	return bestEmptyChild
}

func completionHasEmptyChild(element SyntaxElement, position int) bool {
	if element == nil {
		return false
	}
	if position >= element.End() {
		if _, list := element.(*SyntaxList); list {
			return true
		}
	}
	hasEmptyChild := false
	for index := 0; index < element.ChildCount(); index++ {
		child := element.GetChild(index)
		if child != nil {
			if position <= child.TextStart() && hasEmptyChild {
				return true
			}
			if position < child.End() {
				return false
			}
		}
		hasEmptyChild = child == nil || child.Width() == 0
	}
	return hasEmptyChild
}

func completionNearestAncestorWithEmptyChild(token *SyntaxToken, position int) SyntaxElement {
	if token == nil {
		return nil
	}
	node := token.Parent()
	for node != nil && !completionHasEmptyChild(node, position) {
		node = node.Parent()
	}
	return node
}

func (c *kustoCompleter) tryGetCompletionContext(position int) (kustoCompletionContext, bool) {
	if c == nil || c.code == nil || c.code.Syntax() == nil {
		return kustoCompletionContext{}, false
	}
	token := c.code.Syntax().GetTokenAt(position)
	if token == nil {
		return kustoCompletionContext{}, false
	}
	hasAffinity := HasAffinity(token, position)
	if hasAffinity {
		if previous := token.GetPreviousToken(true); previous != nil && completionElementIsMissing(previous) {
			return kustoCompletionContext{}, false
		}
		if c.completionIsWithinLineConstraint(token, position) && token.Parent() != nil {
			return kustoCompletionContext{node: token.Parent(), childIndex: completionChildIndex(token.Parent(), position)}, true
		}
	}

	tokenNode := completionNearestAncestorWithEmptyChild(token, position)
	if (position <= token.TextStart() && !hasAffinity) || token.Kind() == SyntaxKindEndOfTextToken {
		if previous := token.GetPreviousToken(); previous != nil {
			previousNode := completionNearestAncestorWithEmptyChild(previous, position)
			if previousNode != nil && (tokenNode == nil || SyntaxDepth(previousNode) >= SyntaxDepth(tokenNode)) &&
				c.completionIsWithinLineConstraint(previousNode, position) {
				return kustoCompletionContext{node: previousNode, childIndex: completionChildIndex(previousNode, position)}, true
			}
		}
	} else if position >= token.End() && !hasAffinity {
		if next := token.GetNextToken(); next != nil {
			nextNode := completionNearestAncestorWithEmptyChild(next, position)
			if nextNode != nil && (tokenNode == nil || SyntaxDepth(nextNode) > SyntaxDepth(tokenNode)) &&
				c.completionIsWithinLineConstraint(nextNode, position) {
				return kustoCompletionContext{node: nextNode, childIndex: completionChildIndex(nextNode, position)}, true
			}
		}
	}

	if tokenNode != nil && c.completionIsWithinLineConstraint(tokenNode, position) {
		return kustoCompletionContext{node: tokenNode, childIndex: completionChildIndex(tokenNode, position)}, true
	}

	if position > token.TriviaStart() && position < token.TextStart() && !hasAffinity {
		offset := position - token.TriviaStart()
		lineBreak := NextLineBreakStart(token.Trivia(), offset)
		if lineBreak >= 0 && IsWhitespaceOnly(token.Trivia(), 0, lineBreak) {
			if previous := token.GetPreviousToken(); previous != nil {
				common := completionCommonAncestor(previous, token)
				if common != nil {
					index := SyntaxGetDescendantIndex(common, token)
					if index >= 0 {
						return kustoCompletionContext{node: common, childIndex: index}, true
					}
				}
			}
		}
	}
	return kustoCompletionContext{}, false
}

func completionCommonAncestor(first, second SyntaxElement) SyntaxElement {
	seen := map[SyntaxElement]struct{}{}
	for current := first; current != nil; current = current.Parent() {
		seen[current] = struct{}{}
	}
	for current := second; current != nil; current = current.Parent() {
		if _, ok := seen[current]; ok {
			return current
		}
	}
	return nil
}
