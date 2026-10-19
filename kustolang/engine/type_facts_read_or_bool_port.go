package kustolang

// Ported from src/Kusto.Language/Symbols/TypeFacts.cs at
// c94c9e761b3d095e2b9c21e3ec3e92afa0af4856.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// IsAnyScalarExceptReadOrBool preserves the misspelled public name in the
// pinned Kusto.Language API. "Read" means "Real" in the upstream source.
func IsAnyScalarExceptReadOrBool(symbol Symbol) bool {
	return IsAnyScalarExceptRealOrBool(symbol)
}
