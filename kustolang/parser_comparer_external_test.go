package kustolang_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicParserEqualityComparerOverloads(t *testing.T) {
	equalFold := func(left, right string) bool { return strings.EqualFold(left, right) }
	words := kustolang.NewArraySource([]string{"ALPHA", "beta", "TAIL"})

	if got := kustolang.MatchItemWithComparer("alpha", equalFold).Scan(words, 0); got != 1 {
		t.Fatalf("item match length = %d", got)
	}
	if got := kustolang.MatchAnyWithComparer([]string{"gamma", "BETA"}, equalFold).Scan(words, 1); got != 1 {
		t.Fatalf("any match length = %d", got)
	}
	if got := kustolang.MatchSequenceWithComparer([]string{"alpha", "BETA"}, equalFold).Scan(words, 0); got != 2 {
		t.Fatalf("sequence match length = %d", got)
	}
	if got := kustolang.MatchSequenceWithComparer([]string{"alpha", "missing"}, equalFold).Scan(words, 0); got != ^1 {
		t.Fatalf("sequence mismatch = %d, want %d", got, ^1)
	}

	byteSlices := kustolang.NewArraySource([][]byte{[]byte("one"), []byte("TWO")})
	equalBytes := func(left, right []byte) bool { return bytes.EqualFold(left, right) }
	if got := kustolang.MatchSequenceWithComparer([][]byte{[]byte("ONE"), []byte("two")}, equalBytes).Scan(byteSlices, 0); got != 2 {
		t.Fatalf("non-comparable sequence match length = %d", got)
	}
}
