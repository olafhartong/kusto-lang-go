package kustolang

import "testing"

func TestNativeSummarizeGroupFunctionUsesFirstArgumentResultName(t *testing.T) {
	queries := []string{
		`datatable(T:datetime,V:real)[datetime(2024-01-01),1.5] | summarize Average=avg(V) by bin(T, 1h)`,
		`datatable(T:datetime,V:real)[datetime(2024-01-01),1.5] | summarize Average=avg(V) by bin(T, 1h) | top 2 by Average desc`,
	}
	for _, text := range queries {
		code := ParseAndAnalyze(text)
		if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", text, diagnostics)
		}
		result, ok := TableSymbolForType(code.ResultType)
		if !ok {
			t.Fatalf("%q result = %#v", text, code.ResultType)
		}
		assertOrderedColumns(t, result, "T", "Average")
		if column := result.GetColumn("T"); column == nil || column.Type != ScalarTypes.DateTime.Type() {
			t.Fatalf("%q T column = %#v", text, column)
		}
	}
}
