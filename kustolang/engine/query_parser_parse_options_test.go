package kustolang

import "testing"

func TestLexHonorsParseOptions(t *testing.T) {
	withoutEnd := DefaultParseOptions.WithAlwaysProduceEndTokens(false)
	if tokens := LexWithOptions("", withoutEnd); len(tokens) != 0 {
		t.Fatalf("empty token count = %d, want 0", len(tokens))
	}
	if tokens := LexWithOptions(" \t", withoutEnd); len(tokens) != 1 || tokens[0].Kind != SyntaxKindEndOfTextToken || tokens[0].Trivia != " \t" {
		t.Fatalf("trivia-only tokens = %#v, want one end token retaining trivia", tokens)
	}
	query := ParseQueryWithOptions("print x=1", withoutEnd)
	if query == nil || query.EndOfQuery() != nil || len(query.ContainedDiagnostics()) != 0 {
		t.Fatalf("query without forced end token = %#v diagnostics=%#v", query, GetContainedDiagnostics(query, DiagnosticsIncludeDefault))
	}
	queryWithTrailingTrivia := ParseQueryWithOptions("print x=1 ", withoutEnd)
	if queryWithTrailingTrivia == nil || queryWithTrailingTrivia.EndOfQuery() == nil || queryWithTrailingTrivia.EndOfQuery().Trivia() != " " {
		t.Fatalf("query trailing-trivia end token = %#v", queryWithTrailingTrivia)
	}

	text := "long(1\n2)"
	strict := LexWithOptions(text, DefaultParseOptions)
	if len(strict) == 0 || strict[0].Text != "long(1" || len(strict[0].Diagnostics) == 0 {
		t.Fatalf("strict literal tokens = %#v", strict)
	}
	permissive := LexWithOptions(text, DefaultParseOptions.WithAllowLiteralsWithLineBreaks(true))
	if len(permissive) != 2 || permissive[0].Text != text || len(permissive[0].Diagnostics) != 0 || permissive[1].Kind != SyntaxKindEndOfTextToken {
		t.Fatalf("multiline literal tokens = %#v", permissive)
	}
}

func TestStandaloneQueryParserEntryPointsHonorOptions(t *testing.T) {
	options := DefaultParseOptions.WithAllowLiteralsWithLineBreaks(true)
	text := "long(1\n2)"
	literal, ok := ParseLiteralWithOptions(text, options).(*LiteralExpression)
	if !ok || literal.Token().RawText() != text || len(literal.ContainedDiagnostics()) != 0 {
		t.Fatalf("text literal = %#v diagnostics=%#v", literal, GetContainedDiagnostics(literal, DiagnosticsIncludeDefault))
	}
	tokenLiteral, ok := ParseLiteralTokensWithOptions(LexWithOptions(text, options), options).(*LiteralExpression)
	if !ok || tokenLiteral.Token().RawText() != text || len(tokenLiteral.ContainedDiagnostics()) != 0 {
		t.Fatalf("token literal = %#v diagnostics=%#v", tokenLiteral, GetContainedDiagnostics(tokenLiteral, DiagnosticsIncludeDefault))
	}

	queryText := "print x=" + text
	if diagnostics := ParseQueryWithOptions(queryText, options).ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("option-aware query diagnostics = %#v", diagnostics)
	}
	if diagnostics := ParseQuery(queryText).ContainedDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("default query unexpectedly accepted a parenthesized literal spanning lines")
	}

	expression := ParseExpressionWithOptions("union Sec * 2024", DefaultParseOptions.WithAllowNonAdjacentWildcardParts(true))
	patterns := queryParserWildcardPatterns(expression)
	if len(patterns) != 1 || patterns[0].RawText() != "Sec * 2024" || patterns[0].ValueText() != "Sec*2024" {
		t.Fatalf("option-aware expression wildcards = %#v", patterns)
	}
	tokenExpression := ParseExpressionTokensWithOptions(
		Lex("union Sec * 2024"),
		DefaultParseOptions.WithAllowNonAdjacentWildcardParts(true),
	)
	if patterns = queryParserWildcardPatterns(tokenExpression); len(patterns) != 1 || patterns[0].ValueText() != "Sec*2024" {
		t.Fatalf("option-aware token expression wildcards = %#v", patterns)
	}
}

