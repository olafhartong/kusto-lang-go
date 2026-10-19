package kustolang

import "testing"

func TestSyntaxExtensionsGetParameterAndLiteralValue(t *testing.T) {
	kind := NewQueryOperatorParameter(
		"kind",
		QueryOperatorParameterValueBoolLiteral,
		WithQueryOperatorParameterAliases([]string{"k"}),
	)
	parameter := MustNewNamedParameterTyped(
		MustNewNameDeclarationTyped(MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "k")))),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindEqualToken, "", "=")),
		MustNewLiteralExpressionTyped(SyntaxKindBooleanLiteralExpression, SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindBooleanLiteralToken, "", "true"))),
		CompletionHintScalar,
	)
	list := AsSyntaxListOf[*NamedParameter](NewTypedSyntaxList("NamedParameter", parameter))

	if got := GetParameter(list, kind); got != parameter {
		t.Fatalf("GetParameter alias = %#v, want parameter", got)
	}
	if got := GetParameterLiteralValue(list, kind, false); got != true {
		t.Fatalf("bool literal value = %v, want true", got)
	}
	if got := GetParameterLiteralValue(list, kind, 99); got != 99 {
		t.Fatalf("mismatched literal default = %v, want 99", got)
	}
	missing := NewQueryOperatorParameter("missing", QueryOperatorParameterValueBoolLiteral)
	if got := GetParameterLiteralValue(list, missing, true); got != true {
		t.Fatalf("missing literal default = %v, want true", got)
	}
}

func TestSyntaxExtensionsGetParameterNameValue(t *testing.T) {
	target := NewQueryOperatorParameter("target", QueryOperatorParameterValueNameDeclaration)
	stringParam := NewQueryOperatorParameter("stringy", QueryOperatorParameterValueStringLiteral)
	compoundParam := NewQueryOperatorParameter("compound", QueryOperatorParameterValueStringLiteral)
	numberParam := NewQueryOperatorParameter("number", QueryOperatorParameterValueIntegerLiteral)

	nameValue := MustNewNamedParameterTyped(
		MustNewNameDeclarationTyped(MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "target")))),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindEqualToken, "", "=")),
		MustNewNameReferenceTyped(MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "Account"))), SymbolMatchAny),
		CompletionHintScalar,
	)
	stringValue := MustNewNamedParameterTyped(
		MustNewNameDeclarationTyped(MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "stringy")))),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindEqualToken, "", "=")),
		MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindStringLiteralToken, "", "'abc'"))),
		CompletionHintScalar,
	)
	compoundValue := MustNewNamedParameterTyped(
		MustNewNameDeclarationTyped(MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "compound")))),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindEqualToken, "", "=")),
		MustNewCompoundStringLiteralExpressionTyped(AsSyntaxListOf[*SyntaxToken](NewTypedSyntaxList("SyntaxToken",
			SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindStringLiteralToken, "", "'a'")),
			SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindStringLiteralToken, "", "'b'")),
		))),
		CompletionHintScalar,
	)
	numberValue := MustNewNamedParameterTyped(
		MustNewNameDeclarationTyped(MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "number")))),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindEqualToken, "", "=")),
		MustNewLiteralExpressionTyped(SyntaxKindLongLiteralExpression, SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindLongLiteralToken, "", "42"))),
		CompletionHintScalar,
	)
	list := AsSyntaxListOf[*NamedParameter](NewTypedSyntaxList("NamedParameter", nameValue, stringValue, compoundValue, numberValue))

	if got := GetParameterNameValue(list, target); got != "Account" {
		t.Fatalf("name parameter value = %q, want Account", got)
	}
	if got := GetParameterNameValue(list, stringParam); got != "abc" {
		t.Fatalf("string parameter value = %q, want abc", got)
	}
	if got := GetParameterNameValue(list, compoundParam); got != "ab" {
		t.Fatalf("compound parameter value = %q, want ab", got)
	}
	if got := GetParameterNameValue(list, numberParam); got != "" {
		t.Fatalf("non-name parameter value = %q, want empty", got)
	}
}

func TestNameSimpleNameMirrorsUpstreamCases(t *testing.T) {
	tokenName := MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "Account")))
	if got := NameSimpleName(tokenName); got != "Account" {
		t.Fatalf("token simple name = %q", got)
	}

	bracketed := MustNewBracketedNameTyped(
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindOpenBracketToken, "", "[")),
		MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindStringLiteralToken, "", "'bad name'"))),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindCloseBracketToken, "", "]")),
	)
	if got := NameSimpleName(bracketed); got != "bad name" {
		t.Fatalf("bracketed simple name = %q", got)
	}

	wildcard := MustNewWildcardedNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "Sec*")))
	if got := NameSimpleName(wildcard); got != "Sec*" {
		t.Fatalf("wildcard simple name = %q", got)
	}
}
