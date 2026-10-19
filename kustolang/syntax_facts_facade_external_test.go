package kustolang_test

import (
	"slices"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicSyntaxFactsFacadeForwardsKindFacts(t *testing.T) {
	tests := []kustolang.SyntaxKind{
		kustolang.SyntaxKindWhereKeyword,
		kustolang.SyntaxKindCommaToken,
		kustolang.SyntaxKindPlusToken,
		kustolang.SyntaxKindLongLiteralToken,
		kustolang.SyntaxKindStringKeyword,
	}

	for _, kind := range tests {
		if got, want := kustolang.SyntaxFacts.GetText(kind), kustolang.SyntaxText(kind); got != want {
			t.Errorf("GetText(%v) = %q, want %q", kind, got, want)
		}
		if got, want := kustolang.SyntaxFacts.GetCategory(kind), kustolang.SyntaxCategoryOf(kind); got != want {
			t.Errorf("GetCategory(%v) = %v, want %v", kind, got, want)
		}
		if got, want := kustolang.SyntaxFacts.IsKeyword(kind), kustolang.IsKeywordKind(kind); got != want {
			t.Errorf("IsKeyword(%v) = %t, want %t", kind, got, want)
		}
		if got, want := kustolang.SyntaxFacts.IsPunctuation(kind), kustolang.IsPunctuationKind(kind); got != want {
			t.Errorf("IsPunctuation(%v) = %t, want %t", kind, got, want)
		}
		if got, want := kustolang.SyntaxFacts.IsOperator(kind), kustolang.IsOperatorKind(kind); got != want {
			t.Errorf("IsOperator(%v) = %t, want %t", kind, got, want)
		}
		if got, want := kustolang.SyntaxFacts.IsLiteral(kind), kustolang.IsLiteralKind(kind); got != want {
			t.Errorf("IsLiteral(%v) = %t, want %t", kind, got, want)
		}
		if got, want := kustolang.SyntaxFacts.IsType(kind), kustolang.IsTypeKeyword(kind); got != want {
			t.Errorf("IsType(%v) = %t, want %t", kind, got, want)
		}
		if got, want := kustolang.SyntaxFacts.GetOperatorKind(kind), kustolang.SyntaxOperatorKind(kind); got != want {
			t.Errorf("GetOperatorKind(%v) = %v, want %v", kind, got, want)
		}
		if got, want := kustolang.SyntaxFacts.CanBeIdentifier(kind), kustolang.SyntaxKindCanBeIdentifier(kind); got != want {
			t.Errorf("CanBeIdentifier(%v) = %t, want %t", kind, got, want)
		}
	}
}

func TestPublicSyntaxFactsFacadeForwardsLookupsAndInventories(t *testing.T) {
	for _, text := range []string{"where", ",", "+", "not-a-kind"} {
		gotKind, gotOK := kustolang.SyntaxFacts.GetKind(text)
		wantKind, wantOK := kustolang.SyntaxKindFromText(text)
		if gotKind != wantKind || gotOK != wantOK {
			t.Errorf("GetKind(%q) = (%v, %t), want (%v, %t)", text, gotKind, gotOK, wantKind, wantOK)
		}
	}

	gotKind, gotOK := kustolang.SyntaxFacts.GetKindFromRange("xxwhereyy", 2, 5)
	wantKind, wantOK := kustolang.SyntaxKindFromTextRange("xxwhereyy", 2, 5)
	if gotKind != wantKind || gotOK != wantOK {
		t.Fatalf("GetKindFromRange() = (%v, %t), want (%v, %t)", gotKind, gotOK, wantKind, wantOK)
	}
	if _, ok := kustolang.SyntaxFacts.GetKindFromRange("where", -1, 5); ok {
		t.Fatal("GetKindFromRange accepted an invalid range")
	}

	keywordCategory := kustolang.SyntaxCategoryOf(kustolang.SyntaxKindWhereKeyword)
	if got, want := kustolang.SyntaxFacts.GetKinds(keywordCategory), kustolang.SyntaxKinds(keywordCategory); !slices.Equal(got, want) {
		t.Fatal("GetKinds differs from SyntaxKinds")
	}
	if got, want := kustolang.SyntaxFacts.GetKindsWithFixedText(), kustolang.SyntaxKindsWithFixedText(); !slices.Equal(got, want) {
		t.Fatal("GetKindsWithFixedText differs from SyntaxKindsWithFixedText")
	}
	if got, want := kustolang.SyntaxFacts.Keywords(), kustolang.SyntaxKeywordTexts(); !slices.Equal(got, want) {
		t.Fatal("Keywords differs from SyntaxKeywordTexts")
	}
	if got, want := kustolang.SyntaxFacts.Punctuation(), kustolang.SyntaxPunctuationTexts(); !slices.Equal(got, want) {
		t.Fatal("Punctuation differs from SyntaxPunctuationTexts")
	}
	if got, want := kustolang.SyntaxFacts.Operators(), kustolang.SyntaxOperatorTexts(); !slices.Equal(got, want) {
		t.Fatal("Operators differs from SyntaxOperatorTexts")
	}
}

func TestPublicSyntaxFactsFacadeDistinguishesKeywordOverloads(t *testing.T) {
	for _, text := range []string{"where", "day", "not-a-keyword"} {
		if got, want := kustolang.SyntaxFacts.IsKeywordText(text), kustolang.IsKeywordText(text); got != want {
			t.Errorf("IsKeywordText(%q) = %t, want %t", text, got, want)
		}
		if got, want := kustolang.SyntaxFacts.IsKeywordThatCanBeIdentifier(text), kustolang.IsKeywordThatCanBeIdentifier(text); got != want {
			t.Errorf("IsKeywordThatCanBeIdentifier(%q) = %t, want %t", text, got, want)
		}
	}
}
