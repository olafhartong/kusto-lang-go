package kustolang

import "testing"

func TestNativeEmptyRecoveryNameDoesNotAddUnknownNameDiagnostic(t *testing.T) {
	query := "datatable(s:string)['a=1 b=x'] | parse-kv s as (a:long,b:string) with (pair_delimiter=' ',"
	code := ParseAndAnalyze(query)
	want := []Diagnostic{
		{Code: "KS006", Message: "Missing parameter", Start: 90, Length: 0},
		{Code: "KS005", Message: "Expected: )", Start: 90, Length: 0},
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
