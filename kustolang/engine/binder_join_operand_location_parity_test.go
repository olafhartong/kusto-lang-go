package kustolang

import "testing"

func TestNativeMalformedJoinOperandDiagnosticHasSourceLocation(t *testing.T) {
	code := ParseAndAnalyze("datatable(k:long,a:string)[1,'x'] | join kind")
	diagnostics := code.GetDiagnostics()
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	want := Diagnostic{Code: "KS204", Message: "The name 'kind' does not refer to any known table, tabular variable or function.", Start: 41, Length: 4}
	got := diagnostics[1]
	if got.Code != want.Code || got.Message != want.Message || got.Start != want.Start || got.Length != want.Length {
		t.Fatalf("join operand diagnostic = %#v, want %#v", got, want)
	}
}
