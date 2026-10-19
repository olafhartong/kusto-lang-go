package kustolang

import "testing"

func TestNativeMissingParameterEqualPreservesSyntaxDiagnosticOrder(t *testing.T) {
	query := "datatable(x:long)[1,2,3] | partition hint.strategy"
	code := ParseAndAnalyze(query)
	wantMessages := []string{
		"Query operator expected.",
		"Expected: =",
		"Expected one of: shuffle, native, legacy",
		"Expected: by",
		"Missing name",
		"Expected: (",
		"Missing expression",
		"Expected: )",
	}
	diagnostics := code.GetDiagnostics()
	if len(diagnostics) != len(wantMessages) {
		t.Fatalf("diagnostics = %#v, want messages %v", diagnostics, wantMessages)
	}
	for i, message := range wantMessages {
		if diagnostics[i].Message != message || diagnostics[i].Start != 50 || diagnostics[i].Length != 0 {
			t.Fatalf("diagnostic %d = %#v, want %q at 50:0", i, diagnostics[i], message)
		}
	}
}
