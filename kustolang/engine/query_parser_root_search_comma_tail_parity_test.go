package kustolang

import "testing"

func TestNativeRootSearchExpressionCommaTailRecoveryParity(t *testing.T) {
	const text = `datatablek:long,a:string)[1,'x'] | join kind=inner (datatable(k:long,b:string)[1,'y']) on k`
	code := ParseAndAnalyze(text)

	expression := queryBlockExpressionAtForTest(code.Query, 0)
	search, ok := expression.(*BinaryExpression)
	if !ok {
		t.Fatalf("root expression = %T, want BinaryExpression", expression)
	}
	if search.Kind() != SyntaxKindSearchExpression {
		t.Fatalf("root expression kind = %v, want SearchExpression", search.Kind())
	}
	assertDiagnosticLocations(t, code.GetDiagnostics(),
		diagnosticLocation{"KS142", 0, 10},
		diagnosticLocation{"KS228", 11, 4},
		diagnosticLocation{"KS142", 11, 4},
		diagnosticLocation{"KS198", 15, 1},
	)

	skipped := code.Query.SkippedTokens()
	if skipped == nil || skipped.Tokens() == nil || skipped.Tokens().Count() == 0 {
		t.Fatal("comma tail was not preserved as skipped tokens")
	}
	first := skipped.Tokens().At(0)
	if first.Kind() != SyntaxKindCommaToken || first.ByteTextStart() != 15 {
		t.Fatalf("first skipped token = %v at %d, want comma at 15", first.Kind(), first.ByteTextStart())
	}
	if table := terminalResultTable(code); table != nil && len(table.Columns()) != 0 {
		t.Fatalf("result columns = %#v, want empty result", table.Columns())
	}
}

func TestNativeRootSearchExpressionCommaTailRecoveryControls(t *testing.T) {
	t.Run("valid datatable source", func(t *testing.T) {
		const text = `datatable(k:long,a:string)[1,'x'] | join kind=inner (datatable(k:long,b:string)[1,'y']) on k`
		code := ParseAndAnalyze(text)
		if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
			t.Fatalf("diagnostics = %#v, want none", diagnostics)
		}
		if _, ok := queryBlockExpressionAtForTest(code.Query, 0).(*PipeExpression); !ok {
			t.Fatalf("root expression = %T, want PipeExpression", queryBlockExpressionAtForTest(code.Query, 0))
		}
		if code.Query.SkippedTokens() != nil {
			t.Fatalf("skipped tokens = %#v, want none", code.Query.SkippedTokens())
		}
	})

	t.Run("non-search binary tail", func(t *testing.T) {
		query := ParseQuery(`1 + 2, 3`)
		statement, ok := query.Statements().At(0).Element().(*ExpressionStatement)
		if !ok || statement.GetChild(0) == nil || statement.GetChild(0).Kind() != SyntaxKindNone {
			t.Fatalf("root statement = %#v, want existing opaque fallback", query.Statements().At(0).Element())
		}
		if query.SkippedTokens() != nil {
			t.Fatalf("skipped tokens = %#v, want opaque fallback ownership", query.SkippedTokens())
		}
	})
}
