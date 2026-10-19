package kustolang

import (
	"strings"
	"unicode/utf8"
)

// Ported from src/Kusto.Language/Editor/EditString.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type EditString struct {
	OriginalText string
	CurrentText  string
	changes      []editStringEdit
}

var EditStringEmpty = NewEditString("")

func NewEditString(text string) EditString {
	return newEditString(text, text, nil)
}

func newEditString(originalText string, currentText string, changes []editStringEdit) EditString {
	return EditString{
		OriginalText: originalText,
		CurrentText:  currentText,
		changes:      ListToReadOnly(changes),
	}
}

func (s EditString) Length() int {
	return len(s.CurrentText)
}

func (s EditString) At(index int) byte {
	return s.CurrentText[index]
}

func (s EditString) IndexOf(value string) int {
	return strings.Index(s.CurrentText, value)
}

func (s EditString) String() string {
	return s.CurrentText
}

func (s EditString) GetChanges() []TextEdit {
	textChanges := make([]TextEdit, 0, len(s.changes))
	delta := 0
	for _, edit := range s.changes {
		insertStart := edit.Start + delta
		insertEnd := insertStart + edit.InsertLength
		insertText := ""
		if insertStart >= 0 && insertStart <= len(s.CurrentText) && insertEnd <= len(s.CurrentText) {
			insertText = s.CurrentText[insertStart:insertEnd]
		}
		textChanges = append(textChanges, NewTextReplacement(edit.Start, edit.DeleteLength, insertText))
		delta = delta - edit.DeleteLength + edit.InsertLength
	}
	return ListToReadOnly(textChanges)
}

func ConvertToSequentialChanges(collectiveChanges []TextEdit) []TextEdit {
	sequentialChanges := make([]TextEdit, 0, len(collectiveChanges))
	delta := 0
	for _, edit := range collectiveChanges {
		sequentialChanges = append(sequentialChanges, NewTextReplacement(edit.Start+delta, edit.DeleteLength, edit.InsertText))
		delta = delta - edit.DeleteLength + len(edit.InsertText)
	}
	return ListToReadOnly(sequentialChanges)
}

func (s EditString) ReplaceAt(start int, length int, text string) EditString {
	return s.Apply(NewTextReplacement(start, length, text))
}

func (s EditString) Insert(position int, text string) EditString {
	return s.ReplaceAt(position, 0, text)
}

func (s EditString) Append(text string) EditString {
	return s.ReplaceAt(len(s.CurrentText), 0, text)
}

func (s EditString) Prepend(text string) EditString {
	return s.ReplaceAt(0, 0, text)
}

func (s EditString) Remove(start int, length int) EditString {
	return s.ReplaceAt(start, length, "")
}

func (s EditString) Substring(start int, length ...int) EditString {
	subLength := len(s.CurrentText) - start
	if len(length) > 0 {
		subLength = length[0]
	}
	newText := s.CurrentText[start : start+subLength]
	newEdits := make([]editStringEdit, 0, 2)
	endDeleteStart := start + subLength
	endDeleteLength := len(s.CurrentText) - endDeleteStart
	if start > 0 {
		newEdits = append(newEdits, newEditStringEdit(0, start, 0))
	}
	if endDeleteStart > 0 {
		newEdits = append(newEdits, newEditStringEdit(endDeleteStart, endDeleteLength, 0))
	}
	return s.applyEdits(newText, newEdits)
}

func (s EditString) TrimLeft() EditString {
	length := WhitespaceCount(s.CurrentText, 0)
	if length > 0 {
		return s.Remove(0, length)
	}
	return s
}

func (s EditString) TrimRight() EditString {
	length := whitespaceCountBefore(s.CurrentText, len(s.CurrentText))
	if length > 0 {
		return s.Remove(len(s.CurrentText)-length, length)
	}
	return s
}

func (s EditString) Trim() EditString {
	return s.TrimLeft().TrimRight()
}

func (s EditString) Replace(oldText string, newText string) EditString {
	if oldText == "" {
		panic(ArgumentError{ParameterName: "oldText", Message: "oldText must not be empty"})
	}
	newCurrentText := strings.ReplaceAll(s.CurrentText, oldText, newText)
	if newCurrentText == s.CurrentText {
		return s
	}

	startIndex := 0
	newEdits := []editStringEdit{}
	for {
		oldValueStart := strings.Index(s.CurrentText[startIndex:], oldText)
		if oldValueStart < 0 {
			break
		}
		oldValueStart += startIndex
		newEdits = append(newEdits, newEditStringEdit(oldValueStart, len(oldText), len(newText)))
		startIndex = oldValueStart + len(oldText)
	}

	return s.applyEdits(newCurrentText, newEdits)
}

