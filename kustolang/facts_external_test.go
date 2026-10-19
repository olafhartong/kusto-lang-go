package kustolang_test

import (
	"slices"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicTextFactsAppendLineStartsOverload(t *testing.T) {
	starts := []int{99}
	kustolang.AppendLineStarts("a\r\nb\n", &starts)
	want := []int{99, 0, 3, 5}
	if !slices.Equal(starts, want) {
		t.Fatalf("appended line starts = %#v, want %#v", starts, want)
	}
}

func TestPublicKustoFactsNamesAndLiterals(t *testing.T) {
	if !kustolang.CanBeIdentifier("Account") || kustolang.CanBeIdentifier("where") || kustolang.CanBeIdentifier("true") {
		t.Fatal("query identifier classification differs from KustoFacts")
	}
	if kustolang.CanBeIdentifier("Account", kustolang.KustoDialectDataManagerCommand) {
		t.Fatal("data-manager command identifier did not require brackets")
	}
	if got := kustolang.BracketNameIfNecessary("where"); got != "['where']" {
		t.Fatalf("bracketed keyword = %q", got)
	}
	if got := kustolang.GetStringLiteral("can't\n"); got != `"can't\n"` {
		t.Fatalf("escaped string literal = %q", got)
	}
	if got := kustolang.GetStringLiteralValue(`@'can''t'`); got != "can't" {
		t.Fatalf("verbatim string value = %q", got)
	}
	if got := kustolang.GetMultiLineStringLiteral("a\nb"); got != "```a\nb```" {
		t.Fatalf("multiline string literal = %q", got)
	}
	if !kustolang.KustoFactsMatches("Storm*", "StormEvents") || kustolang.KustoFactsMatches("Storm*", "stormevents") {
		t.Fatal("wildcard matching differs from KustoFacts")
	}
	if !kustolang.KustoFactsMatches("Storm*", "stormevents", true) {
		t.Fatal("case-insensitive wildcard matching failed")
	}
}

func TestPublicSyntaxFactsInventories(t *testing.T) {
	kind, ok := kustolang.SyntaxKindFromTextRange("xxwhereyy", 2, 5)
	if !ok || kind != kustolang.SyntaxKindWhereKeyword {
		t.Fatalf("kind from range = %v, %t", kind, ok)
	}
	if _, ok := kustolang.SyntaxKindFromTextRange("where", -1, 5); ok {
		t.Fatal("invalid text range unexpectedly resolved")
	}
	if !kustolang.IsPunctuationKind(kustolang.SyntaxKindCommaToken) || !kustolang.IsTypeKeyword(kustolang.SyntaxKindStringKeyword) {
		t.Fatal("syntax category classification differs from SyntaxFacts")
	}
	if !slices.Contains(kustolang.SyntaxKeywordTexts(), "where") || !slices.Contains(kustolang.SyntaxPunctuationTexts(), ",") {
		t.Fatal("syntax text inventories are incomplete")
	}
	if len(kustolang.SyntaxKindsWithFixedText()) == 0 || len(kustolang.KustoFacts.JoinKinds) == 0 || len(kustolang.KnownQueryOperatorParameterNames()) == 0 {
		t.Fatal("public Kusto/SyntaxFacts inventories are empty")
	}
}

func TestPublicTokenScannersAcceptAndRejectSyntax(t *testing.T) {
	tests := []struct {
		name string
		got  int
		want int
	}{
		{"boolean", kustolang.ScanBooleanLiteral("true"), 4},
		{"boolean offset", kustolang.ScanBooleanLiteral("xFALSE!", 1), 5},
		{"boolean identifier boundary", kustolang.ScanBooleanLiteral("trueValue"), -1},
		{"identifier", kustolang.ScanIdentifier("Account_1"), 9},
		{"bad identifier", kustolang.ScanIdentifier("-Account"), -1},
		{"long", kustolang.ScanLongLiteral("0x2a"), 4},
		{"bad long boundary", kustolang.ScanLongLiteral("12abc"), -1},
		{"real", kustolang.ScanRealLiteral("1.25e2"), 6},
		{"bad real", kustolang.ScanRealLiteral("1e+"), -1},
		{"timespan", kustolang.ScanTimespanLiteral("2.5hours"), 8},
		{"bad timespan suffix", kustolang.ScanTimespanLiteral("2fortnights"), -1},
		{"raw GUID", kustolang.ScanRawGuidLiteral("01234567-89ab-cdef-0123-456789abcdef", 0), 36},
		{"bad raw GUID", kustolang.ScanRawGuidLiteral("01234567-89ab-cdef-0123-456789abcdeg", 0), -1},
		{"closed string", kustolang.ScanStringLiteral(`"ok"`, 0, true), 4},
		{"unclosed string strict", kustolang.ScanStringLiteral(`"broken`, 0, true), -1},
	}
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("%s = %d, want %d", test.name, test.got, test.want)
		}
	}

	length, nameStart, nameLength, indexStart, indexLength := kustolang.ScanClientParameterParts("x{user[-2]}", 1)
	if length != 10 || nameStart != 2 || nameLength != 4 || indexStart != 7 || indexLength != 2 {
		t.Fatalf("client parameter parts = %d, %d, %d, %d, %d", length, nameStart, nameLength, indexStart, indexLength)
	}
	if kustolang.ScanClientParameter("{user[bad]}", 0) != -1 {
		t.Fatal("malformed client parameter was accepted")
	}
	if got := kustolang.ParseTrivia("x \t// comment\r\nvalue", 1); got != " \t// comment\r\n" {
		t.Fatalf("ParseTrivia = %q", got)
	}
}

