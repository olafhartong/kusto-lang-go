package kustolang

import (
	"reflect"
	"strings"
)

// Ported from the token-parser helper surface in
// src/Kusto.Language/Parser/SyntaxParsers.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func CreateMissingSyntaxToken(kind SyntaxKind, diagnostic ...Diagnostic) *SyntaxToken {
	if len(diagnostic) > 0 {
		return NewMissingToken(kind, diagnostic[0])
	}
	if kind == SyntaxKindIdentifierToken {
		return NewMissingToken(kind, DiagnosticMissingName())
	}
	return NewMissingToken(kind, DiagnosticTermsExpected(missingTokenText(kind)))
}

func CreateMissingSyntaxTokenForKinds(kinds []SyntaxKind) *SyntaxToken {
	if len(kinds) == 0 {
		return CreateMissingSyntaxToken(SyntaxKindNone)
	}
	return NewMissingToken(kinds[0], DiagnosticTermsExpected(syntaxKindsExpectedTerms(kinds)...))
}

func CreateMissingSyntaxTokenForText(text string, diagnostic ...Diagnostic) *SyntaxToken {
	kind, ok := SyntaxKindFromText(text)
	if !ok {
		kind = SyntaxKindIdentifierToken
	}
	if len(diagnostic) > 0 {
		return NewMissingToken(kind, diagnostic[0])
	}
	return NewMissingToken(kind, DiagnosticTermsExpected(text))
}

func CreateMissingSyntaxTokenForTexts(texts []string) *SyntaxToken {
	if len(texts) == 0 {
		return CreateMissingSyntaxToken(SyntaxKindIdentifierToken)
	}
	return NewMissingToken(SyntaxKindIdentifierToken, DiagnosticTermsExpected(texts...))
}

func AnySyntaxTokenParser() ResultParser[LexicalToken, *SyntaxToken] {
	return asResultParser[LexicalToken, *SyntaxToken](
		MatchValue(func(LexicalToken) bool { return true }, func(token LexicalToken) *SyntaxToken {
			return SyntaxTokenFromLexical(token)
		}).WithTag("<any>"),
	)
}

func AnySyntaxTokenButEndParser() ResultParser[LexicalToken, *SyntaxToken] {
	return asResultParser[LexicalToken, *SyntaxToken](
		MatchValue(func(token LexicalToken) bool { return token.Kind != SyntaxKindEndOfTextToken }, func(token LexicalToken) *SyntaxToken {
			return SyntaxTokenFromLexical(token)
		}).WithTag("<any!end>"),
	)
}

func EndOfTextSyntaxParser() Parser[LexicalToken] {
	return Match(func(token LexicalToken) bool {
		return token.Kind == SyntaxKindEndOfTextToken
	})
}

func SyntaxTokenKindParser(kind SyntaxKind, items ...CompletionItem) ResultParser[LexicalToken, *SyntaxToken] {
	rule := asResultParser[LexicalToken, *SyntaxToken](
		MatchValue(func(token LexicalToken) bool { return token.Kind == kind }, func(token LexicalToken) *SyntaxToken {
			return SyntaxTokenFromLexical(token)
		}).WithTag(defaultSyntaxParserTagForKind(kind)),
	)
	if len(items) == 0 {
		if item, ok := CreateSyntaxCompletionItemForKind(kind, syntaxCompletionKindForKind(kind), CompletionPriorityNormal); ok {
			items = append(items, item)
		}
	}
	if len(items) > 0 {
		return WithSyntaxCompletion(rule, items...)
	}
	return rule
}

func SyntaxTokenKindsParser(kinds []SyntaxKind, defaultKind ...CompletionKind) ResultParser[LexicalToken, *SyntaxToken] {
	set := map[SyntaxKind]struct{}{}
	for _, kind := range kinds {
		set[kind] = struct{}{}
	}
	rule := asResultParser[LexicalToken, *SyntaxToken](
		MatchValue(func(token LexicalToken) bool {
			_, ok := set[token.Kind]
			return ok
		}, func(token LexicalToken) *SyntaxToken {
			return SyntaxTokenFromLexical(token)
		}).WithTag(strings.Join(syntaxKindTags(kinds), " | ")),
	)
	items := completionItemsForSyntaxKinds(kinds, firstCompletionKind(defaultKind), CompletionPriorityNormal)
	if len(items) > 0 {
		return WithSyntaxCompletion(rule, items...)
	}
	return rule
}

