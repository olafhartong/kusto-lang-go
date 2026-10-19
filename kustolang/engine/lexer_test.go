package kustolang

import "testing"

func TestLexPreservesTriviaAndRecognizesKQLTokens(t *testing.T) {
	tokens := Lex("// comment\nlet x = 'abc';\nSigninLogs | where ResultType in~ (0, 1) | project-away Foo")
	kinds := tokenKinds(tokens)
	want := []SyntaxKind{
		SyntaxKindLetKeyword,
		SyntaxKindIdentifierToken,
		SyntaxKindEqualToken,
		SyntaxKindStringLiteralToken,
		SyntaxKindSemicolonToken,
		SyntaxKindIdentifierToken,
		SyntaxKindBarToken,
		SyntaxKindWhereKeyword,
		SyntaxKindIdentifierToken,
		SyntaxKindInCsKeyword,
		SyntaxKindOpenParenToken,
		SyntaxKindLongLiteralToken,
		SyntaxKindCommaToken,
		SyntaxKindLongLiteralToken,
		SyntaxKindCloseParenToken,
		SyntaxKindBarToken,
		SyntaxKindProjectAwayKeyword,
		SyntaxKindIdentifierToken,
		SyntaxKindEndOfTextToken,
	}
	if len(kinds) != len(want) {
		t.Fatalf("token count = %d, want %d: %#v", len(kinds), len(want), kinds)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("token %d kind = %v, want %v", i, kinds[i], want[i])
		}
	}
	if tokens[0].Trivia != "// comment\n" {
		t.Fatalf("first trivia = %q", tokens[0].Trivia)
	}
}

func TestLexHintAndBangKeywords(t *testing.T) {
	tokens := Lex("summarize hint.shufflekey = Recipient count() by Subject | where Subject !in ('x')")
	assertTokenKind(t, tokens, 0, SyntaxKindSummarizeKeyword)
	assertTokenKind(t, tokens, 1, SyntaxKindHintDotShuffleKeyKeyword)
	assertTokenKind(t, tokens, 2, SyntaxKindEqualToken)
	if !containsKind(tokens, SyntaxKindNotInKeyword) {
		t.Fatalf("expected !in token in %#v", tokenKinds(tokens))
	}
}

func TestLexClientDirectiveToken(t *testing.T) {
	tokens := Lex("#set query_now=datetime(2024-01-01)\nT | count")
	assertTokenKind(t, tokens, 0, SyntaxKindDirectiveToken)
	if tokens[0].Text != "#set query_now=datetime(2024-01-01)" {
		t.Fatalf("directive text = %q", tokens[0].Text)
	}
	assertTokenKind(t, tokens, 1, SyntaxKindIdentifierToken)
	if tokens[1].Trivia != "\n" {
		t.Fatalf("post-directive trivia = %q", tokens[1].Trivia)
	}
}

func TestLexUnexpectedCharacterUsesUpstreamDiagnostic(t *testing.T) {
	tokens := Lex("^")
	assertTokenKind(t, tokens, 0, SyntaxKindBadToken)
	if tokens[0].Text != "^" {
		t.Fatalf("bad token text = %q", tokens[0].Text)
	}
	if len(tokens[0].Diagnostics) != 1 ||
		tokens[0].Diagnostics[0].Code != "KS002" ||
		tokens[0].Diagnostics[0].Message != "Unexpected: ^" {
		t.Fatalf("bad token diagnostics = %#v", tokens[0].Diagnostics)
	}
}

