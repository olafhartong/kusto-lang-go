package kustolang

import (
	"sort"
	"unicode/utf8"
)

// Ported from src/Kusto.Language/Parser/TextFacts.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func IsWhitespace(ch rune) bool {
	switch ch {
	case '\t', ' ', '\r', '\n', '\u000c', '\u00a0', '\u1680', '\u180e',
		'\u2000', '\u2001', '\u2002', '\u2003', '\u2004', '\u2005',
		'\u2006', '\u2007', '\u2008', '\u2009', '\u200a', '\u200b',
		'\u202f', '\u205f', '\u3000', '\uFEFF':
		return true
	default:
		return false
	}
}

func IsLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func IsDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func IsLetterOrDigit(ch rune) bool {
	return IsLetter(ch) || IsDigit(ch)
}

func IsHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func IsLineBreakStart(ch rune) bool {
	switch ch {
	case '\r', '\n', '\u2028', '\u2029':
		return true
	default:
		return false
	}
}

func IsWhitespaceOnly(text string, startAndLength ...int) bool {
	start, length := 0, len(text)
	if len(startAndLength) >= 2 {
		start = startAndLength[0]
		length = startAndLength[1]
	}
	end := minInt(len(text), start+length)
	for pos := start; pos < end; {
		ch, size := utf8.DecodeRuneInString(text[pos:end])
		if !IsWhitespace(ch) {
			return false
		}
		pos += size
	}
	return true
}

func IsWhitespaceLine(text string, start int) bool {
	return IsWhitespaceOnly(text, start, LineLength(text, start))
}

func GetWhitespace(text string, start int) string {
	count := WhitespaceCount(text, start)
	if count == 0 {
		return ""
	}
	return text[start : start+count]
}

func LineBreakLength(text string, position int) int {
	if position >= len(text) {
		return 0
	}
	switch text[position] {
	case '\r':
		if position+1 < len(text) && text[position+1] == '\n' {
			return 2
		}
		return 1
	case '\n':
		return 1
	}
	ch, size := utf8.DecodeRuneInString(text[position:])
	if ch == '\u2028' || ch == '\u2029' {
		return size
	}
	return 0
}

func HasLineBreaks(text string) bool {
	for pos := 0; pos < len(text); pos++ {
		if LineBreakLength(text, pos) > 0 {
			return true
		}
	}
	return false
}

func WhitespaceCount(text string, start int) int {
	pos := start
	for pos < len(text) {
		ch, size := utf8.DecodeRuneInString(text[pos:])
		if !IsWhitespace(ch) {
			break
		}
		pos += size
	}
	return pos - start
}

func WhitespaceCountBefore(text string, start int) int {
	// The source returns the distance from start to the first preceding
	// non-whitespace position, so a non-empty input has a minimum of one.
	pos := start - 1
	for pos >= 0 {
		ch, size := utf8.DecodeLastRuneInString(text[:pos+1])
		if !IsWhitespace(ch) {
			break
		}
		pos -= size
	}
	return start - pos
}

func LineLength(text string, start int, includeLineBreak ...bool) int {
	pos := start
	for pos < len(text) {
		if LineBreakLength(text, pos) > 0 {
			break
		}
		pos++
	}
	if len(includeLineBreak) > 0 && includeLineBreak[0] && pos < len(text) {
		pos += LineBreakLength(text, pos)
	}
	return pos - start
}

func LineEnd(text string, position int) int {
	return position + LineLength(text, position)
}

func NextLineBreakStart(text string, start int) int {
	if start < 0 {
		start = 0
	}
	for pos := start; pos < len(text); pos++ {
		if LineBreakLength(text, pos) > 0 {
			return pos
		}
	}
	return -1
}

func NextLineBreakEnd(text string, start int) int {
	lineBreakStart := NextLineBreakStart(text, start)
	if lineBreakStart < 0 {
		return -1
	}
	return lineBreakStart + LineBreakLength(text, lineBreakStart)
}

func LineStart(text string, position int) int {
	if previousLineEnd, ok := TryGetPreviousLineEnd(text, position); ok {
		return previousLineEnd + LineBreakLength(text, previousLineEnd)
	}
	return 0
}

func TryGetPreviousLineEnd(text string, position int) (previousLineEnd int, ok bool) {
	for ; position >= 0; position-- {
		lineBreakLength := LineBreakLength(text, position)
		if lineBreakLength == 0 {
			continue
		}
		if lineBreakLength == 1 && position > 0 && LineBreakLength(text, position-1) == lineBreakLength+1 {
			return position - 1, true
		}
		return position, true
	}
	return 0, false
}

func TryGetNextLineStart(text string, position int) (nextLineStart int, ok bool) {
	lineEnd := LineEnd(text, position)
	lineBreakLength := LineBreakLength(text, lineEnd)
	if lineBreakLength == 0 {
		return 0, false
	}
	return lineEnd + lineBreakLength, true
}

func FirstLineBreakStart(text string) int {
	return NextLineBreakStart(text, 0)
}

func FirstLineBreakEnd(text string) int {
	return NextLineBreakEnd(text, 0)
}

func LastLineBreakStart(text string, start ...int) int {
	result := -1
	lastLineBreakEnd := 0
	if len(start) > 0 {
		lastLineBreakEnd = start[0]
	}
	for lastLineBreakEnd >= 0 {
		nextLineBreakStart := NextLineBreakStart(text, lastLineBreakEnd)
		if nextLineBreakStart < 0 {
			break
		}
		lastLineBreakEnd = nextLineBreakStart + LineBreakLength(text, nextLineBreakStart)
		result = nextLineBreakStart
	}
	return result
}

