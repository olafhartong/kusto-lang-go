package kustolang

import "testing"

func TestColumnMapBucketsDifferentTypesByName(t *testing.T) {
	account := NewColumnSymbol("Value", ScalarTypes.String.Type())
	count := NewColumnSymbol("Value", ScalarTypes.Long.Type())
	total := NewColumnSymbol("Value", ScalarTypes.Long.Type())
	status := NewColumnSymbol("Status", ScalarTypes.String.Type())

	columnMap := NewColumnMapFromColumns([]*ColumnSymbol{account, count, total, status})

	if !columnMap.HasColumns("Value") || columnMap.HasColumns("Missing") {
		t.Fatal("HasColumns mismatch")
	}
	if !columnMap.HasMultipleTypes("Value") {
		t.Fatal("different type buckets should report multiple types")
	}
	if !columnMap.HasMutipleColumns("Value") {
		t.Fatal("different type buckets should report multiple columns")
	}
	if columnMap.HasMultipleColumns("Value", ScalarTypes.String.Type()) {
		t.Fatal("single string column should not report multiple columns for its type")
	}
	if !columnMap.HasMultipleColumns("Value", ScalarTypes.Long.Type()) {
		t.Fatal("two long columns should report multiple columns for their type")
	}

	requireColumnMapTypes(t, columnMap.GetTypes("Value"), ScalarTypes.String.Type(), ScalarTypes.Long.Type())
	requireColumnMapColumns(t, columnMap.GetColumns("Value"), account, count, total)
	requireColumnMapColumns(t, columnMap.GetColumnsByType("Value", ScalarTypes.String.Type()), account)
	requireColumnMapColumns(t, columnMap.GetColumnsByType("Value", ScalarTypes.Long.Type()), count, total)

	columnMap.Remove("Value")
	if columnMap.HasColumns("Value") || len(columnMap.GetColumns("Value")) != 0 || len(columnMap.GetTypes("Value")) != 0 {
		t.Fatal("Remove should clear all columns for the name")
	}
	requireColumnMapColumns(t, columnMap.GetColumns("Status"), status)
}

func TestColumnMapSameTypeListMirrorsUpstreamDifferentTypeEdge(t *testing.T) {
	first := NewColumnSymbol("Value", ScalarTypes.String.Type())
	second := NewColumnSymbol("Value", ScalarTypes.String.Type())
	ignored := NewColumnSymbol("Value", ScalarTypes.Long.Type())

	columnMap := NewColumnMap()
	columnMap.Add(first)
	columnMap.Add(second)
	columnMap.Add(ignored)

	if columnMap.HasMultipleTypes("Value") {
		t.Fatal("upstream does not publish a type dictionary when a different type is added to an existing same-type list")
	}
	if !columnMap.HasMutipleColumns("Value") || !columnMap.HasMultipleColumns("Value", ScalarTypes.String.Type()) {
		t.Fatal("same-type duplicate list should remain visible")
	}
	if columnMap.HasMultipleColumns("Value", ScalarTypes.Long.Type()) {
		t.Fatal("ignored long column should not affect typed duplicate checks")
	}
	requireColumnMapTypes(t, columnMap.GetTypes("Value"), ScalarTypes.String.Type())
	requireColumnMapColumns(t, columnMap.GetColumns("Value"), first, second)
	requireColumnMapColumns(t, columnMap.GetColumnsByType("Value", ScalarTypes.Long.Type()), first, second)
}

func TestColumnMapDynamicTypeUnification(t *testing.T) {
	bag := NewColumnSymbol("Payload", ScalarTypes.DynamicBag.Type())
	dynamicString := NewColumnSymbol("Payload", ScalarTypes.DynamicString.Type())

	unified := NewColumnMap()
	unified.Add(bag)
	unified.Add(dynamicString)
	if unified.HasMultipleTypes("Payload") {
		t.Fatal("dynamic subclasses should unify by default")
	}
	if !unified.HasMultipleColumns("Payload", ScalarTypes.Dynamic.Type()) {
		t.Fatal("unified dynamic duplicates should be found through dynamic type key")
	}
	requireColumnMapColumns(t, unified.GetColumnsByType("Payload", ScalarTypes.Dynamic.Type()), bag, dynamicString)
	requireColumnMapTypes(t, unified.GetTypes("Payload"), ScalarTypes.DynamicBag.Type())

	separated := NewColumnMap(false)
	separated.Add(bag)
	separated.Add(dynamicString)
	if !separated.HasMultipleTypes("Payload") {
		t.Fatal("dynamic subclasses should remain separate when unification is disabled")
	}
	if separated.HasMultipleColumns("Payload", ScalarTypes.Dynamic.Type()) {
		t.Fatal("untyped dynamic key should not match typed dynamic buckets when unification is disabled")
	}
	requireColumnMapTypes(t, separated.GetTypes("Payload"), ScalarTypes.DynamicBag.Type(), ScalarTypes.DynamicString.Type())
	requireColumnMapColumns(t, separated.GetColumnsByType("Payload", ScalarTypes.DynamicBag.Type()), bag)
	requireColumnMapColumns(t, separated.GetColumnsByType("Payload", ScalarTypes.DynamicString.Type()), dynamicString)
}

func requireColumnMapColumns(t *testing.T, got []*ColumnSymbol, want ...*ColumnSymbol) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("columns count = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("columns[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func requireColumnMapTypes(t *testing.T, got []*TypeSymbol, want ...*TypeSymbol) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("types count = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("types[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}