func (s EditString) Apply(edit TextEdit) EditString {
	if edit.Start < 0 || edit.Start > len(s.CurrentText) {
		panic(ArgumentOutOfRangeError{ParameterName: "Start"})
	}
	if edit.Start+edit.DeleteLength > len(s.CurrentText) {
		panic(ArgumentOutOfRangeError{ParameterName: "DeleteLength"})
	}

	newText := s.CurrentText
	if edit.DeleteLength > 0 {
		newText = newText[:edit.Start] + newText[edit.Start+edit.DeleteLength:]
	}
	if len(edit.InsertText) > 0 {
		newText = newText[:edit.Start] + edit.InsertText + newText[edit.Start:]
	}

	insertLength := len(edit.InsertText)
	unchangedReplacement := insertLength == edit.DeleteLength &&
		edit.Start+insertLength <= len(s.CurrentText) &&
		s.CurrentText[edit.Start:edit.Start+insertLength] == edit.InsertText
	if !unchangedReplacement {
		return s.applyEdits(newText, []editStringEdit{newEditStringEdit(edit.Start, edit.DeleteLength, insertLength)})
	}
	return newEditString(s.OriginalText, newText, s.changes)
}

func (s EditString) ApplyAll(edits []TextEdit) EditString {
	if len(edits) == 0 {
		return s
	}
	ordered := getOrderedTextEdits(edits)
	if !s.CanApplyAll(ordered) {
		panic(InvalidOperationError{Message: "An edit occurs out of order, overlaps a prior edit or is specified out of bounds of the current text."})
	}
	newText := editStringNewText(s.CurrentText, ordered)
	newEdits := make([]editStringEdit, 0, len(ordered))
	for _, edit := range ordered {
		newEdits = append(newEdits, newEditStringEdit(edit.Start, edit.DeleteLength, len(edit.InsertText)))
	}
	return s.applyEdits(newText, newEdits)
}

func (s EditString) CanApplyAll(edits []TextEdit) bool {
	ordered := getOrderedTextEdits(edits)
	priorEnd := 0
	for _, edit := range ordered {
		if edit.Start < priorEnd || edit.Start > len(s.CurrentText) || edit.Start+edit.DeleteLength > len(s.CurrentText) {
			return false
		}
		priorEnd = edit.Start + edit.DeleteLength
	}
	return true
}

func getOrderedTextEdits(edits []TextEdit) []TextEdit {
	result := append([]TextEdit{}, edits...)
	if textEditsInOrder(result) {
		return result
	}
	stableSortTextEditsByStart(result)
	return result
}

func textEditsInOrder(edits []TextEdit) bool {
	lastStart := 0
	for _, edit := range edits {
		if edit.Start < lastStart {
			return false
		}
		lastStart = edit.Start
	}
	return true
}

func stableSortTextEditsByStart(edits []TextEdit) {
	for i := 1; i < len(edits); i++ {
		item := edits[i]
		j := i - 1
		for ; j >= 0 && edits[j].Start > item.Start; j-- {
			edits[j+1] = edits[j]
		}
		edits[j+1] = item
	}
}

func editStringNewText(text string, edits []TextEdit) string {
	var builder strings.Builder
	priorEnd := 0
	for _, edit := range edits {
		if edit.Start > priorEnd {
			builder.WriteString(text[priorEnd:edit.Start])
		}
		priorEnd = edit.Start + edit.DeleteLength
		if len(edit.InsertText) > 0 {
			builder.WriteString(edit.InsertText)
		}
	}
	if priorEnd < len(text) {
		builder.WriteString(text[priorEnd:])
	}
	return builder.String()
}

func (s EditString) applyEdits(newText string, newEdits []editStringEdit) EditString {
	combinedEdits := combineEditStringEdits(s.changes, newEdits)
	return newEditString(s.OriginalText, newText, combinedEdits)
}

type editStringEdit struct {
	Start        int
	DeleteLength int
	InsertLength int
}

func newEditStringEdit(start int, deleteLength int, insertLength int) editStringEdit {
	if start < 0 {
		panic(ArgumentOutOfRangeError{ParameterName: "start", Message: "negative start position"})
	}
	if deleteLength < 0 {
		panic(ArgumentOutOfRangeError{ParameterName: "deleteLength", Message: "negative deletion length"})
	}
	if insertLength < 0 {
		panic(ArgumentOutOfRangeError{ParameterName: "insertLength", Message: "negative insertion length"})
	}
	return editStringEdit{Start: start, DeleteLength: deleteLength, InsertLength: insertLength}
}

func (e editStringEdit) DeleteEnd() int {
	return e.Start + e.DeleteLength
}

func (e editStringEdit) InsertEnd() int {
	return e.Start + e.InsertLength
}

