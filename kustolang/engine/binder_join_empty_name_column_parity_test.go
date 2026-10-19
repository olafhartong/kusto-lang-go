package kustolang

import "testing"

func TestNativeJoinPreservesRecoveredEmptyNameColumns(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  map[string]string
	}{
		{
			name:  "right schema",
			query: "datatable(k:long,a:string)[1,'x'] | join kind=inner (datatable(k:longb:string)[1,'y']) on k",
			want:  map[string]string{"": "string", "a": "string", "k": "long", "k1": "unknown"},
		},
		{
			name:  "left schema",
			query: "datatable(k:longa:string)[1,'x'] | join kind=inner (datatable(k:long,b:string)[1,'y']) on k",
			want:  map[string]string{"": "string", "b": "string", "k": "unknown", "k1": "long"},
		},
		{
			name:  "valid duplicate key remains unique",
			query: "datatable(k:long,a:string)[1,'x'] | join kind=inner (datatable(k:long,b:string)[1,'y']) on k",
			want:  map[string]string{"a": "string", "b": "string", "k": "long", "k1": "long"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := ParseAndAnalyze(tt.query)
			table, ok := TableSymbolForType(code.ResultType)
			if !ok {
				t.Fatalf("result type = %#v, want table", code.ResultType)
			}
			if len(table.Columns()) != len(tt.want) {
				t.Fatalf("columns = %#v, want %v", table.Columns(), tt.want)
			}
			for _, column := range table.Columns() {
				if tt.want[column.SymbolName()] != column.Type.SymbolName() {
					t.Fatalf("column %q type = %q, want %q", column.SymbolName(), column.Type.SymbolName(), tt.want[column.SymbolName()])
				}
			}
		})
	}
}
