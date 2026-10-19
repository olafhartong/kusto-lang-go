package kustolang

// Ported from the public syntax construction helpers under
// src/Kusto.Language/Syntax and src/Kusto.Language/Parser.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

type CustomElementDescriptor = internal.CustomElementDescriptor
type CustomElementDescriptorFromOptions = internal.CustomElementDescriptorFromOptions
type CustomNode = internal.CustomNode
type GeneratedSyntaxNodeChild = internal.GeneratedSyntaxNodeChild
type GeneratedSyntaxNodePayload = internal.GeneratedSyntaxNodePayload
type GeneratedSyntaxNodeMetadata = internal.GeneratedSyntaxNodeMetadata

var CustomElementDescriptorDefault = internal.CustomElementDescriptorDefault

func NewLexicalToken(kind SyntaxKind, trivia, text string, diagnostics ...Diagnostic) LexicalToken {
	return internal.NewLexicalToken(kind, trivia, text, diagnostics...)
}

func SyntaxTokenFromLexical(token LexicalToken, diagnostics ...Diagnostic) *SyntaxToken {
	return internal.SyntaxTokenFromLexical(token, diagnostics...)
}

func NewIdentifierToken(trivia, text string, diagnostics ...Diagnostic) *SyntaxToken {
	return internal.NewIdentifierToken(trivia, text, diagnostics...)
}

func NewIdentifierTokenWithValueText(trivia, text, valueText string, diagnostics ...Diagnostic) *SyntaxToken {
	return internal.NewIdentifierTokenWithValueText(trivia, text, valueText, diagnostics...)
}

func NewKeywordToken(trivia string, kind SyntaxKind, diagnostics ...Diagnostic) (*SyntaxToken, error) {
	return internal.NewKeywordToken(trivia, kind, diagnostics...)
}

func NewKeywordTokenWithValueText(trivia string, kind SyntaxKind, valueText string, diagnostics ...Diagnostic) (*SyntaxToken, error) {
	return internal.NewKeywordTokenWithValueText(trivia, kind, valueText, diagnostics...)
}

func NewPunctuationToken(trivia string, kind SyntaxKind, diagnostics ...Diagnostic) (*SyntaxToken, error) {
	return internal.NewPunctuationToken(trivia, kind, diagnostics...)
}

func NewOperatorToken(trivia string, kind SyntaxKind, diagnostics ...Diagnostic) (*SyntaxToken, error) {
	return internal.NewOperatorToken(trivia, kind, diagnostics...)
}

func NewLiteralToken(trivia, text string, kind SyntaxKind, diagnostics ...Diagnostic) (*SyntaxToken, error) {
	return internal.NewLiteralToken(trivia, text, kind, diagnostics...)
}

func NewOtherToken(trivia, text string, kind SyntaxKind, diagnostics ...Diagnostic) *SyntaxToken {
	return internal.NewOtherToken(trivia, text, kind, diagnostics...)
}

func NewMissingTokenWithTrivia(trivia string, kind SyntaxKind, diagnostics ...Diagnostic) *SyntaxToken {
	return internal.NewMissingTokenWithTrivia(trivia, kind, diagnostics...)
}

func NewMissingToken(kind SyntaxKind, diagnostic Diagnostic) *SyntaxToken {
	return internal.NewMissingToken(kind, diagnostic)
}

func CreateMissingSyntaxToken(kind SyntaxKind, diagnostic ...Diagnostic) *SyntaxToken {
	return internal.CreateMissingSyntaxToken(kind, diagnostic...)
}

func CreateMissingSyntaxTokenForKinds(kinds []SyntaxKind) *SyntaxToken {
	return internal.CreateMissingSyntaxTokenForKinds(kinds)
}

func CreateMissingSyntaxTokenForText(text string, diagnostic ...Diagnostic) *SyntaxToken {
	return internal.CreateMissingSyntaxTokenForText(text, diagnostic...)
}

func CreateMissingSyntaxTokenForTexts(texts []string) *SyntaxToken {
	return internal.CreateMissingSyntaxTokenForTexts(texts)
}

func NewTypedSyntaxList(elementType string, elements ...SyntaxElement) *SyntaxList {
	return internal.NewTypedSyntaxList(elementType, elements...)
}

func NewSyntaxList(elements ...SyntaxElement) *SyntaxList {
	return internal.NewSyntaxList(elements...)
}

func NewSyntaxListWithDiagnostics(elements []SyntaxElement, diagnostics ...Diagnostic) *SyntaxList {
	return internal.NewSyntaxListWithDiagnostics(elements, diagnostics...)
}

