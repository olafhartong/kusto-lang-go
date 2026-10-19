package kustolang

import "testing"

func TestNativeUnknownFunctionInJoinOperandBindsAsOpenTable(t *testing.T) {
	query := "datatable(k:long,a:string)[1,'x'] | join kindinner (datatable(k:long,b:string)[1,'y']) on k"
	code := ParseAndAnalyze(query)
	diagnostics := code.GetDiagnostics()
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS211" || diagnostics[0].Start != 41 || diagnostics[0].Length != 9 {
		t.Fatalf("diagnostics = %#v, want KS211 at 41:9", diagnostics)
	}

	table, ok := TableSymbolForType(code.ResultType)
	if !ok {
		t.Fatalf("result type = %#v, want table", code.ResultType)
	}
	want := map[string]string{"a": "string", "k": "long", "k1": "unknown"}
	if len(table.Columns()) != len(want) {
		t.Fatalf("columns = %#v, want %v", table.Columns(), want)
	}
	for _, column := range table.Columns() {
		if want[column.SymbolName()] != column.Type.SymbolName() {
			t.Fatalf("column %q type = %q, want %q", column.SymbolName(), column.Type.SymbolName(), want[column.SymbolName()])
		}
	}
}
