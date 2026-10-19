package kustolang_test

// Behavioral parity cases derived from the Apache-2.0 Kusto.Language sources.
// See third_party/kusto/LICENSE and third_party/kusto/UPSTREAM.json.

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicPrintTupleProjectionParity(t *testing.T) {
	tests := []struct {
		query   string
		columns map[string]string
		code    string
	}{
		{
			query:   `print Stats=series_stats(dynamic([1,2,3]))`,
			columns: map[string]string{"Stats": "real"},
		},
		{
			query: `print series_stats(dynamic([1,2,3]))`,
			columns: map[string]string{
				"series_stats__min":      "real",
				"series_stats__min_idx":  "long",
				"series_stats__max":      "real",
				"series_stats__max_idx":  "long",
				"series_stats__avg":      "real",
				"series_stats__stdev":    "real",
				"series_stats__variance": "real",
			},
		},
		{
			query:   `print array_sort_asc(dynamic([2,1]))`,
			columns: map[string]string{"print_0": "dynamic"},
		},
		{
			query:   `print Stats=series_stats(dynamic([1,2,3])`,
			columns: map[string]string{"Stats": "real"},
			code:    "KS005",
		},
	}

	for _, test := range tests {
		code := kustolang.ParseAndAnalyze(test.query)
		result, ok := kustolang.TableSymbolForType(code.ResultType)
		if !ok || len(result.Columns()) != len(test.columns) {
			t.Fatalf("query %q result = %#v, want %d columns", test.query, code.ResultType, len(test.columns))
		}
		for name, typeName := range test.columns {
			column := result.GetColumn(name)
			if column == nil || column.Type == nil || column.Type.SymbolName() != typeName {
				t.Fatalf("query %q column %q = %#v, want %s", test.query, name, column, typeName)
			}
		}
		diagnostics := code.GetDiagnostics()
		if test.code == "" {
			if len(diagnostics) != 0 {
				t.Fatalf("query %q diagnostics = %#v", test.query, diagnostics)
			}
		} else if len(diagnostics) != 1 || diagnostics[0].Code != test.code {
			t.Fatalf("query %q diagnostics = %#v, want %s", test.query, diagnostics, test.code)
		}
	}
}
