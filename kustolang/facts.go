package kustolang

// Ported from src/Kusto.Language/Parser/KustoFacts.cs,
// src/Kusto.Language/Parser/TokenParser.cs, and
// src/Kusto.Language/Syntax/SyntaxFacts.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

const (
	MultiLineStringQuote          = internal.MultiLineStringQuote
	AlternateMultiLineStringQuote = internal.AlternateMultiLineStringQuote
	KustoWindowsNet               = internal.KustoWindowsNet
)

// KustoFacts exposes the source-compatible inventories used by the lexer,
// grammar, completer, and semantic analyzer.
var KustoFacts = internal.KustoFacts

func KnownQueryOperatorParameterNames() []string {
	return internal.KnownQueryOperatorParameterNames()
}

func HasPipedInput(queryOperator QueryOperator) bool {
	return internal.HasPipedInput(queryOperator)
}

func CanBeIdentifier(name string, dialect ...KustoDialect) bool {
	return internal.CanBeIdentifier(name, dialect...)
}

func BracketNameIfNecessary(name string, dialect ...KustoDialect) string {
	return internal.BracketNameIfNecessary(name, dialect...)
}

func GetBracketedName(name string) string { return internal.GetBracketedName(name) }

func GetStringLiteral(text string) string { return internal.GetStringLiteral(text) }

func GetSingleQuotedStringLiteral(text string) string {
	return internal.GetSingleQuotedStringLiteral(text)
}

func GetDoubleQuotedStringLiteral(text string) string {
	return internal.GetDoubleQuotedStringLiteral(text)
}

func GetMultiLineStringLiteral(text string) string {
	return internal.GetMultiLineStringLiteral(text)
}

func GetStringLiteralValue(literal string) string {
	return internal.GetStringLiteralValue(literal)
}

func GetExpressionResultName(expression Expression, defaultName string, rowScope *TableSymbol, globals ...*GlobalState) string {
	return internal.GetExpressionResultName(expression, defaultName, rowScope, globals...)
}

func IsHostName(name, hostName string) bool { return internal.IsHostName(name, hostName) }

func GetHostName(clusterURIOrName string) string {
	return internal.GetHostName(clusterURIOrName)
}

func GetHostAndPath(clusterURIOrName string) (string, string) {
	return internal.GetHostAndPath(clusterURIOrName)
}

func GetFullHostName(hostname, defaultDomainSuffix string) string {
	return internal.GetFullHostName(hostname, defaultDomainSuffix)
}

func HasShortHostName(hostname, defaultDomainSuffix string) bool {
	return internal.HasShortHostName(hostname, defaultDomainSuffix)
}

func IsShortHostName(name, hostName, defaultDomainSuffix string) bool {
	return internal.IsShortHostName(name, hostName, defaultDomainSuffix)
}

func IsPossibleShortHostName(name string) bool {
	return internal.IsPossibleShortHostName(name)
}

func GetShortHostName(hostName, defaultDomainSuffix string) string {
	return internal.GetShortHostName(hostName, defaultDomainSuffix)
}

func KustoFactsMatches(pattern, text string, ignoreCase ...bool) bool {
	return internal.KustoFactsMatches(pattern, text, ignoreCase...)
}

func GetSyntaxTabularity(element SyntaxElement, globals ...*GlobalState) Tabularity {
	return internal.GetSyntaxTabularity(element, globals...)
}

func GetLiteralType(kind SyntaxKind) *ScalarSymbol { return internal.GetLiteralType(kind) }

func SyntaxOperatorKind(kind SyntaxKind) OperatorKind {
	return internal.SyntaxOperatorKind(kind)
}

func SyntaxKindFromTextRange(text string, offset, length int) (SyntaxKind, bool) {
	return internal.SyntaxKindFromTextRange(text, offset, length)
}

func SyntaxKinds(category SyntaxCategory) []SyntaxKind {
	return internal.SyntaxKinds(category)
}

func SyntaxKindsWithFixedText() []SyntaxKind { return internal.SyntaxKindsWithFixedText() }

func SyntaxKeywordTexts() []string { return internal.SyntaxKeywordTexts() }

func SyntaxPunctuationTexts() []string { return internal.SyntaxPunctuationTexts() }

func SyntaxOperatorTexts() []string { return internal.SyntaxOperatorTexts() }

func IsPunctuationKind(kind SyntaxKind) bool { return internal.IsPunctuationKind(kind) }

func IsTypeKeyword(kind SyntaxKind) bool { return internal.IsTypeKeyword(kind) }

func SyntaxKindCanBeIdentifier(kind SyntaxKind) bool {
	return internal.SyntaxKindCanBeIdentifier(kind)
}

