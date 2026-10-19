package kustolang

import "testing"

func TestNativePartitionByRecoveredSyntaxPublishesTableResult(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		columns []string
	}{
		{name: "operator only", query: `datatable(Tenant:string,Value:long)["a",1] | __partitionby`},
		{name: "missing entity", query: `datatable(Tenant:string,Value:long)["a",1] | __partitionby (summarize Total=sum(Value))`, columns: []string{"Total"}},
		{name: "missing operand", query: `datatable(Tenant:string,Value:long)["a",1] | __partitionby Tenant`},
		{name: "empty subquery", query: `datatable(Tenant:string,Value:long)["a",1] | __partitionby Tenant ()`},
		{name: "truncated summarize", query: `datatable(Tenant:string,Value:long)["a",1] | __partitionby Tenant (summarize`},
		{name: "truncated named aggregate", query: `datatable(Tenant:string,Value:long)["a",1] | __partitionby Tenant (summarize Total`, columns: []string{"Total"}},
		{name: "truncated function name", query: `datatable(Tenant:string,Value:long)["a",1] | __partitionby Tenant (summarize Total=sum`, columns: []string{"Total"}},
		{name: "truncated function argument", query: `datatable(Tenant:string,Value:long)["a",1] | __partitionby Tenant (summarize Total=sum(Value`, columns: []string{"Total"}},
		{name: "missing outer close", query: `datatable(Tenant:string,Value:long)["a",1] | __partitionby Tenant (summarize Total=sum(Value)`, columns: []string{"Total"}},
		{name: "missing open", query: `datatable(Tenant:string,Value:long)["a",1] | __partitionby Tenant summarize Total=sum(Value))`, columns: []string{"Total"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.query)
			if len(code.GetSyntaxDiagnostics()) == 0 {
				t.Fatal("malformed query did not retain syntax diagnostics")
			}
			result := terminalResultTable(code)
			if result == nil {
				t.Fatalf("result type = %#v, want table", code.ResultType)
			}
			if result.IsOpen() {
				t.Fatal("recovered partition-by result is open")
			}
			assertOrderedColumns(t, result, test.columns...)
		})
	}
}
