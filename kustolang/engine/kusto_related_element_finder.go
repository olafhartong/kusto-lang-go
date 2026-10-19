package kustolang

import "unicode/utf8"

// Ported from src/Kusto.Language/Editor/Kusto/KustoRelatedElementFinder.cs.
// Structural pairs and identity-based names with attached binder semantics are
// supported. Unsupported/unbound subtrees intentionally produce no semantic
// related elements.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func findKustoRelatedElements(code *KustoCode, utf16Position int, options FindRelatedOptions, cancellationToken CancellationToken) RelatedInfo {
	if code == nil || code.Syntax() == nil {
		return RelatedInfoEmpty
	}
	position := utf16PositionToByteOffset(code.Text, utf16Position)
	refs := kustoServiceTokenRefs(code.Syntax())
	index := relatedTokenIndex(refs, position)
	if index < 0 {
		return RelatedInfoEmpty
	}
	cancellationToken.ThrowIfCancellationRequested()
	ref := refs[index]
	if semantic := relatedSemanticNames(code, ref, options, cancellationToken); len(semantic.Elements) > 0 {
		return semantic
	}
	openKind, closeKind, ok := relatedBracketKinds(ref.token.Kind())
	if !ok || relatedTokenIsBracketedName(ref, openKind) {
		return RelatedInfoEmpty
	}

	for i := len(ref.ancestors) - 1; i >= 0; i-- {
		ancestor := ref.ancestors[i]
		open, close, found := relatedDirectBracketPair(ancestor, ref.token, openKind, closeKind)
		if !found || open.token.Width() == 0 || close.token.Width() == 0 {
			continue
		}
		openElement := relatedElementForToken(code.Text, open)
		closeElement := relatedElementForToken(code.Text, close)
		current := 0
		if ref.token == close.token {
			current = 1
		}
		return NewRelatedInfo([]RelatedElement{openElement, closeElement}, current)
	}
	return RelatedInfoEmpty
}

func relatedTokenIndex(refs []kustoServiceTokenRef, position int) int {
	if len(refs) == 0 {
		return -1
	}
	index := -1
	for i, ref := range refs {
		if ref.token == nil {
			continue
		}
		positioned := ref.positionedToken()
		if ref.triviaStart <= position && position < positioned.End() {
			index = i
			break
		}
		if position == positioned.End() {
			index = i
		}
		if positioned.TextStart() > position {
			break
		}
	}
	if index < 0 {
		return -1
	}
	current := refs[index].positionedToken()
	if position > current.TriviaStart() && position < current.TextStart()-1 {
		return -1
	}
	if position == current.TriviaStart() && index > 0 {
		previousRelatable := relatedTokenRefIsRelatable(refs[index-1])
		currentRelatable := relatedTokenRefIsRelatable(refs[index])
		if current.TriviaWidth() > 0 || previousRelatable && !currentRelatable {
			index--
		}
	}
	return index
}

func relatedTokenRefIsRelatable(ref kustoServiceTokenRef) bool {
	if ref.token == nil {
		return false
	}
	if _, _, ok := relatedBracketKinds(ref.token.Kind()); ok {
		return true
	}
	_, _, ok := relatedBoundName(ref)
	return ok
}

func relatedBoundName(ref kustoServiceTokenRef) (SyntaxElement, Symbol, bool) {
	for i := len(ref.ancestors) - 1; i >= 0; i-- {
		wrapped := WrapGeneratedSyntaxElement(ref.ancestors[i].element)
		switch name := wrapped.(type) {
		case *NameReference:
			symbol := name.ReferencedSymbol()
			if symbol != nil {
				if _, unresolved := symbol.(*ErrorSymbol); !unresolved {
					return name.Name(), symbol, true
				}
			}
		case *NameDeclaration:
			symbol := name.ReferencedSymbol()
			if symbol != nil {
				if _, unresolved := symbol.(*ErrorSymbol); !unresolved {
					return name.Name(), symbol, true
				}
			}
		}
	}
	return nil, nil, false
}

func relatedSemanticNames(code *KustoCode, selected kustoServiceTokenRef, options FindRelatedOptions, cancellationToken CancellationToken) RelatedInfo {
	selectedName, selectedSymbol, ok := relatedBoundName(selected)
	if !ok || selectedName == nil {
		return RelatedInfoEmpty
	}
	selectedStart := -1
	if location := relatedAncestorLocation(selected, selectedName); location != nil {
		selectedStart = elementTextStart(location.element, location.triviaStart)
	}
	if selectedStart < 0 {
		return RelatedInfoEmpty
	}
	seeThrough := options&FindRelatedOptionsSeeThroughVariables != 0
	elements := make([]RelatedElement, 0, 4)
	current := -1
	walkSyntaxWithTriviaStartAndAncestors(code.Syntax(), 0, func(element SyntaxElement, triviaStart int, _ []SyntaxElement) {
		cancellationToken.ThrowIfCancellationRequested()
		wrapped := WrapGeneratedSyntaxElement(element)
		var name SyntaxElement
		var symbol Symbol
		kind := RelatedElementKindReference
		matches := false
		switch node := wrapped.(type) {
		case *NameReference:
			name = node.Name()
			symbol = node.ReferencedSymbol()
			matches = symbol == selectedSymbol || seeThrough && symbol == GetResultType(selectedSymbol)
		case *NameDeclaration:
			name = node.Name()
			symbol = node.ReferencedSymbol()
			kind = RelatedElementKindDeclaration
			matches = symbol == selectedSymbol || seeThrough && GetResultType(symbol) == selectedSymbol
		}
		if !matches || name == nil || symbol == nil {
			return
		}
		location := syntaxLocationWithin(element, triviaStart, name)
		if location == nil {
			return
		}
		byteStart := elementTextStart(name, location.triviaStart)
		byteEnd := elementEnd(name, location.triviaStart)
		start := byteOffsetToUTF16Position(code.Text, byteStart)
		end := byteOffsetToUTF16Position(code.Text, byteEnd)
		if byteStart == selectedStart {
			current = len(elements)
		}
		elements = append(elements, NewRelatedElementWithCursors(start, end-start, kind, start, start))
	})
	if len(elements) <= 1 || current < 0 {
		return RelatedInfoEmpty
	}
	return NewRelatedInfo(elements, current)
}

