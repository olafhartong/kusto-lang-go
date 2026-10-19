package kustolang

// Ported from the Symbol.cs assignability helpers over the currently ported
// native symbol model.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func SymbolAssignableTo(sourceType Symbol, targetType Symbol, allowedConversion ...Conversion) bool {
	conversion := ConversionNone
	if len(allowedConversion) > 0 {
		conversion = allowedConversion[0]
	}
	return symbolsAssignable(targetType, sourceType, conversion)
}

func SymbolAssignableToAny(sourceType Symbol, targetTypes []Symbol, allowedConversion ...Conversion) bool {
	conversion := ConversionNone
	if len(allowedConversion) > 0 {
		conversion = allowedConversion[0]
	}
	for _, targetType := range targetTypes {
		if symbolsAssignable(targetType, sourceType, conversion) {
			return true
		}
	}
	return false
}

func symbolsAssignable(targetType Symbol, sourceType Symbol, allowedConversion Conversion) bool {
	if targetType == sourceType {
		return true
	}
	if targetType == nil || sourceType == nil {
		return false
	}

	sourceScalar := scalarSymbolFromSymbol(sourceType)
	targetScalar := scalarSymbolFromSymbol(targetType)
	if targetScalar != nil && sourceScalar != nil {
		if sameType(sourceScalar.Type(), ScalarTypes.Unknown.Type()) && IsScalar(targetType) {
			return true
		}
		if sameType(targetScalar.Type(), ScalarTypes.Unknown.Type()) && IsScalar(sourceType) {
			return true
		}
	}

	if IsScalar(targetType) {
		if tuple, ok := sourceType.(*TupleSymbol); ok {
			columns := tuple.Columns()
			if len(columns) == 1 {
				return symbolsAssignable(targetType, columns[0].Type, ConversionNone)
			}
		}
	}

	if isDynamicAnySymbol(targetType) {
		if isDynamicTypedSymbol(sourceType) {
			return true
		}
		if sourceScalar != nil && allowedConversion >= ConversionDynamic {
			return true
		}
	}

	if isUntypedDynamicArray(targetType) && dynamicArraySymbolFromSymbol(sourceType) != nil && allowedConversion != ConversionNone {
		return true
	}

	if isUntypedDynamicBag(targetType) && dynamicBagSymbolFromSymbol(sourceType) != nil && allowedConversion != ConversionNone {
		return true
	}

	if targetPrimitive := dynamicPrimitiveSymbolFromSymbol(targetType); targetPrimitive != nil {
		if sourcePrimitive := dynamicPrimitiveSymbolFromSymbol(sourceType); sourcePrimitive != nil {
			return symbolsAssignable(targetPrimitive.UnderlyingType.Type(), sourcePrimitive.UnderlyingType.Type(), allowedConversion)
		}
	}

	if targetArray := dynamicArraySymbolFromSymbol(targetType); targetArray != nil {
		sourceArray := dynamicArraySymbolFromSymbol(sourceType)
		return sourceArray != nil && symbolsAssignable(sourceArray.ElementType, targetArray.ElementType, allowedConversion)
	}

	if targetBag := dynamicBagSymbolFromSymbol(targetType); targetBag != nil {
		sourceBag := dynamicBagSymbolFromSymbol(sourceType)
		return sourceBag != nil && bagsAssignable(targetBag, sourceBag)
	}

	if targetScalar != nil && sourceScalar != nil {
		switch allowedConversion {
		case ConversionPromotable:
			return scalarPromotableTo(sourceScalar, targetScalar)
		case ConversionCompatible:
			return scalarPromotableTo(sourceScalar, targetScalar) || scalarPromotableTo(targetScalar, sourceScalar)
		case ConversionAny:
			return true
		default:
			return false
		}
	}

	if targetType.SymbolKind() != sourceType.SymbolKind() {
		return false
	}

	if targetTable, ok := tableSymbolFromSymbol(targetType); ok {
		sourceTable, ok := tableSymbolFromSymbol(sourceType)
		return ok && tablesAssignable(targetTable, sourceTable)
	}

	switch target := targetType.(type) {
	case *ColumnSymbol:
		source, ok := sourceType.(*ColumnSymbol)
		return ok && target.SymbolName() == source.SymbolName() && symbolsAssignable(target.Type, source.Type, allowedConversion)
	case *TupleSymbol:
		return symbolMembersAssignable(target, sourceType)
	case *GroupSymbol:
		return symbolMembersAssignable(target, sourceType)
	}

	return false
}

