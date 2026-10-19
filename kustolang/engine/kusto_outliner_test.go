package kustolang

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestKustoOutlinerRootOutline(t *testing.T) {
	text := "SigninLogs\n| where ResultType != 0\n| project UserPrincipalName\n\n"
	info := KustoOutlinerGetOutlines(Parse(text), OutliningOptionsDefault)
	if len(info.Ranges) != 1 {
		t.Fatalf("outline count = %d, want 1: %#v", len(info.Ranges), info.Ranges)
	}
	outline := info.Ranges[0]
	if outline.Start != 0 || outline.Length != TrimEnd(text) || outline.CollapsedText != "SigninLogs ..." {
		t.Fatalf("root outline = %#v", outline)
	}
}

func TestKustoOutlinerStatementOutlines(t *testing.T) {
	text := "let threshold = 10;\nSigninLogs\n| where ResultType > threshold\n| project UserPrincipalName"
	options := OutliningOptionsDefault.WithQueries(false).WithStatements(true)
	info := KustoOutlinerGetOutlines(Parse(text), options)
	if len(info.Ranges) != 1 {
		t.Fatalf("statement outline count = %d, want 1: %#v", len(info.Ranges), info.Ranges)
	}
	outline := info.Ranges[0]
	if outline.Start != len("let threshold = 10;\n") || outline.CollapsedText != "SigninLogs ..." {
		t.Fatalf("statement outline = %#v", outline)
	}
}

func TestKustoOutlinerCollapsedTextWithBraces(t *testing.T) {
	text := "let f = (x:long) {\n    range i from 1 to x step 1\n};\nf(3)"
	info := KustoOutlinerGetOutlines(Parse(text), OutliningOptionsDefault)
	if len(info.Ranges) != 1 {
		t.Fatalf("brace outline count = %d, want 1: %#v", len(info.Ranges), info.Ranges)
	}
	if got := info.Ranges[0].CollapsedText; got != "let f = (x:long) { ... }" {
		t.Fatalf("brace collapsed text = %q", got)
	}
}

func TestKustoOutlinerSingleLineStatementSkipped(t *testing.T) {
	text := "let threshold = 10;\nSigninLogs | take 1"
	options := OutliningOptionsDefault.WithQueries(false).WithStatements(true)
	info := KustoOutlinerGetOutlines(Parse(text), options)
	if len(info.Ranges) != 0 {
		t.Fatalf("single-line statement outlines = %#v, want none", info.Ranges)
	}
}

func TestKustoOutlinerCollapsedTextUsesUTF16Budget(t *testing.T) {
	text := "// 😀X\nSigninLogs\n| take 1"
	options := OutliningOptionsDefault.WithMaxCollapsedTextLength(5)
	info := KustoOutlinerGetOutlines(Parse(text), options)
	if len(info.Ranges) != 1 {
		t.Fatalf("outline count = %d, want 1", len(info.Ranges))
	}
	if got := info.Ranges[0].CollapsedText; got != "// 😀 ..." || !utf8.ValidString(got) {
		t.Fatalf("collapsed text = %q, want valid UTF-8 UTF-16 prefix", got)
	}
}

func TestKustoCodeServiceOutlinerRecoversInvalidOptions(t *testing.T) {
	text := "// comment\nSigninLogs\n| take 1"
	options := OutliningOptionsDefault.WithMaxCollapsedTextLength(-1)
	info := NewKustoCodeService(text).GetOutlines(options, CancellationToken{})
	if len(info.Ranges) != 1 || !strings.HasPrefix(info.Ranges[0].CollapsedText, "SigninLogs") {
		t.Fatalf("fallback outlines = %#v", info.Ranges)
	}
}
