package kustolang

import "testing"

func TestNativeMakeSeriesSchemaScopeAndFromToRange(t *testing.T) {
	events := NewTableSymbol("Events",
		NewColumnSymbol("Timestamp", ScalarTypes.DateTime.Type()),
		NewColumnSymbol("Region", ScalarTypes.String.Type()),
		NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	query := ParseQuery(`Events | make-series hint.shufflekey=Region Total=sum(Value) default=0, Average=avg(Value) on Timestamp from ago(1d) to now() step 1h by Region`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", events))).TryBind(query)

	series := firstBoundNode[*MakeSeriesOperator](t, query)
	assertDiagnosticCodes(t, series)
	table := semanticTable(t, series)
	assertOrderedColumns(t, table, "Region", "Total", "Average", "Timestamp")
	for _, column := range table.Columns()[1:] {
		if column.Type != ScalarTypes.Dynamic.Type() {
			t.Fatalf("column %s type = %v, want dynamic", column.SymbolName(), column.Type)
		}
	}
	if table.IsSorted() {
		t.Fatal("make-series result must not retain sorted state")
	}
}

func TestNativeMakeSeriesDeprecatedRangeAndValidation(t *testing.T) {
	events := NewTableSymbol("Events",
		NewColumnSymbol("Timestamp", ScalarTypes.DateTime.Type()),
		NewColumnSymbol("Value", ScalarTypes.Long.Type()),
		NewColumnSymbol("Fallback", ScalarTypes.Long.Type()))
	valid := ParseQuery(`Events | make-series count() on Timestamp in range(ago(1d), now(), 1h)`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", events))).TryBind(valid)
	assertDiagnosticCodes(t, firstBoundNode[*MakeSeriesOperator](t, valid))

	invalid := ParseQuery(`Events | make-series Total=sum(Value) default=Fallback on Timestamp from 1 to 2 step 1`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", events))).TryBind(invalid)
	series := firstBoundNode[*MakeSeriesOperator](t, invalid)
	assertDiagnosticCodes(t, series, "KS141", "KS141", "KS178")
}
