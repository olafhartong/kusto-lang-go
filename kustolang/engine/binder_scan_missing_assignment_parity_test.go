package kustolang

import "testing"

func TestNativeScanMissingAssignmentTargetPreservesResultSchema(t *testing.T) {
	query := "datatable(x:long)[1,2] | scan declare (s:long=0) with (step A: true =>"
	code := ParseAndAnalyze(query)

	wantDiagnostics := []Diagnostic{
		{Code: "KS006", Message: "Missing name", Start: 70, Length: 0},
		{Code: "KS005", Message: "Expected: ;", Start: 70, Length: 0},
		{Code: "KS005", Message: "Expected: )", Start: 70, Length: 0},
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
	wantColumns := map[string]string{"x": "long", "s": "long", "match_id": "long"}
	if len(table.Columns()) != len(wantColumns) {
		t.Fatalf("columns = %#v, want %v", table.Columns(), wantColumns)
	}
	for _, column := range table.Columns() {
		if wantColumns[column.SymbolName()] != column.Type.SymbolName() {
			t.Fatalf("column %q type = %q, want %q", column.SymbolName(), column.Type.SymbolName(), wantColumns[column.SymbolName()])
		}
	}
}
