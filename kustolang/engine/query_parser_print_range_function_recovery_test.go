package kustolang

import "testing"

func TestPrintRangeFunctionArgumentRecoveryCurrentReferenceParity(t *testing.T) {
	tests := []struct {
		query string
		want  []diagnosticLocation
	}{
		{
			query: `print X=column_names_of(range A from 1 to 2 step 1)`,
			want: []diagnosticLocation{
				{"KS119", 8, 15}, {"KS117", 24, 5}, {"KS005", 30, 1},
				{"KS142", 30, 1}, {"KS005", 32, 4}, {"KS142", 32, 4},
				{"KS005", 37, 1}, {"KS005", 39, 2}, {"KS198", 39, 2},
			},
		},
		{
			query: `print X=strlen(range A from 1 to 2 step 1)`,
			want: []diagnosticLocation{
				{"KS119", 8, 6}, {"KS117", 15, 5}, {"KS005", 21, 1},
				{"KS142", 21, 1}, {"KS005", 23, 4}, {"KS142", 23, 4},
				{"KS005", 28, 1}, {"KS005", 30, 2}, {"KS198", 30, 2},
			},
		},
		{
			query: `print X=foo(range A from 1 to 2 step 1)`,
			want: []diagnosticLocation{
				{"KS211", 8, 3}, {"KS117", 12, 5}, {"KS005", 18, 1},
				{"KS142", 18, 1}, {"KS005", 20, 4}, {"KS142", 20, 4},
				{"KS005", 25, 1}, {"KS005", 27, 2}, {"KS198", 27, 2},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			assertDiagnosticLocations(t, ParseAndAnalyze(test.query).GetDiagnostics(), test.want...)
		})
	}
}
