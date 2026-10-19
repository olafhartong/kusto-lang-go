package kustolang

import "testing"

func TestCurrentInlineExternalTableUnsupportedPathFormatTokenRecoversWithoutPanic(t *testing.T) {
	code := ParseAndAnalyze(`inline_external_table(a:long) kind=storage pathformat=(?) dataformat=csv ("u")`)
	want := []Diagnostic{
		DiagnosticUnknownTokenInPathFormatDefinition().WithLocation(55, 0),
		DiagnosticMissingValue().WithLocation(55, 0),
		DiagnosticMissingString().WithLocation(55, 0),
	}
	assertInlineExternalTableParityDiagnostics(t, code.GetDiagnostics(), want)

	result, ok := TableSymbolForType(code.ResultType)
	if !ok {
		t.Fatalf("result type = %#v, want table", code.ResultType)
	}
	assertOrderedColumns(t, result, "a")
	if result.Columns()[0].Type != ScalarTypes.Long.Type() || result.IsOpen() {
		t.Fatalf("result = %#v, want closed (a:long)", result)
	}
}

func TestCurrentInlineExternalTableInvalidKindPreservesPinnedRecovery(t *testing.T) {
	code := ParseAndAnalyze(`inline_external_table(a:long) kind=bogus dataformat=csv ("u")`)
	want := []Diagnostic{
		DiagnosticWrongDataStreamType("").WithLocation(35, 0),
		DiagnosticTermsExpected("storage", "delta").WithLocation(35, 5),
		DiagnosticMissingDataFormat().WithLocation(35, 0),
		DiagnosticMissingConnectionStrings().WithLocation(35, 0),
		DiagnosticTermsExpected(";").WithLocation(35, 5),
		DiagnosticNameDoesNotReferToAnyKnownItem("bogus").WithLocation(35, 5),
		DiagnosticTermsExpected(";").WithLocation(41, 10),
		DiagnosticNameDoesNotReferToAnyKnownItem("dataformat").WithLocation(41, 10),
		DiagnosticIncompleteFragment().WithLocation(51, 1),
	}
	assertInlineExternalTableParityDiagnostics(t, code.GetDiagnostics(), want)

	if code.ResultType != &Error.TypeSymbol {
		t.Fatalf("result type = %#v, want error", code.ResultType)
	}
}

func TestCurrentInlineExternalTableRetainsRobustnessForPinnedCrashCases(t *testing.T) {
	queries := []string{
		`inline_external_table(a:long) kind=storage partition by (p:) pathformat=(p) dataformat=csv ("u")`,
		`inline_external_table(a:long) kind=storage partition by (p:datetime) pathformat=(datetime_pattern(,p)) dataformat=csv ("u")`,
	}
	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
			code := ParseAndAnalyze(query)
			if diagnostics := code.GetDiagnostics(); len(diagnostics) == 0 {
				t.Fatal("diagnostics are empty for malformed query")
			}
		})
	}
}

func assertInlineExternalTableParityDiagnostics(t *testing.T, got, want []Diagnostic) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("diagnostics = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}
