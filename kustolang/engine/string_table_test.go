package kustolang

import (
	"sort"
	"testing"
)

func TestStringTableAddsAndReusesEquivalentRanges(t *testing.T) {
	table := NewStringTable()
	first := table.Add("Alpha")
	if first != "Alpha" {
		t.Fatalf("first add = %q, want Alpha", first)
	}
	second := table.AddRange("__Alpha__", 2, 5)
	if second != first {
		t.Fatalf("range add = %q, want %q", second, first)
	}
	values := table.Values()
	if len(values) != 1 || values[0] != "Alpha" {
		t.Fatalf("values = %#v, want [Alpha]", values)
	}
}

func TestStringTableAddsNewSubrangeAndCollection(t *testing.T) {
	table := NewStringTable()
	if got := table.AddRange("prefixValueSuffix", 6, 5); got != "Value" {
		t.Fatalf("range add = %q, want Value", got)
	}
	table.AddAll([]string{"Beta", "Gamma", "Beta"})
	if !table.Contains("Value") || !table.ContainsKey("__Beta__", 2, 4) {
		t.Fatalf("table did not contain expected entries: %#v", table.Values())
	}
	values := table.Values()
	sort.Strings(values)
	if len(values) != 3 || values[0] != "Beta" || values[1] != "Gamma" || values[2] != "Value" {
		t.Fatalf("values = %#v, want [Beta Gamma Value]", values)
	}
}

func TestStringTableEmptyContainsBehavior(t *testing.T) {
	table := NewStringTable()
	if got := table.Add(""); got != "" {
		t.Fatalf("empty add = %q, want empty", got)
	}
	if table.Contains("") {
		t.Fatal("Contains should reject empty text like upstream StringTable.Contains")
	}
	if got, ok := table.values.TryGetValue("", 0, 0); !ok || got != "" {
		t.Fatalf("underlying empty lookup = %q, %t, want empty, true", got, ok)
	}
}
