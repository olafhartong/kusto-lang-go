package kustolang

// Ported from src/Kusto.Language/Parser/ScannerExtensions.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func ScannerMatches(scanner Parser[byte], text string, offsetAndLength ...int) bool {
	if scanner == nil {
		return false
	}
	offset := 0
	length := len(text)
	if len(offsetAndLength) >= 1 {
		offset = offsetAndLength[0]
	}
	if len(offsetAndLength) >= 2 {
		length = offsetAndLength[1]
	}
	if offset < 0 || offset > len(text) || length < 0 {
		return false
	}
	return scanner.Scan(NewTextSource(text), offset) == length
}