func MatchesLexicalText(source Source[LexicalToken], start int, text string) int {
	textOffset := 0
	i := 0
	for ; !source.IsEnd(start + i); i++ {
		token := source.Peek(start + i)
		if i > 0 && token.Trivia != "" {
			break
		}
		if len(token.Text) > len(text)-textOffset {
			break
		}
		if !strings.HasPrefix(text[textOffset:], token.Text) {
			break
		}
		textOffset += len(token.Text)
		if textOffset == len(text) {
			return i + 1
		}
	}
	return -(i + 1)
}

func ProduceSyntaxToken(source Source[LexicalToken], start int, length int, asKind ...SyntaxKind) *SyntaxToken {
	firstToken := source.Peek(start)
	if firstToken.Kind == SyntaxKindNone && firstToken.Text == "" && firstToken.Trivia == "" {
		return nil
	}

	text := firstToken.Text
	for i := 1; i < length; i++ {
		token := source.Peek(start + i)
		if token.Kind == SyntaxKindNone && token.Text == "" || token.Trivia != "" {
			return nil
		}
		text += token.Text
	}
	return ProduceSyntaxTokenFromText(source, start, length, text, asKind...)
}

func ProduceSyntaxTokenFromText(source Source[LexicalToken], start int, length int, text string, asKind ...SyntaxKind) *SyntaxToken {
	firstToken := source.Peek(start)
	if length == 1 && (len(asKind) == 0 || asKind[0] == firstToken.Kind) {
		return SyntaxTokenFromLexical(firstToken)
	}

	if len(asKind) == 0 {
		return SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, firstToken.Trivia, text))
	}

	kind := asKind[0]
	switch SyntaxCategoryOf(kind) {
	case SyntaxCategoryIdentifier, SyntaxCategoryOther:
		return SyntaxTokenFromLexical(NewLexicalToken(kind, firstToken.Trivia, text))
	default:
		panic(InvalidOperationError{Message: "Cannot produce syntax token for kind: " + missingTokenText(kind)})
	}
}

func CombinedLexicalTokenText(source Source[LexicalToken], start int, length int, includeInnerTrivia ...bool) string {
	if length == 1 {
		return source.Peek(start).Text
	}
	includeTrivia := true
	if len(includeInnerTrivia) > 0 {
		includeTrivia = includeInnerTrivia[0]
	}
	var builder strings.Builder
	for i := 0; i < length; i++ {
		token := source.Peek(start + i)
		if i > 0 && includeTrivia {
			builder.WriteString(token.Trivia)
		}
		builder.WriteString(token.Text)
	}
	return builder.String()
}

func SyntaxMatchTextParser(text string, asKind ...SyntaxKind) ResultParser[LexicalToken, *SyntaxToken] {
	return MatchConsumerValue(
		func(source Source[LexicalToken], start int) int {
			return MatchesLexicalText(source, start, text)
		},
		func(source Source[LexicalToken], start int, length int) *SyntaxToken {
			return ProduceSyntaxTokenFromText(source, start, length, text, asKind...)
		},
	)
}

func SyntaxTokenTextParser(text string, asKind ...SyntaxKind) ResultParser[LexicalToken, *SyntaxToken] {
	rule := asResultParser[LexicalToken, *SyntaxToken](SyntaxMatchTextParser(text, asKind...).WithTag(defaultSyntaxParserTagForText(text)))
	if item, ok := CreateSyntaxCompletionItemForText(text, syntaxCompletionKindForText(text), CompletionPriorityNormal); ok {
		return WithSyntaxCompletion(rule, item)
	}
	return rule
}

func SyntaxTokenTextsParser(texts []string, defaultKind ...CompletionKind) ResultParser[LexicalToken, *SyntaxToken] {
	parsers := make([]ResultParser[LexicalToken, *SyntaxToken], 0, len(texts))
	for _, text := range texts {
		parsers = append(parsers, SyntaxMatchTextParser(text))
	}
	rule := asResultParser[LexicalToken, *SyntaxToken](FirstValue(parsers...).WithTag(strings.Join(syntaxTextTags(texts), " | ")))
	items := completionItemsForSyntaxTexts(texts, firstCompletionKind(defaultKind), CompletionPriorityNormal)
	if len(items) > 0 {
		return WithSyntaxCompletion(rule, items...)
	}
	return rule
}

