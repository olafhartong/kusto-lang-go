package kustolang_test

// External-package parity tests for the pinned CommentFacts API.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicCommentFactsMatchesFullLineCommentBehavior(t *testing.T) {
	text := "😀 first\r\n  // first \t\r\n// second\nx // inline\n\t//\n"
	lineStarts := parserFactsTestLineStarts(text)

	if kustolang.CommentFacts.IsCommentLine(text) {
		t.Fatal("first non-comment line was classified as a comment")
	}
	if !kustolang.CommentFacts.IsCommentLine(text, lineStarts[1]) {
		t.Fatal("indented CRLF comment line was not recognized")
	}
	if got := kustolang.CommentFacts.GetCommentLineText(text, lineStarts[1]); got != "first" {
		t.Fatalf("GetCommentLineText() = %q, want first", got)
	}
	if got := kustolang.CommentFacts.GetCommentLineText("\t// default"); got != "default" {
		t.Fatalf("default-position comment text = %q", got)
	}

	want := []string{"first", "second", ""}
	if got := kustolang.CommentFacts.GetCommentTexts(text); !reflect.DeepEqual(got, want) {
		t.Fatalf("GetCommentTexts() = %#v, want %#v", got, want)
	}
	if got := kustolang.CommentFacts.GetCommentTexts("print 1"); got == nil || len(got) != 0 {
		t.Fatalf("comment-free result = %#v, want non-nil empty slice", got)
	}
}

func TestPublicCommentFactsTrimsOnlyOuterCommentAndWhitespaceLines(t *testing.T) {
	text := "// heading\r\n \t\r\n😀 | count\n\n// trailing\n"
	if got, want := kustolang.CommentFacts.TrimCommentAndWhitespaceLines(text), "😀 | count"; got != want {
		t.Fatalf("TrimCommentAndWhitespaceLines() = %q, want %q", got, want)
	}

	for _, text := range []string{"", " \t\r\n", "// only\n  // comments\r\n"} {
		if got := kustolang.CommentFacts.TrimCommentAndWhitespaceLines(text); got != "" {
			t.Fatalf("TrimCommentAndWhitespaceLines(%q) = %q, want empty", text, got)
		}
	}
}

func TestPublicCommentFactsRejectsMalformedUTF16Positions(t *testing.T) {
	text := "😀\n// valid"
	tests := []int{-1, 1, 100}
	for _, position := range tests {
		if kustolang.CommentFacts.IsCommentLine(text, position) {
			t.Fatalf("IsCommentLine accepted invalid UTF-16 position %d", position)
		}
		if got := kustolang.CommentFacts.GetCommentLineText(text, position); got != "" {
			t.Fatalf("GetCommentLineText at %d = %q, want empty", position, got)
		}
	}

	lineStarts := parserFactsTestLineStarts(text)
	if !kustolang.CommentFacts.IsCommentLine(text, lineStarts[1]) {
		t.Fatalf("valid UTF-16 line start %d was rejected", lineStarts[1])
	}
}