func combineEditStringEdits(oldEdits []editStringEdit, newEdits []editStringEdit) []editStringEdit {
	if len(newEdits) == 0 {
		return oldEdits
	}

	oldDelta := 0
	nextOldIndex := 1
	nextNewIndex := 1
	hasOldEdit := len(oldEdits) > 0
	hasNewEdit := len(newEdits) > 0
	var oldEdit editStringEdit
	var newEdit editStringEdit
	if hasOldEdit {
		oldEdit = oldEdits[0]
	}
	if hasNewEdit {
		newEdit = newEdits[0]
	}

	combinedEdits := make([]editStringEdit, 0, len(oldEdits)+len(newEdits))
	for hasOldEdit && hasNewEdit {
		switch {
		case oldEdit.DeleteLength == 0 && oldEdit.InsertLength == 0:
			hasOldEdit = nextOldIndex < len(oldEdits)
			if hasOldEdit {
				oldEdit = oldEdits[nextOldIndex]
				nextOldIndex++
				continue
			}
		case newEdit.DeleteLength == 0 && newEdit.InsertLength == 0:
			hasNewEdit = nextNewIndex < len(newEdits)
			if hasNewEdit {
				newEdit = newEdits[nextNewIndex]
				nextNewIndex++
				continue
			}
		case newEdit.DeleteEnd() <= oldEdit.Start+oldDelta:
			combinedEdits = append(combinedEdits, newEditStringEdit(newEdit.Start-oldDelta, newEdit.DeleteLength, newEdit.InsertLength))
			hasNewEdit = nextNewIndex < len(newEdits)
			if hasNewEdit {
				newEdit = newEdits[nextNewIndex]
				nextNewIndex++
				continue
			}
		case newEdit.Start >= oldEdit.InsertEnd()+oldDelta:
			combinedEdits = append(combinedEdits, oldEdit)
			oldDelta = oldDelta - oldEdit.DeleteLength + oldEdit.InsertLength
			hasOldEdit = nextOldIndex < len(oldEdits)
			if hasOldEdit {
				oldEdit = oldEdits[nextOldIndex]
				nextOldIndex++
				continue
			}
		case newEdit.Start < oldEdit.Start+oldDelta:
			partialDeleteLength := (oldEdit.Start + oldDelta) - newEdit.Start
			combinedEdits = append(combinedEdits, newEditStringEdit(newEdit.Start-oldDelta, partialDeleteLength, 0))
			newEdit = newEditStringEdit(oldEdit.Start+oldDelta, newEdit.DeleteLength-partialDeleteLength, newEdit.InsertLength)
			continue
		case newEdit.Start > oldEdit.Start+oldDelta:
			partialInsertLength := newEdit.Start - (oldEdit.Start + oldDelta)
			partialDeleteLength := minInt(oldEdit.DeleteLength, partialInsertLength)
			combinedEdits = append(combinedEdits, newEditStringEdit(oldEdit.Start, partialDeleteLength, partialInsertLength))
			oldDelta = oldDelta - partialDeleteLength + partialInsertLength
			oldEdit = newEditStringEdit(newEdit.Start-oldDelta, oldEdit.DeleteLength-partialDeleteLength, oldEdit.InsertLength-partialInsertLength)
			continue
		case newEdit.DeleteLength <= oldEdit.InsertLength:
			oldEdit = newEditStringEdit(oldEdit.Start, oldEdit.DeleteLength, oldEdit.InsertLength-newEdit.DeleteLength)
			oldDelta += newEdit.DeleteLength
			combinedEdits = append(combinedEdits, newEditStringEdit(newEdit.DeleteEnd()-oldDelta, 0, newEdit.InsertLength))
			hasNewEdit = nextNewIndex < len(newEdits)
			if hasNewEdit {
				newEdit = newEdits[nextNewIndex]
				nextNewIndex++
				continue
			}
		default:
			oldDelta = oldDelta - oldEdit.DeleteLength + oldEdit.InsertLength
			newDeletion := newEdit.DeleteLength + oldEdit.DeleteLength - oldEdit.InsertLength
			newEdit = newEditStringEdit(oldEdit.Start+oldDelta, newDeletion, newEdit.InsertLength)
			hasOldEdit = nextOldIndex < len(oldEdits)
			if hasOldEdit {
				oldEdit = oldEdits[nextOldIndex]
				nextOldIndex++
				continue
			}
		}
		break
	}

	if hasNewEdit {
		combinedEdits = append(combinedEdits, newEditStringEdit(newEdit.Start-oldDelta, newEdit.DeleteLength, newEdit.InsertLength))
	}
	for ; nextNewIndex < len(newEdits); nextNewIndex++ {
		newEdit = newEdits[nextNewIndex]
		combinedEdits = append(combinedEdits, newEditStringEdit(newEdit.Start-oldDelta, newEdit.DeleteLength, newEdit.InsertLength))
	}

	if hasOldEdit {
		combinedEdits = append(combinedEdits, oldEdit)
	}
	for ; nextOldIndex < len(oldEdits); nextOldIndex++ {
		combinedEdits = append(combinedEdits, oldEdits[nextOldIndex])
	}

	combinedEdits = combineAdjacentEditStringEdits(combinedEdits)
	return ListToReadOnly(combinedEdits)
}