func TestStandaloneQueryParserOptionSurfaceAndMalformedLiteral(t *testing.T) {
	options := DefaultParseOptions.WithAllowLiteralsWithLineBreaks(true)
	if ParseFunctionParametersWithOptions("(x:long=long(1\n2))", options) == nil ||
		ParseFunctionParametersTokensWithOptions(LexWithOptions("(x:long)", options), options) == nil ||
		ParseFunctionParameterWithOptions("x:long=long(1\n2)", options) == nil ||
		ParseFunctionParameterTokensWithOptions(LexWithOptions("x:long", options), options) == nil ||
		ParseFunctionBodyWithOptions("{ print x=long(1\n2) }", options) == nil ||
		ParseFunctionBodyTokensWithOptions(LexWithOptions("{ print x=1 }", options), options) == nil ||
		ParseEntityPathWithOptions("database('db').T", options) == nil ||
		ParseEntityPathTokensWithOptions(Lex("database('db').T"), options) == nil ||
		ParseEntityGroupWithOptions("entity_group[T]", options) == nil ||
		ParseEntityGroupTokensWithOptions(Lex("entity_group[T]"), options) == nil ||
		ParseRowSchemaWithOptions("(x:long)", options) == nil ||
		ParseRowSchemaTokensWithOptions(Lex("(x:long)"), options) == nil {
		t.Fatal("an option-aware standalone parser entry point returned nil for valid source")
	}

	malformed := ParseLiteralWithOptions("long(1\n2", options)
	if malformed == nil || !diagnosticsContainMessage(malformed.ContainedDiagnostics(), "KS001", "Missing: )") {
		t.Fatalf("malformed multiline literal diagnostics = %#v", GetContainedDiagnostics(malformed, DiagnosticsIncludeDefault))
	}
}

func TestQueryParserAllowNonAdjacentWildcardParts(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		allow      bool
		useDefault bool
		wantRaw    string
		wantValue  string
	}{
		{name: "default adjacent", text: "union Sec*2024", useDefault: true, wantRaw: "Sec*2024", wantValue: "Sec*2024"},
		{name: "default separated", text: "union Sec * 2024", useDefault: true},
		{name: "explicit false adjacent", text: "union Sec*2024", allow: false, wantRaw: "Sec*2024", wantValue: "Sec*2024"},
		{name: "false separated", text: "union Sec * 2024", allow: false},
		{name: "true separated", text: "union Sec * 2024", allow: true, wantRaw: "Sec * 2024", wantValue: "Sec*2024"},
		{name: "true comment trivia", text: "union Sec // gap\n * 2024", allow: true, wantRaw: "Sec // gap\n * 2024", wantValue: "Sec*2024"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var query *QueryBlock
			if test.useDefault {
				query = ParseQuery(test.text)
			} else {
				options := DefaultParseOptions.WithAllowNonAdjacentWildcardParts(test.allow)
				query = ParseQueryWithOptions(test.text, options)
			}
			patterns := queryParserWildcardPatterns(query)
			if test.wantRaw == "" {
				if pattern := queryParserWildcardPatternWithValue(patterns, "Sec*2024"); pattern != nil {
					t.Fatalf("non-adjacent wildcard parts were combined with option disabled: raw=%q value=%q", pattern.RawText(), pattern.ValueText())
				}
				return
			}
			if len(patterns) != 1 {
				t.Fatalf("wildcard patterns = %d, want 1; diagnostics = %#v", len(patterns), query.ContainedDiagnostics())
			}
			if got := patterns[0].RawText(); got != test.wantRaw {
				t.Fatalf("wildcard raw text = %q, want %q", got, test.wantRaw)
			}
			if got := patterns[0].ValueText(); got != test.wantValue {
				t.Fatalf("wildcard value text = %q, want %q", got, test.wantValue)
			}
		})
	}
}

