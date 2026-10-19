package kustolang

import (
	"reflect"
	"testing"
)

func TestCommandGrammarRuntimeCommandLimiterStopsAtLineStartDot(t *testing.T) {
	tokens := Lex(".show first\n  .show second")
	source := NewArraySource(tokens)
	limiter := CommandGrammarCommandLimiterParser()

	if got := limiter.Scan(source, 0); got != 3 {
		t.Fatalf("limiter length = %d, want 3", got)
	}
	if got := limiter.Scan(NewArraySource(Lex(".show first .show second")), 0); got != 7 {
		t.Fatalf("same-line limiter length = %d, want 7", got)
	}
}

func TestCommandGrammarRuntimePossibleCommandStart(t *testing.T) {
	tests := []struct {
		name  string
		token LexicalToken
		first bool
		want  bool
	}{
		{name: "non dot", token: NewLexicalToken(SyntaxKindIdentifierToken, "\n", "show"), want: false},
		{name: "first no trivia", token: NewLexicalToken(SyntaxKindDotToken, "", "."), first: true, want: true},
		{name: "not first no trivia", token: NewLexicalToken(SyntaxKindDotToken, "", "."), want: false},
		{name: "first whitespace", token: NewLexicalToken(SyntaxKindDotToken, " \t", "."), first: true, want: true},
		{name: "new line whitespace", token: NewLexicalToken(SyntaxKindDotToken, "x\r\n \t", "."), want: true},
		{name: "new line non whitespace", token: NewLexicalToken(SyntaxKindDotToken, "\n x", "."), want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := CommandGrammarIsPossibleStartOfCommand(test.token, test.first); got != test.want {
				t.Fatalf("possible start = %t, want %t", got, test.want)
			}
		})
	}
}

func TestCommandGrammarRuntimeBadAndMissingCommand(t *testing.T) {
	parser := CommandGrammarBadCommandParser()
	result := parser.ParseValue(NewArraySource(Lex(".")), 0)
	bad, ok := result.Value.(*BadCommand)
	if result.Length != 1 || !ok || parser.Tag() != "<bad-command>" {
		t.Fatalf("bad command = %#v, length = %d, tag = %q", result.Value, result.Length, parser.Tag())
	}
	if bad.DotToken() == nil || bad.DotToken().IsMissing() || !hasDiagnosticCode(bad.Diagnostics(), "KS300") {
		t.Fatalf("bad command dot/diagnostics = %#v/%#v", bad.DotToken(), bad.Diagnostics())
	}
	if failed := parser.ParseValue(NewArraySource(Lex("show")), 0); failed.Length >= 0 {
		t.Fatalf("non-dot parse = %#v", failed)
	}

	statement, ok := CommandGrammarCreateMissingCommandStatement(nil, 42).(*ExpressionStatement)
	if !ok {
		t.Fatalf("missing statement = %T", CommandGrammarCreateMissingCommandStatement(nil, 42))
	}
	missing, ok := statement.Expression().(*BadCommand)
	if !ok || missing.DotToken() == nil || !missing.DotToken().IsMissing() || !hasDiagnosticCode(missing.Diagnostics(), "KS300") {
		t.Fatalf("missing command = %#v", statement.Expression())
	}
}

func TestCommandGrammarRuntimeFailureTermsAndSyntaxSelection(t *testing.T) {
	parser := First(
		Match[LexicalToken](func(token LexicalToken) bool { return token.Text == "a" }).WithTag(" zeta | alpha "),
		Match[LexicalToken](func(token LexicalToken) bool { return token.Text == "b" }).WithTag("alpha | beta"),
	)
	if got := CommandGrammarExpectedTerms(parser); !reflect.DeepEqual(got, []string{"alpha", "beta", "zeta"}) {
		t.Fatalf("expected terms = %#v", got)
	}

	output := []any{"prefix"}
	CommandGrammarOnFailure(parser, &output)
	missing, ok := output[1].(*SyntaxToken)
	if len(output) != 2 || !ok || !missing.IsMissing() || missing.Kind() != SyntaxKindIdentifierToken {
		t.Fatalf("failure output = %#v", output)
	}
	if diagnostics := missing.Diagnostics(); len(diagnostics) != 1 || diagnostics[0].Description != "Expected one of: alpha, beta, zeta" {
		t.Fatalf("failure diagnostics = %#v", missing.Diagnostics())
	}

	earlier := MustNewBadCommandTyped(createOmittedToken(SyntaxKindDotToken), DiagnosticMissingCommand().WithLocation(1, 0))
	later := MustNewBadCommandTyped(createOmittedToken(SyntaxKindDotToken), DiagnosticMissingCommand().WithLocation(2, 0))
	clean := MustNewBadCommandTyped(createOmittedToken(SyntaxKindDotToken))
	if !CommandGrammarIsBetterSyntax(later, earlier) || CommandGrammarIsBetterSyntax(earlier, later) {
		t.Fatal("diagnostic position ordering was not preserved")
	}
	if !CommandGrammarIsBetterSyntax(clean, earlier) || CommandGrammarIsBetterSyntax(earlier, clean) {
		t.Fatal("diagnostic-free syntax was not preferred")
	}
}
