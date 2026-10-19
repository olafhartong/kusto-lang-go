package kustolang

// Ported from src/Kusto.Language/Parser/TextFacts.cs at
// c94c9e761b3d095e2b9c21e3ec3e92afa0af4856.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"sort"
	"unicode/utf16"
)

// TextFactsAPI mirrors the upstream static TextFacts class. Positions,
// lengths, and ranges are UTF-16 code units, matching System.String.
type TextFactsAPI struct{}

func (TextFactsAPI) IsWhitespace(ch uint16) bool { return textFactsIsWhitespace(ch) }

func (TextFactsAPI) IsWhitespaceOnly(text string, startAndLength ...int) bool {
	units := UTF16CodeUnits(text)
	start, length := textFactsRange(units, startAndLength)
	for i := start; i < start+length; i++ {
		if !textFactsIsWhitespace(units[i]) {
			return false
		}
	}
	return true
}

func (api TextFactsAPI) IsWhitespaceLine(text string, start int) bool {
	return api.IsWhitespaceOnly(text, start, api.GetLineLength(text, start))
}

func (TextFactsAPI) GetWhitespace(text string, start int) string {
	units := UTF16CodeUnits(text)
	count := textFactsWhitespaceCount(units, start)
	return textFactsSubstring(units, start, count)
}

func (TextFactsAPI) GetWhitespaceCount(text string, start int) int {
	return textFactsWhitespaceCount(UTF16CodeUnits(text), start)
}

func (TextFactsAPI) GetWhitespaceCountBefore(text string, start int) int {
	units := UTF16CodeUnits(text)
	if start < 0 || start > len(units) {
		return 0
	}
	pos := start - 1
	for pos >= 0 && textFactsIsWhitespace(units[pos]) {
		pos--
	}
	return start - pos
}

func (TextFactsAPI) IsLineBreakStart(ch uint16) bool { return textFactsIsLineBreakStart(ch) }

func (TextFactsAPI) GetLineBreakLength(text string, position int) int {
	return textFactsLineBreakLength(UTF16CodeUnits(text), position)
}

func (TextFactsAPI) HasLineBreaks(text string) bool {
	units := UTF16CodeUnits(text)
	for _, unit := range units {
		if textFactsIsLineBreakStart(unit) {
			return true
		}
	}
	return false
}

func (TextFactsAPI) GetNextLineBreakStart(text string, start int) int {
	return textFactsNextLineBreakStart(UTF16CodeUnits(text), start)
}

func (TextFactsAPI) GetNextLineBreakEnd(text string, start int) int {
	units := UTF16CodeUnits(text)
	position := textFactsNextLineBreakStart(units, start)
	if position < 0 {
		return -1
	}
	return position + textFactsLineBreakLength(units, position)
}

func (TextFactsAPI) GetLineEnd(text string, position int) int {
	units := UTF16CodeUnits(text)
	return position + textFactsLineLength(units, position, false)
}

func (TextFactsAPI) GetLineStart(text string, position int) int {
	units := UTF16CodeUnits(text)
	if previous, ok := textFactsTryGetPreviousLineEnd(units, position); ok {
		return previous + textFactsLineBreakLength(units, previous)
	}
	return 0
}

func (TextFactsAPI) TryGetPreviousLineEnd(text string, position int) (int, bool) {
	return textFactsTryGetPreviousLineEnd(UTF16CodeUnits(text), position)
}

func (api TextFactsAPI) TryGetNextLineStart(text string, position int) (int, bool) {
	lineEnd := api.GetLineEnd(text, position)
	length := api.GetLineBreakLength(text, lineEnd)
	if length == 0 {
		return 0, false
	}
	return lineEnd + length, true
}

func (api TextFactsAPI) GetFirstLineBreakStart(text string) int {
	return api.GetNextLineBreakStart(text, 0)
}

func (api TextFactsAPI) GetFirstLineBreakEnd(text string) int {
	return api.GetNextLineBreakEnd(text, 0)
}

func (TextFactsAPI) GetLastLineBreakStart(text string, start ...int) int {
	units := UTF16CodeUnits(text)
	position := 0
	if len(start) > 0 {
		position = start[0]
	}
	result := -1
	for position >= 0 {
		next := textFactsNextLineBreakStart(units, position)
		if next < 0 {
			break
		}
		result = next
		position = next + textFactsLineBreakLength(units, next)
	}
	return result
}

func (api TextFactsAPI) GetLastLineBreakEnd(text string, start ...int) int {
	position := api.GetLastLineBreakStart(text, start...)
	if position < 0 {
		return -1
	}
	return position + api.GetLineBreakLength(text, position)
}

func (api TextFactsAPI) IsBlankLine(text string, lineStart int) bool {
	units := UTF16CodeUnits(text)
	position := textFactsPositionAfterLeadingWhitespace(units, lineStart)
	return position == len(units) || (position >= 0 && position < len(units) && textFactsIsLineBreakStart(units[position]))
}