func HiddenSyntaxTokenTextParser(text string, asKind ...SyntaxKind) ResultParser[LexicalToken, *SyntaxToken] {
	return asResultParser[LexicalToken, *SyntaxToken](SyntaxMatchTextParser(text, asKind...).WithTag(defaultSyntaxParserTagForText(text)))
}

func HiddenSyntaxTokenKindParser(kind SyntaxKind) ResultParser[LexicalToken, *SyntaxToken] {
	return asResultParser[LexicalToken, *SyntaxToken](
		MatchValue(func(token LexicalToken) bool { return token.Kind == kind }, func(token LexicalToken) *SyntaxToken {
			return SyntaxTokenFromLexical(token)
		}).WithTag(defaultSyntaxParserTagForKind(kind)),
	)
}

func HiddenSyntaxTokenTextsParser(texts []string) ResultParser[LexicalToken, *SyntaxToken] {
	set := map[string]struct{}{}
	for _, text := range texts {
		set[text] = struct{}{}
	}
	return asResultParser[LexicalToken, *SyntaxToken](
		MatchValue(func(token LexicalToken) bool {
			_, ok := set[token.Text]
			return ok
		}, func(token LexicalToken) *SyntaxToken {
			return SyntaxTokenFromLexical(token)
		}).WithTag(strings.Join(syntaxTextTags(texts), " | ")),
	)
}

func RequiredSyntaxTokenKindParser(kind SyntaxKind, items ...CompletionItem) ResultParser[LexicalToken, *SyntaxToken] {
	return Required(SyntaxTokenKindParser(kind, items...), func(Source[LexicalToken], int) *SyntaxToken {
		return CreateMissingSyntaxToken(kind)
	})
}

func RequiredSyntaxTokenKindsParser(kinds []SyntaxKind, defaultKind ...CompletionKind) ResultParser[LexicalToken, *SyntaxToken] {
	return Required(SyntaxTokenKindsParser(kinds, defaultKind...), func(Source[LexicalToken], int) *SyntaxToken {
		return CreateMissingSyntaxTokenForKinds(kinds)
	})
}

func RequiredSyntaxTokenTextParser(text string, asKind ...SyntaxKind) ResultParser[LexicalToken, *SyntaxToken] {
	return Required(SyntaxTokenTextParser(text, asKind...), func(Source[LexicalToken], int) *SyntaxToken {
		return CreateMissingSyntaxTokenForText(text)
	})
}

func RequiredSyntaxTokenTextsParser(texts []string, defaultKind ...CompletionKind) ResultParser[LexicalToken, *SyntaxToken] {
	return Required(SyntaxTokenTextsParser(texts, defaultKind...), func(Source[LexicalToken], int) *SyntaxToken {
		return CreateMissingSyntaxTokenForTexts(texts)
	})
}

func WithSyntaxCompletion[TElement any](parser ResultParser[LexicalToken, TElement], items ...CompletionItem) ResultParser[LexicalToken, TElement] {
	annotations := stripSyntaxCompletionItemAnnotations(parser.Annotations())
	for _, item := range items {
		annotations = append(annotations, item)
	}
	return asResultParser[LexicalToken, TElement](parser.WithAnnotations(annotations))
}

func WithSyntaxCompletionHint[TElement any](parser ResultParser[LexicalToken, TElement], hint CompletionHint) ResultParser[LexicalToken, TElement] {
	annotations := stripSyntaxCompletionHintAnnotations(parser.Annotations())
	annotations = append(annotations, hint)
	return asResultParser[LexicalToken, TElement](parser.WithAnnotations(annotations))
}

func CreateSyntaxCompletionItemForKind(kind SyntaxKind, completionKind CompletionKind, priority CompletionPriority, applyText ...string) (CompletionItem, bool) {
	return CreateSyntaxCompletionItemForText(SyntaxText(kind), completionKind, priority, applyText...)
}

