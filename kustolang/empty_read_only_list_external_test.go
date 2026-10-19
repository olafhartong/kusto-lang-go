package kustolang_test

// Ported from the behavior of
// src/Kusto.Language/Utils/EmptyReadOnlyList.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicEmptyReadOnlyListReturnsIndependentNonNilEmptySlices(t *testing.T) {
	first := kustolang.EmptyReadOnlyList[string]()
	second := kustolang.EmptyReadOnlyList[string]()
	if first == nil || second == nil || len(first) != 0 || len(second) != 0 {
		t.Fatalf("empty lists = %#v and %#v", first, second)
	}

	first = append(first, "value")
	if len(first) != 1 || len(second) != 0 {
		t.Fatalf("appending to one result changed another: %#v and %#v", first, second)
	}
}
