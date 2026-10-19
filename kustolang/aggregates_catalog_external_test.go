package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicAggregateCatalogMatchesPinnedUpstream(t *testing.T) {
	all := kustolang.Aggregates.All()
	if len(all) != 61 {
		t.Fatalf("aggregate count = %d, want 61", len(all))
	}
	if all[0] != kustolang.Aggregates.Sum || all[1] != kustolang.Aggregates.SumIf ||
		all[len(all)-1] != kustolang.Aggregates.CountDistinctIf {
		t.Fatalf("aggregate order = first:%v second:%v last:%v", all[0], all[1], all[len(all)-1])
	}
	if kustolang.Aggregates.Count.SymbolName() != "count" || len(kustolang.Aggregates.Count.Signatures()) != 2 {
		t.Fatalf("count metadata = %#v", kustolang.Aggregates.Count)
	}
	if !kustolang.Aggregates.MakeListDeprecated.IsObsolete() || !kustolang.Aggregates.MakeListDeprecated.IsHidden() ||
		kustolang.Aggregates.MakeListDeprecated.Alternative() != "make_list" {
		t.Fatalf("deprecated make-list metadata = %#v", kustolang.Aggregates.MakeListDeprecated)
	}
	if kustolang.Aggregates.CountDistinct.OptimizedAlternative() != "dcount" {
		t.Fatalf("count_distinct optimized alternative = %q", kustolang.Aggregates.CountDistinct.OptimizedAlternative())
	}

	all[0] = nil
	if kustolang.Aggregates.All()[0] == nil {
		t.Fatal("Aggregates.All exposed mutable catalog storage")
	}
	signatures := kustolang.Aggregates.Count.Signatures()
	signatures[0] = nil
	if kustolang.Aggregates.Count.Signatures()[0] == nil {
		t.Fatal("FunctionSymbol.Signatures exposed mutable storage")
	}
}

func TestPublicAggregateGetAnyResultValidAndMalformedExpressions(t *testing.T) {
	valid := kustolang.ParseExpression("Score")
	malformed := kustolang.ParseExpression("(")
	if diagnostics := malformed.GetContainedSyntaxDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("malformed aggregate argument produced no syntax diagnostic")
	}

	result := kustolang.Aggregates.GetAnyResult(kustolang.CustomReturnTypeContext{
		Arguments:           []kustolang.Expression{valid, malformed},
		ArgumentTypes:       []*kustolang.TypeSymbol{kustolang.ScalarTypes.Long.Type(), kustolang.ScalarTypes.String.Type()},
		ArgumentResultNames: []string{"Score", ""},
	}, "any_")
	tuple, ok := result.(*kustolang.TupleSymbol)
	if !ok {
		t.Fatalf("GetAnyResult = %T, want *TupleSymbol", result)
	}
	columns := tuple.Columns()
	if len(columns) != 2 || columns[0].SymbolName() != "Score" || columns[0].Type != kustolang.ScalarTypes.Long.Type() {
		t.Fatalf("GetAnyResult columns = %#v", columns)
	}
	if columns[1].SymbolName() == "" {
		t.Fatal("malformed expression result column has no recovery name")
	}

	var nilCatalog *kustolang.AggregateCatalog
	if got := nilCatalog.GetAnyResult(kustolang.CustomReturnTypeContext{}, "any_"); got != nil {
		t.Fatalf("nil aggregate catalog result = %T, want nil", got)
	}
}

func TestPublicAggregateCatalogDrivesValidAndMalformedBinding(t *testing.T) {
	valid := kustolang.ParseAndAnalyze(`datatable(x:long)[1, 2] | summarize total=sum(x), rows=count()`)
	if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid aggregate query diagnostics = %#v", diagnostics)
	}

	malformed := kustolang.ParseAndAnalyze(`datatable(x:long)[1] | summarize total=sum(`)
	if diagnostics := malformed.GetDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("malformed aggregate call produced no diagnostic")
	}
}
