package kustolang

import "testing"

func TestSyntaxCategoryNumericOrderMatchesUpstream(t *testing.T) {
	cases := []struct {
		category SyntaxCategory
		value    int
		name     string
	}{
		{SyntaxCategoryNone, 0, "None"},
		{SyntaxCategoryKeyword, 1, "Keyword"},
		{SyntaxCategoryIdentifier, 2, "Identifier"},
		{SyntaxCategoryPunctuation, 3, "Punctuation"},
		{SyntaxCategoryOperator, 4, "Operator"},
		{SyntaxCategoryLiteral, 5, "Literal"},
		{SyntaxCategoryList, 6, "List"},
		{SyntaxCategoryNode, 7, "Node"},
		{SyntaxCategoryOther, 8, "Other"},
	}

	for _, test := range cases {
		if got := int(test.category); got != test.value {
			t.Fatalf("%s numeric value = %d, want %d", test.name, got, test.value)
		}
		if got := test.category.String(); got != test.name {
			t.Fatalf("category %d string = %q, want %q", test.value, got, test.name)
		}
	}
}

func TestSyntaxCategoryIntegratesWithSyntaxFacts(t *testing.T) {
	if got := SyntaxCategoryOf(SyntaxKindWhereKeyword); got != SyntaxCategoryKeyword {
		t.Fatalf("where category = %s, want Keyword", got)
	}
	if got := SyntaxCategoryOf(SyntaxKindOpenParenToken); got != SyntaxCategoryPunctuation {
		t.Fatalf("open paren category = %s, want Punctuation", got)
	}
	if got := SyntaxCategoryOf(SyntaxKindStringLiteralExpression); got != SyntaxCategoryNode {
		t.Fatalf("string literal expression category = %s, want Node", got)
	}
	if got := SyntaxCategory(99).String(); got != "Unknown" {
		t.Fatalf("unknown category string = %q, want Unknown", got)
	}
}