func TestQueryParserNonAdjacentWildcardMalformedRecovery(t *testing.T) {
	const text = "union [Sec * 2024"
	for _, allow := range []bool{false, true} {
		t.Run(map[bool]string{false: "disabled", true: "enabled"}[allow], func(t *testing.T) {
			query := ParseQueryWithOptions(text, DefaultParseOptions.WithAllowNonAdjacentWildcardParts(allow))
			patterns := queryParserWildcardPatterns(query)
			if allow {
				if len(patterns) != 1 || patterns[0].RawText() != "Sec * 2024" || patterns[0].ValueText() != "Sec*2024" {
					t.Fatalf("recovered wildcard patterns = %#v", patterns)
				}
				if !diagnosticsContainMessage(query.ContainedDiagnostics(), "KS005", "Expected: ]") {
					t.Fatalf("missing close-bracket recovery diagnostic = %#v", query.ContainedDiagnostics())
				}
			} else if len(patterns) != 0 {
				t.Fatalf("strict malformed input produced wildcard = %#v", patterns)
			}
		})
	}
}

func TestParseQueryTokensWithOptionsCarriesWildcardOption(t *testing.T) {
	const text = "union Sec\t* 2024"
	query := ParseQueryTokensWithOptions(Lex(text), DefaultParseOptions.WithAllowNonAdjacentWildcardParts(true))
	patterns := queryParserWildcardPatterns(query)
	if len(patterns) != 1 || patterns[0].RawText() != "Sec\t* 2024" || patterns[0].ValueText() != "Sec*2024" {
		t.Fatalf("token entry wildcard patterns = %#v", patterns)
	}
}

func TestAllowNonAdjacentWildcardPartsAcrossParserKinds(t *testing.T) {
	const text = "union Sec * 2024"
	for _, kind := range []ParserKind{ParserKindDefault, ParserKindGrammar} {
		for _, allow := range []bool{false, true} {
			name := map[ParserKind]string{ParserKindDefault: "default", ParserKindGrammar: "grammar"}[kind] +
				map[bool]string{false: "/disabled", true: "/enabled"}[allow]
			t.Run(name, func(t *testing.T) {
				options := DefaultParseOptions.
					WithParserKind(kind).
					WithAllowNonAdjacentWildcardParts(allow)
				code := ParseWithGlobals(text, NewGlobalState(nil).WithParseOptions(options))
				patterns := queryParserWildcardPatterns(code.Query)
				if !allow {
					if pattern := queryParserWildcardPatternWithValue(patterns, "Sec*2024"); pattern != nil {
						t.Fatalf("parser kind %v combined non-adjacent wildcard parts: raw=%q", kind, pattern.RawText())
					}
					return
				}
				if len(patterns) != 1 || patterns[0].RawText() != "Sec * 2024" || patterns[0].ValueText() != "Sec*2024" {
					t.Fatalf("parser kind %v wildcard patterns = %#v", kind, patterns)
				}
			})
		}
	}
}

func queryParserWildcardPatterns(element SyntaxElement) (patterns []*SyntaxToken) {
	WalkSyntax(element, func(element SyntaxElement) {
		switch name := WrapGeneratedSyntaxElement(element).(type) {
		case *WildcardedName:
			patterns = append(patterns, name.Pattern())
		case *BracketedWildcardedName:
			patterns = append(patterns, name.Pattern())
		}
	})
	return patterns
}

func queryParserWildcardPatternWithValue(patterns []*SyntaxToken, value string) *SyntaxToken {
	for _, pattern := range patterns {
		if pattern.ValueText() == value {
			return pattern
		}
	}
	return nil
}
