package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/Actors/ActorUtilities.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func actorTokenNear(code *KustoCode, position int) *SyntaxToken {
	return extractTokenNear(code, position)
}

func actorAdjustRangeToNode(code *KustoCode, start, length int) (int, int) {
	if code == nil || code.Syntax() == nil || length < 0 {
		return start, length
	}
	node := code.Syntax().GetNodeAt(start, length)
	if node != nil && node.End() > start+length && length > 0 {
		last := code.Syntax().GetTokenAt(start + length - 1)
		if last == nil || last.GetPreviousToken() == nil {
			return start, length
		}
		length = last.GetPreviousToken().End() - start
		node = code.Syntax().GetNodeAt(start, length)
		if node != nil && node.End() > start+length {
			return start, length
		}
	}
	if node != nil {
		return node.TextStart(), node.End() - node.TextStart()
	}
	return start, length
}

func actorGetFirstReference(root SyntaxElement, symbol Symbol) *NameReference {
	return GetFirstDescendant[*NameReference](root, func(reference *NameReference) bool {
		return reference.ReferencedSymbol() == symbol
	})
}

func actorIsTopLevelStatement(element SyntaxElement) bool {
	if element == nil || element.Parent() == nil {
		return false
	}
	separated, ok := element.Parent().(*SeparatedElement)
	if !ok || separated.Parent() == nil {
		return false
	}
	list, ok := separated.Parent().(*SyntaxList)
	if !ok || list.Parent() == nil {
		return false
	}
	return list.Parent().Kind() == SyntaxKindQueryBlock
}

func actorNearestTopLevelStatementInsertionPosition(code *KustoCode, position int) (int, bool) {
	if code == nil || code.Syntax() == nil {
		return 0, false
	}
	token := code.Syntax().GetTokenAt(position)
	if token == nil {
		return 0, false
	}
	for element := token.Parent(); element != nil; element = element.Parent() {
		if actorIsTopLevelStatement(element) {
			return extractInsertionPositionBefore(element), true
		}
	}
	return 0, false
}

func actorIsEntityFunction(function *FunctionSymbol) bool {
	return function != nil && (function == Functions.Table ||
		function == Functions.ExternalTable ||
		function == Functions.MaterializedView ||
		function == Functions.EntityGroup ||
		function == Functions.StoredQueryResult)
}

func actorIsClusterPart(part Expression) bool {
	call, ok := part.(*FunctionCallExpression)
	return ok && call.ReferencedSymbol() == Functions.Cluster
}

func actorIsDatabasePart(part Expression) bool {
	call, ok := part.(*FunctionCallExpression)
	return ok && call.ReferencedSymbol() == Functions.Database
}

func actorIsEntityGroupElementPart(part Expression) bool {
	reference, ok := part.(*NameReference)
	if !ok {
		return false
	}
	_, ok = reference.ReferencedSymbol().(*EntityGroupElementSymbol)
	return ok
}

func actorIsEntityPart(part Expression, globals *GlobalState) bool {
	if part == nil {
		return false
	}
	if call, ok := part.(*FunctionCallExpression); ok {
		function, _ := call.ReferencedSymbol().(*FunctionSymbol)
		if actorIsEntityFunction(function) {
			return true
		}
	}
	if globals != nil {
		if reference, ok := part.(*NameReference); ok {
			if symbol := reference.ReferencedSymbol(); symbol != nil && globals.GetDatabase(symbol) != nil {
				return true
			}
		}
	}
	if path, ok := WrapGeneratedSyntaxElement(part.Parent()).(*PathExpression); ok && extractSameElement(path.Selector(), part) {
		return actorIsDatabasePart(path.Expression())
	}
	return false
}

func actorGetEntityExpressionPart(element SyntaxElement, globals *GlobalState) Expression {
	if element == nil {
		return nil
	}
	if token, ok := element.(*SyntaxToken); ok {
		element = token.Parent()
	}
	if parent, ok := syntaxElementAs[Name](element.Parent()); ok {
		element = parent
	}
	if parent, ok := syntaxElementAs[*NameReference](element.Parent()); ok {
		element = parent
	}
	if reference, ok := syntaxElementAs[*NameReference](element); ok {
		if call, ok := syntaxElementAs[*FunctionCallExpression](reference.Parent()); ok {
			element = call
		}
	}
	expression, ok := syntaxElementAs[Expression](element)
	if !ok {
		return nil
	}
	if actorIsEntityPart(expression, globals) || actorIsDatabasePart(expression) || actorIsClusterPart(expression) || actorIsEntityGroupElementPart(expression) {
		return expression
	}
	return nil
}

