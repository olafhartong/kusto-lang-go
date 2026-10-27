package kustolang

import (
	"strings"
	"unicode/utf8"
)

// Ported from src/Kusto.Language/Parser/LexicalToken.cs and the lexical
// scanner behavior in TokenParser/CharScanners. This is an incremental lexer:
// it preserves trivia and recognizes the token set currently represented in
// syntax_facts.go.

type LexicalToken struct {
	Kind        SyntaxKind
	Trivia      string
	Text        string
	Diagnostics []Diagnostic
}

func NewLexicalToken(kind SyntaxKind, trivia, text string, diagnostics ...Diagnostic) LexicalToken {
	if trivia == "" {
		trivia = ""
	}
	if text == "" {
		text = ""
	}
	if diagnostics == nil {
		diagnostics = []Diagnostic{}
	}
	return LexicalToken{Kind: kind, Trivia: trivia, Text: text, Diagnostics: diagnostics}
}

func (t LexicalToken) Length() int {
	return len(t.Trivia) + len(t.Text)
}

// ByteLength is the native scanner width used to advance through Go strings.
func (t LexicalToken) ByteLength() int {
	return t.Length()
}

// UTF16Length matches LexicalToken.Length in Kusto.Language, where .NET
// strings are indexed in UTF-16 code units.
func (t LexicalToken) UTF16Length() int {
	return utf16TextWidth(t.Trivia) + utf16TextWidth(t.Text)
}

func (t LexicalToken) DebugText() string {
	if t.Text != "" {
		return t.Text
	}
	return SyntaxText(t.Kind)
}

func Lex(text string) []LexicalToken {
	return LexWithOptions(text, DefaultParseOptions)
}

// LexWithOptions tokenizes text using the source-compatible parser options.
func LexWithOptions(text string, options ParseOptions) []LexicalToken {
	return lex(text, true, options, CancellationToken{})
}

// LexWithOptionsAndCancellation tokenizes text while periodically observing
// cancellation both between tokens and during long single-token scans.
func LexWithOptionsAndCancellation(text string, options ParseOptions, cancellationToken CancellationToken) []LexicalToken {
	return lex(text, true, options, cancellationToken)
}

// ParseToken returns the lexical token beginning at the native byte offset.
func ParseToken(text string, start int) *LexicalToken {
	return ParseTokenWithOptions(text, start, DefaultParseOptions)
}

// ParseTokenWithOptions returns one lexical token beginning at the native byte
// offset, including trivia encountered at that position.
func ParseTokenWithOptions(text string, start int, options ParseOptions) *LexicalToken {
	return parseLexicalToken(text, start, true, options)
}

// ParseTokens is the source-shaped name for Lex.
func ParseTokens(text string) []LexicalToken {
	return Lex(text)
}

// ParseTokensWithOptions is the source-shaped name for LexWithOptions.
func ParseTokensWithOptions(text string, options ParseOptions) []LexicalToken {
	return LexWithOptions(text, options)
}

const lexerCancellationPollInterval = 256

type lexerCancellationPoller struct {
	cancellation CancellationToken
	work         uint64
}

func (p *lexerCancellationPoller) poll() {
	if p == nil {
		return
	}
	p.work++
	if p.work%lexerCancellationPollInterval == 0 {
		p.cancellation.ThrowIfCancellationRequested()
	}
}

func lex(text string, allowAlternateMultiLineString bool, options ParseOptions, cancellationToken CancellationToken) []LexicalToken {
	var tokens []LexicalToken
	pos := 0
	poller := &lexerCancellationPoller{cancellation: cancellationToken}
	for {
		cancellationToken.ThrowIfCancellationRequested()
		token, ok := scanLexicalTokenWithCancellation(text, pos, allowAlternateMultiLineString, options, poller)
		if !ok {
			return tokens
		}
		if tokens == nil {
			// Estimate from the remaining text after the first token and its
			// trivia, so a long initial comment or literal does not inflate the
			// buffer. Cap speculative capacity for large inputs.
			capacity := min(1+(len(text)-token.Length())/4, 1024)
			tokens = make([]LexicalToken, 0, capacity)
		}
		tokens = append(tokens, token)
		if token.Kind == SyntaxKindEndOfTextToken {
			return tokens
		}
		pos += token.Length()
	}
}

