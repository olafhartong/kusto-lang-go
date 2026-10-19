package kustolang

import "testing"

func TestNativeSummarizeIncompleteBracketedGroupPreservesProjection(t *testing.T) {
	query := "range x from 1 to 5 step 1 | extend y=x*x | summarize s=sum(y) by x ["
	code := ParseAndAnalyze(query)

	wantDiagnostics := []Diagnostic{
		{Code: "KS006", Message: "Missing name", Start: 69, Length: 0},
		{Code: "KS005", Message: "Expected: ]", Start: 69, Length: 0},
	}
	gotDiagnostics := code.GetDiagnostics()
	if len(gotDiagnostics) != len(wantDiagnostics) {
		t.Fatalf("diagnostics = %#v, want %#v", gotDiagnostics, wantDiagnostics)
	}
	for i := range wantDiagnostics {
		got, want := gotDiagnostics[i], wantDiagnostics[i]
		if got.Code != want.Code || got.Message != want.Message || got.Start != want.Start || got.Length != want.Length {
			t.Fatalf("diagnostic %d = %#v, want %#v", i, got, want)
		}
	}

	table, ok := TableSymbolForType(code.ResultType)
	if !ok {
		t.Fatalf("result type = %#v, want table", code.ResultType)
	}
	wantColumns := map[string]string{"Column1": "unknown", "s": "long"}
	if len(table.Columns()) != len(wantColumns) {
		t.Fatalf("columns = %#v, want %v", table.Columns(), wantColumns)
	}
	for _, column := range table.Columns() {
		if wantColumns[column.SymbolName()] != column.Type.SymbolName() {
			t.Fatalf("column %q type = %q, want %q", column.SymbolName(), column.Type.SymbolName(), wantColumns[column.SymbolName()])
		}
	}
}
