package kustolang

import "testing"

func TestCharScannersMatchTextAndCharacters(t *testing.T) {
	if !ScannerMatches(Chars("Storm", true), "storm") {
		t.Fatal("Chars should match ignore-case text")
	}
	if ScannerMatches(Chars("Storm"), "storm") {
		t.Fatal("Chars should preserve case-sensitive matching by default")
	}
	if !ScannerMatches(Char('K', true), "k") {
		t.Fatal("Char should match ignore-case characters")
	}
	if !ScannerMatches(Char('é'), "é") {
		t.Fatal("Char should match UTF-8 encoded rune text")
	}
}

func TestCharScannersCommonCharacterClasses(t *testing.T) {
	tests := []struct {
		name    string
		parser  Parser[byte]
		text    string
		wantLen int
	}{
		{name: "letter", parser: Letter, text: "A", wantLen: 1},
		{name: "letter rejects underscore", parser: Letter, text: "_", wantLen: -1},
		{name: "digit", parser: Digit, text: "7", wantLen: 1},
		{name: "digit rejects letter", parser: Digit, text: "A", wantLen: -1},
		{name: "hex uppercase", parser: HexDigit, text: "F", wantLen: 1},
		{name: "hex rejects non hex", parser: HexDigit, text: "G", wantLen: -1},
		{name: "space", parser: Whitespace, text: " ", wantLen: 1},
		{name: "unicode no-break space", parser: Whitespace, text: "\u00a0", wantLen: len("\u00a0")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.parser.Scan(NewTextSource(tt.text), 0); got != tt.wantLen {
				t.Fatalf("scan length = %d, want %d", got, tt.wantLen)
			}
		})
	}
	if Letter.Tag() != "<letter>" || Digit.Tag() != "<digit>" || HexDigit.Tag() != "<hex-digit>" || Whitespace.Tag() != "<whitespace>" {
		t.Fatalf("scanner tags = %q %q %q %q", Letter.Tag(), Digit.Tag(), HexDigit.Tag(), Whitespace.Tag())
	}
}

func TestCharScannersLineBreakMatchesUpstreamAlternatives(t *testing.T) {
	tests := []struct {
		text    string
		wantLen int
	}{
		{text: "\r\n", wantLen: 2},
		{text: "\r", wantLen: 1},
		{text: "\n", wantLen: 1},
		{text: "\u2028", wantLen: len("\u2028")},
		{text: "\u2029", wantLen: len("\u2029")},
		{text: "x", wantLen: -1},
	}
	for _, tt := range tests {
		if got := LineBreak.Scan(NewTextSource(tt.text), 0); got != tt.wantLen {
			t.Fatalf("LineBreak %q length = %d, want %d", tt.text, got, tt.wantLen)
		}
	}
	if LineBreak.Tag() != "<line-break>" {
		t.Fatalf("LineBreak tag = %q", LineBreak.Tag())
	}
}

func TestScannerMatchesUsesOffsetAndExactLength(t *testing.T) {
	if !ScannerMatches(Chars("Events"), "StormEvents", 5, 6) {
		t.Fatal("ScannerMatches should match at supplied offset")
	}
	if ScannerMatches(Chars("Events"), "StormEvents", 5, 5) {
		t.Fatal("ScannerMatches should require the supplied length")
	}
	if ScannerMatches(Chars("Events"), "StormEvents", 99, 6) {
		t.Fatal("ScannerMatches should reject offsets past the source")
	}
	if ScannerMatches(nil, "Events") {
		t.Fatal("ScannerMatches should reject nil scanners")
	}
}
