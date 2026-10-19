package kustolang

// Ported from TypeFacts.cs helpers used by return-kind, custom return metadata,
// binder foundations, and the public Go facade.

func PromoteToLong(typ *TypeSymbol) *TypeSymbol {
	if sameType(typ, ScalarTypes.Int.Type()) {
		return ScalarTypes.Long.Type()
	}
	return typ
}

func IsNumericType(typ *TypeSymbol) bool {
	return IsNumeric(typ)
}

func IsDynamicArrayType(typ *TypeSymbol) bool {
	return IsDynamicArray(typ)
}

func GetElementType(typ *TypeSymbol) *TypeSymbol {
	if array := dynamicArraySymbolFromSymbol(typ); array != nil {
		return ScalarTypes.GetDynamic(array.ElementType)
	}
	if sameType(typ, ScalarTypes.Dynamic.Type()) {
		return ScalarTypes.Dynamic.Type()
	}
	return nil
}

func GetWidestScalarType(types []*TypeSymbol) *ScalarSymbol {
	var widest *ScalarSymbol
	for _, typ := range types {
		scalar := scalarFromType(typ)
		if scalar != nil && scalar.Flags()&ScalarFlagNumeric != 0 && scalar != widest {
			if widest == nil || scalar.IsWiderThan(widest) {
				widest = scalar
			}
		}
	}
	return widest
}

func GetCommonType(types []*TypeSymbol, include func(*TypeSymbol) bool, allowedConversions Conversion, defaultType *TypeSymbol) *TypeSymbol {
	var commonType *TypeSymbol
	for _, typ := range types {
		if typ == nil || (include != nil && !include(typ)) {
			continue
		}
		var ok bool
		commonType, ok = TryGetCommonType(commonType, typ, allowedConversions)
		if !ok {
			return defaultType
		}
	}
	if commonType == nil {
		return defaultType
	}
	return commonType
}

func GetCommonScalarType(types []*TypeSymbol, allowedConversions ...Conversion) *TypeSymbol {
	allowed := optionalConversion(ConversionPromotable, allowedConversions)
	return GetCommonType(types, func(typ *TypeSymbol) bool { return typ != nil && typ.SymbolTabularity() == TabularityScalar }, allowed, nil)
}

func GetCommonResultType(types []*TypeSymbol, allowedConversions ...Conversion) *TypeSymbol {
	allowed := optionalConversion(ConversionPromotable, allowedConversions)
	return GetCommonResultTypeWithDefault(types, allowed, nil, false)
}

func GetCommonResultTypeWithDefault(types []*TypeSymbol, allowedConversions Conversion, defaultType *TypeSymbol, ignoreDynamic bool) *TypeSymbol {
	commonType := (*TypeSymbol)(nil)
	for _, typ := range types {
		if typ == nil || (ignoreDynamic && isDynamicTypedSymbol(typ)) {
			continue
		}
		var ok bool
		commonType, ok = TryGetCommonType(commonType, typ, allowedConversions)
		if !ok {
			return defaultType
		}
	}
	if commonType == nil && ignoreDynamic {
		return GetCommonResultTypeWithDefault(types, allowedConversions, defaultType, false)
	}
	if commonType == nil {
		return defaultType
	}
	return commonType
}

// IsCommonArgumentTypeKind identifies parameters whose argument types are
// constrained to agree with their sibling arguments.
func IsCommonArgumentTypeKind(kind ParameterTypeKind) bool {
	switch kind {
	case ParameterTypeCommonScalar, ParameterTypeCommonNumber,
		ParameterTypeCommonSummable, ParameterTypeCommonOrderable,
		ParameterTypeCommonScalarOrDynamic:
		return true
	default:
		return false
	}
}

func GetCommonArgumentType(argumentParameters []*Parameter, argumentTypes []*TypeSymbol) *TypeSymbol {
	return GetCommonArgumentTypeWithDefault(argumentParameters, argumentTypes, nil, false)
}

