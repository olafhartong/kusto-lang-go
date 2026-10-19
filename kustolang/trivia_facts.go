package kustolang

// Ported from src/Kusto.Language/Parser/TriviaFacts.cs at
// c94c9e761b3d095e2b9c21e3ec3e92afa0af4856.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// TriviaFactsAPI mirrors the upstream static TriviaFacts class. Positions and
// lengths are UTF-16 code units, matching System.String.
type TriviaFactsAPI struct{}

// TryGetCommentSpan gets the comment containing or immediately preceding the
// position. The returned span uses UTF-16 code units.
func (TriviaFactsAPI) TryGetCommentSpan(trivia string, position int) (start, length int, ok bool) {
	units := UTF16CodeUnits(trivia)
	for current := 0; current < len(units) && current <= position; current++ {
		commentLength := textFactsCommentLength(units, current)
		if commentLength == 0 {
			continue
		}
		if position >= current && position <= current+commentLength {
			return current, commentLength, true
		}
		current += commentLength - 1
	}
	return 0, 0, false
}

// TriviaFacts exposes source-shaped access to the upstream static API.
var TriviaFacts TriviaFactsAPI

func textFactsCommentLength(units []uint16, start int) int {
	if start < 0 || start >= len(units) || textFactsIsWhitespace(units[start]) {
		return 0
	}
	end := start
	for end < len(units) {
		if length := textFactsLineBreakLength(units, end); length > 0 {
			end += length
			break
		}
		end++
	}
	return end - start
}