func (TextFactsAPI) GetPositionAfterLeadingWhitespace(text string, lineStart int) int {
	return textFactsPositionAfterLeadingWhitespace(UTF16CodeUnits(text), lineStart)
}

func (TextFactsAPI) GetLineLength(text string, lineStart int, includeLineBreak ...bool) int {
	include := len(includeLineBreak) > 0 && includeLineBreak[0]
	return textFactsLineLength(UTF16CodeUnits(text), lineStart, include)
}

func (api TextFactsAPI) GetNextLineStart(text string, start int) int {
	position := api.GetNextLineBreakStart(text, start)
	if position < 0 {
		return -1
	}
	return position + api.GetLineBreakLength(text, position)
}

func (TextFactsAPI) IsLetter(ch uint16) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z'
}

func (TextFactsAPI) IsDigit(ch uint16) bool { return ch >= '0' && ch <= '9' }

func (api TextFactsAPI) IsLetterOrDigit(ch uint16) bool {
	return api.IsLetter(ch) || api.IsDigit(ch)
}

func (TextFactsAPI) IsHexDigit(ch uint16) bool {
	return ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F'
}

func (TextFactsAPI) GetLineStarts(text string) []int {
	return textFactsLineStarts(UTF16CodeUnits(text))
}

// AppendLineStarts is the Go spelling of the List<int> overload.
func (api TextFactsAPI) AppendLineStarts(text string, lineStarts *[]int) {
	if lineStarts == nil {
		panic("kustolang: nil line-start destination")
	}
	*lineStarts = append(*lineStarts, api.GetLineStarts(text)...)
}

func (TextFactsAPI) TryGetLineStart(text string, line int) (int, bool) {
	return textFactsTryGetLineStart(UTF16CodeUnits(text), line)
}

func (api TextFactsAPI) TryGetLineAndOffset(text string, position int) (line, lineOffset int, ok bool) {
	return api.TryGetLineAndOffsetFromStarts(api.GetLineStarts(text), position)
}

func (TextFactsAPI) TryGetLineAndOffsetFromStarts(lineStarts []int, position int) (line, lineOffset int, ok bool) {
	index := sort.SearchInts(lineStarts, position)
	if index >= len(lineStarts) || lineStarts[index] != position {
		index--
	}
	if index < 0 || index >= len(lineStarts) {
		return 0, 0, false
	}
	return index + 1, position - lineStarts[index] + 1, true
}

func (api TextFactsAPI) TryGetPosition(text string, line, lineOffset int) (int, bool) {
	if line < 1 || lineOffset < 1 {
		return 0, false
	}
	lineStart, ok := api.TryGetLineStart(text, line)
	if !ok {
		return 0, false
	}
	position := lineStart + lineOffset - 1
	if position > len(UTF16CodeUnits(text)) {
		return 0, false
	}
	return position, true
}

func (TextFactsAPI) TryGetPositionFromStarts(lineStarts []int, line, lineOffset int) (int, bool) {
	if line < 1 || line > len(lineStarts) || lineOffset < 1 {
		return 0, false
	}
	return lineStarts[line-1] + lineOffset - 1, true
}

func (TextFactsAPI) TrimEnd(text string, startAndLength ...int) int {
	units := UTF16CodeUnits(text)
	start, length := textFactsRange(units, startAndLength)
	for length > 0 && textFactsIsWhitespace(units[start+length-1]) {
		length--
	}
	return length
}

func (api TextFactsAPI) TrimRangeStart(text string, start, length int) (int, int) {
	count := api.GetWhitespaceCount(text, start)
	if count > length {
		count = length
	}
	return start + count, length - count
}

func (api TextFactsAPI) TrimRangeEnd(text string, start, length int) int {
	return api.TrimEnd(text, start, length)
}

func (api TextFactsAPI) TrimRange(text string, start, length int) (int, int) {
	start, length = api.TrimRangeStart(text, start, length)
	return start, api.TrimRangeEnd(text, start, length)
}

func (api TextFactsAPI) ExpandRangeToStartOfFirstLine(text string, start, length int) (int, int) {
	// Preserve the pinned source's assignment order: length is unchanged.
	start = api.GetLineStart(text, start)
	return start, length
}

func (api TextFactsAPI) ExpandRangeToEndOfLastLine(text string, start, length int) int {
	return api.GetLineEnd(text, start+length) - start
}

func (api TextFactsAPI) ExpandRangeToStartAndEndOfLines(text string, start, length int) (int, int) {
	start, length = api.ExpandRangeToStartOfFirstLine(text, start, length)
	return start, api.ExpandRangeToEndOfLastLine(text, start, length)
}

