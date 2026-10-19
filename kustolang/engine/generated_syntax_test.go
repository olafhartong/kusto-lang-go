package kustolang

import (
	"slices"
	"testing"
)

func TestGeneratedSyntaxKindCoverage(t *testing.T) {
	if int(SyntaxKindDirective) != 604 {
		t.Fatalf("SyntaxKindDirective = %d, want upstream index 604", SyntaxKindDirective)
	}
	if len(syntaxFacts) != 605 {
		t.Fatalf("syntaxFacts count = %d, want 605", len(syntaxFacts))
	}
	if SyntaxText(SyntaxKindMakeSeriesOperator) != "" || SyntaxCategoryOf(SyntaxKindMakeSeriesOperator) != SyntaxCategoryNode {
		t.Fatalf("MakeSeriesOperator facts = text %q category %s", SyntaxText(SyntaxKindMakeSeriesOperator), SyntaxCategoryOf(SyntaxKindMakeSeriesOperator))
	}
	if SyntaxText(SyntaxKindContains_CsKeyword) != "contains_cs" {
		t.Fatalf("Contains_CsKeyword text = %q", SyntaxText(SyntaxKindContains_CsKeyword))
	}
	if kind, ok := SyntaxKindFromTextRange("xxprojectyy", 2, 7); !ok || kind != SyntaxKindProjectKeyword {
		t.Fatalf("range text lookup = %v %v, want ProjectKeyword true", kind, ok)
	}
	if kind, ok := SyntaxKindFromText("PROJECT"); ok || kind != SyntaxKindNone {
		t.Fatalf("uppercase text lookup = %v %v, want None false", kind, ok)
	}
	if kind, ok := SyntaxKindFromTextRange("project", 4, 99); ok || kind != SyntaxKindNone {
		t.Fatalf("invalid range text lookup = %v %v, want None false", kind, ok)
	}
	if !SyntaxKindCanBeIdentifier(SyntaxKindAccessKeyword) {
		t.Fatal("AccessKeyword should be allowed as identifier per upstream facts")
	}
	if !IsTypeKeyword(SyntaxKindUuidKeyword) {
		t.Fatal("UuidKeyword should be a type keyword per upstream facts")
	}
	if !IsOperatorKind(SyntaxKindPlusToken) || IsOperatorKind(SyntaxKindAddExpression) {
		t.Fatalf("operator category helpers returned unexpected values")
	}
	if SyntaxOperatorKind(SyntaxKindPlusToken) != OperatorKindAdd ||
		SyntaxOperatorKind(SyntaxKindLessThanGreaterThanToken) != OperatorKindNotEqual ||
		SyntaxOperatorKind(SyntaxKindContains_CsKeyword) != OperatorKindContainsCs ||
		SyntaxOperatorKind(SyntaxKindInCsKeyword) != OperatorKindInCs {
		t.Fatalf("operator-kind helper returned unexpected token/keyword values")
	}
	if SyntaxOperatorKind(SyntaxKindAddExpression) != OperatorKindNone ||
		SyntaxOperatorKind(SyntaxKindIdentifierToken) != OperatorKindNone {
		t.Fatalf("operator-kind helper should return None for non-operator-kind syntax")
	}
	if !IsLiteralKind(SyntaxKindStringLiteralToken) || !IsLiteralKind(SyntaxKindLongLiteralToken) || IsLiteralKind(SyntaxKindStringLiteralExpression) {
		t.Fatalf("literal category helpers returned unexpected values")
	}
	if !IsKeywordText("project") || IsKeywordText("+") {
		t.Fatalf("keyword text helpers returned unexpected values")
	}
	if !IsKeywordThatCanBeIdentifier("access") || IsKeywordThatCanBeIdentifier("project") {
		t.Fatalf("keyword identifier helper returned unexpected values")
	}
}

func TestSyntaxFactsEnumerableHelpers(t *testing.T) {
	keywordKinds := SyntaxKinds(SyntaxCategoryKeyword)
	if !slices.Contains(keywordKinds, SyntaxKindProjectKeyword) ||
		!slices.Contains(keywordKinds, SyntaxKindContains_CsKeyword) ||
		slices.Contains(keywordKinds, SyntaxKindPlusToken) {
		t.Fatalf("keyword kind enumeration returned unexpected values")
	}

	fixedTextKinds := SyntaxKindsWithFixedText()
	if !slices.Contains(fixedTextKinds, SyntaxKindProjectKeyword) ||
		!slices.Contains(fixedTextKinds, SyntaxKindPlusToken) ||
		!slices.Contains(fixedTextKinds, SyntaxKindDashDashToken) ||
		slices.Contains(fixedTextKinds, SyntaxKindStringLiteralToken) {
		t.Fatalf("fixed-text kind enumeration returned unexpected values")
	}

	keywords := SyntaxKeywordTexts()
	if !slices.Contains(keywords, "project") ||
		!slices.Contains(keywords, "contains_cs") ||
		slices.Contains(keywords, "+") {
		t.Fatalf("keyword text enumeration returned unexpected values")
	}

	punctuation := SyntaxPunctuationTexts()
	if !slices.Contains(punctuation, "(") ||
		!slices.Contains(punctuation, "=>") ||
		slices.Contains(punctuation, "+") {
		t.Fatalf("punctuation text enumeration returned unexpected values")
	}

	wantOperators := []string{"+", "-", "*", "/", "%", "<", "<=", ">", ">=", "==", "!=", "<>", "=~", "!~"}
	if !slices.Equal(SyntaxOperatorTexts(), wantOperators) {
		t.Fatalf("operator text enumeration = %#v, want %#v", SyntaxOperatorTexts(), wantOperators)
	}
}
