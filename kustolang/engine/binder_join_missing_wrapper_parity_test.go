package kustolang

import "testing"

func TestNativeJoinMissingRightWrapperStillBindsDatatable(t *testing.T) {
	query := "datatable(k:long,a:string)[1,'x'] | join kind=inner datatable(k:long,b:string)[1,'y']) on k"
	code := ParseAndAnalyze(query)
	diagnostics := code.GetDiagnostics()
	wantDiagnostics := []Diagnostic{
		{Code: "KS006", Message: "Missing join on condition clause", Start: 36, Length: 49},
		{Code: "KS198", Message: "The incomplete fragment is unexpected.", Start: 85, Length: 1},
	}
	if len(diagnostics) != len(wantDiagnostics) {
		t.Fatalf("diagnostics = %#v, want %#v", diagnostics, wantDiagnostics)
	}
	for i := range wantDiagnostics {
		got, want := diagnostics[i], wantDiagnostics[i]
		if got.Code != want.Code || got.Message != want.Message || got.Start != want.Start || got.Length != want.Length {
			t.Fatalf("diagnostic %d = %#v, want %#v", i, got, want)
		}
	}

	table, ok := TableSymbolForType(code.ResultType)
	if !ok {
		t.Fatalf("result type = %#v, want table", code.ResultType)
	}
	wantColumns := map[string]string{"a": "string", "b": "string", "k": "long", "k1": "long"}
	if len(table.Columns()) != len(wantColumns) {
		t.Fatalf("columns = %#v, want %v", table.Columns(), wantColumns)
	}
	for _, column := range table.Columns() {
		if wantColumns[column.SymbolName()] != column.Type.SymbolName() {
			t.Fatalf("column %q type = %q, want %q", column.SymbolName(), column.Type.SymbolName(), wantColumns[column.SymbolName()])
		}
	}
}