func parseLexicalToken(text string, start int, allowAlternateMultiLineString bool, options ParseOptions) *LexicalToken {
	token, ok := scanLexicalTokenWithCancellation(text, start, allowAlternateMultiLineString, options, nil)
	if !ok {
		return nil
	}
	return &token
}

func scanLexicalTokenWithCancellation(text string, start int, allowAlternateMultiLineString bool, options ParseOptions, poller *lexerCancellationPoller) (LexicalToken, bool) {
	if start < 0 {
		start = 0
	}
	if start > len(text) {
		start = len(text)
	}
	pos := scanTriviaWithCancellation(text, start, poller)
	trivia := text[start:pos]
	if pos >= len(text) {
		if trivia == "" && !options.AlwaysProduceEndToken {
			return LexicalToken{}, false
		}
		token := NewLexicalToken(SyntaxKindEndOfTextToken, trivia, "")
		return token, true
	}
	kind, tokenText, diagnostics, next := scanTokenWithCancellation(text, pos, allowAlternateMultiLineString, options.AllowLiteralsWithLineBreaks, poller)
	token := NewLexicalToken(kind, trivia, tokenText, diagnostics...)
	if next <= pos {
		return LexicalToken{}, false
	}
	return token, true
}

func scanTriviaWithCancellation(text string, pos int, poller *lexerCancellationPoller) int {
	for pos < len(text) {
		poller.poll()
		ch, size := utf8.DecodeRuneInString(text[pos:])
		if IsWhitespace(ch) {
			pos += size
			continue
		}
		if isCommentStart(text, pos) {
			for pos < len(text) && LineBreakLength(text, pos) == 0 {
				poller.poll()
				_, size := utf8.DecodeRuneInString(text[pos:])
				pos += size
			}
			if pos < len(text) {
				pos += LineBreakLength(text, pos)
			}
			continue
		}
		break
	}
	return pos
}

func scanTokenWithCancellation(text string, pos int, allowAlternateMultiLineString, allowLiteralLineBreaks bool, poller *lexerCancellationPoller) (SyntaxKind, string, []Diagnostic, int) {
	if tokenText, next, ok := scanGuidLiteralWithCancellation(text, pos, poller); ok {
		return SyntaxKindGuidLiteralToken, tokenText, nil, next
	}
	if tokenText, next, ok := scanRawGuidLiteral(text, pos); ok {
		return SyntaxKindRawGuidLiteralToken, tokenText, nil, next
	}
	if kind, tokenText, diagnostics, next, ok := scanGooLiteralWithCancellation(text, pos, allowLiteralLineBreaks, poller); ok {
		return kind, tokenText, diagnostics, next
	}

	if scanStringLiteralStart(text, pos, allowAlternateMultiLineString) {
		return scanStringLiteralWithCancellation(text, pos, allowAlternateMultiLineString, poller)
	}
	if text[pos] == '#' {
		end := pos
		for end < len(text) && LineBreakLength(text, end) == 0 {
			poller.poll()
			_, size := utf8.DecodeRuneInString(text[end:])
			end += size
		}
		return SyntaxKindDirectiveToken, text[pos:end], nil, end
	}

	if kind, tokenText, ok := scanFixedTextToken(text, pos); ok {
		return kind, tokenText, nil, pos + len(tokenText)
	}

	ch, size := utf8.DecodeRuneInString(text[pos:])
	if IsDigit(ch) {
		return scanNumberLiteralWithCancellation(text, pos, poller)
	}
	if isIdentifierStart(ch) {
		end := pos + size
		for end < len(text) {
			poller.poll()
			part, partSize := utf8.DecodeRuneInString(text[end:])
			if !isIdentifierPart(part) {
				break
			}
			end += partSize
		}
		tokenText := text[pos:end]
		if isBooleanLiteralText(tokenText) {
			return SyntaxKindBooleanLiteralToken, tokenText, nil, end
		}
		if len(tokenText) <= maxLexicalKeywordLength {
			if kind, ok := SyntaxKindFromText(tokenText); ok && IsKeywordKind(kind) {
				return kind, tokenText, nil, end
			}
		}
		return SyntaxKindIdentifierToken, tokenText, nil, end
	}

	badText := text[pos : pos+size]
	return SyntaxKindBadToken, badText, []Diagnostic{DiagnosticUnexpectedCharacter(badText)}, pos + size
}

