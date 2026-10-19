package kustolang

import (
	"unicode"
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
	lower := make([]uint16, len(units))
	upper := make([]uint16, len(units))
	for i, unit := range units {
		lower[i] = utf16SimpleCase(unit, unicode.ToLower)
		upper[i] = utf16SimpleCase(unit, unicode.ToUpper)
	}
	return MatchConsumer(func(source Source[uint16], start int) int {
		for i, unit := range units {
			if source.IsEnd(start + i) {
				return ^i
			}
			got := source.Peek(start + i)
			if got != unit && (!ignore || got != lower[i] && got != upper[i]) {
				return ^i
			}
		}
		return len(units)
	})
}

func UTF16Char(ch uint16, ignoreCase ...bool) Parser[uint16] {
	if len(ignoreCase) > 0 && ignoreCase[0] {
		upper := utf16SimpleCase(ch, unicode.ToUpper)
		lower := utf16SimpleCase(ch, unicode.ToLower)
		return Match(func(input uint16) bool { return input == ch || input == upper || input == lower })
	}
	return MatchItem(ch)
}

func utf16SimpleCase(ch uint16, convert func(rune) rune) uint16 {
	mapped := convert(rune(ch))
	if mapped < 0 || mapped > 0xffff {
		return ch
	}
	return uint16(mapped)
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
