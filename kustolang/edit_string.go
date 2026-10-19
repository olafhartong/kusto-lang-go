package kustolang

// Ported from src/Kusto.Language/Editor/EditString.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"strings"
	"unicode/utf16"

	internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"
)

// EditString is an immutable string-like value that records the collective
// edits used to produce its current text. Positions use UTF-16 code units.
type EditString struct {
	value internal.EditString
}

var EditStringEmpty = NewEditString("")

func NewEditString(text string) EditString {
	return wrapEditString(internal.NewEditString(text))
}

func wrapEditString(value internal.EditString) EditString {
	return EditString{value: value}
}

func (s EditString) OriginalText() string { return s.value.OriginalText }

func (s EditString) CurrentText() string { return s.value.CurrentText }

func (s EditString) Length() int {
	return byteOffsetToUTF16Position(s.value.CurrentText, len(s.value.CurrentText))
}

func (s EditString) At(index int) uint16 {
	units := utf16.Encode([]rune(s.value.CurrentText))
	if index < 0 || index >= len(units) {
		panic(internal.ArgumentOutOfRangeError{ParameterName: "index"})
	}
	return units[index]
}

func (s EditString) IndexOf(value string) int {
	byteIndex := strings.Index(s.value.CurrentText, value)
	if byteIndex < 0 {
		return -1
	}
	return byteOffsetToUTF16Position(s.value.CurrentText, byteIndex)
}

func (s EditString) String() string { return s.value.CurrentText }

func (s EditString) GetChanges() []TextEdit {
	changes := s.value.GetChanges()
	result := make([]TextEdit, len(changes))
	for i, change := range changes {
		start := byteOffsetToUTF16Position(s.value.OriginalText, change.Start)
		end := byteOffsetToUTF16Position(s.value.OriginalText, change.Start+change.DeleteLength)
		result[i] = NewTextEdit(start, end-start, change.InsertText)
	}
	return result
}

func ConvertToSequentialChanges(collectiveChanges []TextEdit) []TextEdit {
	result := make([]TextEdit, 0, len(collectiveChanges))
	delta := 0
	for _, change := range collectiveChanges {
		insertLength := byteOffsetToUTF16Position(change.InsertText, len(change.InsertText))
		result = append(result, NewTextEdit(change.Start+delta, change.DeleteLength, change.InsertText))
		delta += insertLength - change.DeleteLength
	}
	return result
}

func (s EditString) ReplaceAt(start, length int, text string) EditString {
	byteStart, byteLength := editStringUTF16Range(s.value.CurrentText, start, length)
	return wrapEditString(s.value.ReplaceAt(byteStart, byteLength, text))
}

func (s EditString) Insert(position int, text string) EditString {
	return s.ReplaceAt(position, 0, text)
}

func (s EditString) Append(text string) EditString {
	return wrapEditString(s.value.Append(text))
}

func (s EditString) Prepend(text string) EditString {
	return wrapEditString(s.value.Prepend(text))
}

func (s EditString) Remove(start, length int) EditString {
	return s.ReplaceAt(start, length, "")
}

func (s EditString) Substring(start int, length ...int) EditString {
	utf16Length := s.Length() - start
	if len(length) > 0 {
		utf16Length = length[0]
	}
	byteStart, byteLength := editStringUTF16Range(s.value.CurrentText, start, utf16Length)
	return wrapEditString(s.value.Substring(byteStart, byteLength))
}

func (s EditString) TrimLeft() EditString { return wrapEditString(s.value.TrimLeft()) }

func (s EditString) TrimRight() EditString { return wrapEditString(s.value.TrimRight()) }

func (s EditString) Trim() EditString { return wrapEditString(s.value.Trim()) }

func (s EditString) Replace(oldText, newText string) EditString {
	return wrapEditString(s.value.Replace(oldText, newText))
}

func (s EditString) Apply(edit TextEdit) EditString {
	byteStart, byteLength := editStringUTF16Range(s.value.CurrentText, edit.Start, edit.DeleteLength)
	return wrapEditString(s.value.Apply(internal.NewTextEdit(byteStart, byteLength, edit.InsertText)))
}

