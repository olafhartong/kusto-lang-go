package kustolang

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Ported from src/Kusto.Language/Syntax/SyntaxToken.cs.

type SyntaxToken struct {
	syntaxOwnership
	kind        SyntaxKind
	trivia      string
	text        string
	valueText   string
	diagnostics []Diagnostic
	missing     bool
}

func SyntaxTokenFromLexical(token LexicalToken, extra ...Diagnostic) *SyntaxToken {
	diagnostics := append([]Diagnostic{}, token.Diagnostics...)
	diagnostics = append(diagnostics, extra...)
	text := token.Text
	if text == "" {
		text = SyntaxText(token.Kind)
	}
	result := &SyntaxToken{
		kind:        token.Kind,
		trivia:      token.Trivia,
		text:        text,
		valueText:   literalValueText(token.Kind, text),
		diagnostics: diagnostics,
	}
	initializeSyntaxElement(result)
	return result
}

func newSyntaxToken(kind SyntaxKind, trivia, text, valueText string, diagnostics []Diagnostic) *SyntaxToken {
	result := &SyntaxToken{
		kind:        kind,
		trivia:      trivia,
		text:        text,
		valueText:   valueText,
		diagnostics: append([]Diagnostic{}, diagnostics...),
	}
	initializeSyntaxElement(result)
	return result
}

// NewIdentifierToken constructs the public SyntaxToken.Identifier form.
func NewIdentifierToken(trivia, text string, diagnostics ...Diagnostic) *SyntaxToken {
	return newSyntaxToken(SyntaxKindIdentifierToken, trivia, text, text, diagnostics)
}

// NewIdentifierTokenWithValueText preserves a normalized identifier value
// while retaining the original token text.
func NewIdentifierTokenWithValueText(trivia, text, valueText string, diagnostics ...Diagnostic) *SyntaxToken {
	return newSyntaxToken(SyntaxKindIdentifierToken, trivia, text, valueText, diagnostics)
}

func NewKeywordToken(trivia string, kind SyntaxKind, diagnostics ...Diagnostic) (*SyntaxToken, error) {
	return NewKeywordTokenWithValueText(trivia, kind, SyntaxText(kind), diagnostics...)
}

func NewKeywordTokenWithValueText(trivia string, kind SyntaxKind, valueText string, diagnostics ...Diagnostic) (*SyntaxToken, error) {
	if err := checkSyntaxTokenCategory(kind, SyntaxCategoryKeyword); err != nil {
		return nil, err
	}
	return newSyntaxToken(kind, trivia, SyntaxText(kind), valueText, diagnostics), nil
}

func NewPunctuationToken(trivia string, kind SyntaxKind, diagnostics ...Diagnostic) (*SyntaxToken, error) {
	if err := checkSyntaxTokenCategory(kind, SyntaxCategoryPunctuation); err != nil {
		return nil, err
	}
	text := SyntaxText(kind)
	return newSyntaxToken(kind, trivia, text, text, diagnostics), nil
}

func NewOperatorToken(trivia string, kind SyntaxKind, diagnostics ...Diagnostic) (*SyntaxToken, error) {
	if err := checkSyntaxTokenCategory(kind, SyntaxCategoryOperator); err != nil {
		return nil, err
	}
	text := SyntaxText(kind)
	return newSyntaxToken(kind, trivia, text, text, diagnostics), nil
}

func NewLiteralToken(trivia, text string, kind SyntaxKind, diagnostics ...Diagnostic) (*SyntaxToken, error) {
	if err := checkSyntaxTokenCategory(kind, SyntaxCategoryLiteral); err != nil {
		return nil, err
	}
	return newSyntaxToken(kind, trivia, text, literalValueText(kind, text), diagnostics), nil
}

func NewOtherToken(trivia, text string, kind SyntaxKind, diagnostics ...Diagnostic) *SyntaxToken {
	return newSyntaxToken(kind, trivia, text, text, diagnostics)
}

func NewMissingTokenWithTrivia(trivia string, kind SyntaxKind, diagnostics ...Diagnostic) *SyntaxToken {
	result := newSyntaxToken(kind, trivia, "", "", diagnostics)
	result.missing = true
	return result
}

func checkSyntaxTokenCategory(kind SyntaxKind, expected SyntaxCategory) error {
	actual := SyntaxCategoryOf(kind)
	if actual != expected {
		return fmt.Errorf("syntax kind %v has category %s, want %s", kind, actual, expected)
	}
	return nil
}

