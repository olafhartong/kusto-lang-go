package kustolang

import (
	"unicode"
	"unicode/utf8"
)

// Ported from src/Kusto.Language/Editor/Kusto/KustoClassifier.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// KustoClassifierGetClassifications classifies the source ranges under root using the
// same token/trivia categories as the upstream Kusto editor classifier.
func KustoClassifierGetClassifications(root SyntaxElement, start int, length int, clipToRange bool, list []ClassifiedRange, cancellationToken CancellationToken) []ClassifiedRange {
	end := start + length
	limiter := &kustoClassifierLimiter{
		start:       start,
		end:         end,
		clipToRange: clipToRange,
		list:        list,
	}

	for _, ref := range kustoServiceTokenRefs(root) {
		if !kustoClassifierTokenOverlapsRange(ref, start, end) {
			continue
		}

		cancellationToken.ThrowIfCancellationRequested()

		kustoClassifierAddTriviaClassifications(ref, limiter)

		token := ref.positionedToken()
		tokenStart := token.TextStart()
		if tokenStart < end && token.End() > start {
			tokenKind := SyntaxCategoryOf(ref.token.Kind())
			if tokenKind == SyntaxCategoryLiteral {
				rawText := token.RawText()
				if prefixLength := kustoClassifierLiteralPrefixLength(rawText); prefixLength > 0 {
					limiter.add(ClassifiedRange{
						Start:  tokenStart,
						Length: prefixLength,
						Kind:   ClassificationKindKeyword,
					})
					limiter.add(ClassifiedRange{
						Start:  tokenStart + prefixLength,
						Length: len(rawText) - prefixLength,
						Kind:   ClassificationKindLiteral,
					})
					continue
				}
			}

			limiter.add(ClassifiedRange{
				Start:  tokenStart,
				Length: len(token.RawText()),
				Kind:   kustoClassifierGetKind(ref),
			})
		}
	}

	return limiter.list
}

type kustoClassifierLimiter struct {
	start       int
	end         int
	clipToRange bool
	list        []ClassifiedRange
}

func (l *kustoClassifierLimiter) add(cr ClassifiedRange) {
	if cr.Length <= 0 || cr.End() <= l.start || cr.Start >= l.end {
		return
	}

	if l.clipToRange {
		rangeEnd := cr.End()
		if cr.Start < l.start {
			cr.Start = l.start
		}
		if rangeEnd > l.end {
			rangeEnd = l.end
		}
		cr.Length = rangeEnd - cr.Start
	}

	if len(l.list) > 0 {
		last := &l.list[len(l.list)-1]
		if last.Kind == cr.Kind && last.End() == cr.Start {
			last.Length += cr.Length
			return
		}
	}

	l.list = append(l.list, cr)
}

func kustoClassifierTokenOverlapsRange(ref kustoServiceTokenRef, start int, end int) bool {
	return ref.triviaStart < end && ref.end() > start
}

func kustoClassifierAddTriviaClassifications(ref kustoServiceTokenRef, limiter *kustoClassifierLimiter) {
	trivia := ref.token.Trivia()
	pos := 0
	for pos < len(trivia) {
		ch, size := utf8.DecodeRuneInString(trivia[pos:])
		if IsWhitespace(ch) || IsLineBreakStart(ch) {
			pos += size
			continue
		}

		commentStart := pos
		for pos < len(trivia) {
			ch, size = utf8.DecodeRuneInString(trivia[pos:])
			if IsLineBreakStart(ch) {
				break
			}
			pos += size
		}
		limiter.add(ClassifiedRange{
			Start:  ref.triviaStart + commentStart,
			Length: pos - commentStart,
			Kind:   ClassificationKindComment,
		})
	}
}

