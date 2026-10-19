package kustolang

import "unicode/utf8"

// Ported from src/Kusto.Language/Parser/TriviaFacts.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func TryGetCommentSpan(trivia string, position int) (start int, length int, ok bool) {
	for pos := 0; pos < len(trivia) && pos <= position; {
		commentLength := commentLengthAt(trivia, pos)
		if commentLength > 0 {
			if position >= pos && position <= pos+commentLength {
				return pos, commentLength, true
			}
			pos += commentLength
		} else {
			_, size := utf8.DecodeRuneInString(trivia[pos:])
			pos += size
		}
	}
	return 0, 0, false
}

func commentLengthAt(trivia string, start int) int {
	end := start
	if start >= 0 && start < len(trivia) {
		ch, _ := utf8.DecodeRuneInString(trivia[start:])
		if IsWhitespace(ch) {
			return 0
		}
		for ; end < len(trivia); end++ {
			if lineBreakLength := LineBreakLength(trivia, end); lineBreakLength > 0 {
				end += lineBreakLength
				break
			}
		}
	}
	return end - start
}
