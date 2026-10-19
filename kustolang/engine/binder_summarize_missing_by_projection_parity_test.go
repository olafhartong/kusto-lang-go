package kustolang

import "testing"

func TestNativeSummarizeMissingByExpressionDoesNotCreateColumn(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		columns []string
	}{
		{
			name:    "missing first by expression",
			query:   `range x from 1 to 5 step 1 | summarize s=sum(x) by`,
			columns: []string{"s"},
		},
		{
			name:    "missing trailing by expression",
			query:   `range x from 1 to 5 step 1 | summarize s=sum(x) by x,`,
			columns: []string{"x", "s"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.query)
			assertOrderedColumns(t, terminalResultTable(code), test.columns...)
			assertDiagnosticLocations(t, code.GetDiagnostics(), diagnosticLocation{"KS006", len(test.query), 0})
		})
	}
}

func TestNativeSummarizeNonemptyByRecoveryControls(t *testing.T) {
	code := ParseAndAnalyze(`range x from 1 to 5 step 1 | summarize s=sum(x) by Missing`)
	assertOrderedColumns(t, terminalResultTable(code), "Missing", "s")
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 1 || diagnostics[0].Code != "KS142" {
		t.Fatalf("diagnostics = %#v, want one unresolved-name diagnostic", diagnostics)
	}
}
