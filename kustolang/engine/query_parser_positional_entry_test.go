package kustolang

import "testing"

// Ported from the positional public entry contracts in Parser/TokenParser.cs
// and Parser/QueryParser.cs. See third_party/kusto/LICENSE.

func TestTokenParserPositionalEntryPoints(t *testing.T) {
	text := "prefix  // gap\n long(1\n2)"
	options := DefaultParseOptions.WithAllowLiteralsWithLineBreaks(true)
	token := ParseTokenWithOptions(text, len("prefix"), options)
	if token == nil || token.Kind != SyntaxKindLongLiteralToken || token.Trivia != "  // gap\n " || token.Text != "long(1\n2)" || len(token.Diagnostics) != 0 {
		t.Fatalf("positional token = %#v", token)
	}
	if first := ParseToken(text, -10); first == nil || first.Text != "prefix" {
		t.Fatalf("negative-start token = %#v", first)
	}
	if end := ParseToken(text, len(text)+10); end == nil || end.Kind != SyntaxKindEndOfTextToken {
		t.Fatalf("past-end token = %#v", end)
	}
	withoutEnd := DefaultParseOptions.WithAlwaysProduceEndTokens(false)
	if end := ParseTokenWithOptions(text, len(text), withoutEnd); end != nil {
		t.Fatalf("optional end token = %#v, want nil", end)
	}
	if end := ParseTokenWithOptions(text+" ", len(text), withoutEnd); end == nil || end.Kind != SyntaxKindEndOfTextToken || end.Trivia != " " {
		t.Fatalf("trivia-owned end token = %#v", end)
	}

	parsed := ParseTokensWithOptions("long(1\n2)", options)
	lexed := LexWithOptions("long(1\n2)", options)
	if len(parsed) != len(lexed) || len(parsed) != 2 || parsed[0].Text != lexed[0].Text || parsed[1].Kind != SyntaxKindEndOfTextToken {
		t.Fatalf("ParseTokens=%#v Lex=%#v", parsed, lexed)
	}
}

func TestQueryParserTokenStartEntryPoints(t *testing.T) {
	options := DefaultParseOptions.
		WithAllowNonAdjacentWildcardParts(true).
		WithAllowLiteralsWithLineBreaks(true)

	queryTokens := Lex("ignored ; print x=1")
	query := ParseQueryTokensFromWithOptions(queryTokens, 2, options)
	if query == nil || query.Statements().Count() != 1 || len(query.ContainedDiagnostics()) != 0 {
		t.Fatalf("positional query = %#v diagnostics=%#v", query, GetContainedDiagnostics(query, DiagnosticsIncludeDefault))
	}
	if defaultQuery := ParseQueryTokensFrom(queryTokens, 2); defaultQuery == nil || defaultQuery.Statements().Count() != 1 {
		t.Fatalf("default positional query = %#v", defaultQuery)
	}

	expressionTokens := Lex("ignored ; union Sec * 2024")
	expression := ParseExpressionTokensFromWithOptions(expressionTokens, 2, options)
	patterns := queryParserWildcardPatterns(expression)
	if len(patterns) != 1 || patterns[0].RawText() != "Sec * 2024" || patterns[0].ValueText() != "Sec*2024" {
		t.Fatalf("positional expression patterns = %#v", patterns)
	}
	if ParseExpressionTokensFrom(expressionTokens, 2) == nil {
		t.Fatal("default positional expression returned nil")
	}

	literalText := "ignored ; long(1\n2)"
	literalTokens := LexWithOptions(literalText, options)
	literal := ParseLiteralTokensFromWithOptions(literalTokens, 2, options)
	if literal == nil || literal.Text(IncludeTriviaInterior) != "long(1\n2)" || len(literal.ContainedDiagnostics()) != 0 {
		t.Fatalf("positional literal = %#v diagnostics=%#v", literal, GetContainedDiagnostics(literal, DiagnosticsIncludeDefault))
	}
	if ParseLiteralTokensFrom(Lex("ignored ; 1"), 2) == nil {
		t.Fatal("default positional literal returned nil")
	}

	if ParseEntityPathTokensFromWithOptions(Lex("ignored ; database('db').T"), 2, options) == nil ||
		ParseEntityPathTokensFrom(Lex("ignored ; T"), 2) == nil ||
		ParseEntityGroupTokensFromWithOptions(Lex("ignored ; entity_group[T]"), 2, options) == nil ||
		ParseEntityGroupTokensFrom(Lex("ignored ; entity_group[T]"), 2) == nil ||
		ParseRowSchemaTokensFromWithOptions(Lex("ignored ; (x:long)"), 2, options) == nil ||
		ParseRowSchemaTokensFrom(Lex("ignored ; (x:long)"), 2) == nil ||
		ParseFunctionParametersTokensFromWithOptions(Lex("ignored ; (x:long)"), 2, options) == nil ||
		ParseFunctionParametersTokensFrom(Lex("ignored ; (x:long)"), 2) == nil ||
		ParseFunctionParameterTokensFromWithOptions(Lex("ignored ; x:long"), 2, options) == nil ||
		ParseFunctionParameterTokensFrom(Lex("ignored ; x:long"), 2) == nil ||
		ParseFunctionBodyTokensFromWithOptions(Lex("ignored ; { print x=1 }"), 2, options) == nil ||
		ParseFunctionBodyTokensFrom(Lex("ignored ; { print x=1 }"), 2) == nil {
		t.Fatal("a positional standalone parser returned nil for valid source")
	}
}

func TestQueryParserTokenStartMalformedAndBoundaries(t *testing.T) {
	options := DefaultParseOptions.WithAllowLiteralsWithLineBreaks(true)
	malformedTokens := LexWithOptions("ignored ; long(1\n2", options)
	malformed := ParseLiteralTokensFromWithOptions(malformedTokens, 2, options)
	if malformed == nil || !diagnosticsContainMessage(malformed.ContainedDiagnostics(), "KS001", "Missing: )") {
		t.Fatalf("positional malformed diagnostics = %#v", GetContainedDiagnostics(malformed, DiagnosticsIncludeDefault))
	}

	tokens := Lex("print x=1")
	if expression := ParseExpressionTokensFrom(tokens, len(tokens)); expression != nil {
		t.Fatalf("past-end expression = %#v, want nil", expression)
	}
	emptyQuery := ParseQueryTokensFrom(tokens, len(tokens)+5)
	if emptyQuery == nil || emptyQuery.Statements().Count() != 0 || emptyQuery.EndOfQuery() != nil {
		t.Fatalf("past-end query = %#v", emptyQuery)
	}
	fromNegative := ParseQueryTokensFrom(tokens, -5)
	if fromNegative == nil || fromNegative.Statements().Count() != 1 {
		t.Fatalf("negative-start query = %#v", fromNegative)
	}
}