func isBooleanLiteralText(text string) bool {
	switch text {
	case "true", "True", "TRUE", "false", "False", "FALSE":
		return true
	default:
		return false
	}
}

func scanGuidLiteralWithCancellation(text string, pos int, poller *lexerCancellationPoller) (string, int, bool) {
	const prefix = "guid("
	if pos+len(prefix) > len(text) || !ordinalEqualFold(text[pos:pos+len(prefix)], prefix) {
		return "", pos, false
	}
	innerStart := pos + len(prefix)
	end := innerStart
	for end < len(text) && text[end] != ')' {
		poller.poll()
		end++
	}
	if end >= len(text) {
		return "", pos, false
	}
	trimmedStart, trimmedEnd := trimSpaceBoundsWithCancellation(text, innerStart, end, poller)
	inner := text[trimmedStart:trimmedEnd]
	if len(inner) >= 2 && ((inner[0] == '\'' && inner[len(inner)-1] == '\'') || (inner[0] == '"' && inner[len(inner)-1] == '"')) {
		trimmedStart, trimmedEnd = trimSpaceBoundsWithCancellation(inner, 1, len(inner)-1, poller)
		inner = inner[trimmedStart:trimmedEnd]
	}
	if !isGuidText(inner) {
		return "", pos, false
	}
	return text[pos : end+1], end + 1, true
}

func scanRawGuidLiteral(text string, pos int) (string, int, bool) {
	if pos+36 > len(text) {
		return "", pos, false
	}
	candidate := text[pos : pos+36]
	if !isGuidText(candidate) {
		return "", pos, false
	}
	return candidate, pos + 36, true
}

func isGuidText(text string) bool {
	if len(text) != 36 {
		return false
	}
	for i, ch := range text {
		switch i {
		case 8, 13, 18, 23:
			if ch != '-' {
				return false
			}
		default:
			if !isHexDigit(ch) {
				return false
			}
		}
	}
	return true
}

func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

const maxLexicalKeywordLength = 128

func trimSpaceBoundsWithCancellation(text string, start, end int, poller *lexerCancellationPoller) (int, int) {
	for start < end {
		poller.poll()
		ch, size := utf8.DecodeRuneInString(text[start:end])
		if !IsWhitespace(ch) {
			break
		}
		start += size
	}
	for end > start {
		poller.poll()
		ch, size := utf8.DecodeLastRuneInString(text[start:end])
		if !IsWhitespace(ch) {
			break
		}
		end -= size
	}
	return start, end
}

