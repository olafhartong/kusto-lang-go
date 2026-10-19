package kustolang

// Ported from src/Kusto.Language/Symbols/TypeFacts.cs at
// c94c9e761b3d095e2b9c21e3ec3e92afa0af4856.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// IsAnyScalarExceptReadOrBool preserves the pinned upstream API spelling.
// "Read" is an upstream typo for "Real".
func IsAnyScalarExceptReadOrBool(symbol Symbol) bool {
	return internal.IsAnyScalarExceptReadOrBool(symbol)
}
