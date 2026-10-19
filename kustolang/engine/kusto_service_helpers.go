package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/KustoServiceHelpers.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type kustoServiceElementRef struct {
	element     SyntaxElement
	triviaStart int
}

type kustoServiceTokenRef struct {
	token       *SyntaxToken
	triviaStart int
	ancestors   []kustoServiceElementRef
}

func GetTokenWithAffinity(code *KustoCode, position int) *SyntaxToken {
	refs, index, ok := kustoServiceTokenRefAtPosition(code, position)
	if !ok {
		return nil
	}

	token := refs[index].positionedToken()
	if HasAffinity(token, position) {
		return token
	}

	if index > 0 {
		previous := refs[index-1].positionedToken()
		if HasAffinity(previous, position) {
			return previous
		}
	}

	return nil
}

func HasAffinity(token *SyntaxToken, position int) bool {
	if token == nil {
		return false
	}
	return (position > token.TextStart() && position < token.End()) ||
		(position == token.TextStart() && IsNameLikeToken(token.Kind())) ||
		(position == token.End() && IsNameLikeToken(token.Kind()))
}

func IsNameLikeToken(kind SyntaxKind) bool {
	switch SyntaxCategoryOf(kind) {
	case SyntaxCategoryIdentifier, SyntaxCategoryKeyword:
		return true
	default:
		return kind == SyntaxKindBangToken
	}
}

func GetTokenLeftOfPosition(code *KustoCode, position int) *SyntaxToken {
	refs, index, ok := kustoServiceTokenLeftRef(code, position)
	if !ok {
		return nil
	}
	return refs[index].positionedToken()
}

func GetCompleteExpressionLeftOfPosition(code *KustoCode, position int) Expression {
	refs, index, ok := kustoServiceTokenLeftRef(code, position)
	if !ok {
		return nil
	}

	token := refs[index].positionedToken()
	tokenEnd := token.End()
	for i := len(refs[index].ancestors) - 1; i >= 0; i-- {
		ancestor := refs[index].ancestors[i]
		wrapped := WrapGeneratedSyntaxElement(ancestor.element)
		expr, ok := wrapped.(Expression)
		if !ok {
			continue
		}
		if elementEnd(wrapped, ancestor.triviaStart) == tokenEnd && !syntaxElementHasMissingChildren(wrapped) {
			return expr
		}
		return nil
	}

	return nil
}

func kustoServiceTokenLeftRef(code *KustoCode, position int) ([]kustoServiceTokenRef, int, bool) {
	refs, index, ok := kustoServiceTokenRefAtPosition(code, position)
	if !ok {
		return nil, 0, false
	}

	token := refs[index].positionedToken()
	hasAffinity := HasAffinity(token, position)
	if position <= token.TextStart() || !hasAffinity || token.Kind() == SyntaxKindEndOfTextToken {
		index--
	}
	if index < 0 {
		return nil, 0, false
	}
	return refs, index, true
}

func kustoServiceTokenRefAtPosition(code *KustoCode, position int) ([]kustoServiceTokenRef, int, bool) {
	if code == nil || code.Syntax() == nil {
		return nil, 0, false
	}

	refs := kustoServiceTokenRefs(code.Syntax())
	if len(refs) == 0 {
		return nil, 0, false
	}
	if position >= code.Syntax().FullWidth() {
		return refs, len(refs) - 1, true
	}

	for i, ref := range refs {
		if ref.triviaStart <= position && position < ref.end() {
			return refs, i, true
		}
	}

	return nil, 0, false
}

func kustoServiceTokenRefs(root SyntaxElement) []kustoServiceTokenRef {
	var refs []kustoServiceTokenRef
	collectKustoServiceTokenRefs(root, 0, nil, &refs)
	return refs
}

func collectKustoServiceTokenRefs(element SyntaxElement, triviaStart int, ancestors []kustoServiceElementRef, refs *[]kustoServiceTokenRef) {
	if element == nil {
		return
	}
	if token, ok := element.(*SyntaxToken); ok {
		tokenAncestors := append([]kustoServiceElementRef{}, ancestors...)
		*refs = append(*refs, kustoServiceTokenRef{
			token:       token,
			triviaStart: triviaStart,
			ancestors:   tokenAncestors,
		})
		return
	}

	childStart := triviaStart
	childAncestors := append(append([]kustoServiceElementRef{}, ancestors...), kustoServiceElementRef{element: element, triviaStart: triviaStart})
	for _, child := range element.Children() {
		if child != nil {
			collectKustoServiceTokenRefs(child, childStart, childAncestors, refs)
			childStart += child.FullWidth()
		}
	}
}

func (r kustoServiceTokenRef) positionedToken() *SyntaxToken {
	return r.token.withTriviaStart(r.triviaStart)
}

func (r kustoServiceTokenRef) end() int {
	if r.token == nil {
		return r.triviaStart
	}
	return r.triviaStart + r.token.FullWidth()
}

func syntaxElementHasMissingChildren(element SyntaxElement) bool {
	if element == nil {
		return false
	}
	for _, child := range element.Children() {
		if child == nil {
			continue
		}
		if syntaxElementIsMissing(child) || syntaxElementHasMissingChildren(child) {
			return true
		}
	}
	return false
}

func syntaxElementIsMissing(element SyntaxElement) bool {
	return element != nil && element.Width() == 0 && len(element.ContainedDiagnostics()) > 0
}
