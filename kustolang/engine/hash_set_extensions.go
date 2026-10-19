package kustolang

// Ported from src/Kusto.Language/Utils/HashSetExtensions.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func HashSetAddRange[T comparable](hashSet map[T]struct{}, items []T) {
	for _, item := range items {
		hashSet[item] = struct{}{}
	}
}
