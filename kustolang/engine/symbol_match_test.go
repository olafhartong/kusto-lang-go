package kustolang

import "testing"

func TestSymbolMatchFlagValuesMatchUpstream(t *testing.T) {
	tests := []struct {
		name string
		got  SymbolMatch
		want SymbolMatch
	}{
		{"Column", SymbolMatchColumn, 1},
		{"Table", SymbolMatchTable, 2},
		{"ExternalTable", SymbolMatchExternalTable, 4},
		{"Function", SymbolMatchFunction, 8},
		{"View", SymbolMatchView, 16},
		{"Local", SymbolMatchLocal, 32},
		{"Database", SymbolMatchDatabase, 64},
		{"Cluster", SymbolMatchCluster, 128},
		{"EntityGroup", SymbolMatchEntityGroup, 256},
		{"EntityGroupElement", SymbolMatchEntityGroupElement, 512},
		{"Scalar", SymbolMatchScalar, 1024},
		{"Tabular", SymbolMatchTabular, 2048},
		{"NonScalar", SymbolMatchNonScalar, 4096},
		{"MaterializedView", SymbolMatchMaterializedView, 8192},
		{"Option", SymbolMatchOption, 16384},
		{"Graph", SymbolMatchGraph, 32768},
		{"StoredQueryResult", SymbolMatchStoredQueryResult, 65536},
		{"GraphModel", SymbolMatchGraphModel, 131072},
		{"GraphSnapshot", SymbolMatchGraphSnapshot, 262144},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Fatalf("%s = %d, want %d", tt.name, tt.got, tt.want)
		}
	}

	if SymbolMatchAny.Has(SymbolMatchExternalTable) {
		t.Fatal("upstream SymbolMatch.Any does not include ExternalTable")
	}
	if !SymbolMatchDefault.Has(SymbolMatchExternalTable) {
		t.Fatal("upstream SymbolMatch.Default includes ExternalTable")
	}
	if SymbolMatchDefault.Has(SymbolMatchDatabase) {
		t.Fatal("upstream SymbolMatch.Default does not include Database")
	}
}

func TestSymbolMatchesCoreSymbolKinds(t *testing.T) {
	column := NewColumnSymbol("UserPrincipalName", ScalarTypes.String.Type())
	table := NewTableSymbol("SigninLogs", column)
	external := table.WithIsExternal(true)
	externalConcrete := NewExternalTableSymbol("Archive", column)
	materialized := table.WithIsMaterializedView(true)
	materializedConcrete := NewMaterializedViewSymbol("DailySignin", []*ColumnSymbol{column}, "SigninLogs | summarize count() by UserPrincipalName")
	stored := table.WithIsStoredQueryResult(true)
	storedConcrete := NewStoredQueryResultSymbol("RecentSignin", column)
	storedVariable := NewVariableSymbol("rows", storedConcrete, false, nil, nil)
	database := NewDatabaseSymbol("db", table)

	if !SymbolMatches(column, SymbolMatchColumn) || SymbolMatches(column, SymbolMatchTable) {
		t.Fatalf("column matching failed")
	}
	if !SymbolMatches(table, SymbolMatchTable) || SymbolMatches(table, SymbolMatchExternalTable) {
		t.Fatalf("table matching failed")
	}
	if !SymbolMatches(external, SymbolMatchExternalTable) || SymbolMatches(external, SymbolMatchTable) {
		t.Fatalf("external table matching failed")
	}
	if !SymbolMatches(externalConcrete, SymbolMatchExternalTable) || SymbolMatches(externalConcrete, SymbolMatchTable) {
		t.Fatalf("concrete external table matching failed")
	}
	if !SymbolMatches(materialized, SymbolMatchMaterializedView) || SymbolMatches(materialized, SymbolMatchTable) {
		t.Fatalf("materialized view matching failed")
	}
	if !SymbolMatches(materializedConcrete, SymbolMatchMaterializedView) || SymbolMatches(materializedConcrete, SymbolMatchTable) {
		t.Fatalf("concrete materialized view matching failed")
	}
	if converted := external.WithIsMaterializedView(true); converted.IsExternal() || !converted.IsMaterializedView() {
		t.Fatalf("table classification should be exclusive: external=%t materialized=%t", converted.IsExternal(), converted.IsMaterializedView())
	}
	if SymbolMatches(stored, SymbolMatchTable) || SymbolMatches(stored, SymbolMatchExternalTable) {
		t.Fatalf("stored query result should not match normal/external table")
	}
	if !SymbolMatches(stored, SymbolMatchStoredQueryResult) || !SymbolMatches(storedConcrete, SymbolMatchStoredQueryResult) || !SymbolMatches(storedVariable, SymbolMatchStoredQueryResult) {
		t.Fatalf("stored query result matching failed")
	}
	if SymbolMatches(storedConcrete, SymbolMatchTable) || SymbolMatches(storedConcrete, SymbolMatchExternalTable) {
		t.Fatalf("concrete stored query result should not match normal/external table")
	}
	if !SymbolMatches(database, SymbolMatchDatabase) || SymbolMatches(database, SymbolMatchDefault) {
		t.Fatalf("database matching failed")
	}
}

