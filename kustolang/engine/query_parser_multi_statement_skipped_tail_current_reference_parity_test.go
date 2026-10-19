package kustolang

import "testing"

func TestCurrentMultiStatementSkippedTailRetainsDiagnostics(t *testing.T) {
	for _, test := range []struct {
		query       string
		diagnostics []diagnosticLocation
	}{
		{
			query: `distinct a, b`,
			diagnostics: []diagnosticLocation{
				{"KS142", 0, 8},
				{"KS005", 9, 1},
				{"KS142", 9, 1},
				{"KS198", 10, 1},
			},
		},
		{
			query: `parse s with a:string '=' b:long`,
			diagnostics: []diagnosticLocation{
				{"KS142", 0, 5},
				{"KS005", 6, 1},
				{"KS142", 6, 1},
				{"KS005", 8, 4},
				{"KS142", 8, 4},
				{"KS005", 13, 1},
				{"KS142", 13, 1},
				{"KS006", 15, 0},
				{"KS198", 15, 6},
			},
		},
	} {
		t.Run(test.query, func(t *testing.T) {
			query := ParseQuery(test.query)
			NewBinder(NewGlobalState(nil)).TryBind(query)
			assertDiagnosticLocations(t, GetContainedDiagnostics(query, DiagnosticsIncludeDefault), test.diagnostics...)
			if query.SkippedTokens() == nil {
				t.Fatal("recovered tail was not retained as skipped tokens")
			}
		})
	}
}