func GetCommonArgumentTypeIgnoreDynamic(argumentParameters []*Parameter, argumentTypes []*TypeSymbol) *TypeSymbol {
	return GetCommonArgumentTypeWithDefault(argumentParameters, argumentTypes, nil, true)
}

func GetCommonArgumentTypeWithDefault(argumentParameters []*Parameter, argumentTypes []*TypeSymbol, defaultType *TypeSymbol, ignoreDynamic bool) *TypeSymbol {
	var commonType *TypeSymbol
	for i, argType := range argumentTypes {
		if i >= len(argumentParameters) || argType == nil || isErrorTypeSymbol(argType) {
			continue
		}
		parameter := argumentParameters[i]
		if parameter == nil {
			continue
		}
		if ignoreDynamic && isDynamicTypedSymbol(argType) {
			continue
		}

		include := false
		allowed := ConversionPromotable
		switch parameter.TypeKind {
		case ParameterTypeCommonScalar:
			include = IsScalar(argType)
		case ParameterTypeCommonNumber:
			include = IsNumeric(argType)
		case ParameterTypeCommonSummable:
			include = IsSummable(argType)
		case ParameterTypeCommonOrderable:
			include = IsOrderable(argType)
		case ParameterTypeCommonScalarOrDynamic:
			include = IsScalar(argType)
			allowed = ConversionDynamic
		}
		if !include {
			continue
		}
		var ok bool
		commonType, ok = TryGetCommonType(commonType, argType, allowed)
		if !ok {
			return defaultType
		}
	}
	if commonType == nil && ignoreDynamic {
		return GetCommonArgumentTypeWithDefault(argumentParameters, argumentTypes, defaultType, false)
	}
	if commonType == nil {
		return defaultType
	}
	return commonType
}

func GetCommonColumnType(columns []*ColumnSymbol, allowedConversions ...Conversion) *TypeSymbol {
	allowed := optionalConversion(ConversionPromotable, allowedConversions)
	return GetCommonColumnTypeWithDefault(columns, allowed, nil)
}

func GetCommonColumnTypeWithDefault(columns []*ColumnSymbol, allowedConversions Conversion, defaultType *TypeSymbol) *TypeSymbol {
	types := make([]*TypeSymbol, 0, len(columns))
	for _, column := range columns {
		if column != nil {
			types = append(types, column.Type)
		}
	}
	return GetCommonType(types, nil, allowedConversions, defaultType)
}

func UnionColumns(columnsA []*ColumnSymbol, columnsB []*ColumnSymbol) []*ColumnSymbol {
	nameToType := make(map[string]*TypeSymbol, len(columnsA)+len(columnsB))
	for _, column := range columnsA {
		if column != nil {
			nameToType[column.SymbolName()] = column.Type
		}
	}
	for _, column := range columnsB {
		if column == nil {
			continue
		}
		name := column.SymbolName()
		if current, found := nameToType[name]; found {
			if common, ok := TryGetCommonType(current, column.Type, ConversionDynamic); !ok {
				if common == nil {
					common = ScalarTypes.Dynamic.Type()
				}
				nameToType[name] = common
			}
		} else {
			nameToType[name] = column.Type
		}
	}

	columns := make([]*ColumnSymbol, 0, len(nameToType))
	appendSelected := func(source []*ColumnSymbol) {
		for _, column := range source {
			if column == nil {
				continue
			}
			name := column.SymbolName()
			typ, found := nameToType[name]
			if !found {
				continue
			}
			if sameType(column.Type, typ) {
				columns = append(columns, column)
			} else {
				columns = append(columns, NewColumnSymbol(name, typ))
			}
			delete(nameToType, name)
		}
	}
	appendSelected(columnsA)
	appendSelected(columnsB)
	return columns
}

