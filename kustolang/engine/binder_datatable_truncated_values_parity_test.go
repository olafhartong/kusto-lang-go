package kustolang

import (
	"strings"
	"testing"
)

func TestNativeDataTableTruncatedValuesRecoveryParity(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		values      int
		closeWidth  int
		diagnostics []diagnosticLocation
	}{
		{
			name:       "valid",
			query:      `datatable(k:long,a:string)[1,"x"]`,
			values:     2,
			closeWidth: 1,
		},
		{
			name:       "closed short row",
			query:      `datatable(k:long,a:string)[1]`,
			values:     1,
			closeWidth: 1,
			diagnostics: []diagnosticLocation{
				{"KS173", 27, 1},
			},
		},
		{
			name:       "truncated short row",
			query:      `datatable(k:long,a:string)[1`,
			values:     1,
			closeWidth: 0,
			diagnostics: []diagnosticLocation{
				{"KS173", 27, 1},
				{"KS005", 28, 0},
			},
		},
		{
			name:       "truncated empty values",
			query:      `datatable(k:long,a:string)[`,
			values:     0,
			closeWidth: 0,
			diagnostics: []diagnosticLocation{
				{"KS005", 27, 0},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.query)
			dataTable := queryBlockExpressionAtForTest(code.Query, 0).(*DataTableExpression)
			if dataTable.Values().Count() != test.values || dataTable.CloseBracket().Width() != test.closeWidth {
				t.Fatalf("datatable recovery = values:%d close:%d", dataTable.Values().Count(), dataTable.CloseBracket().Width())
			}
			assertDiagnosticLocations(t, code.GetDiagnostics(), test.diagnostics...)
			assertRecoveredDataTableColumns(t, dataTable)
		})
	}
}

func TestNativeJoinRightDataTableTruncatedValuesRecoveryParity(t *testing.T) {
	const query = `datatable(left:long)[1] | join (datatable(k:long,a:string)[1`
	code := ParseAndAnalyze(query)
	dataTables := boundNodes[*DataTableExpression](code.Query)
	if len(dataTables) != 2 {
		t.Fatalf("bound datatables = %d, diagnostics = %#v", len(dataTables), code.GetDiagnostics())
	}
	right := dataTables[1]
	if right.Values().Count() != 1 || right.CloseBracket().Width() != 0 {
		t.Fatalf("right datatable recovery = values:%d close:%d", right.Values().Count(), right.CloseBracket().Width())
	}
	assertRecoveredDataTableColumns(t, right)
	assertOrderedColumns(t, terminalResultTable(code), "left", "k", "a")

	valueStart := strings.LastIndex(query, "[1") + 1
	assertDiagnosticLocations(t, code.GetDiagnostics(),
		diagnosticLocation{"KS006", 26, len(query) - 26},
		diagnosticLocation{"KS173", valueStart, 1},
		diagnosticLocation{"KS005", len(query), 0},
		diagnosticLocation{"KS005", len(query), 0},
	)
}

func TestNativeJoinRightDataTableValuesControls(t *testing.T) {
	tests := []struct {
		name  string
		right string
		short bool
	}{
		{
			name:  "valid",
			right: `datatable(k:long,a:string)[1,"x"]`,
		},
		{
			name:  "closed short row",
			right: `datatable(k:long,a:string)[1]`,
			short: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := `datatable(left:long)[1] | join (` + test.right + `) on $left.left == $right.k`
			code := ParseAndAnalyze(query)
			if test.short {
				assertDiagnosticLocations(t, code.GetDiagnostics(), diagnosticLocation{"KS173", strings.LastIndex(query, "[1") + 1, 1})
			} else {
				assertDiagnosticLocations(t, code.GetDiagnostics())
			}
			dataTables := boundNodes[*DataTableExpression](code.Query)
			if len(dataTables) != 2 {
				t.Fatalf("bound datatables = %d", len(dataTables))
			}
			assertRecoveredDataTableColumns(t, dataTables[1])
			assertOrderedColumns(t, terminalResultTable(code), "left", "k", "a")
		})
	}
}

func assertRecoveredDataTableColumns(t *testing.T, dataTable *DataTableExpression) {
	t.Helper()
	table := semanticTable(t, dataTable)
	assertOrderedColumns(t, table, "k", "a")
	assertColumnTypes(t, table, map[string]*TypeSymbol{
		"k": ScalarTypes.Long.Type(),
		"a": ScalarTypes.String.Type(),
	})
}