func TestLexUnicodeIdentifiersTriviaAndBadTokens(t *testing.T) {
	tokens := Lex("\u2003Míssing 😀")
	if len(tokens) != 5 {
		t.Fatalf("tokens = %#v", tokens)
	}
	if tokens[0].Kind != SyntaxKindIdentifierToken || tokens[0].Trivia != "\u2003" || tokens[0].Text != "M" {
		t.Fatalf("ASCII identifier prefix = %#v", tokens[0])
	}
	if tokens[1].Kind != SyntaxKindBadToken || tokens[1].Text != "í" || len(tokens[1].Diagnostics) != 1 {
		t.Fatalf("BMP bad token = %#v", tokens[1])
	}
	if tokens[2].Kind != SyntaxKindIdentifierToken || tokens[2].Text != "ssing" {
		t.Fatalf("ASCII identifier suffix = %#v", tokens[2])
	}
	if tokens[3].Kind != SyntaxKindBadToken || tokens[3].Trivia != " " || tokens[3].Text != "😀" || len(tokens[3].Diagnostics) != 1 {
		t.Fatalf("astral bad token = %#v", tokens[3])
	}
}

func TestLexUnicodeLineSeparatorEndsComment(t *testing.T) {
	tokens := Lex("// comment\u2028T")
	if len(tokens) != 2 || tokens[0].Kind != SyntaxKindIdentifierToken || tokens[0].Text != "T" || tokens[0].Trivia != "// comment\u2028" {
		t.Fatalf("unicode line separator tokens = %#v", tokens)
	}
}

func TestLexUnicodeLineSeparatorIsNotWhitespace(t *testing.T) {
	tokens := Lex("T\u2028U")
	if len(tokens) != 4 || tokens[1].Kind != SyntaxKindBadToken || tokens[1].Text != "\u2028" {
		t.Fatalf("standalone Unicode line separator tokens = %#v", tokens)
	}
	if len(tokens[1].Diagnostics) != 1 || tokens[1].Diagnostics[0].Code != "KS002" {
		t.Fatalf("standalone Unicode line separator diagnostics = %#v", tokens[1].Diagnostics)
	}
}

func TestLexGuidLiterals(t *testing.T) {
	tokens := Lex(`print A=guid(00000000-0000-0000-0000-000000000001), B=00000000-0000-0000-0000-000000000002`)
	if !containsKind(tokens, SyntaxKindGuidLiteralToken) {
		t.Fatalf("expected guid literal token in %#v", tokenKinds(tokens))
	}
	if !containsKind(tokens, SyntaxKindRawGuidLiteralToken) {
		t.Fatalf("expected raw guid literal token in %#v", tokenKinds(tokens))
	}
}

func TestLexAdditionalUpstreamKeywords(t *testing.T) {
	tokens := Lex("T | make-series count() on Timestamp step 1h | mv-expand Items | where Name contains_cs 'x'")
	if !containsKind(tokens, SyntaxKindMakeSeriesKeyword) {
		t.Fatalf("expected make-series keyword in %#v", tokenKinds(tokens))
	}
	if !containsKind(tokens, SyntaxKindTimespanLiteralToken) {
		t.Fatalf("expected timespan literal in %#v", tokenKinds(tokens))
	}
	if !containsKind(tokens, SyntaxKindMvDashExpandKeyword) {
		t.Fatalf("expected mv-expand keyword in %#v", tokenKinds(tokens))
	}
	if !containsKind(tokens, SyntaxKindContains_CsKeyword) {
		t.Fatalf("expected contains_cs keyword in %#v", tokenKinds(tokens))
	}
}

func TestLexFixedTextIsCaseSensitive(t *testing.T) {
	tokens := Lex("project PROJECT First Missing Title true True TRUE false False FALSE tRuE")
	assertTokenKind(t, tokens, 0, SyntaxKindProjectKeyword)
	for i, text := range []string{"PROJECT", "First", "Missing", "Title"} {
		index := i + 1
		assertTokenKind(t, tokens, index, SyntaxKindIdentifierToken)
		if tokens[index].Text != text {
			t.Fatalf("token %d text = %q, want %q", index, tokens[index].Text, text)
		}
	}
	for index := 5; index <= 10; index++ {
		assertTokenKind(t, tokens, index, SyntaxKindBooleanLiteralToken)
	}
	assertTokenKind(t, tokens, 11, SyntaxKindIdentifierToken)
}