func CreateSyntaxCompletionItemForText(text string, completionKind CompletionKind, priority CompletionPriority, applyText ...string) (CompletionItem, bool) {
	if strings.TrimSpace(text) == "" || strings.HasPrefix(text, "_") {
		return CompletionItem{}, false
	}
	options := []CompletionItemOption{WithCompletionPriority(priority)}
	if len(applyText) > 0 {
		options = append(options, WithCompletionApplyTextMarkers(applyText[0]))
	}
	return NewCompletionItem(completionKind, text, options...), true
}

func SyntaxCompletionKindForText(text string, defaultKind ...CompletionKind) CompletionKind {
	return syntaxCompletionKindForText(text, defaultKind...)
}

func SyntaxCompletionKindForKind(kind SyntaxKind, defaultKind ...CompletionKind) CompletionKind {
	return syntaxCompletionKindForKind(kind, defaultKind...)
}

func ParseAllSyntax[TElement any](parser ResultParser[LexicalToken, TElement], text string, alwaysProduceEndToken ...bool) []TElement {
	return ParseAllSyntaxTokens(parser, syntaxParserTokens(text, alwaysProduceEndToken...))
}

func ParseAllSyntaxTokens[TElement any](parser ResultParser[LexicalToken, TElement], tokens []LexicalToken) []TElement {
	source := NewArraySource(tokens)
	start := 0
	values := []TElement{}
	for !source.IsEnd(start) {
		result := parser.ParseValue(source, start)
		if result.Length <= 0 {
			break
		}
		values = append(values, result.Value)
		start += result.Length
	}
	return values
}

func ParseFirstSyntax[TElement any](parser ResultParser[LexicalToken, TElement], text string, alwaysProduceEndToken ...bool) TElement {
	return ParseFirstSyntaxTokens(parser, syntaxParserTokens(text, alwaysProduceEndToken...))
}

func ParseFirstSyntaxTokens[TElement any](parser ResultParser[LexicalToken, TElement], tokens []LexicalToken) TElement {
	source := NewArraySource(tokens)
	return parser.ParseValue(source, 0).Value
}

func parseFirstSyntaxTokensWithCancellation[TElement any](parser ResultParser[LexicalToken, TElement], tokens []LexicalToken, cancellationToken CancellationToken) TElement {
	cancellationToken.ThrowIfCancellationRequested()
	source := newCancellationSource[LexicalToken](NewArraySource(tokens), cancellationToken)
	value := parser.ParseValue(source, 0).Value
	cancellationToken.ThrowIfCancellationRequested()
	return value
}

func ScanFirstSyntax(parser Parser[LexicalToken], text string, alwaysProduceEndToken ...bool) int {
	return parser.Scan(NewArraySource(syntaxParserTokens(text, alwaysProduceEndToken...)), 0)
}

func syntaxParserTokens(text string, alwaysProduceEndToken ...bool) []LexicalToken {
	produceEndToken := false
	if len(alwaysProduceEndToken) > 0 {
		produceEndToken = alwaysProduceEndToken[0]
	}
	return LexWithOptions(text, DefaultParseOptions.WithAlwaysProduceEndTokens(produceEndToken))
}

func SyntaxListParser[TElement SyntaxElement](
	elementParser ResultParser[LexicalToken, TElement],
	fnMissingElement func(Source[LexicalToken], int) TElement,
	oneOrMore bool,
) ResultParser[LexicalToken, *SyntaxListOf[TElement]] {
	return ListElementsWithMissing(
		elementParser,
		fnMissingElement,
		oneOrMore,
		func(elements []TElement) *SyntaxListOf[TElement] {
			list := make([]SyntaxElement, 0, len(elements))
			for _, element := range elements {
				list = append(list, element)
			}
			return AsSyntaxListOf[TElement](NewTypedSyntaxList(syntaxElementTypeName[TElement](), list...))
		},
	)
}

func SyntaxSeparatedListParser[TElement SyntaxElement](
	primaryElementParser ResultParser[LexicalToken, TElement],
	separatorKind SyntaxKind,
	fnMissingElement func(Source[LexicalToken], int) TElement,
	endOfList Parser[LexicalToken],
	oneOrMore bool,
	allowTrailingSeparator bool,
) ResultParser[LexicalToken, *SeparatedSyntaxListOf[TElement]] {
	return SyntaxSeparatedListFullParser(
		primaryElementParser,
		separatorKind,
		primaryElementParser,
		fnMissingElement,
		endOfList,
		oneOrMore,
		allowTrailingSeparator,
	)
}

