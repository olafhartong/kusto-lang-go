package kustolang

import (
	"strings"
	"unicode/utf8"
)

// Ported from the static scanner helpers in src/Kusto.Language/Parser/TokenParser.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func ScanWhitespace(text string, start int) int {
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

func ScanTrivia(text string, start int) int {
	pos := start
	for pos < len(text) {
		ch, size := utf8.DecodeRuneInString(text[pos:])
		if IsWhitespace(ch) {
			pos += size
			continue
		}
		if commentLength := ScanComment(text, pos); commentLength > 0 {
			pos += commentLength
			continue
		}
		break
	}
	return pos - start
}

func ParseTrivia(text string, start int) string {
	if start < 0 || start > len(text) {
		return ""
	}
	length := ScanTrivia(text, start)
	return text[start : start+length]
}

func ScanComment(text string, start int) int {
	if isCommentStart(text, start) {
		nextLineStart := NextLineStart(text, start)
		if nextLineStart >= 0 {
			return nextLineStart - start
		}
		return len(text) - start
	}
	return 0
}

func ScanIdentifier(text string, start int) int {
	pos := start
	if pos < 0 || pos >= len(text) {
		return -1
	}
	ch, size := utf8.DecodeRuneInString(text[pos:])
	if tokenParserIdentifierStartChar(ch) {
		pos += size
		for pos < len(text) {
			ch, size = utf8.DecodeRuneInString(text[pos:])
			if !tokenParserIdentifierChar(ch) {
				break
			}
			pos += size
		}
	} else if IsDigit(ch) {
		length := scanDigits(text, pos)
		if length > 0 {
			next, nextSize := utf8.DecodeRuneInString(text[pos+length:])
			if IsLetter(next) || next == '_' {
				pos += length
				pos += nextSize
				for pos < len(text) {
					next, nextSize = utf8.DecodeRuneInString(text[pos:])
					if !tokenParserIdentifierChar(next) {
						break
					}
					pos += nextSize
				}
			}
		}
	}
	if pos > start {
		return pos - start
	}
	return -1
}

// ScanBooleanLiteral scans the raw true/false spellings accepted by the
// upstream token parser. Constructor forms such as bool(true) are outside
// this scanner's scope.
func ScanBooleanLiteral(text string, start int) int {
	if start < 0 || start >= len(text) {
		return -1
	}
	for _, literal := range []string{"false", "False", "FALSE", "true", "True", "TRUE"} {
		if !strings.HasPrefix(text[start:], literal) {
			continue
		}
		end := start + len(literal)
		if end < len(text) {
			ch, _ := utf8.DecodeRuneInString(text[end:])
			if tokenParserIdentifierChar(ch) {
				continue
			}
		}
		return len(literal)
	}
	return -1
}

func ScanLongLiteral(text string, start int) int {
	if hexLen := scanTokenParserHexIntegerLiteral(text, start); hexLen > 0 {
		return hexLen
	}
	length := scanDigits(text, start)
	if length > 0 {
		if start+length >= len(text) {
			return length
		}
		ch, _ := utf8.DecodeRuneInString(text[start+length:])
		if !tokenParserIdentifierChar(ch) {
			return length
		}
	}
	return -1
}

func ScanClientParameter(text string, start int) int {
	length, _, _, _, _ := ScanClientParameterParts(text, start)
	return length
}

func ScanClientParameterParts(text string, start int) (length int, nameStart int, nameLength int, indexStart int, indexLength int) {
	nameStart = -1

	if peekByte(text, start) == '{' {
		nameStart = start + 1
		nameLength = ScanIdentifier(text, nameStart)
		if nameLength > 0 {
			nameEnd := nameStart + nameLength
			switch peekByte(text, nameEnd) {
			case '}':
				return nameLength + 2, nameStart, nameLength, indexStart, indexLength
			case '[':
				indexStart = nameEnd + 1
				indexEnd := indexStart
				if peekByte(text, indexEnd) == '-' {
					indexEnd++
				}
				literalLength := ScanLongLiteral(text, indexEnd)
				if literalLength <= 0 {
					return -1, nameStart, nameLength, indexStart, indexLength
				}
				indexEnd += literalLength
				if peekByte(text, indexEnd) == ']' && peekByte(text, indexEnd+1) == '}' {
					indexLength = indexEnd - indexStart
					return (indexEnd + 2) - start, nameStart, nameLength, indexStart, indexLength
				}
			}
		}
	}

	return -1, nameStart, nameLength, indexStart, indexLength
}

