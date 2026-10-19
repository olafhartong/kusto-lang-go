package kustolang

import "testing"

func TestCurrentDateTimeKeywordTailPreservesParsedSourcePrefix(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantCodes   []string
		wantColumns []string
	}{
		{
			name:        "datatable value tail",
			query:       `datatable(Timestamp:datetime,Value:long)[datetime`,
			wantCodes:   []string{"KS005", "KS198"},
			wantColumns: []string{"Timestamp", "Value"},
		},
		{
			name:        "make-series from tail",
			query:       `datatable(Timestamp:datetime,Value:long)[datetime(2024-01-01),1] | make-series Total=sum(Value) default=0 on Timestamp from datetime`,
			wantCodes:   []string{"KS006", "KS005", "KS198"},
			wantColumns: []string{"Total", "Timestamp"},
		},
		{
			name:        "make-series to tail",
			query:       `datatable(Timestamp:datetime,Value:long)[datetime(2024-01-01),1] | make-series Total=sum(Value) default=0 on Timestamp from datetime(2024-01-01) to datetime`,
			wantCodes:   []string{"KS006", "KS005", "KS198"},
			wantColumns: []string{"Total", "Timestamp"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.query)
			diagnostics := code.GetDiagnostics()
			if len(diagnostics) != len(test.wantCodes) {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for i, want := range test.wantCodes {
				if diagnostics[i].Code != want {
					t.Fatalf("diagnostic %d = %#v, want %s", i, diagnostics[i], want)
				}
			}
			table, ok := TableSymbolForType(code.ResultType)
			if !ok {
				t.Fatalf("result type = %v, want table", code.ResultType)
			}
			assertOrderedColumns(t, table, test.wantColumns...)
			if code.Query.SkippedTokens() == nil {
				t.Fatal("trailing datetime keyword was not retained as skipped syntax")
			}
		})
	}
}
