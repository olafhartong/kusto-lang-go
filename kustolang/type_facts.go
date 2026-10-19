package kustolang

// Ported from src/Kusto.Language/Symbols/TypeFacts.cs. See
// third_party/kusto/LICENSE and third_party/kusto/UPSTREAM.json.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

func GetElementType(typ *TypeSymbol) *TypeSymbol { return internal.GetElementType(typ) }

func GetWidestScalarType(types ...*TypeSymbol) *ScalarSymbol {
	return internal.GetWidestScalarType(types)
}

func GetCommonType(types []*TypeSymbol, include func(*TypeSymbol) bool, allowedConversions Conversion, defaultType *TypeSymbol) *TypeSymbol {
	return internal.GetCommonType(types, include, allowedConversions, defaultType)
}

func TryGetCommonType(typeA, typeB *TypeSymbol, allowedConversions Conversion) (*TypeSymbol, bool) {
	return internal.TryGetCommonType(typeA, typeB, allowedConversions)
}

// TryGetCommonSymbolType is the Go symbol-wide adaptation of the upstream
// TypeSymbol overload. It retains table-derived symbols that cannot be passed
// through Go's concrete *TypeSymbol representation.
func TryGetCommonSymbolType(symbolA, symbolB Symbol, allowedConversions Conversion) (Symbol, bool) {
	return internal.TryGetCommonSymbolType(symbolA, symbolB, allowedConversions)
}

func GetCommonScalarType(types []*TypeSymbol, allowedConversions ...Conversion) *TypeSymbol {
	return internal.GetCommonScalarType(types, allowedConversions...)
}

func GetCommonResultType(types []*TypeSymbol, allowedConversions ...Conversion) *TypeSymbol {
	return internal.GetCommonResultType(types, allowedConversions...)
}

func GetCommonResultTypeWithDefault(types []*TypeSymbol, allowedConversions Conversion, defaultType *TypeSymbol, ignoreDynamic bool) *TypeSymbol {
	return internal.GetCommonResultTypeWithDefault(types, allowedConversions, defaultType, ignoreDynamic)
}

// GetCommonExpressionResultType is the Go spelling of the C# expression-list
// overload of GetCommonResultType.
func GetCommonExpressionResultType(expressions []Expression, allowedConversions ...Conversion) *TypeSymbol {
	types := make([]*TypeSymbol, 0, len(expressions))
	for _, expression := range expressions {
		types = append(types, ExpressionResultType(expression))
	}
	return internal.GetCommonResultType(types, allowedConversions...)
}

// GetCommonExpressionResultTypeWithDefault is the Go slice equivalent of the
// C# separated-expression-list overload.
func GetCommonExpressionResultTypeWithDefault(expressions []Expression, allowedConversions Conversion, defaultType *TypeSymbol, ignoreDynamic bool) *TypeSymbol {
	types := make([]*TypeSymbol, 0, len(expressions))
	for _, expression := range expressions {
		types = append(types, ExpressionResultType(expression))
	}
	return internal.GetCommonResultTypeWithDefault(types, allowedConversions, defaultType, ignoreDynamic)
}

func GetCommonColumnType(columns []*ColumnSymbol, allowedConversions ...Conversion) *TypeSymbol {
	return internal.GetCommonColumnType(columns, allowedConversions...)
}

func IsConversionAllowed(conversion Conversion, allowedConversions Conversion) bool {
	return internal.IsConversionAllowed(conversion, allowedConversions)
}

func IsCommonArgumentTypeKind(kind ParameterTypeKind) bool {
	return internal.IsCommonArgumentTypeKind(kind)
}

func GetCommonArgumentType(argumentParameters []*Parameter, argumentTypes []*TypeSymbol) *TypeSymbol {
	return internal.GetCommonArgumentType(argumentParameters, argumentTypes)
}

func GetCommonArgumentTypeIgnoreDynamic(argumentParameters []*Parameter, argumentTypes []*TypeSymbol) *TypeSymbol {
	return internal.GetCommonArgumentTypeIgnoreDynamic(argumentParameters, argumentTypes)
}

func GetCommonArgumentTypeWithDefault(argumentParameters []*Parameter, argumentTypes []*TypeSymbol, defaultType *TypeSymbol, ignoreDynamic bool) *TypeSymbol {
	return internal.GetCommonArgumentTypeWithDefault(argumentParameters, argumentTypes, defaultType, ignoreDynamic)
}

func GetCommonColumnTypeWithDefault(columns []*ColumnSymbol, allowedConversions Conversion, defaultType *TypeSymbol) *TypeSymbol {
	return internal.GetCommonColumnTypeWithDefault(columns, allowedConversions, defaultType)
}

func PromoteToLong(typ *TypeSymbol) *TypeSymbol { return internal.PromoteToLong(typ) }

func IsPromotableTo(sourceType, targetType *TypeSymbol) bool {
	return internal.IsPromotableTo(sourceType, targetType)
}

func IsInteger(symbol Symbol) bool                { return internal.IsInteger(symbol) }
func IsInterval(symbol Symbol) bool               { return internal.IsInterval(symbol) }
func IsNumeric(symbol Symbol) bool                { return internal.IsNumeric(symbol) }
func IsSummable(symbol Symbol) bool               { return internal.IsSummable(symbol) }
func IsOrderable(symbol Symbol) bool              { return internal.IsOrderable(symbol) }
func IsAnyScalarExceptDynamic(symbol Symbol) bool { return internal.IsAnyScalarExceptDynamic(symbol) }
func IsAnyScalarExceptBool(symbol Symbol) bool    { return internal.IsAnyScalarExceptBool(symbol) }

func IsAnyScalarExceptRealOrBool(symbol Symbol) bool {
	return internal.IsAnyScalarExceptRealOrBool(symbol)
}

func IsNumericOrBool(symbol Symbol) bool    { return internal.IsNumericOrBool(symbol) }
func IsRealOrDecimal(symbol Symbol) bool    { return internal.IsRealOrDecimal(symbol) }
func IsStringOrDynamic(symbol Symbol) bool  { return internal.IsStringOrDynamic(symbol) }
func IsStringOrArray(symbol Symbol) bool    { return internal.IsStringOrArray(symbol) }
func IsIntegerOrArray(symbol Symbol) bool   { return internal.IsIntegerOrArray(symbol) }
func IsIntegerOrDynamic(symbol Symbol) bool { return internal.IsIntegerOrDynamic(symbol) }
func IsDynamicArray(symbol Symbol) bool     { return internal.IsDynamicArray(symbol) }
func IsDynamicBag(symbol Symbol) bool       { return internal.IsDynamicBag(symbol) }

func IsDynamicArrayOrBag(symbol Symbol) bool {
	return internal.IsDynamicArrayOrBag(symbol)
}

func Union(columnsA, columnsB []*ColumnSymbol) []*ColumnSymbol {
	return internal.UnionColumns(columnsA, columnsB)
}

func Intersect(columnsA, columnsB []*ColumnSymbol) []*ColumnSymbol {
	return internal.IntersectColumns(columnsA, columnsB)
}
