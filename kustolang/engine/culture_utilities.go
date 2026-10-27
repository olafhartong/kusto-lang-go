package kustolang

import (
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// Ported from src/Kusto.Language/Utils/CultureUtilities.cs (13.0.0).
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// ordinalCaseKey follows .NET ordinal casing, including supplementary scalars.
// Unicode simple folding (strings.EqualFold) also equates ASCII with long s and
// Kelvin sign; ordinal casing deliberately keeps those characters distinct.
func ordinalCaseKey(ch rune) rune {
	upper := unicode.ToUpper(ch)
	if ch > unicode.MaxASCII && upper <= unicode.MaxASCII {
		return ch
	}
	return upper
}

func ordinalEqualFold(left, right string) bool {
	if left == right {
		return true
	}
	for len(left) > 0 && len(right) > 0 {
		l, lw := utf8.DecodeRuneInString(left)
		r, rw := utf8.DecodeRuneInString(right)
		if ordinalCaseKey(l) != ordinalCaseKey(r) {
			return false
		}
		left, right = left[lw:], right[rw:]
	}
	return len(left) == len(right)
}

func ordinalCompare(left, right string, ignoreCase bool) int {
	if left == right {
		return 0
	}
	return ordinalCompareUnits(utf16.Encode([]rune(left)), utf16.Encode([]rune(right)), ignoreCase)
}

func ordinalCompareUnits(left, right []uint16, ignoreCase bool) int {
	if ignoreCase {
		left = ordinalFoldUnits(left)
		right = ordinalFoldUnits(right)
		for len(left) > 0 && len(right) > 0 {
			l, lw := ordinalScalar(left)
			r, rw := ordinalScalar(right)
			if l != r {
				return int(l - r)
			}
			left, right = left[lw:], right[rw:]
		}
		return len(left) - len(right)
	}
	for i := 0; i < len(left) && i < len(right); i++ {
		if left[i] != right[i] {
			return int(left[i]) - int(right[i])
		}
	}
	return len(left) - len(right)
}

func ordinalScalar(units []uint16) (rune, int) {
	if units[0] >= 0xd800 && units[0] <= 0xdbff && len(units) > 1 && units[1] >= 0xdc00 && units[1] <= 0xdfff {
		return utf16.DecodeRune(rune(units[0]), rune(units[1])), 2
	}
	return rune(units[0]), 1
}

func ordinalFoldUnits(units []uint16) []uint16 {
	result := append([]uint16(nil), units...)
	for i := 0; i < len(result); i++ {
		ch := rune(result[i])
		if ch >= 0xd800 && ch <= 0xdbff && i+1 < len(result) && result[i+1] >= 0xdc00 && result[i+1] <= 0xdfff {
			key := ordinalCaseKey(utf16.DecodeRune(ch, rune(result[i+1])))
			upper, lower := utf16.EncodeRune(key)
			result[i], result[i+1] = uint16(upper), uint16(lower)
			i++
		} else {
			result[i] = uint16(ordinalCaseKey(ch))
		}
	}
	return result
}

func invariantLower(text string) string {
	return strings.Map(func(ch rune) rune {
		// Invariant .NET casing leaves LATIN CAPITAL LETTER I WITH DOT ABOVE
		// unchanged instead of applying the Turkish or Unicode simple mapping.
		if ch == '\u0130' {
			return ch
		}
		return unicode.ToLower(ch)
	}, text)
}

func ordinalIgnoreCaseKey(text string) string {
	return strings.Map(ordinalCaseKey, text)
}

func ordinalHasPrefix(text, prefix string) bool {
	units, pattern := utf16.Encode([]rune(text)), utf16.Encode([]rune(prefix))
	return len(pattern) <= len(units) && ordinalCompareUnits(units[:len(pattern)], pattern, true) == 0
}

func ordinalHasSuffix(text, suffix string) bool {
	units, pattern := utf16.Encode([]rune(text)), utf16.Encode([]rune(suffix))
	return len(pattern) <= len(units) && ordinalCompareUnits(units[len(units)-len(pattern):], pattern, true) == 0
}
