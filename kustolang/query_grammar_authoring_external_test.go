package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicQueryGrammarMissingFactorySignatures(t *testing.T) {
	var source kustolang.Source[kustolang.LexicalToken]
	var name *kustolang.NameDeclaration = kustolang.CreateMissingNameDeclaration(source, 0)
	var nameExpression kustolang.Expression = kustolang.CreateMissingNameDeclarationExpression(source, 0)
	var nameReference kustolang.Expression = kustolang.CreateMissingNameReference(source, 0)
	var token *kustolang.SyntaxToken = kustolang.CreateMissingNameToken(source, 0)
	var expression kustolang.Expression = kustolang.CreateMissingExpression(source, 0)
	var value kustolang.Expression = kustolang.CreateMissingValue(source, 0)
	var typ kustolang.TypeExpression = kustolang.CreateMissingType(source, 0)
	var longLiteral kustolang.Expression = kustolang.CreateMissingLongLiteral(source, 0)
	var realLiteral kustolang.Expression = kustolang.CreateMissingRealLiteral(source, 0)
	var stringLiteral kustolang.Expression = kustolang.CreateMissingStringLiteral(source, 0)
	var boolLiteral kustolang.Expression = kustolang.CreateMissingBooleanLiteral(source, 0)
	var typeOf kustolang.Expression = kustolang.CreateMissingTypeOfLiteral(source, 0)
	var jsonValue kustolang.Expression = kustolang.CreateMissingJSONValue(source, 0)
	var jsonPair *kustolang.JsonPair = kustolang.CreateMissingJSONPair(source, 0)
	var queryOperator kustolang.QueryOperator = kustolang.CreateMissingQueryOperator(source, 0)
	var queryOperatorExpression kustolang.Expression = kustolang.CreateMissingQueryOperatorExpression(source, 0)

	for label, element := range map[string]kustolang.SyntaxElement{
		"name": name, "name expression": nameExpression, "name reference": nameReference,
		"token": token, "expression": expression, "value": value, "type": typ,
		"long": longLiteral, "real": realLiteral, "string": stringLiteral,
		"boolean": boolLiteral, "typeof": typeOf, "json value": jsonValue,
		"json pair": jsonPair, "query operator": queryOperator,
		"query operator expression": queryOperatorExpression,
	} {
		if element == nil || element.Width() != 0 {
			t.Fatalf("%s = %#v, width %d", label, element, element.Width())
		}
	}
}

func TestCreateMissingExpressionKeywordDiagnostic(t *testing.T) {
	plain := kustolang.CreateMissingExpression(nil, 0)
	keywordSource := kustolang.NewArraySource(kustolang.Lex("to"))
	keyword := kustolang.CreateMissingExpression(keywordSource, 0)
	if !hasSyntaxDiagnosticDescription(plain.Diagnostics(), "Missing expression") {
		t.Fatalf("plain diagnostics = %#v", plain.Diagnostics())
	}
	if !hasSyntaxDiagnosticDescription(keyword.Diagnostics(), "Missing expression: If the keyword 'to' is intended to be part of an expression it needs to be bracketted as ['to'].") {
		t.Fatalf("keyword diagnostics = %#v", keyword.Diagnostics())
	}
	if plain == kustolang.CreateMissingExpression(nil, 0) {
		t.Fatal("missing expression factory reused a syntax node")
	}
}

