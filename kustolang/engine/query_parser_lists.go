package kustolang

// Ported from src/Kusto.Language/Parser/QueryParser.cs ParseCommaList<TElement>.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func parseCommaList[T SyntaxElement](
	p *queryParserState,
	elementType string,
	parseElement func() T,
	createMissingElement func() T,
	scanEnd func() bool,
	oneOrMore bool,
	allowTrailingComma bool,
) *SeparatedSyntaxListOf[T] {
	var elements []SyntaxElement
	var element T
	if scanEnd == nil || !scanEnd() {
		element = parseElement()
	}
	if isNilSyntaxElement(element) && p.peekToken().Kind == SyntaxKindCommaToken && createMissingElement != nil {
		element = createMissingElement()
	}

	for !isNilSyntaxElement(element) {
		kind := p.peekToken().Kind
		switch {
		case kind == SyntaxKindEndOfTextToken || kind == SyntaxKindNone || (scanEnd != nil && scanEnd()):
			elements = append(elements, NewSeparatedElement(element, nil))
			element = *new(T)
		case kind == SyntaxKindCommaToken:
			comma := p.parseToken()
			elements = append(elements, NewSeparatedElement(element, comma))
			if createMissingElement != nil {
				if scanEnd == nil || !scanEnd() {
					element = parseElement()
				} else {
					element = *new(T)
				}
				if !isNilSyntaxElement(element) {
					continue
				}
				if allowTrailingComma {
					break
				}
				elements = append(elements, NewSeparatedElement(createMissingElement(), nil))
			}
			element = *new(T)
		case scanEnd != nil:
			next := parseElement()
			if !isNilSyntaxElement(next) {
				elements = append(elements, NewSeparatedElement(element, CreateMissingSyntaxToken(SyntaxKindCommaToken)))
			} else {
				elements = append(elements, NewSeparatedElement(element, nil))
			}
			element = next
		default:
			elements = append(elements, NewSeparatedElement(element, nil))
			element = *new(T)
		}
	}

	if oneOrMore && len(elements) == 0 && createMissingElement != nil {
		elements = append(elements, NewSeparatedElement(createMissingElement(), nil))
	}

	return AsSeparatedSyntaxListOf[T](NewTypedSyntaxList("SeparatedElement<"+elementType+">", elements...))
}