func combineAdjacentEditStringEdits(edits []editStringEdit) []editStringEdit {
	for i := len(edits) - 1; i >= 1; i-- {
		edit := edits[i]
		prevEdit := edits[i-1]
		if prevEdit.DeleteEnd() == edit.Start {
			edits[i-1] = newEditStringEdit(prevEdit.Start, prevEdit.DeleteLength+edit.DeleteLength, prevEdit.InsertLength+edit.InsertLength)
			copy(edits[i:], edits[i+1:])
			edits = edits[:len(edits)-1]
		}
	}
	return edits
}

func (s EditString) GetCurrentPosition(originalPosition int, bias ...PositionBias) int {
	positionBias := PositionBiasRight
	if len(bias) > 0 {
		positionBias = bias[0]
	}
	currentPosition := originalPosition
	delta := 0
	for _, edit := range s.changes {
		if currentPosition < edit.Start+delta {
			break
		}
		currentPosition = translateEditStringPosition(currentPosition, edit.Start+delta, edit.DeleteLength, edit.InsertLength, positionBias)
		delta = delta - edit.DeleteLength + edit.InsertLength
	}
	return currentPosition
}

func (s EditString) GetCurrentRange(originalStart int, originalLength int) TextRange {
	originalEnd := originalStart + originalLength
	currentStart := s.GetCurrentPosition(originalStart, PositionBiasRight)
	currentEnd := s.GetCurrentPosition(originalEnd, PositionBiasLeft)
	return NewTextRange(currentStart, currentEnd-currentStart)
}

func (s EditString) GetOriginalPosition(currentPosition int, bias ...PositionBias) int {
	positionBias := PositionBiasRight
	if len(bias) > 0 {
		positionBias = bias[0]
	}
	originalPosition := currentPosition
	delta := 0
	for _, edit := range s.changes {
		if originalPosition < edit.Start+delta {
			break
		}
		originalPosition = translateEditStringPosition(originalPosition, edit.Start+delta, edit.InsertLength, edit.DeleteLength, positionBias)
	}
	return originalPosition
}

func translateEditStringPosition(position int, start int, deleteLength int, insertLength int, bias PositionBias) int {
	if deleteLength > 0 && position > start {
		if position > start+deleteLength {
			position -= deleteLength
		} else {
			position = start
		}
	}
	if insertLength > 0 {
		if position > start || (bias == PositionBiasRight && position == start) {
			position += insertLength
		}
	}
	return position
}

func (s EditString) GetChangeRange() TextRange {
	if len(s.changes) == 0 {
		return NewTextRange(0, 0)
	}
	firstStart := s.GetCurrentPosition(s.changes[0].Start, PositionBiasLeft)
	lastEnd := s.GetCurrentPosition(s.changes[len(s.changes)-1].Start, PositionBiasRight)
	return TextRangeFromBounds(firstStart, lastEnd)
}

func (s EditString) RemoveBlankLines() EditString {
	pos := 0
	text := s.CurrentText
	var edits []TextEdit
	for pos < len(text) {
		length := LineLength(text, pos, true)
		if IsBlankLine(text, pos) {
			edits = append(edits, NewTextDeletion(pos, length))
		}
		pos += length
	}
	return s.ApplyAll(edits)
}

func (s EditString) ReplaceLineBreaks(newValue string) EditString {
	text := s.CurrentText
	var edits []TextEdit
	for pos := 0; pos < len(text); {
		ch, size := utf8.DecodeRuneInString(text[pos:])
		if IsLineBreakStart(ch) {
			length := LineBreakLength(text, pos)
			if length != len(newValue) || text[pos:pos+length] != newValue {
				edits = append(edits, NewTextReplacement(pos, length, newValue))
			}
			pos += length
		} else {
			pos += size
		}
	}
	return s.ApplyAll(edits)
}

func whitespaceCountBefore(text string, position int) int {
	pos := position
	for pos > 0 {
		ch, size := utf8.DecodeLastRuneInString(text[:pos])
		if !IsWhitespace(ch) {
			break
		}
		pos -= size
	}
	return position - pos
}