func TestQueryGrammarStringOperatorMapParity(t *testing.T) {
	got := kustolang.QueryGrammarStringOperatorMap()
	want := map[kustolang.SyntaxKind]kustolang.SyntaxKind{
		kustolang.SyntaxKindEqualTildeToken:          kustolang.SyntaxKindEqualTildeExpression,
		kustolang.SyntaxKindBangTildeToken:           kustolang.SyntaxKindBangTildeExpression,
		kustolang.SyntaxKindHasKeyword:               kustolang.SyntaxKindHasExpression,
		kustolang.SyntaxKindColonToken:               kustolang.SyntaxKindSearchExpression,
		kustolang.SyntaxKindNotHasKeyword:            kustolang.SyntaxKindNotHasExpression,
		kustolang.SyntaxKindHasCsKeyword:             kustolang.SyntaxKindHasCsExpression,
		kustolang.SyntaxKindNotHasCsKeyword:          kustolang.SyntaxKindNotHasCsExpression,
		kustolang.SyntaxKindHasPrefixKeyword:         kustolang.SyntaxKindHasPrefixExpression,
		kustolang.SyntaxKindNotHasPrefixKeyword:      kustolang.SyntaxKindNotHasPrefixExpression,
		kustolang.SyntaxKindHasPrefixCsKeyword:       kustolang.SyntaxKindHasPrefixCsExpression,
		kustolang.SyntaxKindNotHasPrefixCsKeyword:    kustolang.SyntaxKindNotHasPrefixCsExpression,
		kustolang.SyntaxKindHasSuffixKeyword:         kustolang.SyntaxKindHasSuffixExpression,
		kustolang.SyntaxKindNotHasSuffixKeyword:      kustolang.SyntaxKindNotHasSuffixExpression,
		kustolang.SyntaxKindHasSuffixCsKeyword:       kustolang.SyntaxKindHasSuffixCsExpression,
		kustolang.SyntaxKindNotHasSuffixCsKeyword:    kustolang.SyntaxKindNotHasSuffixCsExpression,
		kustolang.SyntaxKindLikeKeyword:              kustolang.SyntaxKindLikeExpression,
		kustolang.SyntaxKindNotLikeKeyword:           kustolang.SyntaxKindNotLikeExpression,
		kustolang.SyntaxKindLikeCsKeyword:            kustolang.SyntaxKindLikeCsExpression,
		kustolang.SyntaxKindNotLikeCsKeyword:         kustolang.SyntaxKindNotLikeCsExpression,
		kustolang.SyntaxKindContainsKeyword:          kustolang.SyntaxKindContainsExpression,
		kustolang.SyntaxKindNotContainsKeyword:       kustolang.SyntaxKindNotContainsExpression,
		kustolang.SyntaxKindNotBangContainsKeyword:   kustolang.SyntaxKindNotContainsExpression,
		kustolang.SyntaxKindContainsCsKeyword:        kustolang.SyntaxKindContainsCsExpression,
		kustolang.SyntaxKindContains_CsKeyword:       kustolang.SyntaxKindContainsCsExpression,
		kustolang.SyntaxKindNotContainsCsKeyword:     kustolang.SyntaxKindNotContainsCsExpression,
		kustolang.SyntaxKindNotBangContainsCsKeyword: kustolang.SyntaxKindNotContainsCsExpression,
		kustolang.SyntaxKindStartsWithKeyword:        kustolang.SyntaxKindStartsWithExpression,
		kustolang.SyntaxKindNotStartsWithKeyword:     kustolang.SyntaxKindNotStartsWithExpression,
		kustolang.SyntaxKindStartsWithCsKeyword:      kustolang.SyntaxKindStartsWithCsExpression,
		kustolang.SyntaxKindNotStartsWithCsKeyword:   kustolang.SyntaxKindNotStartsWithCsExpression,
		kustolang.SyntaxKindEndsWithKeyword:          kustolang.SyntaxKindEndsWithExpression,
		kustolang.SyntaxKindNotEndsWithKeyword:       kustolang.SyntaxKindNotEndsWithExpression,
		kustolang.SyntaxKindEndsWithCsKeyword:        kustolang.SyntaxKindEndsWithCsExpression,
		kustolang.SyntaxKindNotEndsWithCsKeyword:     kustolang.SyntaxKindNotEndsWithCsExpression,
		kustolang.SyntaxKindMatchesRegexKeyword:      kustolang.SyntaxKindMatchesRegexExpression,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("string operator map = %#v", got)
	}
	delete(got, kustolang.SyntaxKindHasKeyword)
	if len(kustolang.QueryGrammarStringOperatorMap()) != len(want) {
		t.Fatal("caller mutation changed the canonical string operator map")
	}
}

func hasSyntaxDiagnosticDescription(diagnostics []kustolang.Diagnostic, description string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Description == description {
			return true
		}
	}
	return false
}
