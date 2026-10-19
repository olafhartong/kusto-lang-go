package kustolang

import "testing"

func TestNativeCurrentRangeUnknownTailParity(t *testing.T) {
	tests := []struct {
		name  string
		query string
		start int
	}{
		{name: "bracketed keyword name", query: `range ['print'] from 1 to 3 step 1 MissingTail`, start: 35},
		{name: "function endpoints", query: `range LastWeek from ago(7d) to now() step 1d MissingTail`, start: 45},
		{name: "long then datetime endpoints", query: `range i from 1 to datetime(2024-01-01) step 1 MissingTail`, start: 46},
		{name: "datetime then long endpoints", query: `range i from datetime(2024-01-01) to 3 step 1d MissingTail`, start: 47},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := Parse(test.query).Analyze()
			assertDiagnosticLocations(t, code.GetDiagnostics(),
				diagnosticLocation{"KS005", test.start, 11},
				diagnosticLocation{"KS142", test.start, 11},
			)
			if code.Query.Statements().Count() != 2 {
				t.Fatalf("statements = %d, want 2", code.Query.Statements().Count())
			}
			if _, ok := queryBlockExpressionAtForTest(code.Query, 0).(*RangeOperator); !ok {
				t.Fatalf("first expression = %T, want *RangeOperator", queryBlockExpressionAtForTest(code.Query, 0))
			}
			if tail, ok := queryBlockExpressionAtForTest(code.Query, 1).(*NameReference); !ok || NameReferenceSimpleName(tail) != "MissingTail" {
				t.Fatalf("tail expression = %#v, want MissingTail", queryBlockExpressionAtForTest(code.Query, 1))
			}
			if table := terminalResultTable(code); table != nil && len(table.Columns()) != 0 {
				t.Fatalf("terminal result = %#v, want no output columns", table.Columns())
			}
		})
	}
}

func queryBlockExpressionAtForTest(block *QueryBlock, index int) Expression {
	if block == nil || block.Statements() == nil || index < 0 || index >= block.Statements().Count() {
		return nil
	}
	statement, ok := block.Statements().At(index).Element().(*ExpressionStatement)
	if !ok {
		return nil
	}
	return statement.Expression()
}
