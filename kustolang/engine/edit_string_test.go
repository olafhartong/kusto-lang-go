package kustolang

import "testing"

func TestEditStringBasicStringOperationsAndChanges(t *testing.T) {
	value := NewEditString("abc")
	if value.OriginalText != "abc" || value.CurrentText != "abc" || value.Length() != 3 || value.At(1) != 'b' || value.IndexOf("bc") != 1 || value.String() != "abc" {
		t.Fatalf("initial edit string = %#v len=%d at1=%q index=%d string=%q", value, value.Length(), value.At(1), value.IndexOf("bc"), value.String())
	}

	edited := value.Insert(1, "X").Append("Y").Prepend("Z").Remove(2, 1)
	if got, want := edited.CurrentText, "ZabcY"; got != want {
		t.Fatalf("edited text = %q, want %q", got, want)
	}
	if edited.OriginalText != "abc" {
		t.Fatalf("original text = %q, want abc", edited.OriginalText)
	}

	changes := edited.GetChanges()
	if len(changes) == 0 {
		t.Fatalf("expected tracked changes for %#v", edited)
	}
	applied := NewEditString(value.OriginalText).ApplyAll(changes)
	if applied.CurrentText != edited.CurrentText {
		t.Fatalf("changes applied to original = %q, want %q; changes=%#v", applied.CurrentText, edited.CurrentText, changes)
	}
}

func TestEditStringApplyAllOrdersAndValidatesCollectiveEdits(t *testing.T) {
	value := NewEditString("abcdef")
	edited := value.ApplyAll([]TextEdit{
		NewTextReplacement(4, 1, "E"),
		NewTextInsertion(1, "-"),
		NewTextReplacement(1, 2, "BC"),
	})
	if got, want := edited.CurrentText, "a-BCdEf"; got != want {
		t.Fatalf("ApplyAll sorted edits text = %q, want %q", got, want)
	}

	if !value.CanApplyAll([]TextEdit{NewTextReplacement(1, 2, "x"), NewTextInsertion(3, "y")}) {
		t.Fatalf("expected non-overlapping edits to be applicable")
	}
	if value.CanApplyAll([]TextEdit{NewTextReplacement(1, 3, "x"), NewTextReplacement(2, 1, "y")}) {
		t.Fatalf("expected overlapping edits to be rejected")
	}
	if value.CanApplyAll([]TextEdit{NewTextReplacement(7, 0, "x")}) {
		t.Fatalf("expected out-of-bounds edit to be rejected")
	}

	err := recoverPanic(t, func() {
		value.ApplyAll([]TextEdit{NewTextReplacement(1, 3, "x"), NewTextReplacement(2, 1, "y")})
	})
	if _, ok := err.(InvalidOperationError); !ok {
		t.Fatalf("overlapping ApplyAll panic = %#v, want InvalidOperationError", err)
	}
}

func TestEditStringConvertToSequentialChanges(t *testing.T) {
	collective := []TextEdit{
		NewTextReplacement(1, 2, "BC"),
		NewTextInsertion(4, "X"),
		NewTextDeletion(5, 1),
	}
	sequential := ConvertToSequentialChanges(collective)
	want := []TextEdit{
		NewTextReplacement(1, 2, "BC"),
		NewTextInsertion(4, "X"),
		NewTextDeletion(6, 1),
	}
	if len(sequential) != len(want) {
		t.Fatalf("sequential length = %d, want %d: %#v", len(sequential), len(want), sequential)
	}
	for i := range want {
		if sequential[i] != want[i] {
			t.Fatalf("sequential[%d] = %#v, want %#v", i, sequential[i], want[i])
		}
	}
}

func TestEditStringSubstringTrimAndReplace(t *testing.T) {
	value := NewEditString("  one two one  ")
	trimmed := value.Trim()
	if got, want := trimmed.CurrentText, "one two one"; got != want {
		t.Fatalf("Trim text = %q, want %q", got, want)
	}
	if got, want := value.TrimLeft().CurrentText, "one two one  "; got != want {
		t.Fatalf("TrimLeft text = %q, want %q", got, want)
	}
	if got, want := value.TrimRight().CurrentText, "  one two one"; got != want {
		t.Fatalf("TrimRight text = %q, want %q", got, want)
	}

	replaced := trimmed.Replace("one", "1")
	if got, want := replaced.CurrentText, "1 two 1"; got != want {
		t.Fatalf("Replace text = %q, want %q", got, want)
	}
	if same := replaced.Replace("missing", "x"); same.CurrentText != replaced.CurrentText || len(same.GetChanges()) != len(replaced.GetChanges()) {
		t.Fatalf("Replace with no matches should preserve observable state: same=%#v replaced=%#v", same, replaced)
	}
	err := recoverPanic(t, func() {
		replaced.Replace("", "x")
	})
	if _, ok := err.(ArgumentError); !ok {
		t.Fatalf("empty old text panic = %#v, want ArgumentError", err)
	}

	sub := NewEditString("abcdef").Substring(2, 3)
	if got, want := sub.CurrentText, "cde"; got != want {
		t.Fatalf("Substring text = %q, want %q", got, want)
	}
	applied := NewEditString("abcdef").ApplyAll(sub.GetChanges())
	if applied.CurrentText != "cde" {
		t.Fatalf("substring changes applied = %q, want cde; changes=%#v", applied.CurrentText, sub.GetChanges())
	}
}

