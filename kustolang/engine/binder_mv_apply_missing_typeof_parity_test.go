package kustolang

import "testing"

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func TestNativeMvApplyMissingTypeOfContinuesRecovery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{
			name:  "missing typeof literal",
			query: "datatable(x:dynamic)[dynamic([1,2])] | mv-apply x to",
			want:  []string{"KS006", "KS005", "KS176", "KS005", "KS006", "KS005"},
		},
		{
			name:  "bare typeof token remains a recovered expression",
			query: "datatable(x:dynamic)[dynamic([1,2])] | mv-apply x to typeof",
			want:  []string{"KS006", "KS005", "KS142", "KS005", "KS176", "KS005", "KS006", "KS005"},
		},
		{
			name:  "identifier beginning with typeof remains a recovered expression",
			query: "datatable(x:dynamic)[dynamic([1,2])] | mv-apply x to typeoflong) on (summarize c=count())",
			want:  []string{"KS006", "KS005", "KS142", "KS005", "KS176", "KS005", "KS006", "KS005", "KS198"},
		},
		{
			name:  "valid type literal",
			query: "datatable(x:dynamic)[dynamic([1,2])] | mv-apply x to typeof(long) on (summarize c=count())",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := ParseAndAnalyze(tt.query)
			got := mvApplyMissingTypeOfDiagnosticCodes(code.GetDiagnostics())
			if !equalStringSlices(got, tt.want) {
				t.Fatalf("diagnostic codes = %v, want %v; diagnostics = %#v", got, tt.want, code.GetDiagnostics())
			}
		})
	}
}

func mvApplyMissingTypeOfDiagnosticCodes(diagnostics []Diagnostic) []string {
	codes := make([]string, len(diagnostics))
	for i, diagnostic := range diagnostics {
		codes[i] = diagnostic.Code
	}
	return codes
}