func (api TextFactsAPI) GetIndentationLength(text string, position int) int {
	return api.GetWhitespaceCount(text, api.GetLineStart(text, position))
}

func (api TextFactsAPI) GetIndentationText(text string, position int) string {
	return api.GetWhitespace(text, api.GetLineStart(text, position))
}

func (api TextFactsAPI) GetLineText(text string, line int) string {
	units := UTF16CodeUnits(text)
	lineStart, ok := textFactsTryGetLineStart(units, line)
	if !ok {
		return ""
	}
	return textFactsSubstring(units, lineStart, textFactsLineLength(units, lineStart, false))
}

func (api TextFactsAPI) GetLineTexts(text string) []string {
	starts := api.GetLineStarts(text)
	lines := make([]string, 0, len(starts))
	units := UTF16CodeUnits(text)
	for _, start := range starts {
		lines = append(lines, textFactsSubstring(units, start, textFactsLineLength(units, start, false)))
	}
	return lines
}

// TextFacts exposes source-shaped access to the upstream static API.
var TextFacts TextFactsAPI

func textFactsIsWhitespace(ch uint16) bool {
	switch ch {
	case '\t', ' ', '\r', '\n', 0x000c, 0x00a0, 0x1680, 0x180e,
		0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, 0x2006,
		0x2007, 0x2008, 0x2009, 0x200a, 0x200b, 0x202f, 0x205f,
		0x3000, 0xfeff:
		return true
	default:
		return false
	}
}

func textFactsIsLineBreakStart(ch uint16) bool {
	return ch == '\r' || ch == '\n' || ch == 0x2028 || ch == 0x2029
}

func textFactsLineBreakLength(units []uint16, position int) int {
	if position < 0 || position >= len(units) {
		return 0
	}
	if units[position] == '\r' && position+1 < len(units) && units[position+1] == '\n' {
		return 2
	}
	if textFactsIsLineBreakStart(units[position]) {
		return 1
	}
	return 0
}

func textFactsWhitespaceCount(units []uint16, start int) int {
	if start < 0 || start > len(units) {
		return 0
	}
	end := start
	for end < len(units) && textFactsIsWhitespace(units[end]) {
		end++
	}
	return end - start
}

func textFactsNextLineBreakStart(units []uint16, start int) int {
	if start < 0 {
		start = 0
	}
	for i := start; i < len(units); i++ {
		if textFactsIsLineBreakStart(units[i]) {
			return i
		}
	}
	return -1
}

func textFactsTryGetPreviousLineEnd(units []uint16, position int) (int, bool) {
	if position >= len(units) {
		position = len(units) - 1
	}
	for ; position >= 0; position-- {
		length := textFactsLineBreakLength(units, position)
		if length == 0 {
			continue
		}
		if length == 1 && position > 0 && textFactsLineBreakLength(units, position-1) == 2 {
			return position - 1, true
		}
		return position, true
	}
	return 0, false
}

func textFactsLineLength(units []uint16, start int, includeLineBreak bool) int {
	if start < 0 || start > len(units) {
		return 0
	}
	position := start
	for position < len(units) && !textFactsIsLineBreakStart(units[position]) {
		position++
	}
	if includeLineBreak {
		position += textFactsLineBreakLength(units, position)
	}
	return position - start
}

func textFactsPositionAfterLeadingWhitespace(units []uint16, start int) int {
	if start < 0 || start > len(units) {
		return start
	}
	position := start
	for position < len(units) && textFactsIsWhitespace(units[position]) && !textFactsIsLineBreakStart(units[position]) {
		position++
	}
	return position
}

func textFactsLineStarts(units []uint16) []int {
	starts := []int{0}
	for position := 0; position < len(units); {
		if length := textFactsLineBreakLength(units, position); length > 0 {
			position += length
			starts = append(starts, position)
		} else {
			position++
		}
	}
	return starts
}

func textFactsTryGetLineStart(units []uint16, line int) (int, bool) {
	if line < 1 {
		return 0, false
	}
	if line == 1 {
		return 0, true
	}
	count := 1
	for position := 0; position < len(units); {
		if length := textFactsLineBreakLength(units, position); length > 0 {
			position += length
			count++
			if count == line {
				return position, true
			}
		} else {
			position++
		}
	}
	return 0, false
}

func textFactsRange(units []uint16, values []int) (int, int) {
	start, length := 0, len(units)
	if len(values) >= 2 {
		start, length = values[0], values[1]
	}
	if start < 0 {
		start = 0
	}
	if start > len(units) {
		start = len(units)
	}
	if length < 0 {
		length = 0
	}
	if length > len(units)-start {
		length = len(units) - start
	}
	return start, length
}

func textFactsSubstring(units []uint16, start, length int) string {
	start, length = textFactsRange(units, []int{start, length})
	return string(utf16.Decode(units[start : start+length]))
}
