package kustolang

// Ported from src/Kusto.Language/Symbols/SchemaDisplay.cs,
// ScalarSymbol.cs, TableSymbol.cs, src/Kusto.Language/ServerKinds.cs and
// src/Kusto.Language/Editor/CompletionDisplay.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// PrimitiveSymbol is the concrete scalar symbol used for primitive Kusto
// types. The Go port combines the upstream ScalarSymbol base and
// PrimitiveSymbol implementation in one native type.
type PrimitiveSymbol = internal.ScalarSymbol

type ScalarFlags = internal.ScalarFlag
type MaterializedViewKind = internal.MaterializedViewKind

const (
	ScalarFlagsNone      = internal.ScalarFlagNone
	ScalarFlagsInteger   = internal.ScalarFlagInteger
	ScalarFlagsNumeric   = internal.ScalarFlagNumeric
	ScalarFlagsInterval  = internal.ScalarFlagInterval
	ScalarFlagsSummable  = internal.ScalarFlagSummable
	ScalarFlagsOrderable = internal.ScalarFlagOrderable
	ScalarFlagsAll       = internal.ScalarFlagAll

	MaterializedViewKindUnknown      = internal.MaterializedViewKindUnknown
	MaterializedViewKindDownsampling = internal.MaterializedViewKindDownsampling
	MaterializedViewKindOther        = internal.MaterializedViewKindOther

	ServerKindEngine         = internal.ServerKindEngine
	ServerKindDataManager    = internal.ServerKindDataManager
	ServerKindClusterManager = internal.ServerKindClusterManager
	ServerKindAriaBridge     = internal.ServerKindAriaBridge
)

func NewPrimitiveSymbol(name string, aliases []string, flags ScalarFlags, widerThan ...*PrimitiveSymbol) *PrimitiveSymbol {
	return internal.NewScalarSymbol(name, aliases, flags, widerThan...)
}

func PrimitiveSymbolFrom(typeName string) *PrimitiveSymbol {
	return internal.ScalarTypes.Get(typeName)
}

// SchemaDisplayAPI mirrors the static SchemaDisplay class.
type SchemaDisplayAPI struct{}

func (SchemaDisplayAPI) GetText(symbol Symbol) string {
	return internal.SchemaDisplayText(symbol)
}

func (SchemaDisplayAPI) GetScalarTypeText(typ *TypeSymbol) string {
	return internal.SchemaDisplayScalarTypeText(typ)
}

func (SchemaDisplayAPI) GetParameterTypeText(parameter *Parameter) string {
	return internal.SchemaDisplayParameterTypeText(parameter)
}

// CompletionDisplayAPI mirrors the static CompletionDisplay class.
type CompletionDisplayAPI struct{}

func (CompletionDisplayAPI) GetText(symbol Symbol, nameOnly ...bool) string {
	return internal.CompletionDisplayText(symbol, nameOnly...)
}

var (
	SchemaDisplay     SchemaDisplayAPI
	CompletionDisplay CompletionDisplayAPI
)
