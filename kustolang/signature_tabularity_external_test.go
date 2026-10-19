package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicComputedTabularityAuthoring(t *testing.T) {
	function := kustolang.NewComputedFunctionSymbolInferred("Rows", "print Value=1")
	signature := function.Signatures()[0]
	signature.ComputeTabularity(kustolang.DefaultGlobalState)
	if signature.Tabularity != kustolang.TabularityTabular || function.SymbolTabularity() != kustolang.TabularityTabular {
		t.Fatalf("public inferred tabularity = signature:%s function:%s", signature.Tabularity, function.SymbolTabularity())
	}

	scalar := kustolang.NewComputedSignatureInferred("value + 1", kustolang.NewParameter("value", kustolang.ScalarTypes.Long.Type()))
	scalar.ComputeTabularity(kustolang.DefaultGlobalState)
	if scalar.Tabularity != kustolang.TabularityScalar {
		t.Fatalf("public scalar tabularity = %s", scalar.Tabularity)
	}
}