func scanGooLiteralWithCancellation(text string, pos int, allowLineBreaks bool, poller *lexerCancellationPoller) (SyntaxKind, string, []Diagnostic, int, bool) {
	if pos >= len(text) {
		return SyntaxKindNone, "", nil, pos, false
	}
	ch, size := utf8.DecodeRuneInString(text[pos:])
	if !isIdentifierStart(ch) {
		return SyntaxKindNone, "", nil, pos, false
	}
	end := pos + size
	for end < len(text) {
		poller.poll()
		ch, size = utf8.DecodeRuneInString(text[end:])
		if !isIdentifierPart(ch) {
			break
		}
		end += size
	}
	if end >= len(text) || text[end] != '(' {
		return SyntaxKindNone, "", nil, pos, false
	}
	if end-pos > maxLexicalKeywordLength {
		return SyntaxKindNone, "", nil, pos, false
	}
	keywordKind, ok := SyntaxKindFromText(text[pos:end])
	if !ok {
		return SyntaxKindNone, "", nil, pos, false
	}
	literalKind := gooLiteralTokenKind(keywordKind)
	if literalKind == SyntaxKindNone {
		return SyntaxKindNone, "", nil, pos, false
	}
	next, closed := scanGooLiteralBodyWithCancellation(text, end, allowLineBreaks, poller)
	if next <= end {
		return SyntaxKindNone, "", nil, pos, false
	}
	var diagnostics []Diagnostic
	if !closed {
		diagnostics = []Diagnostic{DiagnosticMissingText(")").WithLocationKind(DiagnosticLocationRelativeEnd)}
	}
	return literalKind, text[pos:next], diagnostics, next, true
}

func gooLiteralTokenKind(kind SyntaxKind) SyntaxKind {
	switch kind {
	case SyntaxKindLongKeyword, SyntaxKindInt64Keyword:
		return SyntaxKindLongLiteralToken
	case SyntaxKindIntKeyword, SyntaxKindInt32Keyword:
		return SyntaxKindIntLiteralToken
	case SyntaxKindRealKeyword, SyntaxKindDoubleKeyword:
		return SyntaxKindRealLiteralToken
	case SyntaxKindDecimalKeyword:
		return SyntaxKindDecimalLiteralToken
	case SyntaxKindBoolKeyword:
		return SyntaxKindBooleanLiteralToken
	case SyntaxKindDateTimeKeyword, SyntaxKindDateKeyword:
		return SyntaxKindDateTimeLiteralToken
	case SyntaxKindTimeKeyword, SyntaxKindTimespanKeyword:
		return SyntaxKindTimespanLiteralToken
	case SyntaxKindGuidKeyword:
		return SyntaxKindGuidLiteralToken
	default:
		return SyntaxKindNone
	}
}

func scanGooLiteralBodyWithCancellation(text string, pos int, allowLineBreaks bool, poller *lexerCancellationPoller) (int, bool) {
	if pos >= len(text) || text[pos] != '(' {
		return pos, false
	}
	pos++
	for pos < len(text) && text[pos] != ')' && (allowLineBreaks || LineBreakLength(text, pos) == 0) {
		poller.poll()
		pos++
	}
	if pos < len(text) && text[pos] == ')' {
		pos++
		return pos, true
	}
	return pos, false
}

func scanFixedTextToken(text string, pos int) (SyntaxKind, string, bool) {
	if pos >= len(text) {
		return SyntaxKindNone, "", false
	}
	remaining := text[pos:]
	for _, data := range syntaxTextsByFirstByte[text[pos]] {
		if len(data.Text) > len(remaining) {
			continue
		}
		if remaining[:len(data.Text)] != data.Text {
			continue
		}
		if data.Category == SyntaxCategoryKeyword && !keywordBoundary(text, pos, len(data.Text)) {
			continue
		}
		return data.Kind, remaining[:len(data.Text)], true
	}
	return SyntaxKindNone, "", false
}

func scanStringLiteralStart(text string, pos int, allowAlternateMultiLineString bool) bool {
	if pos >= len(text) {
		return false
	}
	ch := text[pos]
	if ch == 'h' || ch == 'H' {
		pos++
		if pos >= len(text) {
			return false
		}
		ch = text[pos]
	}
	if ch == '@' {
		pos++
		if pos >= len(text) {
			return false
		}
		ch = text[pos]
	}
	return isStringLiteralQuote(ch) || strings.HasPrefix(text[pos:], "```") ||
		(allowAlternateMultiLineString && strings.HasPrefix(text[pos:], "~~~"))
}

