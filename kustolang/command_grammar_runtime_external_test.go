package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicCommandGrammarRuntimeParserFacade(t *testing.T) {
	badResult := kustolang.BadCommandParser().ParseValue(kustolang.NewArraySource(kustolang.Lex(".")), 0)
	bad, ok := badResult.Value.(*kustolang.BadCommand)
	if badResult.Length != 1 || !ok || bad.DotToken() == nil || bad.DotToken().IsMissing() {
		t.Fatalf("bad command = %#v, length = %d", badResult.Value, badResult.Length)
	}

	statement, ok := kustolang.CreateMissingCommandStatement(nil, 0).(*kustolang.ExpressionStatement)
	if !ok {
		t.Fatalf("missing statement = %T", kustolang.CreateMissingCommandStatement(nil, 0))
	}
	missing, ok := statement.Expression().(*kustolang.BadCommand)
	if !ok || missing.DotToken() == nil || !missing.DotToken().IsMissing() {
		t.Fatalf("missing command = %#v", statement.Expression())
	}

	limiter := kustolang.CommandLimiterParser()
	if got := limiter.Scan(kustolang.NewArraySource(kustolang.Lex(".show one\n.show two")), 0); got != 3 {
		t.Fatalf("limiter length = %d, want 3", got)
	}
}

func TestPublicCommandGrammarRuntimeFailureFacade(t *testing.T) {
	first := kustolang.Match[kustolang.LexicalToken](func(token kustolang.LexicalToken) bool { return token.Text == "a" }).WithTag("beta | alpha")
	second := kustolang.Match[kustolang.LexicalToken](func(token kustolang.LexicalToken) bool { return token.Text == "b" }).WithTag("gamma")
	parser := kustolang.First(first, second)

	if got := kustolang.CommandParserExpectedTerms(parser); !reflect.DeepEqual(got, []string{"alpha", "beta", "gamma"}) {
		t.Fatalf("expected terms = %#v", got)
	}
	if got := kustolang.CommandGrammarExpectedTerms(parser); !reflect.DeepEqual(got, []string{"alpha", "beta", "gamma"}) {
		t.Fatalf("owner-qualified expected terms = %#v", got)
	}

	var output []any
	kustolang.CommandParserOnFailure(parser, &output)
	missing, ok := output[0].(*kustolang.SyntaxToken)
	if len(output) != 1 || !ok || !missing.IsMissing() {
		t.Fatalf("failure output = %#v", output)
	}

	firstDot := kustolang.NewLexicalToken(kustolang.SyntaxKindDotToken, "", ".")
	lineDot := kustolang.NewLexicalToken(kustolang.SyntaxKindDotToken, "\n  ", ".")
	if !kustolang.IsPossibleStartOfCommand(firstDot, true) || !kustolang.CommandGrammarIsPossibleStartOfCommand(lineDot) {
		t.Fatal("possible-command-start facade returned false")
	}
}

func TestPublicCommandGrammarRuntimeOwnerQualifiedAliases(t *testing.T) {
	if kustolang.CommandGrammarCommandLimiterParser() == nil || kustolang.CommandGrammarBadCommandParser() == nil {
		t.Fatal("owner-qualified parser facade returned nil")
	}
	if _, ok := kustolang.CommandGrammarCreateMissingCommandStatement(nil, 0).(*kustolang.ExpressionStatement); !ok {
		t.Fatal("owner-qualified missing statement has the wrong type")
	}

	clean := kustolang.MustNewBadCommandTyped(kustolang.NewMissingToken(kustolang.SyntaxKindDotToken, kustolang.Diagnostic{}))
	bad := kustolang.MustNewBadCommandTyped(
		kustolang.NewMissingToken(kustolang.SyntaxKindDotToken, kustolang.Diagnostic{}),
		kustolang.DiagnosticMissingCommand(),
	)
	if !kustolang.CommandGrammarIsBetterSyntax(clean, bad) || !kustolang.IsBetterCommandSyntax(clean, bad) {
		t.Fatal("syntax selection facade did not prefer diagnostic-free syntax")
	}
}
