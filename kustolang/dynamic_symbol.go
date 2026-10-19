package kustolang

// Ported from src/Kusto.Language/Symbols/DynamicSymbol.cs. See
// third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// DynamicAnySymbolInstance is the canonical upstream DynamicAnySymbol.Instance.
var DynamicAnySymbolInstance = internal.ScalarTypes.DynamicAny

// DynamicBagEmpty is the Go spelling of DynamicBagSymbol.Empty.
var DynamicBagEmpty = internal.ScalarTypes.DynamicBag

// NewDynamicBagSymbol constructs a dynamic object type with known properties.
func NewDynamicBagSymbol(properties ...*ColumnSymbol) *DynamicBagSymbol {
	return internal.NewDynamicBagSymbol(properties...)
}
