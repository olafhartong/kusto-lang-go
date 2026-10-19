package kustolang

import (
	"strings"
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
			return strings.EqualFold(string(input), string(ch))
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