func SyntaxSeparatedListFullParser[TElement SyntaxElement](
	primaryElementParser ResultParser[LexicalToken, TElement],
	separatorKind SyntaxKind,
	secondaryElementParser ResultParser[LexicalToken, TElement],
	fnMissingElement func(Source[LexicalToken], int) TElement,
	endOfList Parser[LexicalToken],
	oneOrMore bool,
	allowTrailingSeparator bool,
) ResultParser[LexicalToken, *SeparatedSyntaxListOf[TElement]] {
	return OListFull(
		primaryElementParser,
		SyntaxTokenKindParser(separatorKind),
		secondaryElementParser,
		fnMissingElement,
		func(Source[LexicalToken], int) *SyntaxToken { return CreateMissingSyntaxToken(separatorKind) },
		fnMissingElement,
		endOfList,
		oneOrMore,
		allowTrailingSeparator,
		MakeSyntaxSeparatedList[TElement],
	)
}

func EndOfCommaListParser() Parser[LexicalToken] {
	return Match(func(token LexicalToken) bool {
		switch token.Kind {
		case SyntaxKindEndOfTextToken,
			SyntaxKindCloseParenToken,
			SyntaxKindCloseBracketToken,
			SyntaxKindCloseBraceToken,
			SyntaxKindBarToken,
			SyntaxKindSemicolonToken:
			return true
		default:
			return false
		}
	})
}

func SyntaxCommaListParser[TElement SyntaxElement](
	elementParser ResultParser[LexicalToken, TElement],
	fnMissingElement func(Source[LexicalToken], int) TElement,
	oneOrMore bool,
	allowTrailingComma bool,
	endKinds ...SyntaxKind,
) ResultParser[LexicalToken, *SeparatedSyntaxListOf[TElement]] {
	endOfList := EndOfCommaListParser()
	if len(endKinds) > 0 {
		set := map[SyntaxKind]struct{}{}
		for _, kind := range endKinds {
			set[kind] = struct{}{}
		}
		endOfList = First(
			endOfList,
			Match(func(token LexicalToken) bool {
				_, ok := set[token.Kind]
				return ok
			}),
		)
	}
	return SyntaxSeparatedListParser(elementParser, SyntaxKindCommaToken, fnMissingElement, endOfList, oneOrMore, allowTrailingComma)
}

func SyntaxOneOrMoreCommaListParser[TElement SyntaxElement](
	elementParser ResultParser[LexicalToken, TElement],
	fnMissingElement func(Source[LexicalToken], int) TElement,
) ResultParser[LexicalToken, *SeparatedSyntaxListOf[TElement]] {
	return SyntaxCommaListParser(elementParser, fnMissingElement, true, false)
}

func MakeSyntaxSeparatedList[TElement SyntaxElement](elements []any) *SeparatedSyntaxListOf[TElement] {
	if elements == nil {
		return AsSeparatedSyntaxListOf[TElement](NewTypedSyntaxList("SeparatedElement<" + syntaxElementTypeName[TElement]() + ">"))
	}
	separated := make([]SyntaxElement, 0, (len(elements)+1)/2)
	for i := 0; i < len(elements); i += 2 {
		element, _ := elements[i].(TElement)
		var separator *SyntaxToken
		if i < len(elements)-1 {
			separator, _ = elements[i+1].(*SyntaxToken)
		}
		separated = append(separated, NewSeparatedElement(element, separator))
	}
	return AsSeparatedSyntaxListOf[TElement](NewTypedSyntaxList("SeparatedElement<"+syntaxElementTypeName[TElement]()+">", separated...))
}

func WithSyntaxExamples[TElement any](parser ResultParser[LexicalToken, TElement], values ...string) ResultParser[LexicalToken, TElement] {
	annotations := parser.Annotations()
	for _, value := range values {
		annotations = append(annotations, NewCompletionItem(CompletionKindExample, value))
	}
	return asResultParser[LexicalToken, TElement](parser.WithAnnotations(annotations))
}

func defaultSyntaxParserTagForText(text string) string {
	if kind, ok := SyntaxKindFromText(text); ok {
		return defaultSyntaxParserTagForKind(kind)
	}
	return text
}