func IntersectColumns(columnsA []*ColumnSymbol, columnsB []*ColumnSymbol) []*ColumnSymbol {
	nameToType := make(map[string]*TypeSymbol, len(columnsB))
	for _, column := range columnsB {
		if column != nil {
			nameToType[column.SymbolName()] = column.Type
		}
	}

	columns := make([]*ColumnSymbol, 0)
	for _, column := range columnsA {
		if column == nil {
			continue
		}
		typeB, found := nameToType[column.SymbolName()]
		if !found {
			continue
		}
		common, ok := TryGetCommonType(column.Type, typeB, ConversionDynamic)
		if !ok || common == nil {
			common = ScalarTypes.Dynamic.Type()
		}
		if sameType(common, column.Type) {
			columns = append(columns, column)
		} else {
			columns = append(columns, NewColumnSymbol(column.SymbolName(), common))
		}
	}
	return columns
}

func IsPromotableTo(sourceType *TypeSymbol, targetType *TypeSymbol) bool {
	return isPromotableTypeTo(sourceType, targetType)
}

func IsInteger(symbol Symbol) bool {
	return scalarHasTypeFactFlag(symbol, ScalarFlagInteger)
}

func IsInterval(symbol Symbol) bool {
	return scalarHasTypeFactFlag(symbol, ScalarFlagInterval)
}

func IsNumeric(symbol Symbol) bool {
	return scalarHasTypeFactFlag(symbol, ScalarFlagNumeric)
}

func IsSummable(symbol Symbol) bool {
	return scalarHasTypeFactFlag(symbol, ScalarFlagSummable)
}

func IsOrderable(symbol Symbol) bool {
	return scalarHasTypeFactFlag(symbol, ScalarFlagOrderable)
}

func IsAnyScalarExceptDynamic(symbol Symbol) bool {
	return isScalarOrDynamicSymbol(symbol) && !isDynamicTypedSymbol(symbol)
}

func IsAnyScalarExceptBool(symbol Symbol) bool {
	return isScalarOrDynamicSymbol(symbol) &&
		!symbolTypeEquals(symbol, ScalarTypes.Bool.Type()) &&
		!symbolTypeEquals(symbol, ScalarTypes.DynamicBool.Type())
}

func IsAnyScalarExceptRealOrBool(symbol Symbol) bool {
	return isScalarOrDynamicSymbol(symbol) &&
		!symbolTypeEquals(symbol, ScalarTypes.Real.Type()) &&
		!symbolTypeEquals(symbol, ScalarTypes.Bool.Type()) &&
		!symbolTypeEquals(symbol, ScalarTypes.DynamicReal.Type()) &&
		!symbolTypeEquals(symbol, ScalarTypes.DynamicBool.Type())
}

func IsNumericOrBool(symbol Symbol) bool {
	return symbolTypeEquals(symbol, ScalarTypes.Bool.Type()) || IsNumeric(symbol)
}

func IsRealOrDecimal(symbol Symbol) bool {
	return symbolTypeEquals(symbol, ScalarTypes.Real.Type()) || symbolTypeEquals(symbol, ScalarTypes.Decimal.Type())
}

func IsStringOrDynamic(symbol Symbol) bool {
	return symbolTypeEquals(symbol, ScalarTypes.String.Type()) || isDynamicTypedSymbol(symbol)
}

func IsStringOrArray(symbol Symbol) bool {
	return symbolTypeEquals(symbol, ScalarTypes.String.Type()) ||
		symbolTypeEquals(symbol, ScalarTypes.DynamicString.Type()) ||
		symbolTypeEquals(symbol, ScalarTypes.DynamicArrayOfString.Type()) ||
		symbolTypeEquals(symbol, ScalarTypes.DynamicArray.Type()) ||
		symbolTypeEquals(symbol, ScalarTypes.Dynamic.Type())
}

func IsIntegerOrArray(symbol Symbol) bool {
	return IsInteger(symbol) ||
		symbolTypeEquals(symbol, ScalarTypes.DynamicLong.Type()) ||
		symbolTypeEquals(symbol, ScalarTypes.DynamicArrayOfLong.Type()) ||
		symbolTypeEquals(symbol, ScalarTypes.DynamicArray.Type()) ||
		symbolTypeEquals(symbol, ScalarTypes.Dynamic.Type())
}