func NewMissingToken(kind SyntaxKind, diagnostic Diagnostic) *SyntaxToken {
	diagnostics := []Diagnostic{}
	if diagnostic.Code != "" || diagnostic.Message != "" {
		diagnostics = append(diagnostics, diagnostic)
	}
	return NewMissingTokenWithTrivia("", kind, diagnostics...)
}

func (t *SyntaxToken) Kind() SyntaxKind {
	if t == nil {
		return SyntaxKindNone
	}
	return t.kind
}

// IsMissing mirrors SyntaxToken.MissingToken.IsMissing, which is true even
// when the missing token was constructed without a diagnostic.
func (t *SyntaxToken) IsMissing() bool {
	return t != nil && (t.missing || t.syntaxOwnership.IsMissing())
}

func (t *SyntaxToken) Trivia() string {
	return t.trivia
}

func (t *SyntaxToken) RawText() string {
	return t.text
}

func (t *SyntaxToken) ValueText() string {
	return t.valueText
}

func (t *SyntaxToken) Value() any {
	if t.kind != SyntaxKindStringLiteralToken && literalInteriorText(t.text) == "null" {
		return nil
	}

	switch t.kind {
	case SyntaxKindBooleanLiteralToken:
		return strings.EqualFold(t.valueText, "true")
	case SyntaxKindIntLiteralToken:
		if t.valueText == "min" {
			return int(math.MinInt32)
		}
		if t.valueText == "max" {
			return int(math.MaxInt32)
		}
		v, _ := strconv.ParseInt(t.valueText, 0, 0)
		return int(v)
	case SyntaxKindLongLiteralToken:
		text := strings.TrimSuffix(strings.TrimSuffix(t.valueText, "L"), "l")
		if text == "min" {
			return int64(math.MinInt64)
		}
		if text == "max" {
			return int64(math.MaxInt64)
		}
		v, _ := strconv.ParseInt(text, 0, 64)
		return v
	case SyntaxKindRealLiteralToken:
		if t.valueText == "min" {
			return -math.MaxFloat64
		}
		if t.valueText == "max" {
			return math.MaxFloat64
		}
		v, _ := strconv.ParseFloat(t.valueText, 64)
		return v
	case SyntaxKindDecimalLiteralToken:
		return decimalLiteralValue(t.valueText)
	case SyntaxKindTimespanLiteralToken:
		return timespanLiteralValue(t.valueText)
	case SyntaxKindDateTimeLiteralToken:
		return dateTimeLiteralValue(t.valueText)
	case SyntaxKindGuidLiteralToken, SyntaxKindRawGuidLiteralToken:
		return guidLiteralValue(t.valueText)
	case SyntaxKindStringLiteralToken:
		return t.valueText
	default:
		return t.text
	}
}

var maxDecimalValue, _ = new(big.Int).SetString("79228162514264337593543950335", 10)

func decimalLiteralValue(valueText string) *big.Rat {
	switch valueText {
	case "min":
		return new(big.Rat).SetInt(new(big.Int).Neg(new(big.Int).Set(maxDecimalValue)))
	case "max":
		return new(big.Rat).SetInt(new(big.Int).Set(maxDecimalValue))
	}
	value, ok := new(big.Rat).SetString(valueText)
	if !ok || strings.ContainsAny(valueText, "/eE") {
		return new(big.Rat)
	}
	limit := new(big.Rat).SetInt(maxDecimalValue)
	if new(big.Rat).Abs(new(big.Rat).Set(value)).Cmp(limit) > 0 {
		return new(big.Rat)
	}
	return value
}

// TimeSpanValue stores the same signed 100-nanosecond tick count as
// System.TimeSpan. A separate type is necessary because time.Duration cannot
// represent the full TimeSpan range.
type TimeSpanValue int64

const (
	timeSpanTicksPerSecond TimeSpanValue = 10_000_000
	timeSpanTicksPerMinute               = 60 * timeSpanTicksPerSecond
	timeSpanTicksPerHour                 = 60 * timeSpanTicksPerMinute
	timeSpanTicksPerDay                  = 24 * timeSpanTicksPerHour
)

func (v TimeSpanValue) Ticks() int64 {
	return int64(v)
}

// Duration returns the equivalent Go duration when it fits in that narrower
// representation.
func (v TimeSpanValue) Duration() (time.Duration, bool) {
	if v > TimeSpanValue(math.MaxInt64/100) || v < TimeSpanValue(math.MinInt64/100) {
		return 0, false
	}
	return time.Duration(v) * 100, true
}

