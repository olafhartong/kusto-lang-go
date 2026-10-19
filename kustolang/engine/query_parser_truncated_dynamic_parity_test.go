package kustolang

import "testing"

func TestNativeTruncatedDynamicTailPreservesParsedRoot(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantCodes   []string
		wantStarts  []int
		wantColumns map[string]string
	}{
		{
			name:        "truncated dynamic value",
			query:       "datatable(x:dynamic)[dynamic",
			wantCodes:   []string{"KS005", "KS198"},
			wantStarts:  []int{21, 21},
			wantColumns: map[string]string{"x": "dynamic"},
		},
		{
			name:        "malformed dynamic bracket value",
			query:       "datatable(x:dynamic)[dynamic[1,2])] | mv-apply x to typeof(long) on (summarize c=count())",
			wantCodes:   []string{"KS005", "KS198"},
			wantStarts:  []int{21, 21},
			wantColumns: map[string]string{"x": "dynamic"},
		},
		{
			name:       "malformed search root",
			query:      "datatablex:dynamic)[dynamic([1,2])] | mv-apply x to typeof(long) on (summarize c=count())",
			wantCodes:  []string{"KS142", "KS006", "KS198"},
			wantStarts: []int{0, 11, 11},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := ParseAndAnalyze(tt.query)
			diagnostics := code.GetDiagnostics()
			if len(diagnostics) != len(tt.wantCodes) {
				t.Fatalf("diagnostics = %#v, want codes %v", diagnostics, tt.wantCodes)
			}
			for i := range tt.wantCodes {
				if diagnostics[i].Code != tt.wantCodes[i] || diagnostics[i].Start != tt.wantStarts[i] {
					t.Fatalf("diagnostic %d = %#v, want %s at %d", i, diagnostics[i], tt.wantCodes[i], tt.wantStarts[i])
				}
			}
			if tt.wantColumns == nil {
				if table, ok := TableSymbolForType(code.ResultType); ok && len(table.Columns()) != 0 {
					t.Fatalf("columns = %#v, want none", table.Columns())
				}
				return
			}
			table, ok := TableSymbolForType(code.ResultType)
			if !ok || len(table.Columns()) != len(tt.wantColumns) {
				t.Fatalf("result type = %#v, want columns %v", code.ResultType, tt.wantColumns)
			}
			for _, column := range table.Columns() {
				if tt.wantColumns[column.SymbolName()] != column.Type.SymbolName() {
					t.Fatalf("column %q type = %q, want %q", column.SymbolName(), column.Type.SymbolName(), tt.wantColumns[column.SymbolName()])
				}
			}
		})
	}
}
