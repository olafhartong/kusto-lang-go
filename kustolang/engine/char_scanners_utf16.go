package kustolang

import (
	"unicode/utf16"
)

// These scanners preserve the UTF-16 code-unit behavior of C# Parser<char>.
// The existing byte scanners remain the native UTF-8 Go API.

func UTF16CodeUnits(text string) []uint16 {
	return utf16.Encode([]rune(text))
}

func UTF16Chars(text string, ignoreCase ...bool) Parser[uint16] {
	units := UTF16CodeUnits(text)
	ignore := len(ignoreCase) > 0 && ignoreCase[0]
	if len(units) == 1 {
		return UTF16Char(units[0], ignore)
	}
	return MatchConsumer(func(source Source[uint16], start int) int {
		for i, unit := range units {
			if source.IsEnd(start + i) {
				return ^i
			}
			got := source.Peek(start + i)
			if got != unit {
				if !ignore {
					return ^i
				}
				candidate := make([]uint16, len(units))
				for j := range candidate {
					if source.IsEnd(start + j) {
						return ^j
					}
					candidate[j] = source.Peek(start + j)
				}
				if ordinalCompareUnits(candidate, units, true) == 0 {
					return len(units)
				}
				return ^i
			}
		}
		return len(units)
	})
}

func UTF16Char(ch uint16, ignoreCase ...bool) Parser[uint16] {
	if len(ignoreCase) > 0 && ignoreCase[0] {
		return Match(func(input uint16) bool { return ordinalCaseKey(rune(input)) == ordinalCaseKey(rune(ch)) })
	}
	return MatchItem(ch)
}

var UTF16Letter Parser[uint16] = Match(func(ch uint16) bool {
	return IsLetter(rune(ch))
}).WithTag("<letter>")

var UTF16Digit Parser[uint16] = Match(func(ch uint16) bool {
	return IsDigit(rune(ch))
}).WithTag("<digit>")

var UTF16HexDigit Parser[uint16] = Match(func(ch uint16) bool {
	return IsHexDigit(rune(ch))
}).WithTag("<hex-digit>")

var UTF16Whitespace Parser[uint16] = Match(func(ch uint16) bool {
	return IsWhitespace(rune(ch))
}).WithTag("<whitespace>")

var UTF16LineBreak Parser[uint16] = Or(
	And(UTF16Char('\r'), ZeroOrOne(UTF16Char('\n'))),
	UTF16Char('\n'),
	UTF16Char('\u2028'),
	UTF16Char('\u2029'),
).WithTag("<line-break>")

func UTF16ScannerMatches(scanner Parser[uint16], text string, offsetAndLength ...int) bool {
	if scanner == nil {
		return false
	}
	units := UTF16CodeUnits(text)
	offset := 0
	length := len(units)
	if len(offsetAndLength) >= 1 {
		offset = offsetAndLength[0]
	}
	if len(offsetAndLength) >= 2 {
		length = offsetAndLength[1]
	}
	if offset < 0 || offset > len(units) || length < 0 {
		return false
	}
	return scanner.Scan(NewArraySource(units), offset) == length
}
