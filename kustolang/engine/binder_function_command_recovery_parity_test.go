package kustolang

import "testing"

func TestFunctionCommandRecoveryMatchesCurrentReference(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantCodes  []string
		wantResult bool
	}{
		{
			name:       "valid create",
			query:      `.create function F () { print x=1 }`,
			wantResult: true,
		},
		{
			name:      "view rejected with skipped command tail",
			query:     `.create function F view () { print x=1 }`,
			wantCodes: []string{"KS005", "KS005", "KS006", "KS005", "KS005", "KS006", "KS005"},
		},
		{
			name:      "partial command has no result",
			query:     `.add function F admins (`,
			wantCodes: []string{"KS142", "KS005"},
		},
		{
			name:       "unknown altered function",
			query:      `.alter function F (x:string) { print y=x }`,
			wantCodes:  []string{"KS142"},
			wantResult: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.query)
			assertKustoCodeDiagnosticCodes(t, code, test.wantCodes...)
			if test.wantResult {
				if _, ok := TableSymbolForType(code.ResultType); !ok {
					t.Fatalf("result type = %#v, want command result table", code.ResultType)
				}
			} else if code.ResultType != nil {
				t.Fatalf("result type = %#v, want nil", code.ResultType)
			}
		})
	}
}

func TestFunctionCommandPropertyValueExpectedTermParity(t *testing.T) {
	code := ParseAndAnalyze(".create function ifnotexists with (folder=")
	diagnostics := code.GetDiagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one diagnostic", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Code != "KS005" || diagnostic.Message != "Expected: <literal>" || diagnostic.Start != 42 || diagnostic.Length != 0 {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}
