package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicReduceExactStringTypeAndOutputLineage(t *testing.T) {
	table := kustolang.NewTableSymbol("T",
		kustolang.NewColumnSymbol("Text", kustolang.ScalarTypes.String.Type()),
		kustolang.NewColumnSymbol("Payload", kustolang.ScalarTypes.Dynamic.Type()))
	globals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("db", table))

	valid := kustolang.ParseAndAnalyze("T | reduce by Text", globals)
	if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid diagnostics = %#v", diagnostics)
	}
	assertPublicReduceSources(t, valid)

	dynamic := kustolang.ParseAndAnalyze("T | reduce by Payload", globals)
	diagnostics := dynamic.GetDiagnostics()
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS141" || diagnostics[0].Start != len("T | reduce by ") || diagnostics[0].Length != len("Payload") {
		t.Fatalf("dynamic diagnostics = %#v", diagnostics)
	}
	assertPublicReduceSources(t, dynamic)

	malformedText := "T | reduce by"
	malformed := kustolang.ParseAndAnalyze(malformedText, globals)
	diagnostics = malformed.GetDiagnostics()
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS006" || diagnostics[0].Start != len(malformedText) || diagnostics[0].Length != 0 {
		t.Fatalf("malformed diagnostics = %#v", diagnostics)
	}
	assertPublicReduceSources(t, malformed)
}

func assertPublicReduceSources(t *testing.T, code *kustolang.KustoCode) {
	t.Helper()
	reduces := kustolang.GetDescendants[*kustolang.ReduceByOperator](code.Syntax())
	if len(reduces) != 1 || !reduces[0].IsBound() {
		t.Fatalf("reduce operators = %#v", reduces)
	}
	table, ok := kustolang.TableSymbolForType(kustolang.ExpressionResultType(reduces[0]))
	if !ok || len(table.Columns()) != 3 {
		t.Fatalf("reduce result = %#v", kustolang.ExpressionResultType(reduces[0]))
	}
	for _, column := range table.Columns() {
		if column.Source != reduces[0].Expression() {
			t.Fatalf("%s source = %#v, want reduce expression", column.SymbolName(), column.Source)
		}
	}
}
