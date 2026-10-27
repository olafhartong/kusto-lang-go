package kustolang

import (
	"unicode/utf8"
)

// Ported from src/Kusto.Language/Parser/CharScanners.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func Chars(text string, ignoreCase ...bool) Parser[byte] {
	return MatchText(text, ignoreCase...)
}

func Char(ch rune, ignoreCase ...bool) Parser[byte] {
	if len(ignoreCase) > 0 && ignoreCase[0] {
		return matchRune(func(input rune) bool {
			return ordinalCaseKey(input) == ordinalCaseKey(ch)
		})
	}
	return matchRune(func(input rune) bool { return input == ch })
}

var Letter Parser[byte] = matchRune(func(ch rune) bool {
	return IsLetter(rune(ch))
}).WithTag("<letter>")

var Digit Parser[byte] = matchRune(func(ch rune) bool {
	return IsDigit(rune(ch))
}).WithTag("<digit>")

var HexDigit Parser[byte] = matchRune(func(ch rune) bool {
	return IsHexDigit(rune(ch))
}).WithTag("<hex-digit>")

var Whitespace Parser[byte] = matchRune(func(ch rune) bool {
	return IsWhitespace(rune(ch))
}).WithTag("<whitespace>")

var LineBreak Parser[byte] = Or(
	Chars("\r\n"),
	Char('\r'),
	Char('\n'),
	Chars("\u2028"),
	Chars("\u2029"),
).WithTag("<line-break>")

func matchRune(predicate func(rune) bool) Parser[byte] {
	return MatchConsumer(func(source Source[byte], start int) int {
		if source.IsEnd(start) {
			return -1
		}
		first := source.Peek(start)
		if first < utf8.RuneSelf {
			if predicate(rune(first)) {
				return 1
			}
			return -1
		}
		var buf [utf8.UTFMax]byte
		buf[0] = first
		for size := 2; size <= utf8.UTFMax; size++ {
			if source.IsEnd(start + size - 1) {
				return -1
			}
			buf[size-1] = source.Peek(start + size - 1)
			if !utf8.FullRune(buf[:size]) {
				continue
			}
			r, width := utf8.DecodeRune(buf[:size])
			if r == utf8.RuneError && width == 1 {
				return -1
			}
			if width == size {
				if predicate(r) {
					return size
				}
				return -1
			}
		}
		return -1
	})
}

// scanOrdinalText retains byte offsets while comparing full Unicode scalars.
func scanOrdinalText(source Source[byte], start int, text string) int {
	offset := start
	mismatch := -1
	for _, want := range text {
		var buf [utf8.UTFMax]byte
		width := 0
		var got rune
		for width < len(buf) {
			if source.IsEnd(offset + width) {
				return ^(offset + width - start)
			}
			buf[width] = source.Peek(offset + width)
			width++
			if utf8.FullRune(buf[:width]) {
				got, _ = utf8.DecodeRune(buf[:width])
				break
			}
		}
		if ordinalCaseKey(got) != ordinalCaseKey(want) && mismatch < 0 {
			mismatch = offset - start
		}
		offset += width
	}
	if mismatch >= 0 {
		return ^mismatch
	}
	return offset - start
}
