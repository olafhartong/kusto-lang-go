package kustolang

import (
	"unicode"
	"unicode/utf16"
)

// Ported from src/Kusto.Language/Utils/StringAndNumberComparer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func StringAndNumberCompareOrdinal(left string, right string) int {
	return StringAndNumberCompare(left, right, true)
}

func StringAndNumberCompareOrdinalIgnoreCase(left string, right string) int {
	return StringAndNumberCompare(left, right, false)
}

func StringAndNumberCompare(left string, right string, caseSensitive bool) int {
	leftUnits := utf16.Encode([]rune(left))
	rightUnits := utf16.Encode([]rune(right))
	leftStart := 0
	rightStart := 0

	for leftStart < len(leftUnits) && rightStart < len(rightUnits) {
		leftSegmentIsNumber := unicode.IsDigit(rune(leftUnits[leftStart]))
		rightSegmentIsNumber := unicode.IsDigit(rune(rightUnits[rightStart]))

		switch {
		case leftSegmentIsNumber && rightSegmentIsNumber:
			leftNumberLength := lengthOfNumberUnits(leftUnits, leftStart)
			rightNumberLength := lengthOfNumberUnits(rightUnits, rightStart)
			comparison := compareNumberSegment(leftUnits, leftStart, leftNumberLength, rightUnits, rightStart, rightNumberLength)
			if comparison != 0 {
				return comparison
			}
			leftStart += leftNumberLength
			rightStart += rightNumberLength
		case !leftSegmentIsNumber && !rightSegmentIsNumber:
			commonLength := commonLengthWithoutDigits(leftUnits, leftStart, rightUnits, rightStart)
			comparison := compareUnitSegment(leftUnits, leftStart, rightUnits, rightStart, commonLength, caseSensitive)
			if comparison != 0 {
				return comparison
			}
			leftStart += commonLength
			rightStart += commonLength
		default:
			return int(leftUnits[leftStart]) - int(rightUnits[rightStart])
		}
	}

	if len(leftUnits)-leftStart > 0 {
		return 1
	}
	if len(rightUnits)-rightStart > 0 {
		return -1
	}
	return 0
}

func commonLengthWithoutDigits(left []uint16, leftStart int, right []uint16, rightStart int) int {
	length := 0
	for leftStart+length < len(left) &&
		!unicode.IsDigit(rune(left[leftStart+length])) &&
		rightStart+length < len(right) &&
		!unicode.IsDigit(rune(right[rightStart+length])) {
		length++
	}
	return length
}

func compareNumberSegment(left []uint16, leftStart int, leftLength int, right []uint16, rightStart int, rightLength int) int {
	for leftStart < len(left) && leftLength > rightLength && left[leftStart] == '0' {
		leftStart++
	}
	for rightStart < len(right) && rightLength > leftLength && right[rightStart] == '0' {
		rightStart++
	}
	if leftLength > 0 && leftLength == rightLength {
		return compareUnitSegment(left, leftStart, right, rightStart, leftLength, true)
	}
	return leftLength - rightLength
}

func lengthOfNumberUnits(text []uint16, startIndex int) int {
	index := startIndex
	for ; index < len(text); index++ {
		if !unicode.IsDigit(rune(text[index])) {
			break
		}
	}
	return index - startIndex
}

func compareUnitSegment(left []uint16, leftStart int, right []uint16, rightStart int, length int, caseSensitive bool) int {
	return ordinalCompareUnits(left[leftStart:leftStart+length], right[rightStart:rightStart+length], !caseSensitive)
}
