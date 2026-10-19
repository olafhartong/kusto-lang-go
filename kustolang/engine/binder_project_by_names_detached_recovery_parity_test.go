package kustolang

import "testing"

func TestCurrentDetachedProjectByNamesReferenceParity(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		diagnostics []diagnosticLocation
	}{
		{
			name:  "detached filter source",
			query: `datatable(A:long,B:string,C:real)[1,"x",2.5]  filter A > 0 | project-by-names dynamic(["A","B"])`,
			diagnostics: []diagnosticLocation{
				{"KS005", 44, 0}, {"KS142", 46, 6}, {"KS005", 53, 1}, {"KS142", 53, 1},
			},
		},
		{
			name:  "detached project by names",
			query: `datatable(A:long,B:string,C:real)[1,"x",2.5] | filter A > 0  project-by-names dynamic(["A","B"])`,
			diagnostics: []diagnosticLocation{
				{"KS005", 59, 0}, {"KS174", 61, 16},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.query)
			assertDiagnosticLocations(t, code.GetDiagnostics(), test.diagnostics...)
			result := terminalResultTable(code)
			if result == nil || result.IsOpen() {
				t.Fatalf("result type = %#v, want closed table", code.ResultType)
			}
			assertOrderedColumns(t, result, "Column1")
			if column := result.GetColumn("Column1"); column == nil || column.Type.SymbolName() != "dynamic" {
				t.Fatalf("Column1 = %#v, want dynamic", column)
			}
		})
	}
}