func scanStringLiteralWithCancellation(text string, pos int, allowAlternateMultiLineString bool, poller *lexerCancellationPoller) (SyntaxKind, string, []Diagnostic, int) {
	start := pos
	if pos < len(text) && (text[pos] == 'h' || text[pos] == 'H') {
		pos++
	}
	isVerbatim := false
	if pos < len(text) && text[pos] == '@' {
		isVerbatim = true
		pos++
	}
	if pos >= len(text) {
		return SyntaxKindStringLiteralToken, text[start:pos], []Diagnostic{missingStringTerminatorDiagnostic("")}, pos
	}
	if strings.HasPrefix(text[pos:], "```") {
		end, ok := scanMultiLineStringLiteralWithCancellation(text, pos, "```", poller)
		if ok {
			return SyntaxKindStringLiteralToken, text[start:end], nil, end
		}
		return SyntaxKindStringLiteralToken, text[start:end], []Diagnostic{missingStringTerminatorDiagnostic("```")}, end
	}
	if allowAlternateMultiLineString && strings.HasPrefix(text[pos:], "~~~") {
		end, ok := scanMultiLineStringLiteralWithCancellation(text, pos, "~~~", poller)
		if ok {
			return SyntaxKindStringLiteralToken, text[start:end], nil, end
		}
		return SyntaxKindStringLiteralToken, text[start:end], []Diagnostic{missingStringTerminatorDiagnostic("~~~")}, end
	}
	quote := text[pos]
	if !isStringLiteralQuote(quote) {
		badText := text[start : start+1]
		return SyntaxKindBadToken, badText, []Diagnostic{DiagnosticUnexpectedCharacter(badText)}, start + 1
	}
	end := pos + 1
	for end < len(text) {
		poller.poll()
		ch := text[end]
		if isVerbatim && ch == quote && end+1 < len(text) && text[end+1] == quote {
			end += 2
			continue
		}
		if !isVerbatim && ch == '\\' {
			escapeLen := scanStringEscape(text, end)
			if escapeLen > 0 {
				end += escapeLen
				continue
			}
			break
		}
		if ch == quote {
			end++
			return SyntaxKindStringLiteralToken, text[start:end], nil, end
		}
		if LineBreakLength(text, end) > 0 {
			break
		}
		end++
	}
	return SyntaxKindStringLiteralToken, text[start:end], []Diagnostic{missingStringTerminatorDiagnostic(string(quote))}, end
}

func missingStringTerminatorDiagnostic(quote string) Diagnostic {
	if quote == "" {
		quote = "string"
	}
	return DiagnosticMissingText(quote).WithLocationKind(DiagnosticLocationRelativeEnd)
}

func isStringLiteralQuote(ch byte) bool {
	return ch == '\'' || ch == '"'
}

func scanMultiLineStringLiteralWithCancellation(text string, pos int, quote string, poller *lexerCancellationPoller) (int, bool) {
	start := pos + len(quote)
	for end := start; end+len(quote) <= len(text); end++ {
		poller.poll()
		if text[end:end+len(quote)] == quote {
			return end + len(quote), true
		}
	}
	return len(text), false
}

func scanStringEscape(text string, pos int) int {
	if pos >= len(text) || text[pos] != '\\' || pos+1 >= len(text) {
		return -1
	}
	switch text[pos+1] {
	case '\\', '\'', '"', 'a', 'b', 'f', 'n', 'r', 't', 'v':
		return 2
	case 'u':
		if hasHexDigits(text, pos+2, 4) {
			return 6
		}
	case 'U':
		if hasHexDigits(text, pos+2, 8) {
			return 10
		}
	case 'x':
		if hasHexDigits(text, pos+2, 2) {
			return 4
		}
	default:
		if length := scanOctalCode(text, pos+1); length > 0 {
			return length + 1
		}
	}
	return -1
}

func hasHexDigits(text string, pos int, length int) bool {
	if pos+length > len(text) {
		return false
	}
	for i := 0; i < length; i++ {
		if !isHexDigit(rune(text[pos+i])) {
			return false
		}
	}
	return true
}

