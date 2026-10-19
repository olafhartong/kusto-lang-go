package kustolang

import (
	"reflect"
	"testing"
)

func TestTextFactsCharacterAndWhitespaceParity(t *testing.T) {
	if !IsLetterOrDigit('Z') || !IsLetterOrDigit('7') {
		t.Fatal("ASCII letters and digits should be accepted")
	}
	if IsLetterOrDigit('Δ') || IsLetterOrDigit('７') {
		t.Fatal("non-ASCII letters and digits should not match upstream TextFacts")
	}
	if IsWhitespace('\u2028') || IsWhitespace('\u2029') {
		t.Fatal("line and paragraph separators are line breaks, not Kusto whitespace")
	}
	if !IsLineBreakStart('\u2028') || !IsLineBreakStart('\u2029') {
		t.Fatal("Unicode separators should remain line breaks")
	}
	if got := GetWhitespace(" \tvalue", 0); got != " \t" {
		t.Fatalf("Whitespace = %q, want space and tab", got)
	}
	if got := GetWhitespace("value", 0); got != "" {
		t.Fatalf("Whitespace without prefix = %q, want empty", got)
	}
	if got := WhitespaceCountBefore("x  ", 3); got != 3 {
		t.Fatalf("WhitespaceCountBefore = %d, want source distance 3", got)
	}
	if got := WhitespaceCountBefore("x", 1); got != 1 {
		t.Fatalf("WhitespaceCountBefore non-whitespace = %d, want source minimum 1", got)
	}
}

func TestTextFactsRangeParity(t *testing.T) {
	const spaced = "  value \t"
	start, length := TrimRangeStart(spaced, 0, len(spaced))
	if start != 2 || length != 7 {
		t.Fatalf("TrimRangeStart = (%d, %d), want (2, 7)", start, length)
	}
	if got := TrimRangeEnd(spaced, start, length); got != 5 {
		t.Fatalf("TrimRangeEnd = %d, want 5", got)
	}
	if start, length = TrimRange(spaced, 0, len(spaced)); start != 2 || length != 5 {
		t.Fatalf("TrimRange = (%d, %d), want (2, 5)", start, length)
	}

	const lines = "aa\n  bb\ncc"
	if start, length = ExpandRangeToStartOfFirstLine(lines, 5, 1); start != 3 || length != 1 {
		t.Fatalf("ExpandRangeToStartOfFirstLine = (%d, %d), want source shape (3, 1)", start, length)
	}
	if length = ExpandRangeToEndOfLastLine(lines, 5, 1); length != 2 {
		t.Fatalf("ExpandRangeToEndOfLastLine = %d, want 2", length)
	}
	if start, length = ExpandRangeToStartAndEndOfLines(lines, 5, 1); start != 3 || length != 4 {
		t.Fatalf("ExpandRangeToStartAndEndOfLines = (%d, %d), want (3, 4)", start, length)
	}
}

func TestTextFactsIndentationAndLineTextParity(t *testing.T) {
	const text = "alpha\r\n  beta\n\nomega"
	if got := IndentationLength(text, 11); got != 2 {
		t.Fatalf("IndentationLength inside line = %d, want 2", got)
	}
	if got := IndentationText(text, 11); got != "  " {
		t.Fatalf("IndentationText inside line = %q, want two spaces", got)
	}
	if got := LineText(text, 2); got != "  beta" {
		t.Fatalf("LineText(2) = %q, want %q", got, "  beta")
	}
	if got := LineText(text, 5); got != "" {
		t.Fatalf("LineText beyond end = %q, want empty", got)
	}
	want := []string{"alpha", "  beta", "", "omega"}
	if got := LineTexts(text); !reflect.DeepEqual(got, want) {
		t.Fatalf("LineTexts = %#v, want %#v", got, want)
	}
	if got := LineTexts("a\n"); !reflect.DeepEqual(got, []string{"a", ""}) {
		t.Fatalf("trailing LineTexts = %#v, want trailing empty line", got)
	}
}
