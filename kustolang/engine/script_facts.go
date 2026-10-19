package kustolang

import "sort"

// Ported from src/Kusto.Language/Parser/ScriptFacts.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func KustoBlockStarts(text string, lineStarts []int) []int {
	blockStarts := LineSeparatedBlockStarts(text, lineStarts)
	blockStarts = RemoveDuplicateBlockStarts(blockStarts)
	blockStarts = RemoveInvalidKustoBlockStarts(blockStarts, text)
	return blockStarts
}

func LineSeparatedBlockStarts(text string, lineStarts []int) []int {
	blockStarts := []int{0}
	newBlockNextWhitespaceLine := false
	newBlockNextNonWhitespaceLine := false

	for i, lineStart := range lineStarts {
		lineEnd := len(text)
		if i+1 < len(lineStarts) {
			lineEnd = lineStarts[i+1]
		}
		allWhitespace := IsWhitespaceOnly(text, lineStart, lineEnd-lineStart)
		if allWhitespace {
			if newBlockNextWhitespaceLine {
				blockStarts = append(blockStarts, lineStart)
			}
			newBlockNextWhitespaceLine = true
			newBlockNextNonWhitespaceLine = true
		} else if newBlockNextNonWhitespaceLine {
			blockStarts = append(blockStarts, lineStart)
			newBlockNextWhitespaceLine = false
			newBlockNextNonWhitespaceLine = false
		}
	}

	return blockStarts
}

func RemoveDuplicateBlockStarts(blockStarts []int) []int {
	starts := append([]int(nil), blockStarts...)
	sort.Ints(starts)
	out := starts[:0]
	for _, start := range starts {
		if len(out) == 0 || out[len(out)-1] != start {
			out = append(out, start)
		}
	}
	return append([]int(nil), out...)
}

func RemoveInvalidKustoBlockStarts(blockStarts []int, text string) []int {
	return RemoveInvalidBlockStarts(blockStarts, InvalidKustoBlockStartRanges(text))
}

func RemoveInvalidBlockStarts(blockStarts []int, invalidRanges []TextRange) []int {
	starts := append([]int(nil), blockStarts...)
	sort.Ints(starts)
	for rangeIndex, blockIndex := 0, 0; rangeIndex < len(invalidRanges) && blockIndex < len(starts); {
		rangeValue := invalidRanges[rangeIndex]
		blockStart := starts[blockIndex]
		switch {
		case rangeValue.End() <= blockStart:
			rangeIndex++
		case rangeValue.Start >= blockStart:
			blockIndex++
		default:
			starts = append(starts[:blockIndex], starts[blockIndex+1:]...)
		}
	}
	return starts
}

func InvalidKustoBlockStartRanges(text string) []TextRange {
	var invalidRanges []TextRange
	for pos := 0; pos < len(text); {
		if stringLength := ScanStringLiteral(text, pos); stringLength > 0 {
			invalidRanges = append(invalidRanges, NewTextRange(pos, stringLength))
			pos += stringLength
			continue
		}
		if commentLength := ScanComment(text, pos); commentLength > 0 {
			invalidRanges = append(invalidRanges, NewTextRange(pos, commentLength))
			pos += commentLength
			continue
		}
		pos++
	}
	return invalidRanges
}