func TestLexGraphOperatorTokens(t *testing.T) {
	tokens := Lex("T | make-graph Source --> Target | make-graph Source -- Target with NodesTable on NodeId | graph-match (a)-[e*1..3]->(b)<--(c)<-[r]-(d) | graph-to-table edges")
	if !containsKind(tokens, SyntaxKindMakeGraphKeyword) {
		t.Fatalf("expected make-graph keyword in %#v", tokenKinds(tokens))
	}
	if !containsKind(tokens, SyntaxKindDashDashGreaterThanToken) {
		t.Fatalf("expected directed graph token in %#v", tokenKinds(tokens))
	}
	if !containsKind(tokens, SyntaxKindDashDashToken) {
		t.Fatalf("expected undirected graph token in %#v", tokenKinds(tokens))
	}
	if !containsKind(tokens, SyntaxKindLessThanDashDashToken) {
		t.Fatalf("expected reverse graph token in %#v", tokenKinds(tokens))
	}
	if !containsKind(tokens, SyntaxKindDashBracketToken) {
		t.Fatalf("expected graph edge-open token in %#v", tokenKinds(tokens))
	}
	if !containsKind(tokens, SyntaxKindBracketDashGreaterThanToken) {
		t.Fatalf("expected graph edge-close token in %#v", tokenKinds(tokens))
	}
	if !containsKind(tokens, SyntaxKindLessThanDashBracketToken) {
		t.Fatalf("expected reverse graph edge-open token in %#v", tokenKinds(tokens))
	}
	if !containsKind(tokens, SyntaxKindBracketDashToken) {
		t.Fatalf("expected graph edge-close undirected token in %#v", tokenKinds(tokens))
	}
	if !containsKind(tokens, SyntaxKindGraphToTableKeyword) {
		t.Fatalf("expected graph-to-table keyword in %#v", tokenKinds(tokens))
	}
	if !containsKind(tokens, SyntaxKindGraphEdgesKeyword) {
		t.Fatalf("expected graph edges keyword in %#v", tokenKinds(tokens))
	}
	if !containsKind(tokens, SyntaxKindWithKeyword) {
		t.Fatalf("expected with keyword in %#v", tokenKinds(tokens))
	}
	if !containsKind(tokens, SyntaxKindOnKeyword) {
		t.Fatalf("expected on keyword in %#v", tokenKinds(tokens))
	}
}

func TestLexTimespanLiteralSuffixes(t *testing.T) {
	tokens := Lex("ago(1h) ago(7d) ago(12h) ago(1.5hours) ago(500ms)")
	var got []string
	for _, token := range tokens {
		if token.Kind == SyntaxKindTimespanLiteralToken {
			got = append(got, token.Text)
		}
	}
	want := []string{"1h", "7d", "12h", "1.5hours", "500ms"}
	if len(got) != len(want) {
		t.Fatalf("timespan count = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("timespan %d = %q, want %q", i, got[i], want[i])
		}
	}

	noSuffixBoundary := Lex("ago(1hourly)")
	if containsKind(noSuffixBoundary, SyntaxKindTimespanLiteralToken) {
		t.Fatalf("unexpected timespan token in %#v", tokenKinds(noSuffixBoundary))
	}
}

