package kustolang

import "strings"

// Ported from src/Kusto.Language/Utils/StringExtensions.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func StringFilter(text string, predicate func(rune) bool) string {
	all := true
	for _, ch := range text {
		if !predicate(ch) {
			all = false
			break
		}
	}
	if all {
		return text
	}

	var builder strings.Builder
	builder.Grow(len(text))
	for _, ch := range text {
		if predicate(ch) {
			builder.WriteRune(ch)
		}
	}
	return builder.String()
}
