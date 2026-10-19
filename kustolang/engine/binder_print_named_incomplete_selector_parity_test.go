package kustolang

import "testing"

func TestNativePrintNamedExpressionWithIncompleteSelectorPreservesColumn(t *testing.T) {
	query := "let F=(x:long){x+1}; print y=F(2) ["
	code := ParseAndAnalyze(query)
	wantDiagnostics := []Diagnostic{
		{Code: "KS006", Message: "Missing name", Start: 35, Length: 0},
		{Code: "KS005", Message: "Expected: ]", Start: 35, Length: 0},
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
	if !ok || len(table.Columns()) != 1 || table.Columns()[0].SymbolName() != "y" || table.Columns()[0].Type.SymbolName() != "unknown" {
		t.Fatalf("result = %#v, want y:unknown", code.ResultType)
	}
}