func actorGetEntityExpressionStart(element SyntaxElement, globals *GlobalState) Expression {
	if element == nil {
		return nil
	}
	token, ok := element.(*SyntaxToken)
	if !ok {
		token = element.GetLastToken()
	}
	expression := actorGetEntityExpressionPart(token, globals)
	if expression == nil {
		return nil
	}
	end := expression.End()
	for {
		path, ok := syntaxElementAs[*PathExpression](expression.Parent())
		if !ok || path.End() > end {
			break
		}
		expression = path
	}
	return expression
}

func actorGetEntityExpression(element SyntaxElement, globals *GlobalState) Expression {
	if element == nil {
		return nil
	}
	for {
		path, ok := syntaxElementAs[*PathExpression](element)
		if !ok {
			break
		}
		element = path.Selector()
	}
	expression := actorGetEntityExpressionPart(element, globals)
	if expression == nil {
		return nil
	}
	for {
		path, ok := syntaxElementAs[*PathExpression](expression.Parent())
		if !ok {
			break
		}
		expression = path
	}
	return expression
}

func actorIsDatabaseQualifiedName(element SyntaxElement) bool {
	if element == nil {
		return false
	}
	if call, ok := syntaxElementAs[*FunctionCallExpression](element.Parent()); ok {
		element = call
	}
	path, ok := syntaxElementAs[*PathExpression](element.Parent())
	if !ok {
		return false
	}
	if actorIsDatabasePart(path.Expression()) {
		return true
	}
	leftPath, ok := path.Expression().(*PathExpression)
	return ok && actorIsDatabasePart(leftPath.Selector())
}

func actorQueryWithDatabaseQualifiersRemoved(query *SyntaxNode, qualifiedSymbol Symbol) EditString {
	if query == nil || qualifiedSymbol == nil {
		return EditStringEmpty
	}
	text := NewEditString(query.Text(IncludeTriviaAll))
	references := GetDescendants[*NameReference](query, func(reference *NameReference) bool {
		return actorIsDatabaseQualifiedName(reference) && reference.ReferencedSymbol() == qualifiedSymbol
	})
	for index := len(references) - 1; index >= 0; index-- {
		reference := references[index]
		element := SyntaxElement(reference)
		if call, ok := syntaxElementAs[*FunctionCallExpression](reference.Parent()); ok {
			element = call
		}
		if actorIsDatabaseQualifiedName(element) && element.Parent() != nil {
			qualifier := element.Parent()
			length := reference.TextStart() - qualifier.TextStart()
			if length > 0 {
				text = text.Remove(qualifier.TextStart(), length)
			}
		}
	}
	return text
}

func actorInsertBrackets(text EditString, openPosition int, prefixText, openBracketText string, closePosition int, closeBracketText string, options FormattingOptions) EditString {
	if openPosition < 0 || closePosition < openPosition || closePosition > len(text.CurrentText) {
		return text
	}
	lineStart := LineStart(text.CurrentText, openPosition)
	indentationLength := ScanWhitespace(text.CurrentText, lineStart)
	indentation := text.CurrentText[lineStart : lineStart+indentationLength]
	var openText, closeText string
	if options.BrackettingStyle == BrackettingStyleDiagonal {
		openText = prefixText + " " + openBracketText + "\n"
	} else {
		openText = prefixText + "\n" + indentation + openBracketText + "\n"
	}
	closeText = "\n" + indentation + closeBracketText + "\n"
	withBrackets := text.ApplyAll([]TextEdit{
		NewTextInsertion(openPosition, openText),
		NewTextInsertion(closePosition, closeText),
	})
	rangeStart := withBrackets.GetCurrentPosition(openPosition, PositionBiasRight)
	rangeEnd := withBrackets.GetCurrentPosition(closePosition, PositionBiasLeft)
	return actorIndentText(withBrackets, rangeStart, rangeEnd-rangeStart, options.IndentationSize, true)
}

func actorIndentText(text EditString, start, length, delta int, includeFirstLine bool) EditString {
	if length <= 0 || delta == 0 {
		return text
	}
	edits := []TextEdit{}
	lineStart := start
	if !includeFirstLine {
		lineStart += LineLength(text.CurrentText, lineStart, true)
	}
	end := start + length
	for lineStart < end && lineStart <= len(text.CurrentText) {
		if delta > 0 {
			edits = append(edits, NewTextInsertion(lineStart, spaces(delta)))
		} else {
			remove := ScanWhitespace(text.CurrentText, lineStart)
			if remove > -delta {
				remove = -delta
			}
			if remove > 0 {
				edits = append(edits, NewTextDeletion(lineStart, remove))
			}
		}
		lineLength := LineLength(text.CurrentText, lineStart, true)
		if lineLength <= 0 {
			break
		}
		lineStart += lineLength
	}
	return text.ApplyAll(edits)
}

func spaces(count int) string {
	if count <= 0 {
		return ""
	}
	buffer := make([]byte, count)
	for index := range buffer {
		buffer[index] = ' '
	}
	return string(buffer)
}
