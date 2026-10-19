package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicIsParentVisitorAcrossConcreteParsers(t *testing.T) {
	identifier := kustolang.NewPredicateMatchParser(func(token kustolang.LexicalToken) bool {
		return token.Kind == kustolang.SyntaxKindIdentifierToken
	})
	comma := kustolang.NewPredicateMatchParser(func(token kustolang.LexicalToken) bool {
		return token.Kind == kustolang.SyntaxKindCommaToken
	})
	sequence := kustolang.NewSequenceParser([]kustolang.Parser[kustolang.LexicalToken]{identifier, comma})
	repeated := kustolang.NewZeroOrMoreParser[kustolang.LexicalToken](identifier)
	fails := kustolang.NewFailsParser[kustolang.LexicalToken](identifier)

	tests := []struct {
		name   string
		parent kustolang.Parser[kustolang.LexicalToken]
		child  kustolang.Parser[kustolang.LexicalToken]
		want   bool
	}{
		{name: "sequence child", parent: sequence, child: comma, want: true},
		{name: "sequence unrelated", parent: sequence, child: repeated},
		{name: "repetition child", parent: repeated, child: identifier, want: true},
		{name: "fails pattern pinned behavior", parent: fails, child: identifier},
		{name: "fails unrelated pinned behavior", parent: fails, child: comma, want: true},
		{name: "nil parent", child: identifier},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := kustolang.AcceptParserVisitorArgResult(test.parent, kustolang.IsParentVisitorInstance, test.child)
			if got != test.want || got != kustolang.IsParserParent(test.parent, test.child) {
				t.Fatalf("parent result = visitor:%t helper:%t, want %t", got, kustolang.IsParserParent(test.parent, test.child), test.want)
			}
		})
	}
}

func TestPublicIsParentVisitorAfterMalformedParse(t *testing.T) {
	identifier := kustolang.NewPredicateMatchParser(func(token kustolang.LexicalToken) bool {
		return token.Kind == kustolang.SyntaxKindIdentifierToken
	})
	comma := kustolang.NewPredicateMatchParser(func(token kustolang.LexicalToken) bool {
		return token.Kind == kustolang.SyntaxKindCommaToken
	})
	sequence := kustolang.NewSequenceParser([]kustolang.Parser[kustolang.LexicalToken]{identifier, comma})
	tokens := kustolang.ParseTokens("OnlyIdentifier")
	output := []any{}
	if length := sequence.ParseList(kustolang.NewArraySource(tokens), 0, &output, 0); length >= 0 {
		t.Fatalf("malformed sequence length = %d, want failure", length)
	}
	if !kustolang.AcceptParserVisitorArgResult(sequence, kustolang.IsParentVisitorInstance, kustolang.Parser[kustolang.LexicalToken](comma)) {
		t.Fatal("failed parse changed parser parent relationship")
	}
}
