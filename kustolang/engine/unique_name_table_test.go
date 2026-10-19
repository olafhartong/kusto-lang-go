package kustolang

import "testing"

func TestUniqueNameTableMirrorsUpstreamSuffixing(t *testing.T) {
	table := NewUniqueNameTable()
	if got := table.GetOrAddName("Name"); got != "Name" {
		t.Fatalf("first name = %q, want Name", got)
	}
	if got := table.GetOrAddName("Name"); got != "Name1" {
		t.Fatalf("second name = %q, want Name1", got)
	}
	if got := table.GetOrAddName("Name"); got != "Name2" {
		t.Fatalf("third name = %q, want Name2", got)
	}

	table.AddName("Other")
	table.AddName("Other")
	if got := table.GetOrAddName("Other"); got != "Other1" {
		t.Fatalf("duplicate added name = %q, want Other1", got)
	}
}

func TestUniqueNameTableBaseNameAndEmptyHandling(t *testing.T) {
	table := NewUniqueNameTable()
	table.AddName("Column")
	if got := table.GetOrAddName("Column", "Base"); got != "Base1" {
		t.Fatalf("base-name duplicate = %q, want Base1", got)
	}
	if got := table.GetOrAddName("Column", ""); got != "2" {
		t.Fatalf("empty-base duplicate = %q, want 2", got)
	}

	empty := NewUniqueNameTable()
	empty.AddName("")
	if got := empty.GetOrAddName(""); got != "" {
		t.Fatalf("first empty name = %q, want empty", got)
	}
	if got := empty.GetOrAddName(""); got != "1" {
		t.Fatalf("duplicate empty name = %q, want 1", got)
	}
}

func TestUniqueNameTableClear(t *testing.T) {
	table := NewUniqueNameTable()
	table.AddName("Name")
	if got := table.GetOrAddName("Name"); got != "Name1" {
		t.Fatalf("duplicate before clear = %q, want Name1", got)
	}
	table.Clear()
	if got := table.GetOrAddName("Name"); got != "Name" {
		t.Fatalf("name after clear = %q, want Name", got)
	}
}