func IsKeywordText(text string) bool { return internal.IsKeywordText(text) }

func IsKeywordThatCanBeIdentifier(text string) bool {
	return internal.IsKeywordThatCanBeIdentifier(text)
}

func IsEndOfText(kind SyntaxKind) bool { return internal.IsEndOfText(kind) }

func ScanBooleanLiteral(text string, start ...int) int {
	return internal.ScanBooleanLiteral(text, firstStart(start))
}

func ScanWhitespace(text string, start int) int { return internal.ScanWhitespace(text, start) }

func ScanTrivia(text string, start int) int { return internal.ScanTrivia(text, start) }

func ParseTrivia(text string, start int) string { return internal.ParseTrivia(text, start) }

func ScanComment(text string, start int) int { return internal.ScanComment(text, start) }

func ScanIdentifier(text string, start ...int) int {
	return internal.ScanIdentifier(text, firstStart(start))
}

func ScanLongLiteral(text string, start ...int) int {
	return internal.ScanLongLiteral(text, firstStart(start))
}

func ScanRealLiteral(text string, start ...int) int {
	return internal.ScanRealLiteral(text, firstStart(start))
}

func ScanTimespanLiteral(text string, start ...int) int {
	return internal.ScanTimespanLiteral(text, firstStart(start))
}

func ScanRawGuidLiteral(text string, start int) int {
	return internal.ScanRawGuidLiteral(text, start)
}

func ScanStringLiteral(text string, start int, failWhenMissingEndQuote ...bool) int {
	return internal.ScanStringLiteral(text, start, failWhenMissingEndQuote...)
}

func ScanClientParameter(text string, start int) int {
	return internal.ScanClientParameter(text, start)
}

func ScanClientParameterParts(text string, start int) (length, nameStart, nameLength, indexStart, indexLength int) {
	return internal.ScanClientParameterParts(text, start)
}

func IsCommentLine(text string, lineStart ...int) bool {
	return internal.IsCommentLine(text, firstStart(lineStart))
}

func GetCommentLineText(text string, lineStart ...int) string {
	return internal.CommentLineText(text, firstStart(lineStart))
}

func GetCommentTexts(text string) []string { return internal.CommentTexts(text) }

func TrimCommentAndWhitespaceLines(text string) string {
	return internal.TrimCommentAndWhitespaceLines(text)
}

func StripLineComments(text string) string { return internal.StripLineComments(text) }

func TryGetCommentSpan(trivia string, position int) (start, length int, ok bool) {
	return internal.TryGetCommentSpan(trivia, position)
}

func IsCommandStartLine(text string, lineStart int) bool {
	return internal.IsCommandStartLine(text, lineStart)
}

func GetCommandBlockStarts(text string, lineStarts ...[]int) []int {
	return internal.CommandBlockStarts(text, lineStarts...)
}

func GetCommandBlockTexts(text string) []string { return internal.CommandBlockTexts(text) }

func GetCommandTexts(text string) []string { return internal.CommandTexts(text) }

func GetKustoBlockStarts(text string, lineStarts []int) []int {
	return internal.KustoBlockStarts(text, lineStarts)
}

func GetLineSeparatedBlockStarts(text string, lineStarts []int) []int {
	return internal.LineSeparatedBlockStarts(text, lineStarts)
}

func RemoveDuplicateBlockStarts(blockStarts []int) []int {
	return internal.RemoveDuplicateBlockStarts(blockStarts)
}

func RemoveInvalidKustoBlockStarts(blockStarts []int, text string) []int {
	return internal.RemoveInvalidKustoBlockStarts(blockStarts, text)
}

func RemoveInvalidBlockStarts(blockStarts []int, invalidRanges []TextRange) []int {
	return internal.RemoveInvalidBlockStarts(blockStarts, invalidRanges)
}

func GetInvalidKustoBlockStartRanges(text string) []TextRange {
	return internal.InvalidKustoBlockStartRanges(text)
}

func IsWhitespace(ch rune) bool { return internal.IsWhitespace(ch) }

func IsLetter(ch rune) bool { return internal.IsLetter(ch) }

func IsDigit(ch rune) bool { return internal.IsDigit(ch) }

func IsLetterOrDigit(ch rune) bool { return internal.IsLetterOrDigit(ch) }

func IsHexDigit(ch rune) bool { return internal.IsHexDigit(ch) }

func IsLineBreakStart(ch rune) bool { return internal.IsLineBreakStart(ch) }

func IsWhitespaceOnly(text string, startAndLength ...int) bool {
	return internal.IsWhitespaceOnly(text, startAndLength...)
}

func IsWhitespaceLine(text string, start int) bool {
	return internal.IsWhitespaceLine(text, start)
}

