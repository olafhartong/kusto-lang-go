package kustolang

import "testing"

func TestNativePartitionDoesNotAddColumnExpectedForErrorTypedKey(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{
			name:  "malformed parameter is recovered as key",
			query: "datatable(x:long)[1,2,3] | partition hint.strategynative by x (summarize c=count())",
			want:  []string{"KS176", "KS005", "KS142", "KS005", "KS006", "KS005", "KS198"},
		},
		{
			name:  "unresolved key",
			query: "datatable(xlong)[1,2,3] | partition hint.strategy=native by x (summarize c=count())",
			want:  []string{"KS005", "KS006", "KS142"},
		},
		{
			name:  "non-column scalar remains invalid",
			query: "let y=1; datatable(x:long)[1,2,3] | partition by y (summarize c=count())",
			want:  []string{"KS109"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := ParseAndAnalyze(tt.query)
			got := partitionDiagnosticCodes(code.GetDiagnostics())
			if !equalStringSlices(got, tt.want) {
				t.Fatalf("diagnostic codes = %v, want %v; diagnostics = %#v", got, tt.want, code.GetDiagnostics())
			}
		})
	}
}

func partitionDiagnosticCodes(diagnostics []Diagnostic) []string {
	codes := make([]string, len(diagnostics))
	for i, diagnostic := range diagnostics {
		codes[i] = diagnostic.Code
	}
	return codes
}
