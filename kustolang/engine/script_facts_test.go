package kustolang

import (
	"reflect"
	"testing"
)

func TestTokenParserScannersMirrorUpstreamCommentAndStringEdges(t *testing.T) {
	if got := ScanComment("// c\r\nnext", 0); got != len("// c\r\n") {
		t.Fatalf("ScanComment CRLF = %d", got)
	}
	if got := ScanComment("  // c", 0); got != 0 {
		t.Fatalf("ScanComment non-comment start = %d", got)
	}

	if got := ScanStringLiteral(`@"a ""quoted"" value"`, 0); got != len(`@"a ""quoted"" value"`) {
		t.Fatalf("ScanStringLiteral verbatim = %d", got)
	}
	if got := ScanStringLiteral("```one\n\ntwo```", 0); got != len("```one\n\ntwo```") {
		t.Fatalf("ScanStringLiteral multiline = %d", got)
	}
	if got := ScanStringLiteral("'unterminated", 0); got != len("'unterminated") {
		t.Fatalf("ScanStringLiteral missing quote = %d", got)
	}
	if got := ScanStringLiteral("'unterminated", 0, true); got != -1 {
		t.Fatalf("ScanStringLiteral fail missing = %d", got)
	}
}

func TestScriptFactsRemoveInvalidBlockStartsInsideStrings(t *testing.T) {
	text := "let s = ```one\n\ninside```;\n\nT\n| count"
	starts := KustoBlockStarts(text, LineStarts(text))
	want := []int{0, len("let s = ```one\n\ninside```;\n\n")}
	if !reflect.DeepEqual(starts, want) {
		t.Fatalf("KustoBlockStarts = %#v, want %#v", starts, want)
	}

	invalid := InvalidKustoBlockStartRanges(text)
	if len(invalid) != 1 || invalid[0].Start != len("let s = ") || invalid[0].Length != len("```one\n\ninside```") {
		t.Fatalf("invalid ranges = %#v", invalid)
	}
}

func TestCommandFactsExtractCommandTextsWithLeadingComments(t *testing.T) {
	text := "// first command\n.show version\n\n// second command\n.show tables\n"

	starts := CommandBlockStarts(text)
	wantStarts := []int{0, len("// first command\n.show version\n\n")}
	if !reflect.DeepEqual(starts, wantStarts) {
		t.Fatalf("CommandBlockStarts = %#v, want %#v", starts, wantStarts)
	}

	texts := CommandTexts(text)
	wantTexts := []string{".show version", ".show tables"}
	if !reflect.DeepEqual(texts, wantTexts) {
		t.Fatalf("CommandTexts = %#v, want %#v", texts, wantTexts)
	}
}

func TestTextRangeOverlapsUsesInclusiveBounds(t *testing.T) {
	left := NewTextRange(0, 5)
	right := NewTextRange(5, 3)
	if !left.Overlaps(right) {
		t.Fatal("adjacent ranges should overlap like upstream TextRange")
	}
	if TextRangesOverlap(0, 2, 3, 1) {
		t.Fatal("separated ranges should not overlap")
	}
	if got := TextRangeFromBounds(2, 7); got.Start != 2 || got.Length != 5 || got.End() != 7 {
		t.Fatalf("TextRangeFromBounds = %#v end %d", got, got.End())
	}
}
