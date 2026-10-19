package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicUTF16CharacterScannersMatchCSharpCharUnits(t *testing.T) {
	tests := []struct {
		name    string
		scanner kustolang.Parser[uint16]
		text    string
		length  int
	}{
		{name: "chars", scanner: kustolang.UTF16CharScannerChars("Kusto"), text: "Kusto", length: 5},
		{name: "chars ignore case", scanner: kustolang.UTF16CharScannerChars("kusto", true), text: "KUSTO", length: 5},
		{name: "BMP char", scanner: kustolang.UTF16CharScannerChar(uint16('β')), text: "β", length: 1},
		{name: "BMP char ignore case", scanner: kustolang.UTF16CharScannerChar(uint16('β'), true), text: "Β", length: 1},
		{name: "supplementary pair", scanner: kustolang.UTF16CharScannerChars("😀"), text: "😀", length: 2},
		{name: "CRLF", scanner: kustolang.UTF16CharScannerLineBreak, text: "\r\n", length: 2},
		{name: "CR", scanner: kustolang.UTF16CharScannerLineBreak, text: "\r", length: 1},
		{name: "line separator", scanner: kustolang.UTF16CharScannerLineBreak, text: "\u2028", length: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			units := kustolang.UTF16CodeUnits(test.text)
			if got := test.scanner.Scan(kustolang.NewArraySource(units), 0); got != test.length {
				t.Fatalf("Scan() = %d, want %d; units = %#v", got, test.length, units)
			}
			if !kustolang.UTF16ScannerMatches(test.scanner, test.text) {
				t.Fatalf("scanner did not match %q", test.text)
			}
		})
	}
}

func TestPublicUTF16ScannerRangesFailuresAndTags(t *testing.T) {
	if got := kustolang.UTF16CodeUnits("A😀β"); !reflect.DeepEqual(got, []uint16{'A', 0xd83d, 0xde00, 'β'}) {
		t.Fatalf("UTF16CodeUnits() = %#v", got)
	}
	if !kustolang.UTF16ScannerMatches(kustolang.UTF16CharScannerChars("😀β"), "A😀βZ", 1, 3) {
		t.Fatal("UTF-16 code-unit range did not match")
	}
	if kustolang.UTF16ScannerMatches(kustolang.UTF16CharScannerChars("😀β"), "A😀βZ", 1, 2) {
		t.Fatal("partial UTF-16 range matched")
	}
	source := kustolang.NewArraySource(kustolang.UTF16CodeUnits("ax"))
	if got := kustolang.UTF16CharScannerChars("ab").Scan(source, 0); got != ^1 {
		t.Fatalf("mismatch failure = %d, want %d", got, ^1)
	}
	if kustolang.UTF16ScannerMatches(nil, "x") || kustolang.UTF16ScannerMatches(kustolang.UTF16CharScannerLetter, "x", -1, 1) {
		t.Fatal("invalid UTF-16 scanner range matched")
	}
	if kustolang.UTF16CharScannerLetter.Tag() != "<letter>" || kustolang.UTF16CharScannerLineBreak.Tag() != "<line-break>" {
		t.Fatal("UTF-16 scanner tags differ from CharScanners")
	}
}
