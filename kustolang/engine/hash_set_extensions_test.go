package kustolang

import "testing"

func TestHashSetAddRange(t *testing.T) {
	set := map[string]struct{}{"a": {}}
	HashSetAddRange(set, []string{"a", "b", "c", "b"})
	for _, value := range []string{"a", "b", "c"} {
		if _, ok := set[value]; !ok {
			t.Fatalf("missing set value %q", value)
		}
	}
	if len(set) != 3 {
		t.Fatalf("set count = %d, want 3", len(set))
	}
}
