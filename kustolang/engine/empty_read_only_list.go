package kustolang

// Ported from src/Kusto.Language/Utils/EmptyReadOnlyList.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func EmptyReadOnlyList[T any]() []T {
	return []T{}
}
