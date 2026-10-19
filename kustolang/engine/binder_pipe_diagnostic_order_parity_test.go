package kustolang

import "testing"

func TestNativePipeDoesNotPromoteOperatorDiagnosticsAheadOfInput(t *testing.T) {
	query := "datatable(klong,a:string)[1,'x'] | join kind=inner (datatable(k:long,b:string)[1,'y']) on k"
	code := ParseAndAnalyze(query)
	want := []Diagnostic{
		{Code: "KS005", Message: "Expected: :", Start: 15, Length: 1},
		{Code: "KS006", Message: "Missing type name", Start: 15, Length: 0},
		{Code: "KS145", Message: "The column 'k' must exist on both sides of the join.", Start: 90, Length: 1},
	}
	got := code.GetDiagnostics()
	if len(got) != len(want) {
		t.Fatalf("diagnostics = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i].Code != want[i].Code || got[i].Message != want[i].Message || got[i].Start != want[i].Start || got[i].Length != want[i].Length {
			t.Fatalf("diagnostic %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}
