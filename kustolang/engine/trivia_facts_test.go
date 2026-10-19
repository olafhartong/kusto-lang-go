package kustolang

import "testing"

func TestTryGetCommentSpanFindsCommentInTrivia(t *testing.T) {
	trivia := "  // first comment\r\n\t// second\n  "

	start, length, ok := TryGetCommentSpan(trivia, 5)
	if !ok || start != 2 || length != len("// first comment\r\n") {
		t.Fatalf("first comment span = (%d, %d, %t)", start, length, ok)
	}

	start, length, ok = TryGetCommentSpan(trivia, len("  // first comment\r\n\t// second"))
	if !ok || start != len("  // first comment\r\n\t") || length != len("// second\n") {
		t.Fatalf("second comment span = (%d, %d, %t)", start, length, ok)
	}
}

func TestTryGetCommentSpanAcceptsAdjacentEndPosition(t *testing.T) {
	trivia := "\t// comment\n "

	start, length, ok := TryGetCommentSpan(trivia, len("\t// comment\n"))
	if !ok || start != 1 || length != len("// comment\n") {
		t.Fatalf("adjacent comment span = (%d, %d, %t)", start, length, ok)
	}
}

func TestTryGetCommentSpanRejectsWhitespaceOnlyOrBeforeComment(t *testing.T) {
	if _, _, ok := TryGetCommentSpan(" \t\r\n", 2); ok {
		t.Fatal("whitespace-only trivia should not have a comment span")
	}
	if _, _, ok := TryGetCommentSpan("  // comment", 1); ok {
		t.Fatal("position before comment should not match the comment span")
	}
}

func TestTryGetCommentSpanTreatsUnicodeWhitespaceAsWhitespace(t *testing.T) {
	trivia := "\u2003// comment\u2028 "
	for position := 0; position < len("\u2003"); position++ {
		if _, _, ok := TryGetCommentSpan(trivia, position); ok {
			t.Fatalf("Unicode whitespace byte position %d should not match a comment", position)
		}
	}
	start, length, ok := TryGetCommentSpan(trivia, len("\u2003//"))
	if !ok || start != len("\u2003") || length != len("// comment\u2028") {
		t.Fatalf("Unicode trivia comment span = (%d, %d, %t)", start, length, ok)
	}
}
