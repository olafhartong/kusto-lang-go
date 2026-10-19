package kustolang

// Ported from src/Kusto.Language/Symbols/ColumnSymbol.cs and TableSymbol.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// CombineColumns combines column sets using the same rules as
// Kusto.Language.Symbols.ColumnSymbol.Combine.
func CombineColumns(kind CombineKind, columnSets ...[]*ColumnSymbol) []*ColumnSymbol {
	return internal.CombineColumns(kind, columnSets...)
}

// CombineTableSymbols combines table columns using the same rules as
// Kusto.Language.Symbols.TableSymbol.Combine.
func CombineTableSymbols(kind CombineKind, tables ...*TableSymbol) *TableSymbol {
	return internal.CombineTableSymbolsWithKind(kind, tables...)
}