func ScanRealLiteral(text string, start int) int {
	digitLen := scanDigits(text, start)
	if digitLen <= 0 {
		return -1
	}
	pos := start + digitLen
	if peekByte(text, pos) == '.' && (peekByte(text, pos+1) != '.' || peekByte(text, pos+2) == '.') {
		pos++
		if fractionLen := scanDigits(text, pos); fractionLen > 0 {
			pos += fractionLen
		}
		if exponentLen := scanExponent(text, pos); exponentLen > 0 {
			pos += exponentLen
		}
	} else {
		exponentLen := scanExponent(text, pos)
		if exponentLen <= 0 {
			return -1
		}
		pos += exponentLen
	}
	if pos < len(text) {
		ch, _ := utf8.DecodeRuneInString(text[pos:])
		if tokenParserIdentifierChar(ch) {
			return -1
		}
	}
	return pos - start
}

// ScanTimespanLiteral scans raw timespan literals such as 1d and 2.5hours.
// Constructor forms such as timespan(1d) are outside this scanner's scope.
func ScanTimespanLiteral(text string, start int) int {
	if start < 0 || start >= len(text) {
		return -1
	}
	return scanTimespanLiteral(text, start)
}

// ScanRawGuidLiteral scans the 8-4-4-4-12 raw GUID spelling. Constructor
// forms such as guid(...) are outside this scanner's scope.
func ScanRawGuidLiteral(text string, start int) int {
	if start < 0 || start > len(text)-36 {
		return -1
	}
	_, end, ok := scanRawGuidLiteral(text, start)
	if !ok {
		return -1
	}
	return end - start
}

func tokenParserIdentifierStartChar(ch rune) bool {
	return IsLetter(ch) || ch == '_' || ch == '$'
}

func tokenParserIdentifierChar(ch rune) bool {
	return IsLetter(ch) || IsDigit(ch) || ch == '_'
}

func scanTokenParserHexIntegerLiteral(text string, start int) int {
	pos := start
	if peekByte(text, start) == '0' && (peekByte(text, start+1) == 'x' || peekByte(text, start+1) == 'X') {
		pos += 2
		if !isHexDigit(rune(peekByte(text, pos))) {
			return -1
		}
		pos++
		for isHexDigit(rune(peekByte(text, pos))) {
			pos++
		}
	}
	if pos < len(text) {
		ch, _ := utf8.DecodeRuneInString(text[pos:])
		if tokenParserIdentifierChar(ch) {
			return -1
		}
	}
	if pos > start {
		return pos - start
	}
	return -1
}

func ScanStringLiteral(text string, start int, failWhenMissingEndQuote ...bool) int {
	failMissing := len(failWhenMissingEndQuote) > 0 && failWhenMissingEndQuote[0]
	pos := start

	ch := peekByte(text, pos)
	if ch == 'h' || ch == 'H' {
		pos++
		ch = peekByte(text, pos)
	}

	isVerbatim := false
	if ch == '@' {
		isVerbatim = true
		pos++
		ch = peekByte(text, pos)
	}

	switch {
	case ch == '\'' || ch == '"':
		quote := ch
		pos++
		pos += scanStringLiteralContentForTokenParser(text, pos, quote, isVerbatim)
		if peekByte(text, pos) == quote {
			pos++
		} else if failMissing {
			return -1
		}
	case strings.HasPrefix(text[pos:], "```"):
		pos += len("```")
		pos += scanMultiLineStringLiteralContentForTokenParser(text, pos, "```")
		if strings.HasPrefix(text[pos:], "```") {
			pos += len("```")
		} else if failMissing {
			return -1
		}
	case strings.HasPrefix(text[pos:], "~~~"):
		pos += len("~~~")
		pos += scanMultiLineStringLiteralContentForTokenParser(text, pos, "~~~")
		if strings.HasPrefix(text[pos:], "~~~") {
			pos += len("~~~")
		} else if failMissing {
			return -1
		}
	default:
		return -1
	}

	if pos > start {
		return pos - start
	}
	return -1
}

func scanStringLiteralContentForTokenParser(text string, start int, quote byte, isVerbatim bool) int {
	pos := start
	for pos < len(text) {
		ch := peekByte(text, pos)
		switch {
		case ch == quote && isVerbatim && peekByte(text, pos+1) == quote:
			pos += 2
		case ch == '\\' && !isVerbatim:
			escapeLength := scanStringEscape(text, pos)
			if escapeLength <= 0 {
				return pos - start
			}
			pos += escapeLength
		case ch == quote || ch == '\r' || ch == '\n':
			return pos - start
		default:
			pos++
		}
	}
	return pos - start
}

func scanMultiLineStringLiteralContentForTokenParser(text string, start int, endQuote string) int {
	pos := start
	for pos < len(text) && !strings.HasPrefix(text[pos:], endQuote) {
		pos++
	}
	return pos - start
}

func peekByte(text string, position int) byte {
	if position >= 0 && position < len(text) {
		return text[position]
	}
	return 0
}