func EmptySyntaxList() *SyntaxList { return internal.EmptySyntaxList() }

func NewTypedSyntaxListWithDiagnostics(elementType string, elements []SyntaxElement, diagnostics ...Diagnostic) *SyntaxList {
	return internal.NewTypedSyntaxListWithDiagnostics(elementType, elements, diagnostics...)
}

func AsSyntaxListOf[T SyntaxElement](list *SyntaxList) *SyntaxListOf[T] {
	return internal.AsSyntaxListOf[T](list)
}

func AsSeparatedSyntaxListOf[T SyntaxElement](list *SyntaxList) *SeparatedSyntaxListOf[T] {
	return internal.AsSeparatedSyntaxListOf[T](list)
}

func AsSeparatedElementOf[T SyntaxElement](element *SeparatedElement) *SeparatedElementOf[T] {
	return internal.AsSeparatedElementOf[T](element)
}

func NewSeparatedElement(element SyntaxElement, separator *SyntaxToken) *SeparatedElement {
	return internal.NewSeparatedElement(element, separator)
}

func NewSeparatedElementWithDiagnostics(element SyntaxElement, separator *SyntaxToken, diagnostics ...Diagnostic) *SeparatedElement {
	return internal.NewSeparatedElementWithDiagnostics(element, separator, diagnostics...)
}

func EmptySeparatedElement() *SeparatedElement { return internal.EmptySeparatedElement() }

func NewCustomElementDescriptor(name string, hint CompletionHint, optional bool) CustomElementDescriptor {
	return internal.NewCustomElementDescriptor(name, hint, optional)
}

func CustomElementDescriptorFrom(name string, options ...CustomElementDescriptorFromOptions) CustomElementDescriptor {
	return internal.CustomElementDescriptorFrom(name, options...)
}

func CustomElementDescriptorFromHint(hint CompletionHint, isOptional ...bool) CustomElementDescriptor {
	return internal.CustomElementDescriptorFromHint(hint, isOptional...)
}

func CustomElementDescriptorFromIsOptional(isOptional bool) CustomElementDescriptor {
	return internal.CustomElementDescriptorFromIsOptional(isOptional)
}

func CustomNodeGetDefaultShape(count int) []CustomElementDescriptor {
	return internal.CustomNodeGetDefaultShape(count)
}

func NewCustomNode(elements ...SyntaxElement) *CustomNode {
	return internal.NewCustomNode(elements...)
}

func NewCustomNodeWithShape(shape []CustomElementDescriptor, elements []SyntaxElement, diagnostics ...Diagnostic) (*CustomNode, error) {
	return internal.NewCustomNodeWithShape(shape, elements, diagnostics...)
}

func NewGeneratedSyntaxNode(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) (*SyntaxNode, error) {
	return internal.NewGeneratedSyntaxNode(kind, children, diagnostics...)
}

func NewGeneratedSyntaxNodeByName(name string, children []SyntaxElement, diagnostics ...Diagnostic) (*SyntaxNode, error) {
	return internal.NewGeneratedSyntaxNodeByName(name, children, diagnostics...)
}

func NewGeneratedSyntaxNodeByNameWithKind(name string, kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) (*SyntaxNode, error) {
	return internal.NewGeneratedSyntaxNodeByNameWithKind(name, kind, children, diagnostics...)
}

func MustNewGeneratedSyntaxNode(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) *SyntaxNode {
	return internal.MustNewGeneratedSyntaxNode(kind, children, diagnostics...)
}

func MustNewGeneratedSyntaxNodeByName(name string, children []SyntaxElement, diagnostics ...Diagnostic) *SyntaxNode {
	return internal.MustNewGeneratedSyntaxNodeByName(name, children, diagnostics...)
}

func MustNewGeneratedSyntaxNodeByNameWithKind(name string, kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) *SyntaxNode {
	return internal.MustNewGeneratedSyntaxNodeByNameWithKind(name, kind, children, diagnostics...)
}

func GeneratedSyntaxNodeMetadataForKind(kind SyntaxKind) (GeneratedSyntaxNodeMetadata, bool) {
	return internal.GeneratedSyntaxNodeMetadataForKind(kind)
}

func GeneratedSyntaxNodeMetadataForName(name string) (GeneratedSyntaxNodeMetadata, bool) {
	return internal.GeneratedSyntaxNodeMetadataForName(name)
}

func GeneratedSyntaxNodeMetadataAll() []GeneratedSyntaxNodeMetadata {
	return internal.GeneratedSyntaxNodeMetadataAll()
}