func GetWhitespace(text string, start int) string { return internal.GetWhitespace(text, start) }

func GetWhitespaceCount(text string, start int) int {
	return internal.WhitespaceCount(text, start)
}

func GetWhitespaceCountBefore(text string, start int) int {
	return internal.WhitespaceCountBefore(text, start)
}

func GetLineBreakLength(text string, position int) int {
	return internal.LineBreakLength(text, position)
}

func HasLineBreaks(text string) bool { return internal.HasLineBreaks(text) }

func GetNextLineBreakStart(text string, start int) int {
	return internal.NextLineBreakStart(text, start)
}

func GetNextLineBreakEnd(text string, start int) int {
	return internal.NextLineBreakEnd(text, start)
}

func GetLineEnd(text string, position int) int { return internal.LineEnd(text, position) }

func GetLineStart(text string, position int) int { return internal.LineStart(text, position) }

func TryGetPreviousLineEnd(text string, position int) (previousLineEnd int, ok bool) {
	return internal.TryGetPreviousLineEnd(text, position)
}

func TryGetNextLineStart(text string, position int) (nextLineStart int, ok bool) {
	return internal.TryGetNextLineStart(text, position)
}

func GetFirstLineBreakStart(text string) int { return internal.FirstLineBreakStart(text) }

func GetFirstLineBreakEnd(text string) int { return internal.FirstLineBreakEnd(text) }

func GetLastLineBreakStart(text string, start ...int) int {
	return internal.LastLineBreakStart(text, start...)
}

func GetLastLineBreakEnd(text string, start ...int) int {
	return internal.LastLineBreakEnd(text, start...)
}

func IsBlankLine(text string, lineStart int) bool { return internal.IsBlankLine(text, lineStart) }

func GetPositionAfterLeadingWhitespace(text string, lineStart int) int {
	return internal.PositionAfterLeadingWhitespace(text, lineStart)
}

func GetLineLength(text string, lineStart int, includeLineBreak ...bool) int {
	return internal.LineLength(text, lineStart, includeLineBreak...)
}

func GetNextLineStart(text string, start int) int { return internal.NextLineStart(text, start) }

func GetLineStarts(text string) []int { return internal.LineStarts(text) }

// AppendLineStarts is the Go spelling of TextFacts.GetLineStarts(text, list).
func AppendLineStarts(text string, lineStarts *[]int) {
	if lineStarts == nil {
		panic("kustolang: nil line-start destination")
	}
	*lineStarts = append(*lineStarts, internal.LineStarts(text)...)
}

func TryGetLineStart(text string, line int) (lineStart int, ok bool) {
	return internal.TryGetTextLineStart(text, line)
}

func TryGetTextLineAndOffset(text string, position int) (line, lineOffset int, ok bool) {
	return internal.TryGetTextLineAndOffset(text, position)
}

func TryGetTextPosition(text string, line, lineOffset int) (position int, ok bool) {
	return internal.TryGetTextPosition(text, line, lineOffset)
}

func TryGetLineAndOffset(lineStarts []int, position int) (line, lineOffset int, ok bool) {
	return internal.TryGetLineAndOffset(lineStarts, position)
}

func TryGetPosition(lineStarts []int, line, lineOffset int) (position int, ok bool) {
	return internal.TryGetPosition(lineStarts, line, lineOffset)
}

func TrimEnd(text string, startAndLength ...int) int {
	return internal.TrimEnd(text, startAndLength...)
}

func TrimRangeStart(text string, start, length int) (int, int) {
	return internal.TrimRangeStart(text, start, length)
}

func TrimRangeEnd(text string, start, length int) int {
	return internal.TrimRangeEnd(text, start, length)
}

func TrimRange(text string, start, length int) (int, int) {
	return internal.TrimRange(text, start, length)
}

func ExpandRangeToStartOfFirstLine(text string, start, length int) (int, int) {
	return internal.ExpandRangeToStartOfFirstLine(text, start, length)
}

func ExpandRangeToEndOfLastLine(text string, start, length int) int {
	return internal.ExpandRangeToEndOfLastLine(text, start, length)
}

func ExpandRangeToStartAndEndOfLines(text string, start, length int) (int, int) {
	return internal.ExpandRangeToStartAndEndOfLines(text, start, length)
}

func GetIndentationLength(text string, position int) int {
	return internal.IndentationLength(text, position)
}

func GetIndentationText(text string, position int) string {
	return internal.IndentationText(text, position)
}

func GetLineText(text string, line int) string { return internal.LineText(text, line) }

func GetLineTexts(text string) []string { return internal.LineTexts(text) }

func firstStart(start []int) int {
	if len(start) > 0 {
		return start[0]
	}
	return 0
}
