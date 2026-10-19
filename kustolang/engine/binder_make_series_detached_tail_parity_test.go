package kustolang

import "testing"

func TestNativeMalformedMakeSeriesDetachedTailUsesScalarNameResolution(t *testing.T) {
	query := "datatable(x:long)[1,2,3] | make-series c=count) default=0 on x from 1 to 3 step 1"
	code := ParseAndAnalyze(query)
	wantCodes := []string{"KS117", "KS005", "KS006", "KS005", "KS005", "KS005", "KS005", "KS142", "KS198"}
	gotCodes := make([]string, len(code.GetDiagnostics()))
	for i, diagnostic := range code.GetDiagnostics() {
		gotCodes[i] = diagnostic.Code
	}
	if !equalStringSlices(gotCodes, wantCodes) {
		t.Fatalf("diagnostic codes = %v, want %v; diagnostics = %#v", gotCodes, wantCodes, code.GetDiagnostics())
	}
}
