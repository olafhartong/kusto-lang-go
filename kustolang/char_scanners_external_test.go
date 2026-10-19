package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicCharacterScanners(t *testing.T) {
	tests := []struct {
		name    string
		scanner kustolang.Parser[byte]
		text    string
	}{
		{name: "chars", scanner: kustolang.CharScannerChars("Kusto"), text: "Kusto"},
		{name: "chars ignore case", scanner: kustolang.CharScannerChars("kusto", true), text: "KUSTO"},
		{name: "unicode char", scanner: kustolang.CharScannerChar('β'), text: "β"},
		{name: "unicode char ignore case", scanner: kustolang.CharScannerChar('β', true), text: "Β"},
		{name: "letter", scanner: kustolang.CharScannerLetter, text: "A"},
		{name: "digit", scanner: kustolang.CharScannerDigit, text: "7"},
		{name: "hex digit", scanner: kustolang.CharScannerHexDigit, text: "F"},
		{name: "whitespace", scanner: kustolang.CharScannerWhitespace, text: "\t"},
		{name: "CRLF", scanner: kustolang.CharScannerLineBreak, text: "\r\n"},
		{name: "CR", scanner: kustolang.CharScannerLineBreak, text: "\r"},
		{name: "LF", scanner: kustolang.CharScannerLineBreak, text: "\n"},
		{name: "line separator", scanner: kustolang.CharScannerLineBreak, text: "\u2028"},
		{name: "paragraph separator", scanner: kustolang.CharScannerLineBreak, text: "\u2029"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !kustolang.ScannerMatches(tt.scanner, tt.text) {
				t.Fatalf("scanner did not match %q", tt.text)
			}
		})
	}

	if kustolang.ScannerMatches(kustolang.CharScannerChars("abc"), "ab") {
		t.Fatal("partial text matched")
	}
	if kustolang.ScannerMatches(kustolang.CharScannerDigit, "x") {
		t.Fatal("non-digit matched")
	}
}

func TestPublicScannerMatchesRangeAndTags(t *testing.T) {
	if !kustolang.ScannerMatches(kustolang.CharScannerChars("βc"), "aβcd", 1, len("βc")) {
		t.Fatal("UTF-8 byte range did not match")
	}
	if kustolang.ScannerMatches(kustolang.CharScannerChars("βc"), "aβcd", 1, 2) {
		t.Fatal("incomplete UTF-8 range matched")
	}
	if kustolang.ScannerMatches(kustolang.CharScannerLetter, "x", -1, 1) || kustolang.ScannerMatches(nil, "x") {
		t.Fatal("invalid scanner range matched")
	}
	if got := kustolang.CharScannerLetter.Tag(); got != "<letter>" {
		t.Fatalf("letter tag = %q", got)
	}
	if got := kustolang.CharScannerLineBreak.Tag(); got != "<line-break>" {
		t.Fatalf("line-break tag = %q", got)
	}
}
