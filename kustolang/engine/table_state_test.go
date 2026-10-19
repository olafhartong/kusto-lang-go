package kustolang

import "testing"

func TestTableStateFlags(t *testing.T) {
	state := TableStateNone.New().With(TableStateSorted).Open()
	if !state.Has(TableStateSerialized) || !state.Has(TableStateSorted) || !state.Has(TableStateOpen) {
		t.Fatalf("state flags missing: %v", state)
	}
	unsorted := state.Unsorted()
	if !unsorted.Has(TableStateSerialized) || unsorted.Has(TableStateSorted) || !unsorted.Has(TableStateOpen) {
		t.Fatalf("Unsorted = %v", unsorted)
	}
}

func TestTableSymbolStateCloneHelpers(t *testing.T) {
	table := NewTableSymbol("T").WithIsSerialized(true).WithIsSorted(true).WithIsOpen(true)
	if !table.IsSerialized() || !table.IsSorted() || !table.IsOpen() {
		t.Fatalf("table state not set")
	}
	unsorted := table.WithIsSorted(false)
	if unsorted == table || !table.IsSorted() || unsorted.IsSorted() {
		t.Fatalf("WithIsSorted did not clone or preserve original")
	}
}