func timespanLiteralValue(valueText string) TimeSpanValue {
	valueText = strings.TrimSpace(valueText)
	switch valueText {
	case "min":
		return TimeSpanValue(math.MinInt64)
	case "max":
		return TimeSpanValue(math.MaxInt64)
	}
	if value, ok := parseTimeSpanClock(valueText); ok {
		return value
	}

	split := 0
	for split < len(valueText) && (valueText[split] < 'A' || valueText[split] > 'Z') && (valueText[split] < 'a' || valueText[split] > 'z') {
		split++
	}
	number, ok := new(big.Rat).SetString(valueText[:split])
	if !ok {
		return 0
	}
	var unit TimeSpanValue
	truncate := false
	switch valueText[split:] {
	case "s", "sec", "second", "seconds":
		unit = timeSpanTicksPerSecond
	case "m", "min", "minute", "minutes":
		unit = timeSpanTicksPerMinute
	case "h", "hr", "hrs", "hour", "hours":
		unit = timeSpanTicksPerHour
	case "d", "day", "days":
		unit = timeSpanTicksPerDay
	case "ms", "milli", "millis", "millisec", "millisecond", "milliseconds":
		unit = timeSpanTicksPerSecond / 1_000
	case "micro", "micros", "microsec", "microsecond", "microseconds":
		unit = timeSpanTicksPerSecond / 1_000_000
	case "nano", "nanos", "nanosec", "nanosecond", "nanoseconds":
		return timeSpanFromRatio(number, int64(timeSpanTicksPerSecond), 1_000_000_000, false)
	case "tick", "ticks":
		unit = 1
		truncate = true
	default:
		return 0
	}
	return timeSpanFromRatio(number, int64(unit), 1, truncate)
}

func parseTimeSpanClock(valueText string) (TimeSpanValue, bool) {
	negative := strings.HasPrefix(valueText, "-")
	if negative || strings.HasPrefix(valueText, "+") {
		valueText = valueText[1:]
	}
	parts := strings.Split(valueText, ":")
	if len(parts) == 1 {
		days, ok := new(big.Rat).SetString(valueText)
		if !ok {
			return 0, false
		}
		if negative {
			days.Neg(days)
		}
		return timeSpanFromRatio(days, int64(timeSpanTicksPerDay), 1, false), true
	}
	if len(parts) < 2 || len(parts) > 3 {
		return 0, false
	}

	days := int64(0)
	hoursText := parts[0]
	if dot := strings.IndexByte(hoursText, '.'); dot >= 0 {
		var err error
		days, err = strconv.ParseInt(hoursText[:dot], 10, 64)
		if err != nil {
			return 0, false
		}
		hoursText = hoursText[dot+1:]
	}
	hours, err := strconv.ParseInt(hoursText, 10, 64)
	if err != nil {
		return 0, false
	}
	minutes, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, false
	}
	seconds := new(big.Rat)
	if len(parts) == 3 {
		var ok bool
		seconds, ok = new(big.Rat).SetString(parts[2])
		if !ok {
			return 0, false
		}
	}
	if hours < 0 || hours >= 24 || minutes < 0 || minutes >= 60 || seconds.Sign() < 0 || seconds.Cmp(big.NewRat(60, 1)) >= 0 {
		return 0, false
	}
	total := new(big.Rat).SetInt(new(big.Int).Mul(big.NewInt(days), big.NewInt(int64(timeSpanTicksPerDay))))
	total.Add(total, new(big.Rat).SetInt(new(big.Int).Mul(big.NewInt(hours), big.NewInt(int64(timeSpanTicksPerHour)))))
	total.Add(total, new(big.Rat).SetInt(new(big.Int).Mul(big.NewInt(minutes), big.NewInt(int64(timeSpanTicksPerMinute)))))
	total.Add(total, new(big.Rat).Mul(seconds, new(big.Rat).SetInt64(int64(timeSpanTicksPerSecond))))
	if negative {
		total.Neg(total)
	}
	return roundedTimeSpanTicks(total), true
}

func timeSpanFromRatio(value *big.Rat, numerator, denominator int64, truncate bool) TimeSpanValue {
	ticks := new(big.Rat).Mul(value, big.NewRat(numerator, denominator))
	if truncate {
		quotient := new(big.Int).Quo(ticks.Num(), ticks.Denom())
		if quotient.IsInt64() {
			return TimeSpanValue(quotient.Int64())
		}
		return 0
	}
	return roundedTimeSpanTicks(ticks)
}