func symbolMembersAssignable(target Symbol, source Symbol) bool {
	targetMembers := target.Members()
	sourceMembers := source.Members()
	if len(targetMembers) != len(sourceMembers) {
		return false
	}
	for i := range targetMembers {
		if !symbolsAssignable(targetMembers[i], sourceMembers[i], ConversionNone) {
			return false
		}
	}
	return true
}

func bagsAssignable(targetBag *DynamicBagSymbol, sourceBag *DynamicBagSymbol) bool {
	for _, targetProperty := range targetBag.Properties() {
		sourceProperty, ok := sourceBag.TryGetProperty(targetProperty.SymbolName())
		if !ok || !symbolsAssignable(targetProperty.Type, sourceProperty.Type, ConversionAny) {
			return false
		}
	}
	return true
}

func tablesAssignable(target *TableSymbol, source *TableSymbol) bool {
	for _, targetColumn := range target.Columns() {
		sourceColumn := tableColumn(source, targetColumn.SymbolName())
		if sourceColumn == nil || !symbolsAssignable(targetColumn.Type, sourceColumn.Type, ConversionNone) {
			return false
		}
	}
	return true
}

func tableColumn(table *TableSymbol, name string) *ColumnSymbol {
	if table == nil {
		return nil
	}
	for _, column := range table.Columns() {
		if column != nil && column.SymbolName() == name {
			return column
		}
	}
	return nil
}

func scalarPromotableTo(source *ScalarSymbol, target *ScalarSymbol) bool {
	return source != nil && target != nil && target.IsWiderThan(source)
}

func scalarSymbolFromSymbol(symbol Symbol) *ScalarSymbol {
	switch s := symbol.(type) {
	case *ScalarSymbol:
		return s
	case *TypeSymbol:
		return scalarFromType(s)
	default:
		return nil
	}
}

func dynamicPrimitiveSymbolFromSymbol(symbol Symbol) *DynamicPrimitiveSymbol {
	switch s := symbol.(type) {
	case *DynamicPrimitiveSymbol:
		return s
	case *TypeSymbol:
		if primitive, ok := DynamicPrimitiveForType(s); ok {
			return primitive
		}
	}
	return nil
}

func dynamicArraySymbolFromSymbol(symbol Symbol) *DynamicArraySymbol {
	switch s := symbol.(type) {
	case *DynamicArraySymbol:
		return s
	case *TypeSymbol:
		if array, ok := DynamicArrayForType(s); ok {
			return array
		}
	}
	return nil
}

func dynamicBagSymbolFromSymbol(symbol Symbol) *DynamicBagSymbol {
	switch s := symbol.(type) {
	case *DynamicBagSymbol:
		return s
	case *TypeSymbol:
		if bag, ok := DynamicBagForType(s); ok {
			return bag
		}
		if sameType(s, ScalarTypes.DynamicBag.Type()) {
			return ScalarTypes.DynamicBag
		}
	}
	return nil
}

func isDynamicAnySymbol(symbol Symbol) bool {
	typ := typeSymbolFromSymbol(symbol)
	return sameType(typ, ScalarTypes.Dynamic.Type())
}

func isDynamicTypedSymbol(symbol Symbol) bool {
	return isDynamicAnySymbol(symbol) ||
		dynamicPrimitiveSymbolFromSymbol(symbol) != nil ||
		dynamicArraySymbolFromSymbol(symbol) != nil ||
		dynamicBagSymbolFromSymbol(symbol) != nil
}

func isUntypedDynamicArray(symbol Symbol) bool {
	array := dynamicArraySymbolFromSymbol(symbol)
	return array != nil && sameType(array.Type(), ScalarTypes.DynamicArray.Type())
}

func isUntypedDynamicBag(symbol Symbol) bool {
	bag := dynamicBagSymbolFromSymbol(symbol)
	return bag != nil && sameType(bag.Type(), ScalarTypes.DynamicBag.Type())
}

func typeSymbolFromSymbol(symbol Symbol) *TypeSymbol {
	switch s := symbol.(type) {
	case interface{ Type() *TypeSymbol }:
		return s.Type()
	case *TypeSymbol:
		return s
	case *ColumnSymbol:
		return s.Type
	default:
		return nil
	}
}
