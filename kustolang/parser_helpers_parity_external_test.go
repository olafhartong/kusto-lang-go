package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicSyntaxTextHelpersControlEndToken(t *testing.T) {
	anyToken := kustolang.AnySyntaxTokenParser()

	if got := kustolang.ParseAllSyntax(anyToken, ""); len(got) != 0 {
		t.Fatalf("ParseAllSyntax(empty) = %#v, want no tokens", got)
	}
	if got := kustolang.ParseAllSyntax(anyToken, "alpha"); len(got) != 1 || got[0].Kind() != kustolang.SyntaxKindIdentifierToken {
		t.Fatalf("ParseAllSyntax(non-empty) = %#v, want identifier only", got)
	}
	if got := kustolang.ParseAllSyntax(anyToken, "alpha "); len(got) != 2 || got[1].Kind() != kustolang.SyntaxKindEndOfTextToken || got[1].Trivia() != " " {
		t.Fatalf("ParseAllSyntax(trailing trivia) = %#v, want trivia-bearing EOF", got)
	}
	if got := kustolang.ParseAllSyntax(anyToken, "alpha", true); len(got) != 2 || got[1].Kind() != kustolang.SyntaxKindEndOfTextToken {
		t.Fatalf("ParseAllSyntax(..., true) = %#v, want identifier and EOF", got)
	}

	if got := kustolang.ParseFirstSyntax(anyToken, ""); got != nil {
		t.Fatalf("ParseFirstSyntax(empty) = %#v, want nil", got)
	}
	if got := kustolang.ParseFirstSyntax(anyToken, "", true); got == nil || got.Kind() != kustolang.SyntaxKindEndOfTextToken {
		t.Fatalf("ParseFirstSyntax(..., true) = %#v, want EOF", got)
	}
	if got := kustolang.ScanFirstSyntax(kustolang.EndOfTextSyntaxParser(), "alpha", true); got >= 0 {
		t.Fatalf("ScanFirstSyntax malformed start = %d, want failure", got)
	}

	identifiers := kustolang.ParseAllSyntax(
		kustolang.SyntaxTokenKindParser(kustolang.SyntaxKindIdentifierToken),
		"alpha | beta",
	)
	if len(identifiers) != 1 || identifiers[0].RawText() != "alpha" {
		t.Fatalf("ParseAllSyntax malformed tail = %#v, want successful prefix", identifiers)
	}
}

func TestPublicRepetitionListSuccessFailureAndOwnership(t *testing.T) {
	element := publicByteValue('A', "a")
	one := kustolang.OneOrMoreList(element)

	first := one.ParseValue(kustolang.NewTextSource("AAAB"), 0)
	if first.Length != 3 || !reflect.DeepEqual(first.Value, []string{"a", "a", "a"}) {
		t.Fatalf("OneOrMoreList success = %#v", first)
	}
	if failed := one.ParseValue(kustolang.NewTextSource("B"), 0); failed.Length >= 0 || failed.Value != nil {
		t.Fatalf("OneOrMoreList failure = %#v", failed)
	}
	second := one.ParseValue(kustolang.NewTextSource("AA"), 0)
	first.Value[0] = "changed"
	if !reflect.DeepEqual(second.Value, []string{"a", "a"}) {
		t.Fatalf("OneOrMoreList results share storage: first=%#v second=%#v", first.Value, second.Value)
	}

	zero := kustolang.ZeroOrMoreList(element)
	empty := zero.ParseValue(kustolang.NewTextSource("B"), 0)
	if empty.Length != 0 || empty.Value == nil || len(empty.Value) != 0 {
		t.Fatalf("ZeroOrMoreList empty success = %#v", empty)
	}
	populated := zero.ParseValue(kustolang.NewTextSource("AA"), 0)
	if populated.Length != 2 || !reflect.DeepEqual(populated.Value, []string{"a", "a"}) {
		t.Fatalf("ZeroOrMoreList populated success = %#v", populated)
	}
	again := zero.ParseValue(kustolang.NewTextSource("A"), 0)
	populated.Value[0] = "changed"
	if !reflect.DeepEqual(again.Value, []string{"a"}) {
		t.Fatalf("ZeroOrMoreList results share storage: populated=%#v again=%#v", populated.Value, again.Value)
	}
}

func TestPublicTokenParserDefaultStartAndAppendTokens(t *testing.T) {
	if token := kustolang.ParseToken("alpha beta"); token == nil || token.Text != "alpha" {
		t.Fatalf("ParseToken default start = %#v", token)
	}
	if token := kustolang.ParseToken("alpha beta", len("alpha")); token == nil || token.Text != "beta" || token.Trivia != " " {
		t.Fatalf("ParseToken explicit start = %#v", token)
	}
	if token := kustolang.ParseToken(""); token == nil || token.Kind != kustolang.SyntaxKindEndOfTextToken {
		t.Fatalf("ParseToken empty = %#v, want EOF", token)
	}

	prefix := kustolang.NewLexicalToken(kustolang.SyntaxKindIdentifierToken, "", "prefix")
	tokens := kustolang.AppendTokens([]kustolang.LexicalToken{prefix}, "alpha")
	if len(tokens) != 3 || !reflect.DeepEqual(tokens[0], prefix) || tokens[1].Text != "alpha" || tokens[2].Kind != kustolang.SyntaxKindEndOfTextToken {
		t.Fatalf("AppendTokens default = %#v", tokens)
	}

	withoutEOF := kustolang.DefaultParseOptions.WithAlwaysProduceEndTokens(false)
	tokens = kustolang.AppendTokens(tokens[:1], "beta", withoutEOF)
	if len(tokens) != 2 || !reflect.DeepEqual(tokens[0], prefix) || tokens[1].Text != "beta" {
		t.Fatalf("AppendTokens explicit options = %#v", tokens)
	}
	if got := kustolang.AppendTokens(tokens, ""); !reflect.DeepEqual(got[:len(tokens)], tokens) || len(got) != len(tokens)+1 {
		t.Fatalf("AppendTokens empty text did not preserve prefix: before=%#v after=%#v", tokens, got)
	}
}