func kustoClassifierGetKind(ref kustoServiceTokenRef) ClassificationKind {
	if kustoClassifierIsSkippedToken(ref) {
		return ClassificationKindPlainText
	}

	if kustoClassifierIsCommandVerbToken(ref) {
		return ClassificationKindCommand
	}

	if _, ok := kustoClassifierDirectParent(ref).(*BracedName); ok {
		return ClassificationKindClientParameter
	}

	category := SyntaxCategoryOf(ref.token.Kind())
	switch category {
	case SyntaxCategoryIdentifier:
		if kustoClassifierIsQueryParameter(ref) {
			return ClassificationKindQueryParameter
		}
		return kustoClassifierGetIdentifierKind(ref)
	case SyntaxCategoryLiteral:
		if ref.token.Kind() == SyntaxKindStringLiteralToken {
			return ClassificationKindStringLiteral
		}
		return ClassificationKindLiteral
	case SyntaxCategoryKeyword:
		if kustoClassifierIsIdentifierKeyword(ref) {
			return kustoClassifierGetIdentifierKind(ref)
		}
		if kustoClassifierIsTypeToken(ref) {
			return ClassificationKindType
		}
		if kustoClassifierIsScalarOperatorToken(ref) {
			return ClassificationKindScalarOperator
		}
		if kustoClassifierIsQueryOperatorToken(ref) {
			return ClassificationKindQueryOperator
		}
		if kustoClassifierIsQueryParameter(ref) {
			return ClassificationKindQueryParameter
		}
		if kustoClassifierIsCommandVerbToken(ref) {
			return ClassificationKindCommand
		}
		return ClassificationKindKeyword
	case SyntaxCategoryPunctuation:
		return ClassificationKindPunctuation
	case SyntaxCategoryOperator:
		return ClassificationKindMathOperator
	}

	if ref.token.Kind() == SyntaxKindDirectiveToken {
		return ClassificationKindDirective
	}

	return ClassificationKindPlainText
}

func kustoClassifierGetIdentifierKind(ref kustoServiceTokenRef) ClassificationKind {
	switch kustoClassifierReferencedSymbolKind(ref) {
	case SymbolKindColumn:
		return ClassificationKindColumn
	case SymbolKindTable:
		return ClassificationKindTable
	case SymbolKindDatabase:
		return ClassificationKindDatabase
	case SymbolKindFunction, SymbolKindPattern:
		return ClassificationKindFunction
	case SymbolKindVariable:
		return ClassificationKindVariable
	case SymbolKindParameter:
		return ClassificationKindParameter
	case SymbolKindMaterializedView:
		return ClassificationKindMaterializedView
	case SymbolKindOption:
		return ClassificationKindOption
	}

	if kustoClassifierIsName(ref) {
		return ClassificationKindIdentifier
	}

	return ClassificationKindKeyword
}

func kustoClassifierReferencedSymbolKind(ref kustoServiceTokenRef) SymbolKind {
	for i := len(ref.ancestors) - 1; i >= 0; i-- {
		wrapped := WrapGeneratedSyntaxElement(ref.ancestors[i].element)
		if node, ok := wrapped.(interface{ ReferencedSymbol() Symbol }); ok {
			if symbol := node.ReferencedSymbol(); symbol != nil {
				return symbol.SymbolKind()
			}
		}
	}
	return SymbolKindNone
}

func kustoClassifierLiteralPrefixLength(text string) int {
	openParen := -1
	for i, r := range text {
		if r == '(' {
			openParen = i
			break
		}
		if !unicode.IsLetter(r) {
			return 0
		}
	}
	if openParen <= 0 {
		return 0
	}
	return openParen
}

func kustoClassifierDirectParent(ref kustoServiceTokenRef) SyntaxElement {
	if len(ref.ancestors) == 0 {
		return nil
	}
	return WrapGeneratedSyntaxElement(ref.ancestors[len(ref.ancestors)-1].element)
}