func TestEditStringUnicodeWhitespaceAndLineBreaks(t *testing.T) {
	value := NewEditString("\u2003value\u00a0")
	if got := value.TrimLeft().CurrentText; got != "value\u00a0" {
		t.Fatalf("Unicode TrimLeft = %q", got)
	}
	if got := value.TrimRight().CurrentText; got != "\u2003value" {
		t.Fatalf("Unicode TrimRight = %q", got)
	}
	if got := value.Trim().CurrentText; got != "value" {
		t.Fatalf("Unicode Trim = %q", got)
	}

	lineBreaks := NewEditString("a\u2028b\u2029c\r\nd")
	if got := lineBreaks.ReplaceLineBreaks("|").CurrentText; got != "a|b|c|d" {
		t.Fatalf("Unicode ReplaceLineBreaks = %q", got)
	}
}

func TestEditStringPositionTranslationAndChangeRange(t *testing.T) {
	edited := NewEditString("abcdef").ReplaceAt(2, 2, "XYZ")
	if got, want := edited.CurrentText, "abXYZef"; got != want {
		t.Fatalf("edited text = %q, want %q", got, want)
	}
	if got := edited.GetCurrentPosition(2, PositionBiasLeft); got != 2 {
		t.Fatalf("current position left at edit start = %d, want 2", got)
	}
	if got := edited.GetCurrentPosition(2, PositionBiasRight); got != 5 {
		t.Fatalf("current position right at edit start = %d, want 5", got)
	}
	if got := edited.GetCurrentPosition(4); got != 5 {
		t.Fatalf("current position after replaced range = %d, want 5", got)
	}
	if got := edited.GetOriginalPosition(5, PositionBiasLeft); got != 2 {
		t.Fatalf("original position left at inserted end = %d, want 2", got)
	}
	if got := edited.GetOriginalPosition(5, PositionBiasRight); got != 4 {
		t.Fatalf("original position right at inserted end = %d, want 4", got)
	}

	currentRange := edited.GetCurrentRange(1, 4)
	if currentRange.Start != 1 || currentRange.Length != 5 || currentRange.End() != 6 {
		t.Fatalf("current range = %#v end=%d, want start=1 length=5", currentRange, currentRange.End())
	}
	changeRange := edited.GetChangeRange()
	if changeRange.Start != 2 || changeRange.Length != 3 || changeRange.End() != 5 {
		t.Fatalf("change range = %#v end=%d, want start=2 length=3", changeRange, changeRange.End())
	}
}

func TestEditStringBlankLinesAndLineBreakReplacement(t *testing.T) {
	text := "a\n\n  \r\nb\r\n"
	withoutBlank := NewEditString(text).RemoveBlankLines()
	if got, want := withoutBlank.CurrentText, "a\nb\r\n"; got != want {
		t.Fatalf("RemoveBlankLines text = %q, want %q", got, want)
	}
	applied := NewEditString(text).ApplyAll(withoutBlank.GetChanges())
	if applied.CurrentText != withoutBlank.CurrentText {
		t.Fatalf("blank-line changes applied = %q, want %q; changes=%#v", applied.CurrentText, withoutBlank.CurrentText, withoutBlank.GetChanges())
	}

	replaced := NewEditString("a\r\nb\nc").ReplaceLineBreaks("|")
	if got, want := replaced.CurrentText, "a|b|c"; got != want {
		t.Fatalf("ReplaceLineBreaks text = %q, want %q", got, want)
	}
	if same := NewEditString("a\nb").ReplaceLineBreaks("\n"); same.CurrentText != "a\nb" || len(same.GetChanges()) != 0 {
		t.Fatalf("same line-break replacement = text:%q changes:%#v", same.CurrentText, same.GetChanges())
	}
}