func IsIntegerOrDynamic(symbol Symbol) bool {
	return IsInteger(symbol) ||
		symbolTypeEquals(symbol, ScalarTypes.DynamicLong.Type()) ||
		symbolTypeEquals(symbol, ScalarTypes.Dynamic.Type())
}

func IsDynamicArray(symbol Symbol) bool {
	return dynamicArraySymbolFromSymbol(symbol) != nil || symbolTypeEquals(symbol, ScalarTypes.Dynamic.Type())
}

func IsDynamicBag(symbol Symbol) bool {
	return dynamicBagSymbolFromSymbol(symbol) != nil || symbolTypeEquals(symbol, ScalarTypes.Dynamic.Type())
}

func IsDynamicArrayOrBag(symbol Symbol) bool {
	return symbolTypeEquals(symbol, ScalarTypes.Dynamic.Type()) ||
		dynamicArraySymbolFromSymbol(symbol) != nil ||
		dynamicBagSymbolFromSymbol(symbol) != nil
}

func scalarHasTypeFactFlag(symbol Symbol, flag ScalarFlag) bool {
	if isDynamicTypedSymbol(symbol) {
		return false
	}
	scalar := scalarFromType(returnTypeFromSymbol(symbol))
	return scalar != nil && scalar.Flags()&flag != 0
}

func isScalarOrDynamicSymbol(symbol Symbol) bool {
	if symbol == nil {
		return false
	}
	if isDynamicTypedSymbol(symbol) {
		return true
	}
	return scalarFromType(returnTypeFromSymbol(symbol)) != nil
}

func symbolTypeEquals(symbol Symbol, typ *TypeSymbol) bool {
	return sameType(returnTypeFromSymbol(symbol), typ)
}

func scalarFromType(typ *TypeSymbol) *ScalarSymbol {
	if typ == nil {
		return nil
	}
	for _, scalar := range []*ScalarSymbol{
		ScalarTypes.Dynamic,
		ScalarTypes.Bool,
		ScalarTypes.Int,
		ScalarTypes.Long,
		ScalarTypes.Real,
		ScalarTypes.Decimal,
		ScalarTypes.DateTime,
		ScalarTypes.TimeSpan,
		ScalarTypes.Guid,
		ScalarTypes.TypeLit,
		ScalarTypes.String,
		ScalarTypes.Null,
		ScalarTypes.Unknown,
	} {
		if sameType(typ, scalar.Type()) {
			return scalar
		}
	}
	return nil
}