func kustoClassifierParentAt(ref kustoServiceTokenRef, distance int) SyntaxElement {
	index := len(ref.ancestors) - 1 - distance
	if index < 0 || index >= len(ref.ancestors) {
		return nil
	}
	return WrapGeneratedSyntaxElement(ref.ancestors[index].element)
}

func kustoClassifierIsSkippedToken(ref kustoServiceTokenRef) bool {
	if _, ok := kustoClassifierDirectParent(ref).(*SyntaxList); !ok {
		return false
	}
	_, ok := kustoClassifierParentAt(ref, 1).(*SkippedTokens)
	return ok
}

func kustoClassifierIsName(ref kustoServiceTokenRef) bool {
	_, ok := kustoClassifierDirectParent(ref).(Name)
	return ok
}

func kustoClassifierIsIdentifierKeyword(ref kustoServiceTokenRef) bool {
	_, ok := kustoClassifierDirectParent(ref).(*TokenName)
	return ok
}

func kustoClassifierIsTypeToken(ref kustoServiceTokenRef) bool {
	if _, ok := kustoClassifierDirectParent(ref).(*PrimitiveTypeExpression); !ok {
		return false
	}
	return IsTypeKeyword(ref.token.Kind())
}

func kustoClassifierIsScalarOperatorToken(ref kustoServiceTokenRef) bool {
	parent := kustoClassifierDirectParent(ref)
	switch p := parent.(type) {
	case *BinaryExpression:
		return p.Operator() == ref.token
	case *PrefixUnaryExpression:
		return p.Operator() == ref.token
	default:
		return false
	}
}

func kustoClassifierIsQueryOperatorToken(ref kustoServiceTokenRef) bool {
	parent := kustoClassifierDirectParent(ref)
	if _, ok := parent.(QueryOperator); ok && kustoClassifierIsChild(parent, 0, ref.token) {
		return true
	}
	if _, ok := parent.(*TopNestedClause); ok && kustoClassifierIsChild(parent, 0, ref.token) {
		return true
	}
	return false
}

func kustoClassifierIsQueryParameter(ref kustoServiceTokenRef) bool {
	if _, ok := kustoClassifierDirectParent(ref).(Name); !ok {
		return false
	}
	if _, ok := kustoClassifierParentAt(ref, 1).(*NameDeclaration); !ok {
		return false
	}
	_, ok := kustoClassifierParentAt(ref, 2).(*NamedParameter)
	return ok
}

func kustoClassifierIsCommandVerbToken(ref kustoServiceTokenRef) bool {
	parent := kustoClassifierDirectParent(ref)
	if parent == nil {
		return false
	}

	if _, ok := parent.(Command); ok {
		index := kustoClassifierChildIndex(parent, ref.token)
		return index == 0 || index == 1
	}

	// Generated engine commands store their verb in the first child of a
	// CustomNode, while the CustomCommand wrapper owns that node.
	if custom, ok := parent.(*CustomNode); ok {
		if _, ok := kustoClassifierParentAt(ref, 1).(*CustomCommand); ok {
			return kustoClassifierChildIndex(custom, ref.token) == 0
		}
	}

	// Unknown commands retain their words in a SyntaxList rather than a
	// generated command node. Only their first word is a command verb.
	if _, ok := parent.(*SyntaxList); ok {
		if command, ok := kustoClassifierParentAt(ref, 1).(*UnknownCommand); ok {
			parts := command.Parts()
			return parts != nil && parts.Count() > 0 && parts.At(0) == ref.token
		}
	}

	return false
}

func kustoClassifierIsChild(parent SyntaxElement, index int, token *SyntaxToken) bool {
	if parent == nil || index < 0 || index >= parent.ChildCount() {
		return false
	}
	return parent.GetChild(index) == token
}

func kustoClassifierChildIndex(parent SyntaxElement, token *SyntaxToken) int {
	if parent == nil {
		return -1
	}
	for i := 0; i < parent.ChildCount(); i++ {
		if parent.GetChild(i) == token {
			return i
		}
	}
	return -1
}
