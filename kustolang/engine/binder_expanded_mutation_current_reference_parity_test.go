package kustolang

import "testing"

func TestExpandedMutationMalformedRecoveryCurrentReferenceParity(t *testing.T) {
	for _, test := range []struct {
		name        string
		query       string
		diagnostics []diagnosticLocation
		columns     map[string]*TypeSymbol
	}{
		{
			name:  "standalone facet nested pipeline",
			query: `datatable(a:string,b:long)['x',1]  facet by a with (summarize c=count() by b)`,
			diagnostics: []diagnosticLocation{
				{"KS005", 33, 0}, {"KS174", 35, 5}, {"KS142", 44, 1}, {"KS142", 75, 1},
			},
		},
		{
			name:  "standalone top nested span",
			query: `datatable(a:string,b:long)['x',1]  top-nested 2 of a by sum(b)`,
			diagnostics: []diagnosticLocation{
				{"KS005", 33, 0}, {"KS174", 35, 27}, {"KS142", 51, 1}, {"KS142", 60, 1},
			},
			columns: map[string]*TypeSymbol{"a": ScalarTypes.Unknown.Type(), "aggregated_a": ScalarTypes.Unknown.Type()},
		},
		{
			name:        "project empty bracket selector",
			query:       `datatable(a:string,b:long)['x',1] | search 'x' | project a [`,
			diagnostics: []diagnosticLocation{{"KS006", 60, 0}, {"KS005", 60, 0}},
			columns:     map[string]*TypeSymbol{"Column1": ScalarTypes.Unknown.Type()},
		},
		{
			name:        "top nested trailing clause",
			query:       `datatable(a:string,b:long)['x',1] | top-nested 2 of a by sum(b) ,`,
			diagnostics: []diagnosticLocation{{"KS006", 65, 0}},
			columns: map[string]*TypeSymbol{
				"a": ScalarTypes.String.Type(), "aggregated_a": ScalarTypes.Long.Type(),
				"": ScalarTypes.Unknown.Type(), "aggregated_": ScalarTypes.Unknown.Type(),
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.query)
			NewBinder(nil).TryBind(query)
			assertDiagnosticLocations(t, GetContainedDiagnostics(query, DiagnosticsIncludeDefault), test.diagnostics...)
			if test.columns == nil {
				return
			}
			statement := query.Statements().At(query.Statements().Count() - 1).Element().(*ExpressionStatement)
			result, ok := TableSymbolForType(ExpressionResultType(statement.Expression()))
			if !ok {
				t.Fatalf("query result = %#v, want table", ExpressionResultType(statement.Expression()))
			}
			if len(result.Columns()) != len(test.columns) {
				t.Fatalf("columns = %#v, want %d", result.Columns(), len(test.columns))
			}
			assertColumnTypes(t, result, test.columns)
		})
	}
}
