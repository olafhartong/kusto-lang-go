package kustolang

import "testing"

type unionPrintRecoveryDiagnostic struct {
	code    string
	message string
	start   int
	length  int
}

func TestNativeUnionRecoveredOperandsCurrentReferenceParity(t *testing.T) {
	left := NewTableSymbol("Left", NewColumnSymbol("LeftOnly", ScalarTypes.Long.Type()))
	right := NewTableSymbol("Right", NewColumnSymbol("RightOnly", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", left, right))

	tests := []struct {
		name        string
		query       string
		columns     []string
		diagnostics []unionPrintRecoveryDiagnostic
	}{
		{
			name:    "trailing open parenthesis",
			query:   "union Left, (",
			columns: []string{"LeftOnly"},
			diagnostics: []unionPrintRecoveryDiagnostic{
				{"KS006", "Missing expression", 13, 0},
				{"KS005", "Expected: )", 13, 0},
			},
		},
		{
			name:    "trailing open bracket",
			query:   "union Left, [",
			columns: []string{"LeftOnly"},
			diagnostics: []unionPrintRecoveryDiagnostic{
				{"KS006", "Missing name", 13, 0},
				{"KS005", "Expected: ]", 13, 0},
			},
		},
		{
			name:    "malformed first operand",
			query:   "union (, Left, Right",
			columns: []string{"LeftOnly", "RightOnly"},
			diagnostics: []unionPrintRecoveryDiagnostic{
				{"KS006", "Missing expression", 7, 0},
				{"KS005", "Expected: )", 7, 1},
			},
		},
		{
			name:    "malformed second operand",
			query:   "union Left, [, Right",
			columns: []string{"LeftOnly", "RightOnly"},
			diagnostics: []unionPrintRecoveryDiagnostic{
				{"KS006", "Missing name", 13, 0},
				{"KS005", "Expected: ]", 13, 1},
			},
		},
		{
			name:    "missing withsource equals",
			query:   "union withsource Source Left, Right",
			columns: []string{"LeftOnly", "RightOnly"},
			diagnostics: []unionPrintRecoveryDiagnostic{
				{"KS204", "The name 'withsource' does not refer to any known table, tabular variable or function.", 6, 10},
				{"KS005", "Expected: ,", 17, 6},
				{"KS204", "The name 'Source' does not refer to any known table, tabular variable or function.", 17, 6},
				{"KS005", "Expected: ,", 24, 4},
			},
		},
		{
			name:    "valid control",
			query:   "union withsource=Source Left, Right",
			columns: []string{"Source", "LeftOnly", "RightOnly"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.query, globals)
			assertUnionPrintRecoveryDiagnostics(t, code.GetDiagnostics(), test.diagnostics)
			assertOrderedColumns(t, terminalResultTable(code), test.columns...)
		})
	}
}

func TestNativePrintTrailingBracketCurrentReferenceParity(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		columns     []string
		diagnostics []unionPrintRecoveryDiagnostic
	}{
		{
			name:    "trailing open bracket",
			query:   "print First=1, [",
			columns: []string{"First", "print_1"},
			diagnostics: []unionPrintRecoveryDiagnostic{
				{"KS006", "Missing name", 16, 0},
				{"KS005", "Expected: ]", 16, 0},
			},
		},
		{
			name:    "valid control",
			query:   "print First=1, Second=2",
			columns: []string{"First", "Second"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.query)
			assertUnionPrintRecoveryDiagnostics(t, code.GetDiagnostics(), test.diagnostics)
			table := terminalResultTable(code)
			assertOrderedColumns(t, table, test.columns...)
			if test.name == "trailing open bracket" && !sameType(table.GetColumn("print_1").Type, ScalarTypes.Unknown.Type()) {
				t.Fatalf("print_1 type = %s, want unknown", table.GetColumn("print_1").Type.SymbolName())
			}
		})
	}
}

func assertUnionPrintRecoveryDiagnostics(t *testing.T, got []Diagnostic, want []unionPrintRecoveryDiagnostic) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("diagnostics = %#v, want %#v", got, want)
	}
	for i, expected := range want {
		diagnostic := got[i]
		if diagnostic.Code != expected.code || diagnostic.Message != expected.message ||
			diagnostic.LocationKind != DiagnosticLocationAbsolute || diagnostic.Start != expected.start || diagnostic.Length != expected.length {
			t.Fatalf("diagnostic %d = %#v, want %s %q Absolute %d:%d", i, diagnostic,
				expected.code, expected.message, expected.start, expected.length)
		}
	}
}
