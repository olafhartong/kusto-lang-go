package kustolang

// Ported from src/Kusto.Language/Editor/TextRange.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type TextRange struct {
	Start  int
	Length int
}

var TextRangeEmpty = NewTextRange(0, 0)

func NewTextRange(start int, length int) TextRange {
	return TextRange{Start: start, Length: length}
}

func TextRangeFromBounds(start int, end int) TextRange {
	return TextRange{Start: start, Length: end - start}
}

func (r TextRange) End() int {
	return r.Start + r.Length
}

func (r TextRange) Overlaps(other TextRange) bool {
	return TextRangesOverlap(r.Start, r.Length, other.Start, other.Length)
}

func TextRangesOverlap(startA int, lengthA int, startB int, lengthB int) bool {
	endA := startA + lengthA
	endB := startB + lengthB
	return maxInt(startA, startB) <= minInt(endA, endB)
}
