package kustolang

import "testing"

func TestCurrentCompoundStringDataTableTopNestedRecoveryParity(t *testing.T) {
	query := `datatable(Country:string,City:string,Value:long)['NL''AMS',1] | top-nested 2 of Country by sum(Value), top-nested 2 of City by sum(Value)`
	code := ParseAndAnalyze(query)
	diagnostics := code.GetDiagnostics()
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS173" || diagnostics[0].Start != 49 || diagnostics[0].Length != 11 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	table, ok := TableSymbolForType(code.ResultType)
	if !ok {
		t.Fatalf("result type = %v, want table", code.ResultType)
	}
	assertOrderedColumns(t, table, "Country", "aggregated_Country", "City", "aggregated_City")
}

func TestCurrentSerializeDetachedScanTailRecoveryParity(t *testing.T) {
	query := `datatable(Key:long,Value:long)[1,10,2,20] | serialize  scan declare (state:long=0) with (step s: true => state=state + 1;) | project Key, state`
	code := ParseAndAnalyze(query)
	wantCodes := []string{"KS142", "KS005", "KS211", "KS142", "KS006", "KS005", "KS228", "KS005", "KS211", "KS142", "KS005", "KS142", "KS005", "KS198"}
	diagnostics := code.GetDiagnostics()
	if len(diagnostics) != len(wantCodes) {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for i, want := range wantCodes {
		if diagnostics[i].Code != want {
			t.Fatalf("diagnostic %d = %#v, want %s", i, diagnostics[i], want)
		}
	}
	table, ok := TableSymbolForType(code.ResultType)
	if !ok {
		t.Fatalf("result type = %v, want table", code.ResultType)
	}
	assertOrderedColumns(t, table, "Key", "Value", "scan", "Column1", "Column2")
}
