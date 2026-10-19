package kustolang

import "testing"

func TestEmptyReadOnlyListReturnsNonNilEmptySlice(t *testing.T) {
	empty := EmptyReadOnlyList[string]()
	if empty == nil {
		t.Fatal("empty read-only list should be a non-nil empty slice")
	}
	if len(empty) != 0 {
		t.Fatalf("empty read-only list length = %d, want 0", len(empty))
	}
}

func TestListToReadOnlyUsesEmptyReadOnlyListShape(t *testing.T) {
	empty := ListToReadOnly([]int{})
	if empty == nil {
		t.Fatal("ListToReadOnly(empty) should return a non-nil empty slice")
	}
	if len(empty) != 0 {
		t.Fatalf("ListToReadOnly(empty) length = %d, want 0", len(empty))
	}
}
