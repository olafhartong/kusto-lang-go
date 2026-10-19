package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicQueryParserTokenOverloadFamilies(t *testing.T) {
	options := kustolang.DefaultParseOptions.WithAllowLiteralsWithLineBreaks(true)
	tests := []struct {
		name        string
		text        string
		parse       func([]kustolang.LexicalToken) kustolang.SyntaxElement
		parseOption func([]kustolang.LexicalToken, kustolang.ParseOptions) kustolang.SyntaxElement
		parseFrom   func([]kustolang.LexicalToken, int) kustolang.SyntaxElement
		parseAll    func([]kustolang.LexicalToken, int, kustolang.ParseOptions) kustolang.SyntaxElement
	}{
		{
			name: "literal", text: "prefix 42",
			parse: func(tokens []kustolang.LexicalToken) kustolang.SyntaxElement {
				return kustolang.ParseLiteralTokens(tokens[1:])
			},
			parseOption: func(tokens []kustolang.LexicalToken, options kustolang.ParseOptions) kustolang.SyntaxElement {
				return kustolang.ParseLiteralTokensWithOptions(tokens[1:], options)
			},
			parseFrom: func(tokens []kustolang.LexicalToken, start int) kustolang.SyntaxElement {
				return kustolang.ParseLiteralTokensFrom(tokens, start)
			},
			parseAll: func(tokens []kustolang.LexicalToken, start int, options kustolang.ParseOptions) kustolang.SyntaxElement {
				return kustolang.ParseLiteralTokensFromWithOptions(tokens, start, options)
			},
		},
		{
			name: "entity path", text: `prefix database("db").Events`,
			parse: func(tokens []kustolang.LexicalToken) kustolang.SyntaxElement {
				return kustolang.ParseEntityPathTokens(tokens[1:])
			},
			parseOption: func(tokens []kustolang.LexicalToken, options kustolang.ParseOptions) kustolang.SyntaxElement {
				return kustolang.ParseEntityPathTokensWithOptions(tokens[1:], options)
			},
			parseFrom: func(tokens []kustolang.LexicalToken, start int) kustolang.SyntaxElement {
				return kustolang.ParseEntityPathTokensFrom(tokens, start)
			},
			parseAll: func(tokens []kustolang.LexicalToken, start int, options kustolang.ParseOptions) kustolang.SyntaxElement {
				return kustolang.ParseEntityPathTokensFromWithOptions(tokens, start, options)
			},
		},
		{
			name: "entity group", text: `prefix entity_group("Members")`,
			parse: func(tokens []kustolang.LexicalToken) kustolang.SyntaxElement {
				return kustolang.ParseEntityGroupTokens(tokens[1:])
			},
			parseOption: func(tokens []kustolang.LexicalToken, options kustolang.ParseOptions) kustolang.SyntaxElement {
				return kustolang.ParseEntityGroupTokensWithOptions(tokens[1:], options)
			},
			parseFrom: func(tokens []kustolang.LexicalToken, start int) kustolang.SyntaxElement {
				return kustolang.ParseEntityGroupTokensFrom(tokens, start)
			},
			parseAll: func(tokens []kustolang.LexicalToken, start int, options kustolang.ParseOptions) kustolang.SyntaxElement {
				return kustolang.ParseEntityGroupTokensFromWithOptions(tokens, start, options)
			},
		},
		{
			name: "row schema", text: "prefix (A:long, B:string)",
			parse: func(tokens []kustolang.LexicalToken) kustolang.SyntaxElement {
				return kustolang.ParseRowSchemaTokens(tokens[1:])
			},
			parseOption: func(tokens []kustolang.LexicalToken, options kustolang.ParseOptions) kustolang.SyntaxElement {
				return kustolang.ParseRowSchemaTokensWithOptions(tokens[1:], options)
			},
			parseFrom: func(tokens []kustolang.LexicalToken, start int) kustolang.SyntaxElement {
				return kustolang.ParseRowSchemaTokensFrom(tokens, start)
			},
			parseAll: func(tokens []kustolang.LexicalToken, start int, options kustolang.ParseOptions) kustolang.SyntaxElement {
				return kustolang.ParseRowSchemaTokensFromWithOptions(tokens, start, options)
			},
		},
		{
			name: "function parameters", text: "prefix (x:long, y:string)",
			parse: func(tokens []kustolang.LexicalToken) kustolang.SyntaxElement {
				return kustolang.ParseFunctionParametersTokens(tokens[1:])
			},
			parseOption: func(tokens []kustolang.LexicalToken, options kustolang.ParseOptions) kustolang.SyntaxElement {
				return kustolang.ParseFunctionParametersTokensWithOptions(tokens[1:], options)
			},
			parseFrom: func(tokens []kustolang.LexicalToken, start int) kustolang.SyntaxElement {
				return kustolang.ParseFunctionParametersTokensFrom(tokens, start)
			},
			parseAll: func(tokens []kustolang.LexicalToken, start int, options kustolang.ParseOptions) kustolang.SyntaxElement {
				return kustolang.ParseFunctionParametersTokensFromWithOptions(tokens, start, options)
			},
		},
		{
			name: "function parameter", text: "prefix x:long",
			parse: func(tokens []kustolang.LexicalToken) kustolang.SyntaxElement {
				return kustolang.ParseFunctionParameterTokens(tokens[1:])
			},
			parseOption: func(tokens []kustolang.LexicalToken, options kustolang.ParseOptions) kustolang.SyntaxElement {
				return kustolang.ParseFunctionParameterTokensWithOptions(tokens[1:], options)
			},
			parseFrom: func(tokens []kustolang.LexicalToken, start int) kustolang.SyntaxElement {
				return kustolang.ParseFunctionParameterTokensFrom(tokens, start)
			},
			parseAll: func(tokens []kustolang.LexicalToken, start int, options kustolang.ParseOptions) kustolang.SyntaxElement {
				return kustolang.ParseFunctionParameterTokensFromWithOptions(tokens, start, options)
			},
		},
		{
			name: "function body", text: "prefix { print X=1 }",
			parse: func(tokens []kustolang.LexicalToken) kustolang.SyntaxElement {
				return kustolang.ParseFunctionBodyTokens(tokens[1:])
			},
			parseOption: func(tokens []kustolang.LexicalToken, options kustolang.ParseOptions) kustolang.SyntaxElement {
				return kustolang.ParseFunctionBodyTokensWithOptions(tokens[1:], options)
			},
			parseFrom: func(tokens []kustolang.LexicalToken, start int) kustolang.SyntaxElement {
				return kustolang.ParseFunctionBodyTokensFrom(tokens, start)
			},
			parseAll: func(tokens []kustolang.LexicalToken, start int, options kustolang.ParseOptions) kustolang.SyntaxElement {
				return kustolang.ParseFunctionBodyTokensFromWithOptions(tokens, start, options)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens := kustolang.Lex(test.text)
			results := []kustolang.SyntaxElement{
				test.parse(tokens),
				test.parseOption(tokens, options),
				test.parseFrom(tokens, 1),
				test.parseAll(tokens, 1, options),
			}
			for i, result := range results {
				if result == nil || result.IsMissing() {
					t.Fatalf("overload %d result = %#v", i, result)
				}
			}
			want := results[0].Text(kustolang.IncludeTriviaMinimal)
			for i, result := range results[1:] {
				if got := result.Text(kustolang.IncludeTriviaMinimal); got != want {
					t.Fatalf("overload %d text = %q, want %q", i+1, got, want)
				}
			}
		})
	}
}