func roundedTimeSpanTicks(value *big.Rat) TimeSpanValue {
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(value.Num(), value.Denom(), remainder)
	twiceRemainder := new(big.Int).Lsh(new(big.Int).Abs(remainder), 1)
	comparison := twiceRemainder.Cmp(value.Denom())
	if comparison > 0 || (comparison == 0 && quotient.Bit(0) == 1) {
		if value.Sign() < 0 {
			quotient.Sub(quotient, big.NewInt(1))
		} else {
			quotient.Add(quotient, big.NewInt(1))
		}
	}
	if !quotient.IsInt64() {
		return 0
	}
	return TimeSpanValue(quotient.Int64())
}

func dateTimeLiteralValue(valueText string) time.Time {
	valueText = strings.TrimSpace(valueText)
	switch valueText {
	case "min":
		return time.Time{}
	case "max":
		return time.Date(9999, 12, 31, 23, 59, 59, 999999900, time.UTC)
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05.999999999 -07:00",
	} {
		if value, err := time.Parse(layout, valueText); err == nil {
			return value
		}
	}
	for _, layout := range []string{
		"2006-01-02",
		"2006-1-2",
		"2006/01/02",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02 15:04:05.999999999",
		"2006-1-2 15:04:05.999999999",
	} {
		if value, err := time.ParseInLocation(layout, valueText, time.UTC); err == nil {
			return value
		}
	}
	return time.Time{}
}

func guidLiteralValue(valueText string) [16]byte {
	var value [16]byte
	valueText = strings.TrimSpace(valueText)
	if len(valueText) >= 2 && ((valueText[0] == '{' && valueText[len(valueText)-1] == '}') ||
		(valueText[0] == '(' && valueText[len(valueText)-1] == ')')) {
		valueText = valueText[1 : len(valueText)-1]
	}
	valueText = strings.ReplaceAll(valueText, "-", "")
	if len(valueText) != hex.EncodedLen(len(value)) {
		return value
	}
	if _, err := hex.Decode(value[:], []byte(valueText)); err != nil {
		return [16]byte{}
	}
	return value
}

func (t *SyntaxToken) IsLiteral() bool {
	return t != nil && SyntaxCategoryOf(t.kind) == SyntaxCategoryLiteral
}

func (t *SyntaxToken) Prefix() string {
	if !t.IsLiteral() {
		return ""
	}
	if length := literalPrefixLength(t.text); length > 0 {
		return t.text[:length]
	}
	return ""
}

func literalPrefixLength(text string) int {
	for index, r := range text {
		if r == '(' && index > 0 {
			return index
		}
		if !unicode.IsLetter(r) {
			return 0
		}
	}
	return 0
}

func literalInteriorText(text string) string {
	prefixLength := literalPrefixLength(text)
	if prefixLength == 0 {
		return text
	}
	start := prefixLength + 1
	for start < len(text) && (text[start] == ' ' || text[start] == '\t' || text[start] == '\r' || text[start] == '\n') {
		start++
	}
	end := len(text)
	if strings.HasSuffix(text, ")") {
		end--
	}
	for end > start+1 && (text[end-1] == ' ' || text[end-1] == '\t' || text[end-1] == '\r' || text[end-1] == '\n') {
		end--
	}
	return text[start:end]
}

func (t *SyntaxToken) Width() int {
	return len(t.text)
}

func (t *SyntaxToken) ByteWidth() int {
	return t.Width()
}

func (t *SyntaxToken) UTF16Width() int {
	if t == nil {
		return 0
	}
	return utf16TextWidth(t.text)
}

func (t *SyntaxToken) TriviaWidth() int {
	return len(t.trivia)
}

func (t *SyntaxToken) ByteTriviaWidth() int {
	return t.TriviaWidth()
}

func (t *SyntaxToken) UTF16TriviaWidth() int {
	if t == nil {
		return 0
	}
	return utf16TextWidth(t.trivia)
}

func (t *SyntaxToken) FullWidth() int {
	return len(t.trivia) + len(t.text)
}

func (t *SyntaxToken) ByteFullWidth() int {
	return t.FullWidth()
}

func (t *SyntaxToken) UTF16FullWidth() int {
	if t == nil {
		return 0
	}
	return t.UTF16TriviaWidth() + t.UTF16Width()
}

func (t *SyntaxToken) IsToken() bool {
	return true
}

func (t *SyntaxToken) ChildCount() int {
	return 0
}

func (t *SyntaxToken) GetChild(index int) SyntaxElement {
	panic("syntax token has no children")
}