func scanOctalCode(text string, pos int) int {
	if pos >= len(text) || text[pos] < '0' || text[pos] > '7' {
		return -1
	}
	if pos+1 >= len(text) || text[pos+1] < '0' || text[pos+1] > '7' {
		return 1
	}
	if pos+2 >= len(text) || text[pos+2] < '0' || text[pos+2] > '7' || text[pos] > '3' {
		return 2
	}
	return 3
}

func scanNumberLiteralWithCancellation(text string, pos int, poller *lexerCancellationPoller) (SyntaxKind, string, []Diagnostic, int) {
	if length := scanRealLiteralWithCancellation(text, pos, poller); length > 0 {
		return SyntaxKindRealLiteralToken, text[pos : pos+length], nil, pos + length
	}
	if length := scanTimespanLiteralWithCancellation(text, pos, poller); length > 0 {
		return SyntaxKindTimespanLiteralToken, text[pos : pos+length], nil, pos + length
	}
	if length := scanLongLiteralWithCancellation(text, pos, poller); length > 0 {
		return SyntaxKindLongLiteralToken, text[pos : pos+length], nil, pos + length
	}
	if length := scanDigitStartIdentifierWithCancellation(text, pos, poller); length > 0 {
		return SyntaxKindIdentifierToken, text[pos : pos+length], nil, pos + length
	}
	_, size := utf8.DecodeRuneInString(text[pos:])
	badText := text[pos : pos+size]
	return SyntaxKindBadToken, badText, []Diagnostic{DiagnosticUnexpectedCharacter(badText)}, pos + size
}

func scanRealLiteralWithCancellation(text string, pos int, poller *lexerCancellationPoller) int {
	digitLen := scanDigitsWithCancellation(text, pos, poller)
	if digitLen <= 0 {
		return -1
	}
	end := pos + digitLen
	if end < len(text) && text[end] == '.' && (end+1 >= len(text) || text[end+1] != '.' || (end+2 < len(text) && text[end+2] == '.')) {
		end++
		if fractionLen := scanDigitsWithCancellation(text, end, poller); fractionLen > 0 {
			end += fractionLen
		}
		if exponentLen := scanExponentWithCancellation(text, end, poller); exponentLen > 0 {
			end += exponentLen
		}
	} else {
		exponentLen := scanExponentWithCancellation(text, end, poller)
		if exponentLen <= 0 {
			return -1
		}
		end += exponentLen
	}
	if end < len(text) {
		ch, _ := utf8.DecodeRuneInString(text[end:])
		if isIdentifierPart(ch) {
			return -1
		}
	}
	return end - pos
}

func scanExponent(text string, pos int) int {
	return scanExponentWithCancellation(text, pos, nil)
}

func scanExponentWithCancellation(text string, pos int, poller *lexerCancellationPoller) int {
	end := pos
	if end >= len(text) || (text[end] != 'e' && text[end] != 'E') {
		return -1
	}
	end++
	if end < len(text) && (text[end] == '+' || text[end] == '-') {
		end++
	}
	digitLen := scanDigitsWithCancellation(text, end, poller)
	if digitLen <= 0 {
		return -1
	}
	return end + digitLen - pos
}

func scanTimespanLiteral(text string, pos int) int {
	return scanTimespanLiteralWithCancellation(text, pos, nil)
}

func scanTimespanLiteralWithCancellation(text string, pos int, poller *lexerCancellationPoller) int {
	numberLen := scanDigitsWithCancellation(text, pos, poller)
	if numberLen <= 0 {
		return -1
	}
	end := pos + numberLen
	if end < len(text) && text[end] == '.' {
		fractionLen := scanDigitsWithCancellation(text, end+1, poller)
		if fractionLen > 0 {
			end += fractionLen + 1
		}
	}
	if suffixLen := scanTimespanSuffix(text, end); suffixLen > 0 {
		return end + suffixLen - pos
	}
	return -1
}