func TryGetCommonType(typeA *TypeSymbol, typeB *TypeSymbol, allowedConversions Conversion) (*TypeSymbol, bool) {
	switch {
	case typeA == nil && typeB == nil:
		return nil, false
	case typeB == nil:
		return typeA, true
	case typeA == nil:
		return typeB, true
	case sameType(typeA, typeB):
		return typeA, true
	case sameType(typeA, ScalarTypes.Null.Type()):
		return typeB, true
	case sameType(typeB, ScalarTypes.Null.Type()):
		return typeA, true
	case sameType(typeA, ScalarTypes.Unknown.Type()) || sameType(typeB, ScalarTypes.Unknown.Type()):
		return ScalarTypes.Unknown.Type(), true
	case sameType(typeA, ScalarTypes.Dynamic.Type()) || sameType(typeB, ScalarTypes.Dynamic.Type()):
		return ScalarTypes.Dynamic.Type(), true
	case isPromotableTypeTo(typeA, typeB) && IsConversionAllowed(ConversionPromotable, allowedConversions):
		return typeB, true
	case isPromotableTypeTo(typeB, typeA) && IsConversionAllowed(ConversionPromotable, allowedConversions):
		return typeA, true
	}

	arrayA := dynamicArraySymbolFromSymbol(typeA)
	arrayB := dynamicArraySymbolFromSymbol(typeB)
	if arrayA != nil && arrayB != nil {
		if sameType(arrayA.ElementType, arrayB.ElementType) {
			return arrayA.ElementType, true
		}
		return ScalarTypes.DynamicArray.Type(), true
	}

	bagA := dynamicBagSymbolFromSymbol(typeA)
	bagB := dynamicBagSymbolFromSymbol(typeB)
	if bagA != nil && bagB != nil {
		return ScalarTypes.GetDynamicBag(intersectColumns(bagA.Properties(), bagB.Properties())...).Type(), true
	}

	dynamicPrimitiveA := dynamicPrimitiveSymbolFromSymbol(typeA)
	dynamicPrimitiveB := dynamicPrimitiveSymbolFromSymbol(typeB)
	if dynamicPrimitiveA != nil && dynamicPrimitiveB != nil {
		if common, ok := TryGetCommonType(dynamicPrimitiveA.UnderlyingType.Type(), dynamicPrimitiveB.UnderlyingType.Type(), allowedConversions); ok {
			return ScalarTypes.GetDynamic(common), true
		}
	}
	if dynamicPrimitiveA != nil && scalarFromType(typeB) != nil {
		if common, ok := TryGetCommonType(dynamicPrimitiveA.UnderlyingType.Type(), typeB, allowedConversions); ok {
			return ScalarTypes.GetDynamic(common), true
		}
	}
	if scalarFromType(typeA) != nil && dynamicPrimitiveB != nil {
		if common, ok := TryGetCommonType(typeA, dynamicPrimitiveB.UnderlyingType.Type(), allowedConversions); ok {
			return ScalarTypes.GetDynamic(common), true
		}
	}
	if isDynamicTypedSymbol(typeA) && isDynamicTypedSymbol(typeB) {
		return ScalarTypes.Dynamic.Type(), true
	}
	if typeA.SymbolTabularity() == TabularityScalar &&
		typeB.SymbolTabularity() == TabularityScalar &&
		IsConversionAllowed(ConversionDynamic, allowedConversions) {
		return ScalarTypes.Dynamic.Type(), true
	}
	return nil, false
}

func TryGetCommonSymbolType(symbolA Symbol, symbolB Symbol, allowedConversions Conversion) (Symbol, bool) {
	switch {
	case symbolA == nil && symbolB == nil:
		return nil, false
	case symbolB == nil:
		return symbolA, true
	case symbolA == nil:
		return symbolB, true
	case symbolA == symbolB:
		return symbolA, true
	}

	tableA, okA := tableSymbolFromSymbol(symbolA)
	tableB, okB := tableSymbolFromSymbol(symbolB)
	if okA || okB {
		if okA && okB {
			return NewTableSymbol("", intersectColumns(tableA.Columns(), tableB.Columns())...), true
		}
		return nil, false
	}

	common, ok := TryGetCommonType(returnTypeFromSymbol(symbolA), returnTypeFromSymbol(symbolB), allowedConversions)
	if !ok {
		return nil, false
	}
	return common, true
}

func optionalConversion(defaultValue Conversion, conversions []Conversion) Conversion {
	if len(conversions) == 0 {
		return defaultValue
	}
	return conversions[0]
}

func isPromotableTypeTo(sourceType *TypeSymbol, targetType *TypeSymbol) bool {
	sourceScalar := scalarFromType(sourceType)
	targetScalar := scalarFromType(targetType)
	if sourceScalar != nil && targetScalar != nil && targetScalar.IsWiderThan(sourceScalar) {
		return true
	}
	return (dynamicArraySymbolFromSymbol(sourceType) != nil || dynamicBagSymbolFromSymbol(sourceType) != nil) &&
		sameType(targetType, ScalarTypes.Dynamic.Type())
}

func intersectColumns(columnsA []*ColumnSymbol, columnsB []*ColumnSymbol) []*ColumnSymbol {
	return IntersectColumns(columnsA, columnsB)
}
