package kustolang_test

// Ported from the behavior of src/Kusto.Language/Utils/SafeList.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"reflect"
	"sync"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicSafeListPreservesImmutableSnapshots(t *testing.T) {
	input := []int{1, 2}
	original := kustolang.NewSafeList(input)
	input[0] = 99

	first := original.AddItem(3)
	second := original.AddItems([]int{4, 5})
	if got := original.Values(); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("original snapshot = %#v", got)
	}
	if got := first.Values(); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("first branch = %#v", got)
	}
	if got := second.Values(); !reflect.DeepEqual(got, []int{1, 2, 4, 5}) {
		t.Fatalf("second branch = %#v", got)
	}

	copyOfFirst := first.Values()
	copyOfFirst[0] = 99
	if first.At(0) != 1 || first.Count() != 3 {
		t.Fatalf("mutating Values changed snapshot: %#v", first.Values())
	}
}

func TestPublicSafeListEmptyAndBounds(t *testing.T) {
	empty := kustolang.EmptySafeList[string]()
	if empty.Count() != 0 || len(empty.Values()) != 0 {
		t.Fatalf("empty snapshot = count:%d values:%#v", empty.Count(), empty.Values())
	}
	if got := empty.AddItem("x").Values(); !reflect.DeepEqual(got, []string{"x"}) {
		t.Fatalf("empty append = %#v", got)
	}

	defer func() {
		if recover() == nil {
			t.Fatal("At should panic outside the visible snapshot")
		}
	}()
	empty.At(0)
}

func TestPublicSafeListConcurrentBranchesRemainIndependent(t *testing.T) {
	original := kustolang.NewSafeList([]int{1})
	results := make([][]int, 8)
	var wait sync.WaitGroup
	for i := range results {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			results[index] = original.AddItem(index + 2).Values()
		}(i)
	}
	wait.Wait()

	if got := original.Values(); !reflect.DeepEqual(got, []int{1}) {
		t.Fatalf("concurrent additions changed original = %#v", got)
	}
	for index, got := range results {
		want := []int{1, index + 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("branch %d = %#v, want %#v", index, got, want)
		}
	}
}
