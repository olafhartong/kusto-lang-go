package kustolang

import "testing"

type topLevelRecoveryDiagnosticLocation struct {
	code   string
	start  int
	length int
}

func TestCurrentTopLevelRecognizedStatementRecoveryRetainsDiagnostics(t *testing.T) {
	for _, test := range []struct {
		name        string
		query       string
		diagnostics []topLevelRecoveryDiagnosticLocation
	}{
		{
			name:  "concatenated query parameters keyword",
			query: `declare query_parametersUser:string="guest", Limit:long=10); print User=User, Limit=Limit`,
			diagnostics: []topLevelRecoveryDiagnosticLocation{
				{"KS005", 8, 20},
				{"KS006", 29, 0},
				{"KS198", 29, 6},
			},
		},
		{
			name:  "missing query parameter",
			query: `declare query_parameters(x:long,,y:string); print y`,
			diagnostics: []topLevelRecoveryDiagnosticLocation{
				{"KS006", 32, 0},
				{"KS005", 32, 1},
				{"KS198", 32, 1},
			},
		},
		{
			name:  "restrict without access keyword",
			query: `restrict to (T); print x=1`,
			diagnostics: []topLevelRecoveryDiagnosticLocation{
				{"KS198", 9, 2},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.query)
			assertTopLevelRecoveryDiagnosticLocations(t, GetContainedDiagnostics(query, DiagnosticsIncludeSyntactic), test.diagnostics)
			if query.SkippedTokens() == nil {
				t.Fatal("recovered tail was not retained as skipped tokens")
			}
		})
	}
}

func assertTopLevelRecoveryDiagnosticLocations(t *testing.T, got []Diagnostic, want []topLevelRecoveryDiagnosticLocation) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("diagnostics = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i].Code != want[i].code || got[i].Start != want[i].start || got[i].Length != want[i].length {
			t.Fatalf("diagnostic %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestTopLevelRecoveryGuardsRemainNarrow(t *testing.T) {
	for _, text := range []string{
		`restrict from (T); print x=1`,
		`declare unrelated:string; print x=1`,
	} {
		query := ParseQuery(text)
		if query.SkippedTokens() != nil {
			t.Fatalf("%q unexpectedly retained a skipped-token tail", text)
		}
	}
}
