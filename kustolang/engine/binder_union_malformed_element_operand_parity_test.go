package kustolang

import "testing"

func TestNativeUnionMalformedSearchOperandPreservesRecoverableInputs(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		starts      []int
		wantColumns map[string]string
	}{
		{
			name:        "malformed second operand",
			query:       "union withsource=Source (datatable(a:long)[1]), (datatablea:long)[2])",
			starts:      []int{48, 49, 60, 60, 65, 66, 68},
			wantColumns: map[string]string{"Source": "string", "a": "long"},
		},
		{
			name:        "malformed first operand",
			query:       "union withsource=Source (datatablea:long)[1]), (datatable(a:long)[2])",
			starts:      []int{24, 25, 36, 36, 41, 42, 44},
			wantColumns: map[string]string{"Source": "string"},
		},
	}
	wantCodes := []string{"KS111", "KS142", "KS228", "KS142", "KS005", "KS141", "KS198"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := ParseAndAnalyze(tt.query)
			diagnostics := code.GetDiagnostics()
			if len(diagnostics) != len(wantCodes) {
				t.Fatalf("diagnostics = %#v, want codes %v", diagnostics, wantCodes)
			}
			for i := range wantCodes {
				if diagnostics[i].Code != wantCodes[i] || diagnostics[i].Start != tt.starts[i] {
					t.Fatalf("diagnostic %d = %#v, want %s at %d", i, diagnostics[i], wantCodes[i], tt.starts[i])
				}
			}
			table, ok := TableSymbolForType(code.ResultType)
			if !ok || len(table.Columns()) != len(tt.wantColumns) {
				t.Fatalf("result = %#v, want columns %v", code.ResultType, tt.wantColumns)
			}
			for _, column := range table.Columns() {
				if tt.wantColumns[column.SymbolName()] != column.Type.SymbolName() {
					t.Fatalf("column %q type = %q, want %q", column.SymbolName(), column.Type.SymbolName(), tt.wantColumns[column.SymbolName()])
				}
			}
		})
	}
}
