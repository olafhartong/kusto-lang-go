package kustolang

// Ported from the coordinate semantics of Kusto.Language syntax elements.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"unicode/utf16"
	"unicode/utf8"
)

func utf16TextWidth(text string) int {
	width := 0
	for _, r := range text {
		units := utf16.RuneLen(r)
		if units < 0 {
			units = 1
		}
		width += units
	}
	return width
}

// utf16TextWidthOfBytePrefix maps a byte boundary to a UTF-16 boundary. Invalid
// or partial trailing UTF-8 is excluded, matching the repository's existing
// byte-to-UTF-16 conversion behavior.
func utf16TextWidthOfBytePrefix(text string, byteOffset int) int {
	if byteOffset <= 0 {
		return 0
	}
	if byteOffset > len(text) {
		byteOffset = len(text)
	}
	width := 0
	for index := 0; index < byteOffset; {
		r, size := utf8.DecodeRuneInString(text[index:])
		if size == 0 || index+size > byteOffset {
			break
		}
		units := utf16.RuneLen(r)
		if units < 0 {
			units = 1
		}
		width += units
		index += size
	}
	return width
}
