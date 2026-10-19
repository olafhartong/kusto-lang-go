package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicTableStateExtensionsMatchPinnedUpstream(t *testing.T) {
	state := kustolang.TableStateNone.New().With(kustolang.TableStateSorted).Open()
	if !state.Has(kustolang.TableStateSerialized) || !state.Has(kustolang.TableStateSorted) || !state.Has(kustolang.TableStateOpen) {
		t.Fatalf("composed table state = %08b", state)
	}

	unsorted := state.Unsorted()
	if !unsorted.Has(kustolang.TableStateSerialized) || unsorted.Has(kustolang.TableStateSorted) || !unsorted.Has(kustolang.TableStateOpen) {
		t.Fatalf("unsorted table state = %08b", unsorted)
	}
	if got := unsorted.Without(kustolang.TableStateOpen); got.Has(kustolang.TableStateOpen) || !got.Has(kustolang.TableStateSerialized) {
		t.Fatalf("state without open = %08b", got)
	}
}
