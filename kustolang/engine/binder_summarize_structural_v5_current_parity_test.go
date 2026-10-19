package kustolang

import "testing"

func TestNativeSummarizeTrailingMissingAggregateDoesNotCreateColumn(t *testing.T) {
	const text = `datatable(Key:string,Value:long)['a',1,'b',2] | summarize Total=sum(Value),`
	code := ParseAndAnalyze(text)
	diagnostics := code.GetDiagnostics()
	assertDiagnosticLocations(t, diagnostics, diagnosticLocation{"KS006", 75, 0})
	result, ok := TableSymbolForType(code.ResultType)
	if !ok {
		t.Fatalf("result = %#v", code.ResultType)
	}
	assertOrderedColumns(t, result, "Total")
}

func TestNativeIncompleteAggregateCallOutsideAggregateScopeReportsKS144(t *testing.T) {
	tests := []struct {
		text  string
		name  string
		start int
	}{
		{`range X from 1 to 10 step 1 | sample-distinct 3 of X | count (`, "count", 55},
		{`datatable(Key:string,Value:long)['a',1,'b',2] | summarize Total=sum(Value, Maximum=max(Value) by Key | sort by Total desc nulls last | take 5`, "max", 83},
	}
	for _, test := range tests {
		code := ParseAndAnalyze(test.text)
		found := false
		for _, diagnostic := range code.GetDiagnostics() {
			if diagnostic.Code == "KS144" && diagnostic.Start == test.start && diagnostic.Message == "The aggregate function '"+test.name+"' is not allowed in this context." {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%q diagnostics = %#v", test.text, code.GetDiagnostics())
		}
	}
}
