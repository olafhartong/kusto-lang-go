package kustolang

import "testing"

func TestSummarizeBindsDynamicSelectorOnFunctionResult(t *testing.T) {
	table := NewTableSymbol("AWSCloudTrail", NewColumnSymbol("RequestParameters", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	query := `AWSCloudTrail | summarize n=countif(isnotnull(parse_json(RequestParameters).peerOwnerId)), m=countif(isnotnull((parse_json(RequestParameters).launchPermission).add.items))`
	code := ParseWithGlobals(query, globals).Analyze()
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid dynamic function-result selectors produced diagnostics: %#v", diagnostics)
	}
}