func TestSymbolMatchesNameAndTabularityFilters(t *testing.T) {
	scalarFn := NewScalarFunctionSymbol("MyScalar", ScalarTypes.Bool.Type())
	tabularFn := NewComputedFunctionSymbol("MyView", "T | take 1", TabularityTabular).WithIsView(true)
	local := NewVariableSymbol("threshold", ScalarTypes.Long.Type(), false, nil, nil)
	table := NewTableSymbol("T")
	tabularLocal := NewVariableSymbol("rows", table, false, nil, nil)
	parameter := NewParameterSymbol("tenantId", ScalarTypes.String.Type(), "")
	option := NewOptionSymbolWithType("query_now", "Overrides now().", ScalarTypes.DateTime, nil)

	if !SymbolMatchesName(scalarFn, "MyScalar", SymbolMatchFunction|SymbolMatchScalar, false) {
		t.Fatal("scalar function should match scalar function filter")
	}
	if SymbolMatchesName(scalarFn, "myscalar", SymbolMatchFunction, false) {
		t.Fatal("case-sensitive name comparison should fail")
	}
	if !SymbolMatchesName(scalarFn, "myscalar", SymbolMatchFunction, true) {
		t.Fatal("case-insensitive name comparison should match")
	}
	if SymbolMatches(scalarFn, SymbolMatchFunction|SymbolMatchTabular) {
		t.Fatal("scalar function should not match tabular-only filter")
	}
	if !SymbolMatches(tabularFn, SymbolMatchFunction|SymbolMatchTabular) {
		t.Fatal("tabular function should match tabular function filter")
	}
	if !SymbolMatches(tabularFn, SymbolMatchView) {
		t.Fatal("view function should match view flag")
	}
	if SymbolMatches(tabularFn, SymbolMatchFunction|SymbolMatchScalar) {
		t.Fatal("tabular function should not match scalar-only filter")
	}
	if !SymbolMatches(local, SymbolMatchLocal|SymbolMatchScalar) || SymbolMatches(local, SymbolMatchLocal|SymbolMatchTabular) {
		t.Fatal("scalar local variable tabularity matching failed")
	}
	if !SymbolMatches(tabularLocal, SymbolMatchLocal|SymbolMatchTabular) || SymbolMatches(tabularLocal, SymbolMatchLocal|SymbolMatchScalar) {
		t.Fatal("tabular local variable tabularity matching failed")
	}
	if !SymbolMatches(parameter, SymbolMatchLocal) || !SymbolMatchesName(parameter, "tenantId", SymbolMatchLocal, false) {
		t.Fatal("parameter local matching failed")
	}
	if !SymbolMatches(option, SymbolMatchOption) || SymbolMatches(option, SymbolMatchDefault) {
		t.Fatal("option matching failed")
	}
}

func TestFindMembersUsesSymbolMatch(t *testing.T) {
	table := NewTableSymbol("SigninLogs", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	fn := NewScalarFunctionSymbol("Account", ScalarTypes.String.Type())
	local := NewVariableSymbol("Account", ScalarTypes.Long.Type(), false, nil, nil)
	option := NewOptionSymbolWithType("query_now", "Overrides now().", ScalarTypes.DateTime, nil)
	db := NewDatabaseSymbol("db", table, fn, local, option)

	matches := FindMembers(db, "Account", SymbolMatchFunction, false)
	if len(matches) != 1 || matches[0] != fn {
		t.Fatalf("function members = %#v", matches)
	}
	matches = FindMembers(db, "Account", SymbolMatchLocal, false)
	if len(matches) != 1 || matches[0] != local {
		t.Fatalf("local members = %#v", matches)
	}
	if got := FindFirstMemberMatch(db, "SigninLogs", SymbolMatchTable, false); got != table {
		t.Fatalf("first table member = %#v", got)
	}
	if got := FindFirstMemberMatch(db, "query_now", SymbolMatchOption, false); got != option {
		t.Fatalf("first option member = %#v", got)
	}
	if got := FindFirstMember(db, "query_now", false); got != nil {
		t.Fatalf("default first-member lookup should use upstream SymbolMatch.Any and skip options, got %#v", got)
	}
	if got := FindFirstMemberMatch(db, "signinlogs", SymbolMatchTable, false); got != nil {
		t.Fatalf("case-sensitive table lookup = %#v", got)
	}
	if got := FindFirstMemberMatch(db, "signinlogs", SymbolMatchTable, true); got != table {
		t.Fatalf("case-insensitive table lookup = %#v", got)
	}
	alternateDatabase := NewDatabaseSymbolWithAlternateName("primary", "alias", nil, false)
	root := NewClusterSymbolWithMembers("cluster", []Symbol{alternateDatabase}, false)
	if got := FindFirstMember(root, "alias", false); got != nil {
		t.Fatalf("first-member lookup should match primary names only, got %#v", got)
	}
	if got := FindMembersByMatch(db, SymbolMatchFunction); len(got) != 1 || got[0] != fn {
		t.Fatalf("members by match = %#v", got)
	}
}
