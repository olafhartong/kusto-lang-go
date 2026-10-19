package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicFilterRecoveryPreservesSemanticResults(t *testing.T) {
	a := kustolang.NewColumnSymbol("A", kustolang.ScalarTypes.String.Type())
	table := kustolang.NewTableSymbol("T", a)
	globals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("db", table))

	valid := kustolang.ParseAndAnalyze(`T | where A == "x"`, globals)
	if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid diagnostics = %#v", diagnostics)
	}

	malformed := kustolang.ParseAndAnalyze("T | where A in ( | count", globals)
	diagnostics := malformed.GetDiagnostics()
	if len(diagnostics) != 2 || diagnostics[0].Code != "KS006" || diagnostics[0].Start != 16 || diagnostics[0].Length != 0 ||
		diagnostics[1].Code != "KS005" || diagnostics[1].Start != 17 || diagnostics[1].Length != 1 {
		t.Fatalf("malformed diagnostics = %#v", diagnostics)
	}
	result, ok := kustolang.TableSymbolForType(malformed.ResultType)
	if !ok || len(result.Columns()) != 1 || result.GetColumn("Count") == nil {
		t.Fatalf("malformed result = %#v, want Count table", malformed.ResultType)
	}

	var filterBound, inBound bool
	malformed.Syntax().WalkElements(func(element kustolang.SyntaxElement) {
		switch node := kustolang.WrapGeneratedSyntaxElement(element).(type) {
		case *kustolang.FilterOperator:
			filterBound = node.IsBound()
		case *kustolang.InExpression:
			inBound = node.IsBound() && kustolang.ExpressionResultType(node) == kustolang.ScalarTypes.Bool.Type()
		}
	})
	if !filterBound || !inBound {
		t.Fatalf("recovered semantics = filter:%t in-expression:%t", filterBound, inBound)
	}
}
