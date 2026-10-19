package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicDatabaseSymbolContainsPinnedBehavior(t *testing.T) {
	member := kustolang.NewTableSymbol("Member")
	probe := kustolang.NewTableSymbol("Probe")
	database := kustolang.NewDatabaseSymbol("db", member)

	if !database.Contains(probe) || database.Contains(member) || !database.Contains(probe) {
		t.Fatal("public Contains did not preserve the pinned lazy-cache behavior")
	}
}
