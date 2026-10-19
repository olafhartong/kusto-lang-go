package kustolang

import "strings"

// Ported from src/Kusto.Language/Symbols/DebugDisplay.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func DebugDisplayText(symbol Symbol) string {
	switch s := symbol.(type) {
	case nil:
		return ""
	case *ColumnSymbol:
		return s.SymbolName() + ": " + DebugDisplayText(s.Type)
	case *TableSymbol:
		return debugDisplayTableText(s)
	case *MaterializedViewSymbol:
		if s == nil {
			return ""
		}
		return debugDisplayTableText(&s.TableSymbol)
	case *ExternalTableSymbol:
		if s == nil {
			return ""
		}
		return debugDisplayTableText(&s.TableSymbol)
	case *StoredQueryResultSymbol:
		if s == nil {
			return ""
		}
		return debugDisplayTableText(&s.TableSymbol)
	case *DatabaseSymbol:
		name := s.SymbolName()
		if s.AlternateName() != "" {
			name = s.AlternateName()
		}
		return GetSingleQuotedStringLiteral(name)
	case *ClusterSymbol:
		return GetSingleQuotedStringLiteral(s.SymbolName())
	case *GroupSymbol:
		return "[" + debugDisplaySymbolList(s.Members()) + "]"
	case *FunctionSymbol:
		signatures := s.Signatures()
		if len(signatures) == 0 {
			return s.SymbolName() + " = ()"
		}
		return s.SymbolName() + " = " + DebugDisplaySignatureText(signatures[0], false, false)
	case *ParameterSymbol:
		return s.SymbolName() + ": " + DebugDisplayText(s.Type)
	case *GraphSymbol:
		if s.NodeShape != nil {
			return s.SymbolName() + ": [Edge" + DebugDisplayText(s.EdgeShape) + ", Node" + DebugDisplayText(s.NodeShape) + "]"
		}
		return s.SymbolName() + ": [Edge" + DebugDisplayText(s.EdgeShape) + "]"
	case *EntityGroupSymbol:
		if s.Definition() != "" {
			return s.SymbolName() + ": " + s.Definition()
		}
		return s.SymbolName() + ": " + debugDisplaySymbolList(s.Members())
	case *PatternSymbol:
		return s.SymbolName() + ": (" + debugDisplayParameterList(s.Parameters()) + ")"
	case *VariableSymbol:
		return s.SymbolName() + " = " + DebugDisplayText(s.Type)
	case *TupleSymbol:
		return "[" + debugDisplaySymbolList(s.Members()) + "]"
	case *ScalarSymbol:
		return s.SymbolName()
	case *DynamicAnySymbol:
		return "dynamic"
	case *DynamicPrimitiveSymbol:
		return "dynamic(" + DebugDisplayText(s.UnderlyingType) + ")"
	case *DynamicArraySymbol:
		if sameType(s.ElementType, ScalarTypes.Dynamic.Type()) {
			return "dynamic([])"
		}
		return "dynamic([" + DebugDisplayText(s.ElementType) + "])"
	case *DynamicBagSymbol:
		properties := s.Properties()
		if len(properties) == 0 {
			return "dynamic({})"
		}
		return "dynamic({" + debugDisplayColumnList(properties) + "})"
	default:
		return symbol.SymbolName()
	}
}

func DebugDisplayParameterText(parameter *Parameter) string {
	if parameter == nil {
		return ""
	}
	if parameter.TypeKind == ParameterTypeDeclared {
		parts := make([]string, 0, maxInt(len(parameter.DeclaredSymbols), len(parameter.DeclaredTypes)))
		for i := 0; i < maxInt(len(parameter.DeclaredSymbols), len(parameter.DeclaredTypes)); i++ {
			parts = append(parts, DebugDisplayText(parameter.DeclaredSymbol(i)))
		}
		return parameter.Name + ": " + strings.Join(parts, "|")
	}
	return parameter.Name + ": <" + string(parameter.TypeKind) + ">"
}