func relatedAncestorLocation(ref kustoServiceTokenRef, target SyntaxElement) *kustoReferenceLocation {
	targetNode := simpleBinderSyntaxNode(target)
	for _, ancestor := range ref.ancestors {
		if ancestor.element == target || targetNode != nil && ancestor.element == targetNode {
			return &kustoReferenceLocation{element: target, triviaStart: ancestor.triviaStart}
		}
	}
	return nil
}

func relatedBracketTokenIndex(refs []kustoServiceTokenRef, position int) int {
	ending := -1
	for i, ref := range refs {
		if ref.token == nil {
			continue
		}
		if _, _, ok := relatedBracketKinds(ref.token.Kind()); !ok {
			continue
		}
		token := ref.positionedToken()
		if position == token.TextStart() || position > token.TextStart() && position < token.End() {
			return i
		}
		if position == token.End() {
			ending = i
		}
		if token.TextStart() > position {
			break
		}
	}
	return ending
}

func relatedBracketKinds(kind SyntaxKind) (SyntaxKind, SyntaxKind, bool) {
	switch kind {
	case SyntaxKindOpenParenToken, SyntaxKindCloseParenToken:
		return SyntaxKindOpenParenToken, SyntaxKindCloseParenToken, true
	case SyntaxKindOpenBraceToken, SyntaxKindCloseBraceToken:
		return SyntaxKindOpenBraceToken, SyntaxKindCloseBraceToken, true
	case SyntaxKindOpenBracketToken, SyntaxKindCloseBracketToken:
		return SyntaxKindOpenBracketToken, SyntaxKindCloseBracketToken, true
	default:
		return SyntaxKindNone, SyntaxKindNone, false
	}
}

func relatedTokenIsBracketedName(ref kustoServiceTokenRef, openKind SyntaxKind) bool {
	if openKind != SyntaxKindOpenBracketToken {
		return false
	}
	for _, ancestor := range ref.ancestors {
		switch WrapGeneratedSyntaxElement(ancestor.element).(type) {
		case *BracketedName, *BracketedWildcardedName:
			return true
		}
	}
	return false
}

func relatedDirectBracketPair(ancestor kustoServiceElementRef, selected *SyntaxToken, openKind, closeKind SyntaxKind) (kustoServiceTokenRef, kustoServiceTokenRef, bool) {
	if ancestor.element == nil {
		return kustoServiceTokenRef{}, kustoServiceTokenRef{}, false
	}
	childStart := ancestor.triviaStart
	var open, close kustoServiceTokenRef
	selectedIsChild := false
	for _, child := range ancestor.element.Children() {
		if child == nil {
			continue
		}
		if token, ok := child.(*SyntaxToken); ok {
			ref := kustoServiceTokenRef{token: token, triviaStart: childStart}
			if token == selected {
				selectedIsChild = true
			}
			if token.Kind() == openKind && open.token == nil {
				open = ref
			}
			if token.Kind() == closeKind {
				close = ref
			}
		}
		childStart += child.FullWidth()
	}
	return open, close, selectedIsChild && open.token != nil && close.token != nil
}

func relatedElementForToken(text string, ref kustoServiceTokenRef) RelatedElement {
	token := ref.positionedToken()
	start := byteOffsetToUTF16Position(text, token.TextStart())
	end := byteOffsetToUTF16Position(text, token.End())
	return NewRelatedElementWithCursors(start, end-start, RelatedElementKindSyntax, start, end)
}

func utf16PositionToByteOffset(text string, position int) int {
	if position <= 0 {
		return 0
	}
	units := 0
	for offset, r := range text {
		if units >= position {
			return offset
		}
		width := 1
		if r > 0xffff {
			width = 2
		}
		if units+width > position {
			return offset
		}
		units += width
	}
	return len(text)
}

func byteOffsetToUTF16Position(text string, offset int) int {
	if offset <= 0 {
		return 0
	}
	if offset > len(text) {
		offset = len(text)
	}
	units := 0
	for index := 0; index < offset; {
		r, size := utf8.DecodeRuneInString(text[index:])
		if size == 0 || index+size > offset {
			break
		}
		if r > 0xffff {
			units += 2
		} else {
			units++
		}
		index += size
	}
	return units
}
