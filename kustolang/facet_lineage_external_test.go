package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicFacetPreservesWithIdentityAndCountLineage(t *testing.T) {
	a := kustolang.NewColumnSymbol("A", kustolang.ScalarTypes.String.Type())
	b := kustolang.NewColumnSymbol("B", kustolang.ScalarTypes.Long.Type())
	table := kustolang.NewTableSymbol("T", a, b).WithIsSerialized(true)
	globals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("db", table))
	code := kustolang.ParseAndAnalyze(`T | facet by A with (project B)`, globals)
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("facet diagnostics = %#v", diagnostics)
	}

	var facet *kustolang.FacetOperator
	code.Syntax().WalkElements(func(element kustolang.SyntaxElement) {
		if candidate, ok := kustolang.WrapGeneratedSyntaxElement(element).(*kustolang.FacetOperator); ok {
			facet = candidate
		}
	})
	if facet == nil {
		t.Fatal("facet node not found")
	}
	group, ok := facet.ReferencedSymbol().(*kustolang.GroupSymbol)
	if !ok || len(group.Members()) != 2 {
		t.Fatalf("facet result = %#v", facet.ReferencedSymbol())
	}
	with := facet.WithClause().(*kustolang.FacetWithExpressionClause).Expression()
	withTable, ok := kustolang.TableSymbolForType(kustolang.ExpressionResultType(with))
	if !ok || group.Members()[0] != withTable {
		t.Fatal("public facet did not preserve with-clause table identity")
	}
	facetTable := group.Members()[1].(*kustolang.TableSymbol)
	count := facetTable.GetColumn("count_A")
	expression := facet.Expressions().At(0).Element()
	if facetTable.GetColumn("A") != a || count == nil || len(count.OriginalColumns) != 1 ||
		count.OriginalColumns[0] != a || count.Source != expression {
		t.Fatalf("public facet lineage = A:%#v count:%#v", facetTable.GetColumn("A"), count)
	}
}