func DebugDisplaySignatureText(signature *Signature, includeSymbolName bool, verbose bool) string {
	if signature == nil {
		return "()"
	}
	parts := make([]string, 0, len(signature.Parameters)+1)
	for _, parameter := range signature.Parameters {
		if parameter == nil {
			continue
		}
		text := debugDisplaySignatureParameterText(signature, parameter, verbose)
		if parameter.IsOptional() {
			text = "[" + text + "]"
		}
		parts = append(parts, text)
	}
	if signature.HasRepeatableParameters {
		parts = append(parts, "...")
	}
	parameters := strings.Join(parts, ", ")
	if includeSymbolName && signature.Symbol != nil && signature.Symbol.SymbolName() != "" {
		return signature.Symbol.SymbolName() + "(" + parameters + ")"
	}
	return "(" + parameters + ")"
}

func debugDisplaySignatureParameterText(signature *Signature, parameter *Parameter, verbose bool) string {
	if verbose {
		if typeText := debugDisplayParameterTypeText(signature, parameter); typeText != "" {
			return parameter.Name + ": " + typeText
		}
	}
	return parameter.Name
}

func debugDisplayParameterTypeText(signature *Signature, parameter *Parameter) string {
	if parameter == nil {
		return ""
	}
	switch parameter.TypeKind {
	case ParameterTypeDeclared:
		parts := make([]string, 0, maxInt(len(parameter.DeclaredSymbols), len(parameter.DeclaredTypes)))
		for i := 0; i < maxInt(len(parameter.DeclaredSymbols), len(parameter.DeclaredTypes)); i++ {
			parts = append(parts, DebugDisplayText(parameter.DeclaredSymbol(i)))
		}
		return strings.Join(parts, "|")
	case ParameterTypeInteger:
		return "integer"
	case ParameterTypeIntegerOrArray:
		return "integer|array"
	case ParameterTypeDynamicArray:
		return "array|dynamic"
	case ParameterTypeDynamicBag:
		return "bag|dynamic"
	case ParameterTypeNumber:
		return "number"
	case ParameterTypeNumberOrBool:
		return "number|bool"
	case ParameterTypeRealOrDecimal:
		return "real|decimal"
	case ParameterTypeSummable:
		return "summable"
	case ParameterTypeStringOrDynamic:
		return "string|dynamic"
	case ParameterTypeStringOrArray:
		return "string|array"
	case ParameterTypeParameter0:
		return debugDisplayParameterTypeText(signature, signatureParameterAt(signature, 0))
	case ParameterTypeParameter1:
		return debugDisplayParameterTypeText(signature, signatureParameterAt(signature, 1))
	case ParameterTypeParameter2:
		return debugDisplayParameterTypeText(signature, signatureParameterAt(signature, 2))
	case ParameterTypeTabular:
		return "()"
	case ParameterTypeCluster:
		return "cluster"
	case ParameterTypeDatabase:
		return "database"
	case ParameterTypeScalar:
		fallthrough
	default:
		return "scalar"
	}
}

func signatureParameterAt(signature *Signature, index int) *Parameter {
	if signature == nil || index < 0 || index >= len(signature.Parameters) {
		return nil
	}
	return signature.Parameters[index]
}

func debugDisplayTableText(table *TableSymbol) string {
	if table == nil {
		return ": ()"
	}
	return table.SymbolName() + ": (" + debugDisplayColumnList(table.Columns()) + ")"
}

func debugDisplayColumnList(columns []*ColumnSymbol) string {
	parts := make([]string, 0, len(columns))
	for _, column := range columns {
		if column != nil {
			parts = append(parts, DebugDisplayText(column))
		}
	}
	return strings.Join(parts, ", ")
}

func debugDisplaySymbolList(symbols []Symbol) string {
	parts := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		if symbol != nil {
			parts = append(parts, DebugDisplayText(symbol))
		}
	}
	return strings.Join(parts, ", ")
}

func debugDisplayParameterList(parameters []*Parameter) string {
	parts := make([]string, 0, len(parameters))
	for _, parameter := range parameters {
		if parameter != nil {
			parts = append(parts, DebugDisplayParameterText(parameter))
		}
	}
	return strings.Join(parts, ", ")
}
