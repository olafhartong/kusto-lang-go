package kustolang

import (
	"reflect"
	"strings"
	"testing"
)

func TestQueryParserTokenOwnership(t *testing.T) {
	tokens := Lex("x + 1")
	parser := newQueryParserStateWithOptions(tokens, DefaultParseOptions)
	tokens[0].Text = "changed"
	if got := parser.parseExpression().Text(IncludeTriviaAll); got != "x + 1" {
		t.Fatalf("caller changed parser tokens: %q", got)
	}
	// Verify that the private ownership transfer actually avoids a second buffer.
	owned := Lex("x + 1")
	transferred := newQueryParserStateWithOwnedTokens(owned, DefaultParseOptions, CancellationToken{})
	if &transferred.tokens[0] != &owned[0] {
		t.Fatal("owned tokens were copied")
	}
}

func TestQueryParserOwnedTextMatchesPublicTokens(t *testing.T) {
	parsers := []struct {
		name   string
		text   func(string) SyntaxElement
		tokens func([]LexicalToken) SyntaxElement
		inputs []string
	}{
		{"query", func(s string) SyntaxElement { return ParseQuery(s) }, func(ts []LexicalToken) SyntaxElement { return ParseQueryTokens(ts) }, []string{"", "// 😀\n let x=1; print y=x", "T | project x,", "set querytrace; T | unknown x", "restrict access to (a,,b); T"}},
		{"expression", func(s string) SyntaxElement { return ParseExpression(s) }, func(ts []LexicalToken) SyntaxElement { return ParseExpressionTokens(ts) }, []string{"f(1,,)", "f(a 1)", "// 😀\n x + 'é'", "inline_external_table(x:long) kind=()"}},
		{"literal", func(s string) SyntaxElement { return ParseLiteral(s) }, func(ts []LexicalToken) SyntaxElement { return ParseLiteralTokens(ts) }, []string{"'unterminated", "dynamic([1,])", "long(bad)"}},
		{"entity-group", func(s string) SyntaxElement { return ParseEntityGroup(s) }, func(ts []LexicalToken) SyntaxElement { return ParseEntityGroupTokens(ts) }, []string{"entity_group [T, U]", "entity_group [T,"}},
		{"entity-path", func(s string) SyntaxElement { return ParseEntityPath(s) }, func(ts []LexicalToken) SyntaxElement { return ParseEntityPathTokens(ts) }, []string{"database('db').T", "database("}},
		{"schema", func(s string) SyntaxElement { return ParseRowSchema(s) }, func(ts []LexicalToken) SyntaxElement { return ParseRowSchemaTokens(ts) }, []string{"(x:long, y:string)", "(x:long,,)"}},
		{"parameters", func(s string) SyntaxElement { return ParseFunctionParameters(s) }, func(ts []LexicalToken) SyntaxElement { return ParseFunctionParametersTokens(ts) }, []string{"(x:long, y:string='é')", "(x:long,,)"}},
		{"parameter", func(s string) SyntaxElement { return ParseFunctionParameter(s) }, func(ts []LexicalToken) SyntaxElement { return ParseFunctionParameterTokens(ts) }, []string{"x:long", "x:"}},
	}
	for _, parser := range parsers {
		t.Run(parser.name, func(t *testing.T) {
			for _, text := range parser.inputs {
				tokens := Lex(text)
				before := append([]LexicalToken(nil), tokens...)
				fromTokens, fromText := parser.tokens(tokens), parser.text(text)
				if !reflect.DeepEqual(tokens, before) {
					t.Fatalf("public parser modified input for %q", text)
				}
				want, got := GetDescendantsOrSelf[SyntaxElement](fromTokens), GetDescendantsOrSelf[SyntaxElement](fromText)
				if len(want) != len(got) {
					t.Fatalf("syntax shape changed for %q", text)
				}
				for i, node := range want {
					other := got[i]
					if node.Kind() != other.Kind() || node.Text(IncludeTriviaAll) != other.Text(IncludeTriviaAll) || node.TextStart() != other.TextStart() || node.UTF16TextStart() != other.UTF16TextStart() || node.Width() != other.Width() || node.UTF16Width() != other.UTF16Width() || !reflect.DeepEqual(node.ContainedDiagnostics(), other.ContainedDiagnostics()) {
						t.Fatalf("syntax/diagnostics changed for %q at node %d", text, i)
					}
				}
			}
		})
	}
}

func TestQueryParserOwnedTokensObserveCancellation(t *testing.T) {
	tokens := Lex(strings.Repeat("print x=1; ", 1000))
	canceled := NewCancellationTokenFunc(func() bool { return true })
	assertOperationCanceledPanic(t, func() { newQueryParserStateWithOwnedTokens(tokens, DefaultParseOptions, canceled) })
	checks := 0
	token := NewCancellationTokenFunc(func() bool { checks++; return checks >= 8 })
	assertOperationCanceledPanic(t, func() { parseQueryWithParser(newQueryParserStateWithOwnedTokens(tokens, DefaultParseOptions, token)) })
}
