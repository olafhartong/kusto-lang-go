package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicOperatorCatalogMatchesPinnedUpstream(t *testing.T) {
	all := kustolang.Operators.All()
	if len(all) != 55 {
		t.Fatalf("operator count = %d, want 55", len(all))
	}
	if all[0] != kustolang.Operators.UnaryMinus || all[1] != kustolang.Operators.UnaryPlus ||
		all[len(all)-1] != kustolang.Operators.HasAll {
		t.Fatalf("operator order = first:%v second:%v last:%v", all[0], all[1], all[len(all)-1])
	}
	if kustolang.Operators.Add.OperatorKind != kustolang.OperatorKindAdd ||
		kustolang.Operators.Add.SymbolKind() != kustolang.SymbolKindOperator ||
		kustolang.Operators.Add.SymbolTabularity() != kustolang.TabularityScalar {
		t.Fatalf("add metadata = %#v", kustolang.Operators.Add)
	}
	if len(kustolang.Operators.Add.Signatures()) != 8 || len(kustolang.Operators.Between.Signatures()) != 5 {
		t.Fatalf("signature counts = add:%d between:%d", len(kustolang.Operators.Add.Signatures()), len(kustolang.Operators.Between.Signatures()))
	}

	all[0] = nil
	if kustolang.Operators.All()[0] == nil {
		t.Fatal("Operators.All exposed mutable catalog storage")
	}
	signatures := kustolang.Operators.Add.Signatures()
	signatures[0] = nil
	if kustolang.Operators.Add.Signatures()[0] == nil {
		t.Fatal("OperatorSymbol.Signatures exposed mutable storage")
	}
}

func TestPublicOperatorCatalogDrivesValidAndMalformedBinding(t *testing.T) {
	valid := kustolang.ParseAndAnalyze(`print sum = 1 + 2, matches = "alpha" has "ph"`)
	if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid operator query diagnostics = %#v", diagnostics)
	}

	malformed := kustolang.ParseAndAnalyze(`print broken = 1 +`)
	if diagnostics := malformed.GetDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("malformed operator expression produced no diagnostic")
	}
}
