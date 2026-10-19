package kustolang

import "unicode/utf8"

// parserFactsOptionalUTF16Position converts the first optional public
// position. Additional values are ignored, matching other optional Go facades.
func parserFactsOptionalUTF16Position(text string, positions []int) (int, bool) {
	position := 0
	if len(positions) > 0 {
		position = positions[0]
	}
	return parserFactsUTF16ToByteOffset(text, position)
}

func parserFactsUTF16ToByteOffset(text string, position int) (int, bool) {
	if position < 0 {
		return 0, false
	}

	utf16Position := 0
	for bytePosition := 0; bytePosition < len(text); {
		if utf16Position == position {
			return bytePosition, true
		}
		r, width := utf8.DecodeRuneInString(text[bytePosition:])
		units := 1
		if r > 0xffff {
			units = 2
		}
		if position < utf16Position+units {
			return 0, false
		}
		utf16Position += units
		bytePosition += width
	}
	if utf16Position == position {
		return len(text), true
	}
	return 0, false
}

func parserFactsByteToUTF16Offset(text string, offset int) (int, bool) {
	if offset < 0 || offset > len(text) {
		return 0, false
	}

	utf16Position := 0
	for bytePosition := 0; bytePosition < offset; {
		r, width := utf8.DecodeRuneInString(text[bytePosition:])
		if bytePosition+width > offset {
			return 0, false
		}
		if r > 0xffff {
			utf16Position += 2
		} else {
			utf16Position++
		}
		bytePosition += width
	}
	return utf16Position, true
}