func LastLineBreakEnd(text string, start ...int) int {
	lastLineBreakStart := LastLineBreakStart(text, start...)
	if lastLineBreakStart < 0 {
		return -1
	}
	return lastLineBreakStart + LineBreakLength(text, lastLineBreakStart)
}

func TrimEnd(text string, startAndLength ...int) int {
	start := 0
	length := len(text)
	if len(startAndLength) >= 2 {
		start = startAndLength[0]
		length = startAndLength[1]
	}
	if start < 0 {
		start = 0
	}
	if length < 0 {
		length = 0
	}
	end := minInt(len(text), start+length)
	for end > start {
		ch, size := utf8.DecodeLastRuneInString(text[start:end])
		if !IsWhitespace(ch) {
			break
		}
		end -= size
	}
	return end - start
}

func TrimRangeStart(text string, start int, length int) (int, int) {
	count := WhitespaceCount(text, start)
	return start + count, length - count
}

func TrimRangeEnd(text string, start int, length int) int {
	return TrimEnd(text, start, length)
}

func TrimRange(text string, start int, length int) (int, int) {
	start, length = TrimRangeStart(text, start, length)
	return start, TrimRangeEnd(text, start, length)
}

func ExpandRangeToStartOfFirstLine(text string, start int, length int) (int, int) {
	newStart := LineStart(text, start)
	// Preserve the source assignment order, including its unchanged length.
	start = newStart
	length += start - newStart
	return start, length
}

func ExpandRangeToEndOfLastLine(text string, start int, length int) int {
	return LineEnd(text, start+length) - start
}

func ExpandRangeToStartAndEndOfLines(text string, start int, length int) (int, int) {
	start, length = ExpandRangeToStartOfFirstLine(text, start, length)
	return start, ExpandRangeToEndOfLastLine(text, start, length)
}

func NextLineStart(text string, start int) int {
	lineBreakStart := NextLineBreakStart(text, start)
	if lineBreakStart < 0 {
		return -1
	}
	return lineBreakStart + LineBreakLength(text, lineBreakStart)
}

func PositionAfterLeadingWhitespace(text string, lineStart int) int {
	pos := lineStart
	for pos < len(text) && LineBreakLength(text, pos) == 0 {
		ch, size := utf8.DecodeRuneInString(text[pos:])
		if !IsWhitespace(ch) {
			break
		}
		pos += size
	}
	return pos
}

func IsBlankLine(text string, lineStart int) bool {
	pos := PositionAfterLeadingWhitespace(text, lineStart)
	if pos == len(text) {
		return true
	}
	ch, _ := utf8.DecodeRuneInString(text[pos:])
	return IsLineBreakStart(ch)
}

func LineStarts(text string) []int {
	starts := []int{0}
	for pos := 0; pos < len(text); {
		if length := LineBreakLength(text, pos); length > 0 {
			pos += length
			starts = append(starts, pos)
			continue
		}
		pos++
	}
	return starts
}

func TryGetLineAndOffset(lineStarts []int, position int) (line int, lineOffset int, ok bool) {
	index := sort.SearchInts(lineStarts, position)
	if index >= len(lineStarts) || lineStarts[index] != position {
		index--
	}
	if index < 0 || index >= len(lineStarts) {
		return 0, 0, false
	}
	return index + 1, position - lineStarts[index] + 1, true
}

func TryGetTextLineStart(text string, line int) (lineStart int, ok bool) {
	if line < 1 {
		return 0, false
	}
	if line == 1 {
		return 0, true
	}
	count := 1
	for pos := 0; pos < len(text); {
		if length := LineBreakLength(text, pos); length > 0 {
			pos += length
			count++
			if count == line {
				return pos, true
			}
			continue
		}
		pos++
	}
	return 0, false
}

func TryGetTextLineAndOffset(text string, position int) (line int, lineOffset int, ok bool) {
	return TryGetLineAndOffset(LineStarts(text), position)
}

func TryGetTextPosition(text string, line int, lineOffset int) (position int, ok bool) {
	if line < 1 || lineOffset < 1 {
		return 0, false
	}
	lineStart, ok := TryGetTextLineStart(text, line)
	if !ok {
		return 0, false
	}
	position = lineStart + lineOffset - 1
	if position > len(text) {
		return 0, false
	}
	return position, true
}

func TryGetPosition(lineStarts []int, line int, lineOffset int) (position int, ok bool) {
	if line < 1 || line > len(lineStarts) || lineOffset < 1 {
		return 0, false
	}
	return lineStarts[line-1] + lineOffset - 1, true
}

func IndentationLength(text string, position int) int {
	lineStart := LineStart(text, position)
	return WhitespaceCount(text, lineStart)
}

func IndentationText(text string, position int) string {
	return GetWhitespace(text, LineStart(text, position))
}

func LineText(text string, line int) string {
	lineStart, ok := TryGetTextLineStart(text, line)
	if !ok {
		return ""
	}
	return text[lineStart:LineEnd(text, lineStart)]
}

func LineTexts(text string) []string {
	lineStarts := LineStarts(text)
	lines := make([]string, 0, len(lineStarts))
	for _, lineStart := range lineStarts {
		lines = append(lines, text[lineStart:lineStart+LineLength(text, lineStart)])
	}
	return lines
}
