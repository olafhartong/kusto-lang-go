package kustolang

// Ported from src/Kusto.Language/Utils/EmptyReadOnlyList.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// EmptyReadOnlyList returns the non-nil empty-list shape used by the pinned
// Kusto.Language API. Go slices are values, so appending to the result cannot
// mutate a shared global instance.
func EmptyReadOnlyList[T any]() []T {
	return internal.EmptyReadOnlyList[T]()
}