func TestPublicTextCommentCommandAndScriptFacts(t *testing.T) {
	text := "  // first\r\n.show tables\r\n\r\nStormEvents | count\n"
	lineStarts := kustolang.GetLineStarts(text)
	if len(lineStarts) != 5 || kustolang.GetLineText(text, 2) != ".show tables" {
		t.Fatalf("line facts = starts:%v line:%q", lineStarts, kustolang.GetLineText(text, 2))
	}
	if !kustolang.IsCommentLine(text, 0) || kustolang.GetCommentLineText(text, 0) != "first" {
		t.Fatalf("comment line = %q", kustolang.GetCommentLineText(text, 0))
	}
	if comments := kustolang.GetCommentTexts(text); len(comments) != 1 || comments[0] != "first" {
		t.Fatalf("comment texts = %#v", comments)
	}
	commentStart, commentLength, ok := kustolang.TryGetCommentSpan("  // first\r\n  ", 5)
	if !ok || commentStart != 2 || commentLength != 10 {
		t.Fatalf("comment span = %d, %d, %t", commentStart, commentLength, ok)
	}
	commandText := "// first\r\n.show tables\r\n\r\n.show version"
	if starts := kustolang.GetCommandBlockStarts(commandText); len(starts) != 2 || starts[0] != 0 {
		t.Fatalf("command block starts = %v", starts)
	}
	if commands := kustolang.GetCommandTexts(commandText); len(commands) != 2 || commands[0] != ".show tables" || commands[1] != ".show version" {
		t.Fatalf("command texts = %#v", commands)
	}
	if kustolang.GetLineBreakLength(text, len("  // first")) != 2 || !kustolang.IsWhitespaceOnly(" \t\r\n") {
		t.Fatal("line-break or whitespace facts differ from upstream")
	}
	if _, _, ok := kustolang.TryGetTextLineAndOffset(text, -1); ok {
		t.Fatal("negative text position was accepted")
	}
	if blocks := kustolang.GetLineSeparatedBlockStarts(text, lineStarts); len(blocks) == 0 {
		t.Fatal("line-separated script block discovery returned no blocks")
	}
}
