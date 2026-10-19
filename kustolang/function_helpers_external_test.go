package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicFunctionHelpersConstantsNamesAndBooleans(t *testing.T) {
	if kustolang.MaxRepeat != 32767 {
		t.Fatalf("MaxRepeat = %d, want 32767", kustolang.MaxRepeat)
	}
	if got := kustolang.MakeColumnName("series", "value", "avg"); got != "series_value_avg" {
		t.Fatalf("MakeColumnName = %q", got)
	}
	if got := kustolang.MakeColumnName("single"); got != "single" {
		t.Fatalf("single MakeColumnName = %q", got)
	}
	if got := kustolang.MakeValidNameFragment("a-b c"); got != "a_b_c" {
		t.Fatalf("MakeValidNameFragment = %q", got)
	}

	integer := kustolang.ParseExpression("42")
	boolean := kustolang.ParseExpression("true")
	if got := kustolang.GetConstantValue(integer); got != "42" {
		t.Fatalf("integer constant = %q", got)
	}
	if got := kustolang.GetConstantValue(boolean); got != "True" {
		t.Fatalf("boolean constant = %q", got)
	}
	if !kustolang.IsBoolean(boolean) || kustolang.IsBoolean(integer) {
		t.Fatalf("boolean classification = true:%t integer:%t", kustolang.IsBoolean(boolean), kustolang.IsBoolean(integer))
	}

	comma := kustolang.SyntaxTokenFromLexical(kustolang.NewLexicalToken(kustolang.SyntaxKindCommaToken, "", ","))
	values := kustolang.MakeSyntaxSeparatedList[kustolang.Expression]([]any{integer, comma, boolean})
	if got := kustolang.GetConstantValues(values); len(got) != 2 || got[0] != "42" || got[1] != "True" {
		t.Fatalf("constant values = %#v", got)
	}
}

func TestPublicFunctionHelpersPrefixedTuplePreservesLineage(t *testing.T) {
	parameter := kustolang.NewParameter("value", kustolang.ScalarTypes.String.Type())
	signature := kustolang.NewSignature(kustolang.ScalarTypes.Dynamic.Type(), parameter)
	function := kustolang.NewFunctionSymbol("series_stats", signature).WithResultNamePrefix("stats")
	signature.Symbol = function
	argument := kustolang.ParseExpression("Score")
	context := kustolang.CustomReturnTypeContext{
		Arguments:           []kustolang.Expression{argument},
		ArgumentParameters:  []*kustolang.Parameter{parameter},
		ArgumentResultNames: []string{"Score"},
		Signature:           signature,
	}
	base := kustolang.NewTupleSymbolOf(
		kustolang.NewColumnSymbol("avg", kustolang.ScalarTypes.Real.Type()),
		kustolang.NewColumnSymbol("stdev", kustolang.ScalarTypes.Real.Type()),
	)
	prefixed := kustolang.MakePrefixedTuple(context, "value", base)
	columns := prefixed.Columns()
	if len(columns) != 2 || columns[0].SymbolName() != "stats_Score_avg" || columns[1].SymbolName() != "stats_Score_stdev" {
		t.Fatalf("prefixed columns = %#v", columns)
	}
	if original := columns[0].OriginalColumns; len(original) != 1 || original[0] != base.Columns()[0] {
		t.Fatalf("prefixed lineage = %#v", original)
	}
}

func TestPublicFunctionHelpersRejectMalformedBooleanExpression(t *testing.T) {
	malformed := kustolang.ParseExpression("(")
	if diagnostics := malformed.GetContainedSyntaxDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("malformed expression produced no syntax diagnostic")
	}
	if kustolang.IsBoolean(malformed) {
		t.Fatal("malformed expression reported boolean")
	}
}
