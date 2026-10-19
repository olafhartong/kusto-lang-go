package kustolang

import "testing"

func TestNativeMakeSeriesMissingExpressionsDoNotCreateColumns(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		columns []string
	}{
		{name: "all missing", query: `datatable(x:long)[1,2,3] | make-series`, columns: nil},
		{name: "bare aggregate", query: `datatable(x:long)[1,2,3] | make-series c`, columns: []string{"c"}},
		{name: "bare function", query: `datatable(x:long)[1,2,3] | make-series c=count`, columns: []string{"c"}},
		{name: "unclosed call", query: `datatable(x:long)[1,2,3] | make-series c=count(`, columns: []string{"c"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.query)
			assertOrderedColumns(t, terminalResultTable(code), test.columns...)
			if len(code.GetDiagnostics()) == 0 {
				t.Fatal("malformed make-series has no diagnostics")
			}
		})
	}
}
