package kustolang

import "testing"

func TestMakeSeriesMalformedDateTimeRangeUsesCurrentSourceTypeDiagnostic(t *testing.T) {
	tests := []struct {
		name  string
		query string
		start int
	}{
		{
			name:  "malformed from invocation",
			query: `datatable(Timestamp:datetime,Value:long)[datetime(2024-01-01),1] | make-series Total=sum(Value) default=0 on Timestamp from datetime2024-01-01) to datetime(2024-01-02) step 1h`,
			start: 124,
		},
		{
			name:  "malformed to invocation",
			query: `datatable(Timestamp:datetime,Value:long)[datetime(2024-01-01),1] | make-series Total=sum(Value) default=0 on Timestamp from datetime(2024-01-01) to datetime2024-01-02) step 1h`,
			start: 148,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			diagnostics := ParseAndAnalyze(test.query).GetDiagnostics()
			if len(diagnostics) != 5 {
				t.Fatalf("diagnostics = %#v, want 5", diagnostics)
			}
			wantCodes := []string{"KS141", "KS142", "KS005", "KS006", "KS198"}
			for i, want := range wantCodes {
				if diagnostics[i].Code != want {
					t.Fatalf("diagnostic %d = %#v, want code %s", i, diagnostics[i], want)
				}
			}
			if diagnostics[0].Message != "The expression must have the type datetime." || diagnostics[0].Start != test.start || diagnostics[0].Length != 18 {
				t.Fatalf("type diagnostic = %#v", diagnostics[0])
			}
		})
	}
}