func AnySyntaxTokenParser() ResultParser[LexicalToken, *SyntaxToken] {
	return internal.AnySyntaxTokenParser()
}

func AnySyntaxTokenButEndParser() ResultParser[LexicalToken, *SyntaxToken] {
	return internal.AnySyntaxTokenButEndParser()
}

func EndOfTextSyntaxParser() Parser[LexicalToken] { return internal.EndOfTextSyntaxParser() }

func SyntaxTokenKindParser(kind SyntaxKind, items ...CompletionItem) ResultParser[LexicalToken, *SyntaxToken] {
	return internal.SyntaxTokenKindParser(kind, items...)
}

func SyntaxTokenKindsParser(kinds []SyntaxKind, defaultKind ...CompletionKind) ResultParser[LexicalToken, *SyntaxToken] {
	return internal.SyntaxTokenKindsParser(kinds, defaultKind...)
}

func MatchesLexicalText(source Source[LexicalToken], start int, text string) int {
	return internal.MatchesLexicalText(source, start, text)
}

func ProduceSyntaxToken(source Source[LexicalToken], start, length int, asKind ...SyntaxKind) *SyntaxToken {
	return internal.ProduceSyntaxToken(source, start, length, asKind...)
}

func ProduceSyntaxTokenFromText(source Source[LexicalToken], start, length int, text string, asKind ...SyntaxKind) *SyntaxToken {
	return internal.ProduceSyntaxTokenFromText(source, start, length, text, asKind...)
}

func CombinedLexicalTokenText(source Source[LexicalToken], start, length int, includeInnerTrivia ...bool) string {
	return internal.CombinedLexicalTokenText(source, start, length, includeInnerTrivia...)
}

func SyntaxMatchTextParser(text string, asKind ...SyntaxKind) ResultParser[LexicalToken, *SyntaxToken] {
	return internal.SyntaxMatchTextParser(text, asKind...)
}

func SyntaxTokenTextParser(text string, asKind ...SyntaxKind) ResultParser[LexicalToken, *SyntaxToken] {
	return internal.SyntaxTokenTextParser(text, asKind...)
}

func SyntaxTokenTextsParser(texts []string, defaultKind ...CompletionKind) ResultParser[LexicalToken, *SyntaxToken] {
	return internal.SyntaxTokenTextsParser(texts, defaultKind...)
}

func HiddenSyntaxTokenTextParser(text string, asKind ...SyntaxKind) ResultParser[LexicalToken, *SyntaxToken] {
	return internal.HiddenSyntaxTokenTextParser(text, asKind...)
}

func HiddenSyntaxTokenKindParser(kind SyntaxKind) ResultParser[LexicalToken, *SyntaxToken] {
	return internal.HiddenSyntaxTokenKindParser(kind)
}

func HiddenSyntaxTokenTextsParser(texts []string) ResultParser[LexicalToken, *SyntaxToken] {
	return internal.HiddenSyntaxTokenTextsParser(texts)
}

func RequiredSyntaxTokenKindParser(kind SyntaxKind, items ...CompletionItem) ResultParser[LexicalToken, *SyntaxToken] {
	return internal.RequiredSyntaxTokenKindParser(kind, items...)
}

func RequiredSyntaxTokenKindsParser(kinds []SyntaxKind, defaultKind ...CompletionKind) ResultParser[LexicalToken, *SyntaxToken] {
	return internal.RequiredSyntaxTokenKindsParser(kinds, defaultKind...)
}

func RequiredSyntaxTokenTextParser(text string, asKind ...SyntaxKind) ResultParser[LexicalToken, *SyntaxToken] {
	return internal.RequiredSyntaxTokenTextParser(text, asKind...)
}

func RequiredSyntaxTokenTextsParser(texts []string, defaultKind ...CompletionKind) ResultParser[LexicalToken, *SyntaxToken] {
	return internal.RequiredSyntaxTokenTextsParser(texts, defaultKind...)
}

func WithSyntaxCompletion[TElement any](parser ResultParser[LexicalToken, TElement], items ...CompletionItem) ResultParser[LexicalToken, TElement] {
	return internal.WithSyntaxCompletion(parser, items...)
}

func WithSyntaxCompletionHint[TElement any](parser ResultParser[LexicalToken, TElement], hint CompletionHint) ResultParser[LexicalToken, TElement] {
	return internal.WithSyntaxCompletionHint(parser, hint)
}