func scanLongLiteralWithCancellation(text string, pos int, poller *lexerCancellationPoller) int {
	if length := scanHexIntegerLiteralWithCancellation(text, pos, poller); length > 0 {
		return length
	}
	digitLen := scanDigitsWithCancellation(text, pos, poller)
	if digitLen > 0 {
		end := pos + digitLen
		if end >= len(text) {
			return digitLen
		}
		ch, _ := utf8.DecodeRuneInString(text[end:])
		if !isIdentifierPart(ch) {
			return digitLen
		}
	}
	return -1
}

func scanHexIntegerLiteralWithCancellation(text string, pos int, poller *lexerCancellationPoller) int {
	if pos+2 > len(text) || text[pos] != '0' || (text[pos+1] != 'x' && text[pos+1] != 'X') {
		return -1
	}
	end := pos + 2
	if end >= len(text) || !isHexDigit(rune(text[end])) {
		return -1
	}
	end++
	for end < len(text) && isHexDigit(rune(text[end])) {
		poller.poll()
		end++
	}
	if end < len(text) {
		ch, _ := utf8.DecodeRuneInString(text[end:])
		if isIdentifierPart(ch) {
			return -1
		}
	}
	return end - pos
}

func scanDigits(text string, pos int) int {
	return scanDigitsWithCancellation(text, pos, nil)
}

func scanDigitsWithCancellation(text string, pos int, poller *lexerCancellationPoller) int {
	end := pos
	for end < len(text) {
		poller.poll()
		ch, size := utf8.DecodeRuneInString(text[end:])
		if !IsDigit(ch) {
			break
		}
		end += size
	}
	if end == pos {
		return -1
	}
	return end - pos
}

func scanDigitStartIdentifierWithCancellation(text string, pos int, poller *lexerCancellationPoller) int {
	digitLen := scanDigitsWithCancellation(text, pos, poller)
	if digitLen <= 0 {
		return -1
	}
	end := pos + digitLen
	if end >= len(text) {
		return -1
	}
	ch, size := utf8.DecodeRuneInString(text[end:])
	if !(IsLetter(ch) || ch == '_') {
		return -1
	}
	end += size
	for end < len(text) {
		poller.poll()
		ch, size = utf8.DecodeRuneInString(text[end:])
		if !isIdentifierPart(ch) {
			break
		}
		end += size
	}
	return end - pos
}

var timespanSuffixes = []string{
	"milliseconds", "millisecond", "millisec", "minutes", "minute", "seconds", "second",
	"microseconds", "microsecond", "microsec", "nanoseconds", "nanosecond",
	"nanosec", "hours", "hour", "days", "day", "millis", "milli", "micros", "micro",
	"nanos", "nano", "ticks", "tick", "min", "hrs", "hr", "sec", "ms", "m", "s", "d", "h",
}

func scanTimespanSuffix(text string, pos int) int {
	longest := 0
	for _, suffix := range timespanSuffixes {
		end := pos + len(suffix)
		if end > len(text) || len(suffix) <= longest {
			continue
		}
		boundary := true
		if end < len(text) {
			ch, _ := utf8.DecodeRuneInString(text[end:])
			boundary = !isIdentifierPart(ch)
		}
		if ordinalEqualFold(text[pos:end], suffix) && boundary {
			longest = len(suffix)
		}
	}
	return longest
}

func keywordBoundary(text string, pos, length int) bool {
	beforeOK := true
	if pos > 0 {
		ch, _ := utf8.DecodeLastRuneInString(text[:pos])
		beforeOK = !isIdentifierPart(ch)
	}
	after := pos + length
	afterOK := true
	if after < len(text) {
		ch, _ := utf8.DecodeRuneInString(text[after:])
		afterOK = !isIdentifierPart(ch)
	}
	return beforeOK && afterOK
}

func isIdentifierStart(ch rune) bool {
	return IsLetter(ch) || ch == '_' || ch == '$'
}

func isIdentifierPart(ch rune) bool {
	return IsLetter(ch) || IsDigit(ch) || ch == '_' || ch == '$'
}
