package kustolang

import (
	"reflect"
	"testing"
)

func TestCommandParserExpectedTermsCompletionAnnotationPrecedesParserTag(t *testing.T) {
	parser := MatchByte('a').WithTag("parser-tag").WithAnnotations([]any{
		"ignored",
		NewCompletionItem(CompletionKindSyntax, "annotation-tag"),
		NewCompletionItem(CompletionKindSyntax, "later-annotation"),
	})

	assertCommandTags(t, parser, []string{"annotation-tag"})
}

func TestCommandParserExpectedTermsSplitsTrimsDeduplicatesAndSorts(t *testing.T) {
	parser := First(
		MatchByte('a').WithTag(" zeta | alpha |  | beta "),
		MatchByte('b').WithTag("beta|alpha"),
	)

	assertCommandTags(t, parser, []string{"alpha", "beta", "zeta"})
}

func TestCommandParserExpectedTermsFallsBackToToken(t *testing.T) {
	assertCommandTags(t, MatchByte('a'), []string{"<token>"})
}

func TestCommandParserExpectedTermsValueHidesLiteralTokenKinds(t *testing.T) {
	rules := NewPredefinedRuleParsers()

	assertCommandTags(t, rules.Value, []string{"<literal>"})
}

func TestCommandParserExpectedTermsSequenceUsesFirstParser(t *testing.T) {
	parser := Sequence(
		MatchByte('a').WithTag("first"),
		MatchByte('b').WithTag("second"),
	)

	assertCommandTags(t, parser, []string{"first"})
}

func TestCommandParserExpectedTermsAlternationsUseAllParsers(t *testing.T) {
	tests := []struct {
		name   string
		parser Parser[byte]
	}{
		{name: "first", parser: First(MatchByte('a').WithTag("alpha"), MatchByte('b').WithTag("beta"))},
		{name: "best", parser: Best(MatchByte('a').WithTag("alpha"), MatchByte('b').WithTag("beta"))},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertCommandTags(t, test.parser, []string{"alpha", "beta"})
		})
	}
}

func TestCommandParserExpectedTermsPlainIfUsesTestAndIfValueUsesBody(t *testing.T) {
	test := MatchByte('t').WithTag("test")
	body := MatchValue[byte, string](func(value byte) bool { return value == 'b' }, func(byte) string { return "body" }).WithTag("body")

	assertCommandTags(t, If(test, body), []string{"test"})
	assertCommandTags(t, IfValue(test, asResultParser[byte, string](body)), []string{"body"})
}

func assertCommandTags[TInput any](t *testing.T, parser Parser[TInput], want []string) {
	t.Helper()
	if got := CommandParserExpectedTerms(parser); !reflect.DeepEqual(got, want) {
		t.Fatalf("CommandParserExpectedTerms() = %v, want %v", got, want)
	}
}
