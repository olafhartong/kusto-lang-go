package kustolang

import (
	"strings"
	"testing"
)

// Ported from KustoCompleter.GetGrammarSearchContext in
// src/Kusto.Language/Editor/Kusto/KustoCompleter.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func TestKustoCompleterExtendedQueryGrammarContext(t *testing.T) {
	grammar := QueryGrammarFrom(DefaultGlobalState)
	tests := []struct {
		name      string
		text      string
		position  int
		startText string
		parser    Parser[LexicalToken]
		malformed bool
	}{
		{
			name:      "following pipe operator",
			text:      "T | where X | project ",
			position:  len("T | where X | project "),
			startText: "project",
			parser:    grammar.FollowingPipeElementExpression,
			malformed: false,
		},
		{
			name:      "trailing pipe subexpression",
			text:      "T | where X | project Y | ",
			position:  len("T | where X | project Y | "),
			startText: "project",
			parser:    grammar.PipeSubExpression,
			malformed: true,
		},
		{
			name:      "function body",
			text:      "let F=(){ }; F()",
			position:  strings.Index("let F=(){ }; F()", "{") + 1,
			startText: "{",
			parser:    grammar.FunctionBody,
			malformed: false,
		},
		{
			name:      "statement list",
			text:      "let Existing=1; let Pending =",
			position:  len("let Existing=1; let Pending ="),
			startText: "let Pending",
			parser:    grammar.StatementList,
			malformed: true,
		},
		{
			name:      "function arguments",
			text:      "print x=strcat('a', )",
			position:  strings.Index("print x=strcat('a', )", ")"),
			startText: "strcat",
			parser:    grammar.UnnamedExpression,
			malformed: true,
		},
		{
			name:      "datatable literals",
			text:      "datatable(x:string)[1, ]",
			position:  strings.Index("datatable(x:string)[1, ]", "]"),
			startText: "1",
			parser:    grammar.LiteralList,
			malformed: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := Parse(test.text)
			if malformed := len(code.GetSyntaxDiagnostics()) != 0; malformed != test.malformed {
				t.Fatalf("syntax diagnostics present = %t, want %t: %#v", malformed, test.malformed, code.GetSyntaxDiagnostics())
			}
			completer := newKustoCompleter(code, CompletionOptionsDefault.WithIncludeExtendedSyntax(true), CancellationToken{})
			parser, startIndex, matchIndex, ok := completer.extendedQueryGrammarContext(test.position)
			if !ok {
				t.Fatal("extended query grammar context was not found")
			}
			if parser != test.parser {
				t.Fatalf("parser = %T, want %T", parser, test.parser)
			}
			wantStart := code.GetTokenIndex(strings.Index(test.text, test.startText))
			if startIndex != wantStart {
				t.Fatalf("start token = %d, want %d (%q)", startIndex, wantStart, test.startText)
			}
			if wantMatch := code.GetTokenIndex(test.position); matchIndex != wantMatch {
				t.Fatalf("match token = %d, want %d", matchIndex, wantMatch)
			}
		})
	}
}
