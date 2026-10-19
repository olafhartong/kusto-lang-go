package kustolang

import "testing"

func TestCurrentDetachedAndMalformedGraphOperatorReferenceParity(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		diagnostics []diagnosticLocation
		columns     []string
	}{
		{
			name:  "detached graph-match",
			query: `let G=datatable(Source:long,Target:long,Weight:long)[1,2,3] | make-graph Source --> Target; G  graph-match (a)-[e]->(b) where e.Weight > 0 project Src=a.Source, Weight=e.Weight`,
			diagnostics: []diagnosticLocation{
				{"KS005", 93, 0}, {"KS174", 95, 11}, {"KS219", 95, 11},
				{"KS142", 128, 6}, {"KS142", 153, 6}, {"KS142", 170, 6},
			},
			columns: []string{"Src", "Weight"},
		},
		{
			name:  "detached graph-shortest-paths",
			query: `let G=datatable(Source:long,Target:long,Weight:long)[1,2,3] | make-graph Source --> Target; G  graph-shortest-paths (a)-[e]->(b) where e.Weight > 0 project Src=a.Source, Dst=b.Target`,
			diagnostics: []diagnosticLocation{
				{"KS005", 93, 0}, {"KS174", 95, 20}, {"KS219", 95, 20},
				{"KS142", 137, 6}, {"KS142", 162, 6}, {"KS142", 176, 6},
			},
			columns: []string{"Src", "Dst"},
		},
		{
			name:  "graph-match missing edge close",
			query: `let G=datatable(Source:long,Target:long,Weight:long)[1,2,3] | make-graph Source --> Target; G | graph-match (a)-[e->(b) where e.Weight > 0 project Src=a.Source, Weight=e.Weight`,
			diagnostics: []diagnosticLocation{
				{"KS238", 111, 3}, {"KS005", 114, 1}, {"KS005", 114, 1}, {"KS006", 115, 0},
				{"KS142", 117, 1}, {"KS005", 120, 5}, {"KS174", 120, 5}, {"KS142", 126, 1},
				{"KS005", 139, 7}, {"KS174", 139, 7}, {"KS142", 151, 1}, {"KS142", 168, 1},
			},
			columns: []string{"Src", "Weight"},
		},
		{
			name:  "graph-shortest-paths missing edge close",
			query: `let G=datatable(Source:long,Target:long,Weight:long)[1,2,3] | make-graph Source --> Target; G | graph-shortest-paths (a)-[e->(b) where e.Weight > 0 project Src=a.Source, Dst=b.Target`,
			diagnostics: []diagnosticLocation{
				{"KS238", 120, 3}, {"KS005", 123, 1}, {"KS005", 123, 1}, {"KS006", 124, 0},
				{"KS142", 126, 1}, {"KS005", 129, 5}, {"KS174", 129, 5}, {"KS142", 135, 1},
				{"KS005", 148, 7}, {"KS174", 148, 7}, {"KS142", 160, 1}, {"KS142", 174, 1},
			},
			columns: []string{"Src", "Dst"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := ParseAndAnalyze(tt.query)
			assertDiagnosticLocations(t, code.GetDiagnostics(), tt.diagnostics...)
			assertOrderedColumns(t, terminalResultTable(code), tt.columns...)
		})
	}
}