func CreateSyntaxCompletionItemForKind(kind SyntaxKind, completionKind CompletionKind, priority CompletionPriority, applyText ...string) (CompletionItem, bool) {
	return internal.CreateSyntaxCompletionItemForKind(kind, completionKind, priority, applyText...)
}

func CreateSyntaxCompletionItemForText(text string, completionKind CompletionKind, priority CompletionPriority, applyText ...string) (CompletionItem, bool) {
	return internal.CreateSyntaxCompletionItemForText(text, completionKind, priority, applyText...)
}

func SyntaxCompletionKindForText(text string, defaultKind ...CompletionKind) CompletionKind {
	return internal.SyntaxCompletionKindForText(text, defaultKind...)
}

func SyntaxCompletionKindForKind(kind SyntaxKind, defaultKind ...CompletionKind) CompletionKind {
	return internal.SyntaxCompletionKindForKind(kind, defaultKind...)
}

func ParseAllSyntax[TElement any](parser ResultParser[LexicalToken, TElement], text string, alwaysProduceEndToken ...bool) []TElement {
	return internal.ParseAllSyntax(parser, text, alwaysProduceEndToken...)
}

func ParseAllSyntaxTokens[TElement any](parser ResultParser[LexicalToken, TElement], tokens []LexicalToken) []TElement {
	return internal.ParseAllSyntaxTokens(parser, tokens)
}

func ParseFirstSyntax[TElement any](parser ResultParser[LexicalToken, TElement], text string, alwaysProduceEndToken ...bool) TElement {
	return internal.ParseFirstSyntax(parser, text, alwaysProduceEndToken...)
}

func ParseFirstSyntaxTokens[TElement any](parser ResultParser[LexicalToken, TElement], tokens []LexicalToken) TElement {
	return internal.ParseFirstSyntaxTokens(parser, tokens)
}

func ScanFirstSyntax(parser Parser[LexicalToken], text string, alwaysProduceEndToken ...bool) int {
	return internal.ScanFirstSyntax(parser, text, alwaysProduceEndToken...)
}

func SyntaxListParser[TElement SyntaxElement](
	elementParser ResultParser[LexicalToken, TElement],
	fnMissingElement func(Source[LexicalToken], int) TElement,
	oneOrMore bool,
) ResultParser[LexicalToken, *SyntaxListOf[TElement]] {
	return internal.SyntaxListParser(elementParser, fnMissingElement, oneOrMore)
}

func SyntaxSeparatedListParser[TElement SyntaxElement](
	primaryElementParser ResultParser[LexicalToken, TElement],
	separatorKind SyntaxKind,
	fnMissingElement func(Source[LexicalToken], int) TElement,
	endOfList Parser[LexicalToken],
	oneOrMore bool,
	allowTrailingSeparator bool,
) ResultParser[LexicalToken, *SeparatedSyntaxListOf[TElement]] {
	return internal.SyntaxSeparatedListParser(
		primaryElementParser,
		separatorKind,
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
	return internal.SyntaxSeparatedListFullParser(
		primaryElementParser,
		separatorKind,
		secondaryElementParser,
		fnMissingElement,
		endOfList,
		oneOrMore,
		allowTrailingSeparator,
	)
}

func EndOfCommaListParser() Parser[LexicalToken] {
	return internal.EndOfCommaListParser()
}

func SyntaxCommaListParser[TElement SyntaxElement](
	elementParser ResultParser[LexicalToken, TElement],
	fnMissingElement func(Source[LexicalToken], int) TElement,
	oneOrMore bool,
	allowTrailingComma bool,
	endKinds ...SyntaxKind,
) ResultParser[LexicalToken, *SeparatedSyntaxListOf[TElement]] {
	return internal.SyntaxCommaListParser(
		elementParser,
		fnMissingElement,
		oneOrMore,
		allowTrailingComma,
		endKinds...,
	)
}

func SyntaxOneOrMoreCommaListParser[TElement SyntaxElement](
	elementParser ResultParser[LexicalToken, TElement],
	fnMissingElement func(Source[LexicalToken], int) TElement,
) ResultParser[LexicalToken, *SeparatedSyntaxListOf[TElement]] {
	return internal.SyntaxOneOrMoreCommaListParser(elementParser, fnMissingElement)
}

func MakeSyntaxSeparatedList[TElement SyntaxElement](elements []any) *SeparatedSyntaxListOf[TElement] {
	return internal.MakeSyntaxSeparatedList[TElement](elements)
}

func WithSyntaxExamples[TElement any](parser ResultParser[LexicalToken, TElement], values ...string) ResultParser[LexicalToken, TElement] {
	return internal.WithSyntaxExamples(parser, values...)
}