func defaultSyntaxParserTagForKind(kind SyntaxKind) string {
	text := SyntaxText(kind)
	if SyntaxCategoryOf(kind) == SyntaxCategoryPunctuation {
		return "'" + text + "'"
	}
	return text
}

func syntaxCompletionKindForText(text string, defaultKind ...CompletionKind) CompletionKind {
	if kind, ok := SyntaxKindFromText(text); ok {
		return syntaxCompletionKindForKind(kind, defaultKind...)
	}
	if len(defaultKind) > 0 {
		return defaultKind[0]
	}
	return CompletionKindSyntax
}

func syntaxCompletionKindForKind(kind SyntaxKind, defaultKind ...CompletionKind) CompletionKind {
	switch SyntaxCategoryOf(kind) {
	case SyntaxCategoryKeyword:
		return CompletionKindKeyword
	case SyntaxCategoryOperator:
		return CompletionKindScalarInfix
	case SyntaxCategoryPunctuation:
		return CompletionKindPunctuation
	default:
		if len(defaultKind) > 0 {
			return defaultKind[0]
		}
		return CompletionKindSyntax
	}
}

func firstCompletionKind(kinds []CompletionKind) []CompletionKind {
	if len(kinds) == 0 {
		return nil
	}
	return []CompletionKind{kinds[0]}
}

func completionItemsForSyntaxKinds(kinds []SyntaxKind, defaultKind []CompletionKind, priority CompletionPriority) []CompletionItem {
	items := []CompletionItem{}
	for _, kind := range kinds {
		if item, ok := CreateSyntaxCompletionItemForKind(kind, syntaxCompletionKindForKind(kind, defaultKind...), priority); ok {
			items = append(items, item)
		}
	}
	return items
}

func completionItemsForSyntaxTexts(texts []string, defaultKind []CompletionKind, priority CompletionPriority) []CompletionItem {
	items := []CompletionItem{}
	for _, text := range texts {
		if item, ok := CreateSyntaxCompletionItemForText(text, syntaxCompletionKindForText(text, defaultKind...), priority); ok {
			items = append(items, item)
		}
	}
	return items
}

func syntaxKindTags(kinds []SyntaxKind) []string {
	tags := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		tags = append(tags, defaultSyntaxParserTagForKind(kind))
	}
	return tags
}

func syntaxTextTags(texts []string) []string {
	tags := make([]string, 0, len(texts))
	for _, text := range texts {
		tags = append(tags, defaultSyntaxParserTagForText(text))
	}
	return tags
}

func syntaxKindsExpectedTerms(kinds []SyntaxKind) []string {
	terms := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		terms = append(terms, missingTokenText(kind))
	}
	return terms
}

func syntaxElementTypeName[TElement any]() string {
	var zero TElement
	typ := reflect.TypeOf(zero)
	if typ == nil {
		return "SyntaxElement"
	}
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Name() == "" {
		return "SyntaxElement"
	}
	return typ.Name()
}

func stripSyntaxCompletionItemAnnotations(annotations []any) []any {
	filtered := make([]any, 0, len(annotations))
	for _, annotation := range annotations {
		switch annotation.(type) {
		case CompletionItem, *CompletionItem:
			continue
		default:
			filtered = append(filtered, annotation)
		}
	}
	return filtered
}

func syntaxCompletionItemsFromAnnotations(annotations []any) []CompletionItem {
	items := []CompletionItem{}
	for _, annotation := range annotations {
		switch item := annotation.(type) {
		case CompletionItem:
			items = append(items, item)
		case *CompletionItem:
			if item != nil {
				items = append(items, *item)
			}
		}
	}
	return items
}

func stripSyntaxCompletionHintAnnotations(annotations []any) []any {
	filtered := make([]any, 0, len(annotations))
	for _, annotation := range annotations {
		switch annotation.(type) {
		case CompletionHint:
			continue
		default:
			filtered = append(filtered, annotation)
		}
	}
	return filtered
}

func syntaxCompletionHintFromAnnotations(annotations []any) CompletionHint {
	var hint CompletionHint
	for _, annotation := range annotations {
		if value, ok := annotation.(CompletionHint); ok {
			hint |= value
		}
	}
	return hint
}
