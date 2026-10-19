package kustolang

import "testing"

func TestTokenParserScannersUseUnicodeRunes(t *testing.T) {
	if got := ScanIdentifier("Míssing tail", 0); got != 1 {
		t.Fatalf("ScanIdentifier must stop at non-ASCII letter: %d", got)
	}
	if got := ScanWhitespace("\u2003\u00a0x", 0); got != len("\u2003\u00a0") {
		t.Fatalf("ScanWhitespace unicode length = %d, want %d bytes", got, len("\u2003\u00a0"))
	}
	if got := ScanTrivia("\u2003// comment\u2028β", 0); got != len("\u2003// comment\u2028") {
		t.Fatalf("ScanTrivia unicode length = %d, want %d bytes", got, len("\u2003// comment\u2028"))
	}
}
