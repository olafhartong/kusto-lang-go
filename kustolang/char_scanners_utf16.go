package kustolang

// Ported from src/Kusto.Language/Parser/CharScanners.cs and
// ScannerExtensions.cs. See third_party/kusto/LICENSE.
//
// The UTF-16 facade mirrors C# Parser<char> offsets and lengths. Use the
// byte-based CharScanner API when integrating with native UTF-8 Go text.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

var (
	UTF16CharScannerLetter     Parser[uint16] = internal.UTF16Letter
	UTF16CharScannerDigit      Parser[uint16] = internal.UTF16Digit
	UTF16CharScannerHexDigit   Parser[uint16] = internal.UTF16HexDigit
	UTF16CharScannerWhitespace Parser[uint16] = internal.UTF16Whitespace
	UTF16CharScannerLineBreak  Parser[uint16] = internal.UTF16LineBreak
)

func UTF16CharScannerChars(text string, ignoreCase ...bool) Parser[uint16] {
	return internal.UTF16Chars(text, ignoreCase...)
}

func UTF16CharScannerChar(ch uint16, ignoreCase ...bool) Parser[uint16] {
	return internal.UTF16Char(ch, ignoreCase...)
}

func UTF16ScannerMatches(scanner Parser[uint16], text string, offsetAndLength ...int) bool {
	return internal.UTF16ScannerMatches(scanner, text, offsetAndLength...)
}

func UTF16CodeUnits(text string) []uint16 {
	return internal.UTF16CodeUnits(text)
}