func (t *SyntaxToken) GetName(index int) string {
	return ""
}

func (t *SyntaxToken) CompletionHint(index int) CompletionHint {
	return CompletionHintInherit
}

func (t *SyntaxToken) IsOptional(index int) bool {
	return false
}

func (t *SyntaxToken) Diagnostics() []Diagnostic {
	return append([]Diagnostic{}, t.diagnostics...)
}

func (t *SyntaxToken) ContainedDiagnostics() []Diagnostic {
	if t.diagnostics == nil {
		return []Diagnostic{}
	}
	return append([]Diagnostic{}, t.diagnostics...)
}

func (t *SyntaxToken) GetContainedSyntaxDiagnostics() []Diagnostic {
	return GetContainedDiagnosticsPreserveDuplicates(t, DiagnosticsIncludeSyntactic)
}

func (t *SyntaxToken) Children() []SyntaxElement {
	return nil
}

func (t *SyntaxToken) Text(includeTrivia IncludeTrivia) string {
	return t.TextWithContext(includeTrivia, false, t.kind == SyntaxKindEndOfTextToken)
}

func (t *SyntaxToken) TextWithContext(includeTrivia IncludeTrivia, beforeQuery, afterQuery bool) string {
	return t.textTrivia(includeTrivia, beforeQuery, afterQuery) + t.text
}

func (t *SyntaxToken) textTrivia(includeTrivia IncludeTrivia, beforeQuery, afterQuery bool) string {
	if t.trivia != "" {
		switch includeTrivia {
		case IncludeTriviaAll:
			return t.trivia
		case IncludeTriviaInterior:
			if !beforeQuery {
				return t.trivia
			}
		case IncludeTriviaMinimal:
			if !beforeQuery && !afterQuery {
				if strings.Contains(t.trivia, "\n") {
					return "\n"
				}
				return " "
			}
		case IncludeTriviaSingleLine:
			if !beforeQuery && !afterQuery {
				return " "
			}
		}
	}
	return ""
}

func (t *SyntaxToken) Clone(includeDiagnostics bool) SyntaxElement {
	if t == nil {
		return (*SyntaxToken)(nil)
	}
	clone := &SyntaxToken{
		kind:        t.kind,
		trivia:      t.trivia,
		text:        t.text,
		valueText:   t.valueText,
		diagnostics: cloneDiagnostics(t.diagnostics, includeDiagnostics),
		missing:     t.missing,
	}
	initializeSyntaxElement(clone)
	return clone
}

func (t *SyntaxToken) withTriviaStart(triviaStart int) *SyntaxToken {
	if t == nil {
		return nil
	}
	clone := *t
	clone.syntaxOwnership = syntaxOwnership{
		triviaStart: triviaStart,
		hasPosition: true,
	}
	clone.self = &clone
	return &clone
}

func literalValueText(kind SyntaxKind, text string) string {
	switch kind {
	case SyntaxKindStringLiteralToken:
		return decodeStringLiteral(text)
	case SyntaxKindGuidLiteralToken, SyntaxKindRawGuidLiteralToken:
		return guidLiteralValueText(text)
	case SyntaxKindLongLiteralToken:
		if value, ok := typedLiteralValueText(text, "long", "int64"); ok {
			text = value
		}
		return strings.TrimSuffix(strings.TrimSuffix(text, "L"), "l")
	case SyntaxKindIntLiteralToken:
		if value, ok := typedLiteralValueText(text, "int", "int32"); ok {
			return value
		}
	case SyntaxKindRealLiteralToken:
		if value, ok := typedLiteralValueText(text, "real", "double"); ok {
			return value
		}
	case SyntaxKindDecimalLiteralToken:
		if value, ok := typedLiteralValueText(text, "decimal"); ok {
			return value
		}
	case SyntaxKindBooleanLiteralToken:
		if value, ok := typedLiteralValueText(text, "bool"); ok {
			return value
		}
	case SyntaxKindDateTimeLiteralToken:
		if value, ok := typedLiteralValueText(text, "datetime", "date"); ok {
			return destringifyLiteralValue(value)
		}
	case SyntaxKindTimespanLiteralToken:
		if value, ok := typedLiteralValueText(text, "timespan", "time"); ok {
			return destringifyLiteralValue(value)
		}
	default:
		return text
	}
	return text
}

func guidLiteralValueText(text string) string {
	if inner, ok := typedLiteralValueText(text, "guid"); ok {
		return destringifyLiteralValue(inner)
	}
	return strings.TrimSpace(text)
}