func (s EditString) ApplyAll(edits []TextEdit) EditString {
	converted := s.editsToBytes(edits)
	return wrapEditString(s.value.ApplyAll(converted))
}

func (s EditString) CanApplyAll(edits []TextEdit) bool {
	converted, ok := s.tryEditsToBytes(edits)
	return ok && s.value.CanApplyAll(converted)
}

func (s EditString) GetCurrentPosition(originalPosition int, bias ...PositionBias) int {
	bytePosition := editStringUTF16Position(s.value.OriginalText, originalPosition)
	currentBytePosition := s.value.GetCurrentPosition(bytePosition, bias...)
	return byteOffsetToUTF16Position(s.value.CurrentText, currentBytePosition)
}

func (s EditString) GetCurrentRange(originalStart, originalLength int) TextRange {
	currentStart := s.GetCurrentPosition(originalStart, PositionBiasRight)
	currentEnd := s.GetCurrentPosition(originalStart+originalLength, PositionBiasLeft)
	return internal.NewTextRange(currentStart, currentEnd-currentStart)
}

func (s EditString) GetOriginalPosition(currentPosition int, bias ...PositionBias) int {
	bytePosition := editStringUTF16Position(s.value.CurrentText, currentPosition)
	originalBytePosition := s.value.GetOriginalPosition(bytePosition, bias...)
	return byteOffsetToUTF16Position(s.value.OriginalText, originalBytePosition)
}

func (s EditString) GetChangeRange() TextRange {
	rng := s.value.GetChangeRange()
	start := byteOffsetToUTF16Position(s.value.CurrentText, rng.Start)
	end := byteOffsetToUTF16Position(s.value.CurrentText, rng.End())
	return internal.NewTextRange(start, end-start)
}

func (s EditString) RemoveBlankLines() EditString {
	return wrapEditString(s.value.RemoveBlankLines())
}

func (s EditString) ReplaceLineBreaks(newValue string) EditString {
	return wrapEditString(s.value.ReplaceLineBreaks(newValue))
}

func (s EditString) editsToBytes(edits []TextEdit) []internal.TextEdit {
	converted, ok := s.tryEditsToBytes(edits)
	if !ok {
		panic(internal.InvalidOperationError{Message: "An edit is specified outside a UTF-16 character boundary."})
	}
	return converted
}

func (s EditString) tryEditsToBytes(edits []TextEdit) ([]internal.TextEdit, bool) {
	converted := make([]internal.TextEdit, len(edits))
	for i, edit := range edits {
		start, startOK := editStringTryUTF16Position(s.value.CurrentText, edit.Start)
		end, endOK := editStringTryUTF16Position(s.value.CurrentText, edit.Start+edit.DeleteLength)
		if edit.Start < 0 || edit.DeleteLength < 0 || !startOK || !endOK {
			return nil, false
		}
		converted[i] = internal.NewTextEdit(start, end-start, edit.InsertText)
	}
	return converted, true
}

func editStringUTF16Range(text string, start, length int) (int, int) {
	if start < 0 {
		panic(internal.ArgumentOutOfRangeError{ParameterName: "start"})
	}
	if length < 0 {
		panic(internal.ArgumentOutOfRangeError{ParameterName: "length"})
	}
	byteStart := editStringUTF16Position(text, start)
	byteEnd := editStringUTF16Position(text, start+length)
	return byteStart, byteEnd - byteStart
}

func editStringUTF16Position(text string, position int) int {
	offset, ok := editStringTryUTF16Position(text, position)
	if !ok {
		panic(internal.ArgumentOutOfRangeError{ParameterName: "position", Message: "position is outside the text or splits a UTF-16 surrogate pair"})
	}
	return offset
}

func editStringTryUTF16Position(text string, position int) (int, bool) {
	if position < 0 {
		return 0, false
	}
	offset := utf16PositionToByteOffset(text, position)
	return offset, byteOffsetToUTF16Position(text, offset) == position
}
