package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicSyntaxTokenFactoriesPreserveSourceForms(t *testing.T) {
	diagnostic := kustolang.NewDiagnostic("KS999", "factory diagnostic")

	identifier := kustolang.NewIdentifierTokenWithValueText(" ", "['Account Name']", "Account Name", diagnostic)
	if identifier.Kind() != kustolang.SyntaxKindIdentifierToken || identifier.Trivia() != " " || identifier.RawText() != "['Account Name']" || identifier.ValueText() != "Account Name" {
		t.Fatalf("identifier = kind:%v trivia:%q text:%q value:%q", identifier.Kind(), identifier.Trivia(), identifier.RawText(), identifier.ValueText())
	}
	if got := identifier.Diagnostics(); len(got) != 1 || got[0].Code != "KS999" {
		t.Fatalf("identifier diagnostics = %#v", got)
	}

	keyword, err := kustolang.NewKeywordTokenWithValueText("\n", kustolang.SyntaxKindProjectKeyword, "PROJECT")
	if err != nil || keyword.RawText() != "project" || keyword.ValueText() != "PROJECT" || keyword.IsLiteral() {
		t.Fatalf("keyword = %#v, err = %v", keyword, err)
	}

	punctuation, err := kustolang.NewPunctuationToken("", kustolang.SyntaxKindOpenParenToken)
	if err != nil || punctuation.RawText() != "(" {
		t.Fatalf("punctuation = %#v, err = %v", punctuation, err)
	}

	operator, err := kustolang.NewOperatorToken(" ", kustolang.SyntaxKindPlusToken)
	if err != nil || operator.RawText() != "+" || operator.Trivia() != " " {
		t.Fatalf("operator = %#v, err = %v", operator, err)
	}

	literal, err := kustolang.NewLiteralToken("", `datetime("2024-01-02")`, kustolang.SyntaxKindDateTimeLiteralToken)
	if err != nil || !literal.IsLiteral() || literal.Prefix() != "datetime" || literal.ValueText() != "2024-01-02" {
		t.Fatalf("literal = text:%q value:%q prefix:%q literal:%t, err = %v", literal.RawText(), literal.ValueText(), literal.Prefix(), literal.IsLiteral(), err)
	}

	other := kustolang.NewOtherToken("", "!directive", kustolang.SyntaxKindInputTextToken)
	if other.RawText() != "!directive" || other.ValueText() != "!directive" {
		t.Fatalf("other = %#v", other)
	}
}

func TestPublicSyntaxTokenFactoriesRejectWrongCategories(t *testing.T) {
	tests := []struct {
		name string
		make func() error
	}{
		{"keyword", func() error { _, err := kustolang.NewKeywordToken("", kustolang.SyntaxKindPlusToken); return err }},
		{"punctuation", func() error {
			_, err := kustolang.NewPunctuationToken("", kustolang.SyntaxKindWhereKeyword)
			return err
		}},
		{"operator", func() error { _, err := kustolang.NewOperatorToken("", kustolang.SyntaxKindCommaToken); return err }},
		{"literal", func() error {
			_, err := kustolang.NewLiteralToken("", "where", kustolang.SyntaxKindWhereKeyword)
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.make(); err == nil {
				t.Fatal("expected category error")
			}
		})
	}
}

func TestPublicMissingTokenFactoryPreservesTriviaAndRecoveryDiagnostics(t *testing.T) {
	missing := kustolang.NewMissingTokenWithTrivia(" /* recovery */", kustolang.SyntaxKindCloseParenToken,
		kustolang.NewDiagnostic("KS005", "Expected: )"),
		kustolang.NewDiagnostic("KS006", "Malformed expression"),
	)
	if !missing.IsMissing() || missing.Width() != 0 || missing.Trivia() != " /* recovery */" || missing.FullWidth() != len(" /* recovery */") {
		t.Fatalf("missing = missing:%t width:%d trivia:%q full:%d", missing.IsMissing(), missing.Width(), missing.Trivia(), missing.FullWidth())
	}
	if got := missing.Diagnostics(); len(got) != 2 || got[0].Code != "KS005" || got[1].Code != "KS006" {
		t.Fatalf("missing diagnostics = %#v", got)
	}

	clone, ok := missing.Clone(false).(*kustolang.SyntaxToken)
	if !ok || !clone.IsMissing() || clone.Trivia() != missing.Trivia() || len(clone.Diagnostics()) != 0 {
		t.Fatalf("clone = %#v", clone)
	}
}
