package kustolang

func optionalBool(values []bool) bool {
	return len(values) > 0 && values[0]
}

func getSyntaxSibling(info *syntaxOwnership, direction int, includeZeroWidth bool) SyntaxElement {
	if info == nil || info.parent == nil {
		return nil
	}
	for index := info.indexInParent + direction; index >= 0 && index < info.parent.ChildCount(); index += direction {
		sibling := info.parent.GetChild(index)
		if sibling != nil && (includeZeroWidth || sibling.FullWidth() > 0) {
			return sibling
		}
	}
	return nil
}

func navigationFirstToken(root SyntaxElement, includeZeroWidth bool) *SyntaxToken {
	if root == nil {
		return nil
	}
	if token, ok := root.(*SyntaxToken); ok {
		if includeZeroWidth || token.FullWidth() > 0 {
			return token
		}
		return nil
	}
	for index := 0; index < root.ChildCount(); index++ {
		if token := navigationFirstToken(root.GetChild(index), includeZeroWidth); token != nil {
			return token
		}
	}
	return nil
}

func navigationLastToken(root SyntaxElement, includeZeroWidth bool) *SyntaxToken {
	if root == nil {
		return nil
	}
	if token, ok := root.(*SyntaxToken); ok {
		if includeZeroWidth || token.FullWidth() > 0 {
			return token
		}
		return nil
	}
	for index := root.ChildCount() - 1; index >= 0; index-- {
		if token := navigationLastToken(root.GetChild(index), includeZeroWidth); token != nil {
			return token
		}
	}
	return nil
}

func nextSyntaxToken(token *SyntaxToken, includeZeroWidth bool) *SyntaxToken {
	if token == nil {
		return nil
	}
	for element := SyntaxElement(token); element.Parent() != nil; element = element.Parent() {
		parent := element.Parent()
		for index := element.IndexInParent() + 1; index < parent.ChildCount(); index++ {
			if next := navigationFirstToken(parent.GetChild(index), includeZeroWidth); next != nil {
				return next
			}
		}
	}
	return nil
}

func previousSyntaxToken(token *SyntaxToken, includeZeroWidth bool) *SyntaxToken {
	if token == nil {
		return nil
	}
	for element := SyntaxElement(token); element.Parent() != nil; element = element.Parent() {
		parent := element.Parent()
		for index := element.IndexInParent() - 1; index >= 0; index-- {
			if previous := navigationLastToken(parent.GetChild(index), includeZeroWidth); previous != nil {
				return previous
			}
		}
	}
	return nil
}

func (t *SyntaxToken) GetNextToken(includeZeroWidthTokens ...bool) *SyntaxToken {
	return nextSyntaxToken(t, optionalBool(includeZeroWidthTokens))
}

func (t *SyntaxToken) GetPreviousToken(includeZeroWidthTokens ...bool) *SyntaxToken {
	return previousSyntaxToken(t, optionalBool(includeZeroWidthTokens))
}

func syntaxTokenAt(element SyntaxElement, position int) *SyntaxToken {
	if element == nil {
		return nil
	}
	if token, ok := element.(*SyntaxToken); ok {
		if position >= token.TriviaStart() && position < token.End() {
			return token
		}
		if position >= token.FullWidth() {
			return token
		}
		return nil
	}
	root := element.Root()
	if root == nil {
		return nil
	}
	if position >= root.End() {
		return navigationLastToken(root, true)
	}
	for current := root; current != nil; {
		found := false
		for index := 0; index < current.ChildCount(); index++ {
			child := current.GetChild(index)
			if child == nil || position < child.TriviaStart() || position >= child.End() {
				continue
			}
			if token, ok := child.(*SyntaxToken); ok {
				return token
			}
			current = child
			found = true
			break
		}
		if !found {
			return nil
		}
	}
	return nil
}

func syntaxTokenAtUTF16(element SyntaxElement, position int) *SyntaxToken {
	if element == nil {
		return nil
	}
	if token, ok := element.(*SyntaxToken); ok {
		if position >= token.UTF16TriviaStart() && position < token.UTF16End() {
			return token
		}
		if position >= token.UTF16FullWidth() {
			return token
		}
		return nil
	}
	root := element.Root()
	if root == nil {
		return nil
	}
	if position >= root.UTF16End() {
		return navigationLastToken(root, true)
	}
	for current := root; current != nil; {
		found := false
		for index := 0; index < current.ChildCount(); index++ {
			child := current.GetChild(index)
			if child == nil || position < child.UTF16TriviaStart() || position >= child.UTF16End() {
				continue
			}
			if token, ok := child.(*SyntaxToken); ok {
				return token
			}
			current = child
			found = true
			break
		}
		if !found {
			return nil
		}
	}
	return nil
}

func syntaxNodeAt(element SyntaxElement, position, length int) *SyntaxNode {
	token := syntaxTokenAt(element, position)
	if token == nil {
		return nil
	}
	for parent := token.Parent(); parent != nil; parent = parent.Parent() {
		node := syntaxNodeForBinderLocation(parent)
		if node != nil && parent.End() >= position+length {
			return node
		}
	}
	return nil
}

func syntaxNodeAtUTF16(element SyntaxElement, position, length int) *SyntaxNode {
	token := syntaxTokenAtUTF16(element, position)
	if token == nil {
		return nil
	}
	for parent := token.Parent(); parent != nil; parent = parent.Parent() {
		node := syntaxNodeForBinderLocation(parent)
		if node != nil && parent.UTF16End() >= position+length {
			return node
		}
	}
	return nil
}