func destringifyLiteralValue(text string) string {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) >= 2 && ((trimmed[0] == '\'' && trimmed[len(trimmed)-1] == '\'') ||
		(trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"') ||
		(trimmed[0] == '`' && trimmed[len(trimmed)-1] == '`')) {
		return decodeStringLiteral(trimmed)
	}
	return text
}

func typedLiteralValueText(text string, prefixes ...string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	for _, prefix := range prefixes {
		fullPrefix := prefix + "("
		if strings.HasPrefix(lower, fullPrefix) && strings.HasSuffix(trimmed, ")") {
			return strings.TrimSpace(trimmed[len(fullPrefix) : len(trimmed)-1]), true
		}
	}
	return "", false
}

func decodeStringLiteral(text string) string {
	if text == "" {
		return text
	}
	start := 0
	if text[start] == 'h' || text[start] == 'H' {
		start++
	}
	isVerbatim := false
	if start < len(text) && text[start] == '@' {
		start++
		isVerbatim = true
	}
	if value, ok := multilineStringLiteralValue(text[start:], "```"); ok {
		return value
	}
	if value, ok := multilineStringLiteralValue(text[start:], "~~~"); ok {
		return value
	}
	if start >= len(text) {
		return ""
	}
	quote := text[start]
	if quote != '\'' && quote != '"' && quote != '`' && quote != '~' {
		return text
	}
	end := len(text)
	if end > start && text[end-1] == quote {
		end--
	}
	start++
	if end <= start {
		return ""
	}
	body := text[start:end]
	if isVerbatim {
		return strings.ReplaceAll(body, string([]byte{quote, quote}), string(quote))
	}
	return decodeStringEscapes(body)
}

func multilineStringLiteralValue(text string, quote string) (string, bool) {
	if !strings.HasPrefix(text, quote) {
		return "", false
	}
	start := len(quote)
	if len(text) >= len(quote)*2 && strings.HasSuffix(text, quote) {
		return text[start : len(text)-len(quote)], true
	}
	return text[start:], true
}

func decodeStringEscapes(text string) string {
	if !strings.Contains(text, `\`) {
		return text
	}
	var b strings.Builder
	for i := 0; i < len(text); {
		if text[i] != '\\' || i+1 >= len(text) {
			b.WriteByte(text[i])
			i++
			continue
		}
		ch := text[i+1]
		switch ch {
		case '\'', '"', '\\':
			b.WriteByte(ch)
			i += 2
		case 'a':
			b.WriteByte('\a')
			i += 2
		case 'b':
			b.WriteByte('\b')
			i += 2
		case 'f':
			b.WriteByte('\f')
			i += 2
		case 'n':
			b.WriteByte('\n')
			i += 2
		case 'r':
			b.WriteByte('\r')
			i += 2
		case 't':
			b.WriteByte('\t')
			i += 2
		case 'v':
			b.WriteByte('\v')
			i += 2
		case 'u':
			value, length := decodeHexEscape(text, i+2, 4)
			b.WriteRune(rune(value))
			i += 2 + length
		case 'U':
			value, length := decodeHexEscape(text, i+2, 8)
			b.WriteRune(rune(value))
			i += 2 + length
		case 'x':
			value, length := decodeHexEscape(text, i+2, 2)
			b.WriteRune(rune(value))
			i += 2 + length
		default:
			if ch >= '0' && ch <= '9' {
				value, length := decodeOctalEscape(text, i+1)
				b.WriteRune(rune(value))
				i += 1 + length
			} else {
				b.WriteByte(ch)
				i += 2
			}
		}
	}
	return b.String()
}

func decodeHexEscape(text string, pos int, length int) (int64, int) {
	value := int64(0)
	count := 0
	for count < length && pos+count < len(text) {
		ch := text[pos+count]
		var digit int64
		switch {
		case ch >= 'a' && ch <= 'f':
			digit = int64(ch-'a') + 10
		case ch >= 'A' && ch <= 'F':
			digit = int64(ch-'A') + 10
		case ch >= '0' && ch <= '9':
			digit = int64(ch - '0')
		default:
			return value, count
		}
		value = value<<4 + digit
		count++
	}
	return value, count
}

func decodeOctalEscape(text string, pos int) (int64, int) {
	end := pos
	value := int64(0)
	for end < len(text) && end-pos < 3 && text[end] >= '0' && text[end] <= '9' {
		value = value<<3 + int64(text[end]-'0')
		end++
	}
	return value, end - pos
}
