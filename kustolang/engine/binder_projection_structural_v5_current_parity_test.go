package kustolang

import (
	"reflect"
	"testing"
)

func TestNativeStructuralMutationV5ProjectionRecoveryParity(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		diagnostics []diagnosticLocation
		columns     map[string]*TypeSymbol
	}{
		{
			name:  "detached project keep retains downstream recovery",
			query: `datatable(A:long,B:long)[1,2,3,4] | extend C=A+B  project-keep A,C | distinct A,C | sample 1`,
			diagnostics: []diagnosticLocation{
				{"KS005", 48, 0}, {"KS174", 50, 12}, {"KS109", 63, 1}, {"KS109", 65, 1},
				{"KS142", 63, 1}, {"KS142", 65, 1}, {"KS142", 78, 1}, {"KS142", 80, 1},
			},
			columns: map[string]*TypeSymbol{"A": ScalarTypes.Unknown.Type(), "C": ScalarTypes.Unknown.Type()},
		},
		{
			name:  "unresolved project keep feeds distinct",
			query: `datatable(A:long,B:long)[1,2,3,4] | extend C=A+B | project-keep AC | distinct A,C | sample 1`,
			diagnostics: []diagnosticLocation{
				{"KS109", 64, 2}, {"KS142", 64, 2}, {"KS142", 78, 1}, {"KS142", 80, 1},
			},
			columns: map[string]*TypeSymbol{"A": ScalarTypes.Unknown.Type(), "C": ScalarTypes.Unknown.Type()},
		},
		{
			name:        "project reorder missing comma keeps schema",
			query:       `datatable(A:long,B:string,C:real)[1,'x',2.5] | project-away B | project-rename X=A | project-reorder X asc C desc`,
			diagnostics: []diagnosticLocation{{"KS005", 107, 1}},
			columns:     map[string]*TypeSymbol{"X": ScalarTypes.Long.Type(), "C": ScalarTypes.Real.Type()},
		},
		{
			name:  "malformed rename produces recoverable reorder column",
			query: `datatable(A:long,B:string,C:real)[1,'x',2.5] | project-away B | project-rename XA | project-reorder X asc, C desc`,
			diagnostics: []diagnosticLocation{
				{"KS110", 79, 2}, {"KS142", 79, 2}, {"KS142", 100, 1},
			},
			columns: map[string]*TypeSymbol{"A": ScalarTypes.Long.Type(), "C": ScalarTypes.Real.Type(), "X": ScalarTypes.Unknown.Type()},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.query)
			gotLocations := make([]diagnosticLocation, 0, len(code.GetDiagnostics()))
			for _, diagnostic := range code.GetDiagnostics() {
				gotLocations = append(gotLocations, diagnosticLocation{diagnostic.Code, diagnostic.Start, diagnostic.Length})
			}
			if !reflect.DeepEqual(gotLocations, test.diagnostics) {
				t.Fatalf("diagnostics = %#v, want %#v", gotLocations, test.diagnostics)
			}
			result, ok := TableSymbolForType(code.ResultType)
			if !ok {
				t.Fatalf("result = %#v", code.ResultType)
			}
			if len(result.Columns()) != len(test.columns) {
				t.Fatalf("columns = %#v, want %d", result.Columns(), len(test.columns))
			}
			assertColumnTypes(t, result, test.columns)
		})
	}
}
