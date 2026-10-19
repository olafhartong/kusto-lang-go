package kustolang_test

// External-package parity tests for the pinned TextFacts and TriviaFacts APIs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicTextFactsFacadeUsesUTF16Coordinates(t *testing.T) {
	text := "😀 \tA\r\n\u2028β\n"
	wantStarts := []int{0, 7, 8, 10}
	if got := kustolang.TextFacts.GetLineStarts(text); !reflect.DeepEqual(got, wantStarts) {
		t.Fatalf("GetLineStarts() = %#v, want %#v", got, wantStarts)
	}
	if got := kustolang.TextFacts.GetLineLength(text, 0); got != 5 {
		t.Fatalf("first line length = %d, want 5 UTF-16 units", got)
	}
	if got := kustolang.TextFacts.GetLineLength(text, 0, true); got != 7 {
		t.Fatalf("first line length with CRLF = %d, want 7", got)
	}
	if got := kustolang.TextFacts.GetLineText(text, 1); got != "😀 \tA" {
		t.Fatalf("first line text = %q", got)
	}
	if got, want := kustolang.TextFacts.GetLineTexts(text), []string{"😀 \tA", "", "β", ""}; !reflect.DeepEqual(got, want) {
		t.Fatalf("GetLineTexts() = %#v, want %#v", got, want)
	}

	// System.String permits positions between the surrogate code units.
	line, offset, ok := kustolang.TextFacts.TryGetLineAndOffset(text, 1)
	if !ok || line != 1 || offset != 2 {
		t.Fatalf("surrogate-interior line/offset = %d, %d, %t", line, offset, ok)
	}
	if position, ok := kustolang.TextFacts.TryGetPosition(text, 1, 2); !ok || position != 1 {
		t.Fatalf("surrogate-interior position = %d, %t", position, ok)
	}
	if line, offset, ok := kustolang.TextFacts.TryGetLineAndOffset(text, 8); !ok || line != 3 || offset != 1 {
		t.Fatalf("third-line position = %d, %d, %t", line, offset, ok)
	}
	if start, ok := kustolang.TextFacts.TryGetLineStart(text, 3); !ok || start != 8 {
		t.Fatalf("third-line start = %d, %t", start, ok)
	}
	if start := kustolang.TextFacts.GetLineStart(text, 8); start != 8 {
		t.Fatalf("GetLineStart(beta) = %d, want 8", start)
	}
}

func TestPublicTextFactsFacadeMatchesCharacterWhitespaceAndRangeBehavior(t *testing.T) {
	if !kustolang.TextFacts.IsWhitespace('\u200b') || kustolang.TextFacts.IsWhitespace('x') {
		t.Fatal("whitespace character classification differs from TextFacts")
	}
	if !kustolang.TextFacts.IsLetter('Z') || kustolang.TextFacts.IsLetter('β') ||
		!kustolang.TextFacts.IsDigit('7') || !kustolang.TextFacts.IsHexDigit('f') ||
		kustolang.TextFacts.IsHexDigit('g') {
		t.Fatal("ASCII character classification differs from TextFacts")
	}
	if !kustolang.TextFacts.IsLineBreakStart('\u2029') || kustolang.TextFacts.GetLineBreakLength("\r\nx", 0) != 2 {
		t.Fatal("line-break classification differs from TextFacts")
	}

	text := "😀  x\n\tvalue"
	if got := kustolang.TextFacts.GetWhitespaceCount(text, 2); got != 2 {
		t.Fatalf("whitespace count = %d, want 2", got)
	}
	if got := kustolang.TextFacts.GetWhitespaceCountBefore(text, 4); got != 3 {
		t.Fatalf("whitespace-before count = %d, want pinned source value 3", got)
	}
	if start, length := kustolang.TextFacts.TrimRange(text, 2, 3); start != 4 || length != 1 {
		t.Fatalf("TrimRange() = %d, %d, want 4, 1", start, length)
	}
	if got := kustolang.TextFacts.TrimEnd("😀  "); got != 2 {
		t.Fatalf("TrimEnd() = %d, want two UTF-16 units", got)
	}
	if got := kustolang.TextFacts.GetIndentationLength(text, 8); got != 1 || kustolang.TextFacts.GetIndentationText(text, 8) != "\t" {
		t.Fatalf("indentation = %d, %q", got, kustolang.TextFacts.GetIndentationText(text, 8))
	}
	if !kustolang.TextFacts.IsBlankLine("😀\n\u2028x", 3) {
		t.Fatal("Unicode-separator-only line was not blank")
	}
}

func TestPublicTextFactsFacadeListOverloadAdaptations(t *testing.T) {
	starts := []int{99}
	kustolang.TextFacts.AppendLineStarts("😀\r\nx", &starts)
	if want := []int{99, 0, 4}; !reflect.DeepEqual(starts, want) {
		t.Fatalf("AppendLineStarts() = %#v, want %#v", starts, want)
	}
	line, offset, ok := kustolang.TextFacts.TryGetLineAndOffsetFromStarts(starts[1:], 4)
	if !ok || line != 2 || offset != 1 {
		t.Fatalf("list line/offset = %d, %d, %t", line, offset, ok)
	}
	if position, ok := kustolang.TextFacts.TryGetPositionFromStarts(starts[1:], 2, 2); !ok || position != 5 {
		t.Fatalf("list position = %d, %t", position, ok)
	}
	if _, _, ok := kustolang.TextFacts.TryGetLineAndOffsetFromStarts(starts[1:], -1); ok {
		t.Fatal("negative list position was accepted")
	}
	if _, ok := kustolang.TextFacts.TryGetPositionFromStarts(starts[1:], 0, 1); ok {
		t.Fatal("zero line was accepted")
	}
}

func TestPublicTriviaFactsFacadeReturnsUTF16CommentSpan(t *testing.T) {
	trivia := "  😀 note\r\n\t"
	start, length, ok := kustolang.TriviaFacts.TryGetCommentSpan(trivia, 5)
	if !ok || start != 2 || length != 9 {
		t.Fatalf("comment span = %d, %d, %t, want 2, 9, true", start, length, ok)
	}
	if adjacentStart, adjacentLength, adjacent := kustolang.TriviaFacts.TryGetCommentSpan(trivia, 11); !adjacent || adjacentStart != start || adjacentLength != length {
		t.Fatalf("adjacent comment span = %d, %d, %t", adjacentStart, adjacentLength, adjacent)
	}
	if _, _, ok := kustolang.TriviaFacts.TryGetCommentSpan(" \t", 1); ok {
		t.Fatal("whitespace trivia was classified as a comment")
	}
	if _, _, ok := kustolang.TriviaFacts.TryGetCommentSpan(trivia, -1); ok {
		t.Fatal("negative trivia position was accepted")
	}
}