func TestLexRawIntegerLiteralForms(t *testing.T) {
	tests := []struct {
		text      string
		want      SyntaxKind
		valueText string
		value     any
	}{
		{text: "42", want: SyntaxKindLongLiteralToken, valueText: "42", value: int64(42)},
		{text: "0x10", want: SyntaxKindLongLiteralToken, valueText: "0x10", value: int64(16)},
		{text: "1.0", want: SyntaxKindRealLiteralToken, valueText: "1.0", value: 1.0},
		{text: "1.", want: SyntaxKindRealLiteralToken, valueText: "1.", value: 1.0},
		{text: "1e3", want: SyntaxKindRealLiteralToken, valueText: "1e3", value: 1000.0},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			tokens := Lex(tt.text)
			assertTokenKind(t, tokens, 0, tt.want)
			token := SyntaxTokenFromLexical(tokens[0])
			if token.ValueText() != tt.valueText {
				t.Fatalf("literal value text = %q, want %q", token.ValueText(), tt.valueText)
			}
			if got := token.Value(); got != tt.value {
				t.Fatalf("literal value = %#v, want %#v", got, tt.value)
			}
		})
	}
}

func TestLexDigitStartIdentifierFallback(t *testing.T) {
	tokens := Lex("123abc 1e")
	assertTokenKind(t, tokens, 0, SyntaxKindIdentifierToken)
	if tokens[0].Text != "123abc" {
		t.Fatalf("digit-start identifier text = %q", tokens[0].Text)
	}
	assertTokenKind(t, tokens, 1, SyntaxKindIdentifierToken)
	if tokens[1].Text != "1e" {
		t.Fatalf("malformed exponent identifier text = %q", tokens[1].Text)
	}
}

func TestLexGooLiteralTokens(t *testing.T) {
	tests := []struct {
		text      string
		want      SyntaxKind
		valueText string
	}{
		{text: "datetime(1910-06-11)", want: SyntaxKindDateTimeLiteralToken, valueText: "1910-06-11"},
		{text: "date(1910-06-11)", want: SyntaxKindDateTimeLiteralToken, valueText: "1910-06-11"},
		{text: "timespan(1d)", want: SyntaxKindTimespanLiteralToken, valueText: "1d"},
		{text: "time(01:02:03)", want: SyntaxKindTimespanLiteralToken, valueText: "01:02:03"},
		{text: "decimal(1.25)", want: SyntaxKindDecimalLiteralToken, valueText: "1.25"},
		{text: "long(42)", want: SyntaxKindLongLiteralToken, valueText: "42"},
		{text: "int(42)", want: SyntaxKindIntLiteralToken, valueText: "42"},
		{text: "real(1.5)", want: SyntaxKindRealLiteralToken, valueText: "1.5"},
		{text: "bool(true)", want: SyntaxKindBooleanLiteralToken, valueText: "true"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			tokens := Lex(tt.text)
			assertTokenKind(t, tokens, 0, tt.want)
			if tokens[0].Text != tt.text {
				t.Fatalf("literal text = %q, want %q", tokens[0].Text, tt.text)
			}
			token := SyntaxTokenFromLexical(tokens[0])
			if token.ValueText() != tt.valueText {
				t.Fatalf("literal value text = %q, want %q", token.ValueText(), tt.valueText)
			}
		})
	}
}

func TestLexUnterminatedGooLiteralDiagnostic(t *testing.T) {
	tests := []struct {
		text string
		want SyntaxKind
	}{
		{text: "datetime(2024-01-01", want: SyntaxKindDateTimeLiteralToken},
		{text: "long(42", want: SyntaxKindLongLiteralToken},
		{text: "guid(00000000-0000-0000-0000-000000000001", want: SyntaxKindGuidLiteralToken},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			tokens := Lex(tt.text)
			assertTokenKind(t, tokens, 0, tt.want)
			if len(tokens[0].Diagnostics) != 1 ||
				tokens[0].Diagnostics[0].Code != "KS001" ||
				tokens[0].Diagnostics[0].Message != "Missing: )" ||
				tokens[0].Diagnostics[0].LocationKind != DiagnosticLocationRelativeEnd {
				t.Fatalf("goo literal diagnostics = %#v", tokens[0].Diagnostics)
			}
		})
	}
}

