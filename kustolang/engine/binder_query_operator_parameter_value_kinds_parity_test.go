package kustolang

import "testing"

func TestNativeQueryOperatorParameterSyntheticMetadataParity(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		declaration *QueryOperatorParameter
		wantCodes   []string
	}{
		{
			name:        "unconstrained word accepts token",
			value:       "anything",
			declaration: NewQueryOperatorParameter("hint.strategy", QueryOperatorParameterValueWord),
		},
		{
			name:        "summable datetime",
			value:       "datetime(2024-01-01)",
			declaration: NewQueryOperatorParameter("hint.num_partitions", QueryOperatorParameterValueSummableLiteral),
		},
		{
			name:        "summable bool rejected",
			value:       "true",
			declaration: NewQueryOperatorParameter("hint.num_partitions", QueryOperatorParameterValueSummableLiteral),
			wantCodes:   []string{"KS224"},
		},
		{
			name:        "nonword allowed literal",
			value:       "2",
			declaration: NewQueryOperatorParameter("hint.num_partitions", QueryOperatorParameterValueIntegerLiteral, values([]string{"1"})),
			wantCodes:   []string{"KS140"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery("datatable(x:long)[1] | distinct " + test.declaration.SymbolName() + "=" + test.value + " x")
			binder := NewBinder(nil)
			binder.TryBind(query)
			parameters := firstBoundNode[*DistinctOperator](t, query).Parameters()
			diagnostics := queryOperatorParameterDiagnostics(binder, parameters, []*QueryOperatorParameter{test.declaration})
			got := make([]string, len(diagnostics))
			for i := range diagnostics {
				got[i] = diagnostics[i].Code
			}
			if len(got) != len(test.wantCodes) {
				t.Fatalf("diagnostic codes = %v, want %v; diagnostics=%#v", got, test.wantCodes, diagnostics)
			}
			for i := range got {
				if got[i] != test.wantCodes[i] {
					t.Fatalf("diagnostic codes = %v, want %v", got, test.wantCodes)
				}
			}
		})
	}
}

func TestNativeTakeAndSampleRejectParsedParametersWithWarning(t *testing.T) {
	for _, operator := range []string{"take", "sample"} {
		query := ParseQuery("datatable(x:long)[1] | " + operator + " hint.strategy=shuffle 1")
		NewBinder(nil).TryBind(query)
		var diagnostics []Diagnostic
		if operator == "take" {
			diagnostics = firstBoundNode[*TakeOperator](t, query).SemanticDiagnostics()
		} else {
			diagnostics = firstBoundNode[*SampleOperator](t, query).SemanticDiagnostics()
		}
		if len(diagnostics) != 1 || diagnostics[0].Code != "KS179" || diagnostics[0].Severity != DiagnosticSeverityWarning {
			t.Fatalf("%s diagnostics = %#v", operator, diagnostics)
		}
	}
}
