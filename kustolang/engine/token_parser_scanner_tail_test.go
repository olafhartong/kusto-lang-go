package kustolang

import "testing"

func TestTokenParserParseTriviaTailParity(t *testing.T) {
	tests := []struct {
		text  string
		start int
		want  string
	}{
		{text: "value", want: ""},
		{text: "  value", want: "  "},
		{text: "x \t\u2003// comment\r\nnext", start: 1, want: " \t\u2003// comment\r\n"},
		{text: "// first\n\t// second\u2028next", want: "// first\n\t// second\u2028"},
		{text: "// eof", want: "// eof"},
		{text: "/ value", want: ""},
		{text: "", start: -1, want: ""},
		{text: "", start: 1, want: ""},
	}

	for _, test := range tests {
		if got := ParseTrivia(test.text, test.start); got != test.want {
			t.Errorf("ParseTrivia(%q, %d) = %q, want %q", test.text, test.start, got, test.want)
		}
	}
}

func TestTokenParserScanRawGuidLiteralTailParity(t *testing.T) {
	const lower = "01234567-89ab-cdef-0123-456789abcdef"
	const upper = "ABCDEF01-2345-6789-ABCD-EF0123456789"
	tests := []struct {
		text  string
		start int
		want  int
	}{
		{text: lower, want: 36},
		{text: upper, want: 36},
		{text: "x" + lower, start: 1, want: 36},
		{text: lower + "f", want: 36},
		{text: "01234567_89ab-cdef-0123-456789abcdef", want: -1},
		{text: "01234567-89ab-cdef-0123-456789abcdeg", want: -1},
		{text: lower[:35], want: -1},
		{text: "guid(" + lower + ")", want: -1},
		{text: "{" + lower + "}", want: -1},
		{text: lower, start: -1, want: -1},
		{text: lower, start: len(lower) + 1, want: -1},
	}

	for _, test := range tests {
		if got := ScanRawGuidLiteral(test.text, test.start); got != test.want {
			t.Errorf("ScanRawGuidLiteral(%q, %d) = %d, want %d", test.text, test.start, got, test.want)
		}
	}

	tokens := Lex(lower)
	if len(tokens) == 0 || tokens[0].Kind != SyntaxKindRawGuidLiteralToken || tokens[0].Text != lower {
		t.Fatalf("lexer raw GUID = %#v", tokens)
	}
}
