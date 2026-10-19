package kustolang

// Ported from src/Kusto.Language/Utils/Optional.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type Optional[T any] struct {
	Value    T
	HasValue bool
}

func NewOptional[T any](value T) Optional[T] {
	return Optional[T]{Value: value, HasValue: true}
}
