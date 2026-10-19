package kustolang

import "testing"

func TestNativePartitionMissingPipeStillBindsSubqueryProjection(t *testing.T) {
	query := "datatable(x:long)[1,2,3]  partition hint.strategy=native by x (summarize c=count())"
	code := ParseAndAnalyze(query)
	wantDiagnostics := []Diagnostic{
		{Code: "KS005", Message: "Expected: ;", Start: 24, Length: 0},
		{Code: "KS174", Message: "The operator cannot be the first operator in a query.", Start: 26, Length: 9},
		{Code: "KS142", Message: "The name 'x' does not refer to any known column, table, variable or function.", Start: 60, Length: 1},
	}
	got := code.GetDiagnostics()
	if len(got) != len(wantDiagnostics) {
		t.Fatalf("diagnostics = %#v, want %#v", got, wantDiagnostics)
	}
	for i := range wantDiagnostics {
		want := wantDiagnostics[i]
		if got[i].Code != want.Code || got[i].Message != want.Message || got[i].Start != want.Start || got[i].Length != want.Length {
			t.Fatalf("diagnostic %d = %#v, want %#v", i, got[i], want)
		}
	}

	table, ok := TableSymbolForType(code.ResultType)
	if !ok || len(table.Columns()) != 1 || table.Columns()[0].SymbolName() != "c" || table.Columns()[0].Type.SymbolName() != "long" {
		t.Fatalf("result = %#v, want c:long", code.ResultType)
	}
}