func TestLexStringLiteralForms(t *testing.T) {
	tests := []struct {
		text      string
		valueText string
	}{
		{text: `'a\nb'`, valueText: "a\nb"},
		{text: `@"C:\Temp\file.txt"`, valueText: `C:\Temp\file.txt`},
		{text: `@"a ""quoted"" value"`, valueText: `a "quoted" value`},
		{text: `h'secret'`, valueText: "secret"},
		{text: "H@'don''t escape \\'", valueText: `don't escape \`},
		{text: "```line1\nline2```", valueText: "line1\nline2"},
		{text: "~~~line1\nline2~~~", valueText: "line1\nline2"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			tokens := Lex(tt.text)
			assertTokenKind(t, tokens, 0, SyntaxKindStringLiteralToken)
			if len(tokens[0].Diagnostics) != 0 {
				t.Fatalf("unexpected string diagnostics = %#v", tokens[0].Diagnostics)
			}
			token := SyntaxTokenFromLexical(tokens[0])
			if token.ValueText() != tt.valueText {
				t.Fatalf("value text = %q, want %q", token.ValueText(), tt.valueText)
			}
		})
	}
}

func TestLexSingleBacktickAndTildeAreNotStringDelimiters(t *testing.T) {
	for _, text := range []string{"`value`", "~value~"} {
		t.Run(text, func(t *testing.T) {
			tokens := Lex(text)
			if tokens[0].Kind != SyntaxKindBadToken || len(tokens[0].Diagnostics) != 1 || tokens[0].Diagnostics[0].Code != "KS002" {
				t.Fatalf("first token = %#v", tokens[0])
			}
		})
	}
}

func TestLexUnterminatedStringDiagnostic(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		message string
	}{
		{name: "single quote", text: "print x = 'unterminated", message: "Missing: '"},
		{name: "double quote", text: `print x = "unterminated`, message: `Missing: "`},
		{name: "multiline backtick", text: "print x = ```unterminated", message: "Missing: ```"},
		{name: "multiline tilde", text: "print x = ~~~unterminated", message: "Missing: ~~~"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := Lex(tt.text)
			var found bool
			for _, token := range tokens {
				if token.Kind == SyntaxKindStringLiteralToken {
					found = true
					if len(token.Diagnostics) != 1 || token.Diagnostics[0].Code != "KS001" || token.Diagnostics[0].Message != tt.message || token.Diagnostics[0].LocationKind != DiagnosticLocationRelativeEnd {
						t.Fatalf("string diagnostics = %#v", token.Diagnostics)
					}
				}
			}
			if !found {
				t.Fatal("string token not found")
			}
		})
	}
}

func TestStripLineCommentsIgnoresStringContent(t *testing.T) {
	got := StripLineComments("print x='http://example', y=1 // trailing\n| project x")
	want := "print x='http://example', y=1 \n| project x"
	if got != want {
		t.Fatalf("StripLineComments = %q, want %q", got, want)
	}
	multiline := "print x=```a//b\nc``` // trailing\r\n| project x"
	multilineWant := "print x=```a//b\nc``` \r\n| project x"
	if got := StripLineComments(multiline); got != multilineWant {
		t.Fatalf("StripLineComments multiline = %q, want %q", got, multilineWant)
	}
}

func tokenKinds(tokens []LexicalToken) []SyntaxKind {
	out := make([]SyntaxKind, len(tokens))
	for i, token := range tokens {
		out[i] = token.Kind
	}
	return out
}

func assertTokenKind(t *testing.T, tokens []LexicalToken, index int, want SyntaxKind) {
	t.Helper()
	if index >= len(tokens) {
		t.Fatalf("missing token %d in %#v", index, tokenKinds(tokens))
	}
	if tokens[index].Kind != want {
		t.Fatalf("token %d = %v, want %v", index, tokens[index].Kind, want)
	}
}

func containsKind(tokens []LexicalToken, want SyntaxKind) bool {
	for _, token := range tokens {
		if token.Kind == want {
			return true
		}
	}
	return false
}
