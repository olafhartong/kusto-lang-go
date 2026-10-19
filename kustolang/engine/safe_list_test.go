package kustolang

import "testing"

func TestSafeListCopiesInputAndIndexesVisibleLength(t *testing.T) {
	source := []string{"a", "b"}
	list := NewSafeList(source)
	source[0] = "changed"

	if list.Count() != 2 {
		t.Fatalf("count = %d, want 2", list.Count())
	}
	if got := list.At(0); got != "a" {
		t.Fatalf("At(0) = %q, want a", got)
	}
	if !list.IsOwner() {
		t.Fatal("new SafeList should own its backing list")
	}
}

func TestSafeListAddItemKeepsPreviousViewImmutable(t *testing.T) {
	original := NewSafeList([]int{1, 2})
	added := original.AddItem(3)

	if original.IsOwner() {
		t.Fatal("original should no longer own backing list after AddItem")
	}
	if !added.IsOwner() {
		t.Fatal("new list should own backing list")
	}
	if original.Count() != 2 || added.Count() != 3 {
		t.Fatalf("counts original/added = %d/%d, want 2/3", original.Count(), added.Count())
	}
	if got := original.Values(); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("original values = %#v, want [1 2]", got)
	}
	if got := added.Values(); len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("added values = %#v, want [1 2 3]", got)
	}
}

func TestSafeListAddItemFromNonOwnerCopiesVisibleItems(t *testing.T) {
	original := NewSafeList([]int{1})
	first := original.AddItem(2)
	second := original.AddItem(3)

	if got := first.Values(); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("first values = %#v, want [1 2]", got)
	}
	if got := second.Values(); len(got) != 2 || got[0] != 1 || got[1] != 3 {
		t.Fatalf("second values = %#v, want [1 3]", got)
	}
}

func TestSafeListAddItemsAndEmpty(t *testing.T) {
	empty := EmptySafeList[string]()
	if empty.Count() != 0 || empty.IsOwner() {
		t.Fatalf("empty count/owner = %d/%t, want 0/false", empty.Count(), empty.IsOwner())
	}

	added := empty.AddItems([]string{"a", "b"})
	if empty.Count() != 0 {
		t.Fatalf("empty view count after AddItems = %d, want 0", empty.Count())
	}
	if got := added.Values(); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("added values = %#v, want [a b]", got)
	}
}

func TestSafeListAtPanicsOutOfRange(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("At should panic for out-of-range index")
		}
	}()
	NewSafeList([]int{1}).At(1)
}
