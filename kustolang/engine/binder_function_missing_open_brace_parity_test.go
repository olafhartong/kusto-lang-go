package kustolang

import "testing"

func TestNativeMissingFunctionBraceKeepsRecoveredPipeTabular(t *testing.T) {
	query := "let F=(T:(k:long))T | summarize c=count()}; F(datatable(k:long)[1,2])"
	code := ParseAndAnalyze(query)
	wantCodes := []string{"KS005", "KS005", "KS005", "KS204", "KS198"}
	wantStarts := []int{18, 18, 18, 18, 41}
	diagnostics := code.GetDiagnostics()
	if len(diagnostics) != len(wantCodes) {
		t.Fatalf("diagnostics = %#v, want codes %v", diagnostics, wantCodes)
	}
	for i := range wantCodes {
		if diagnostics[i].Code != wantCodes[i] || diagnostics[i].Start != wantStarts[i] {
			t.Fatalf("diagnostic %d = %#v, want %s at %d", i, diagnostics[i], wantCodes[i], wantStarts[i])
		}
	}
	table, ok := TableSymbolForType(code.ResultType)
	if !ok || len(table.Columns()) != 1 || table.Columns()[0].SymbolName() != "c" || table.Columns()[0].Type.SymbolName() != "long" {
		t.Fatalf("result = %#v, want c:long", code.ResultType)
	}
}
