package kustolang

import "testing"

var (
	globalStateDatabaseSink *DatabaseSymbol
	globalStateTableSink    *TableSymbol
)

func TestGlobalStateSymbolOwnershipLookupsDoNotAllocate(t *testing.T) {
	column := NewColumnSymbol("Account", ScalarTypes.String.Type())
	table := NewTableSymbol("Events", column)
	database := NewDatabaseSymbol("db", table)
	globals := NewGlobalState(database)

	allocations := testing.AllocsPerRun(100, func() {
		globalStateDatabaseSink = globals.GetDatabase(table)
		globalStateTableSink = globals.GetTable(column)
	})
	if allocations != 0 {
		t.Fatalf("symbol ownership lookups allocate %v times per run, want 0", allocations)
	}
	if globalStateDatabaseSink != database || globalStateTableSink != table {
		t.Fatalf("ownership lookup returned database %p and table %p, want %p and %p", globalStateDatabaseSink, globalStateTableSink, database, table)
	}
}
