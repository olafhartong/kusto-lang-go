package environment

import (
	"sort"
	"strings"
	"unicode"

	kustolang "github.com/olafhartong/kusto-lang-go/kustolang/engine"
)

func schemaFromTableDetails(details TableDetails, origin bool) TableSchema {
	out := NewTableSchema()
	out.MissingColumnType = kustolang.Error.SymbolName()
	keys := make([]string, 0, len(details))
	for k := range details {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		out.Set(k, details[k], origin)
	}
	return out
}

func publicTypeName(typ *kustolang.TypeSymbol) string {
	if typ == nil {
		return "dynamic"
	}
	return normalizeType(typ.SymbolName())
}

const nonScalarNamedTypePrefix = "\x00non_scalar_type:"

const emptyTableTypeName = "\x00empty_table_type"

func columnFromTypeSymbol(typ *kustolang.TypeSymbol, origin bool) Column {
	properties, elementProperties := dynamicPropertiesFromTypeSymbol(typ)
	elementElementProperties, elementElementType := dynamicNestedElementMetadataFromTypeSymbol(typ)
	column := Column{
		Type:                     publicTypeName(typ),
		Origin:                   origin,
		ElementType:              dynamicScalarElementTypeFromTypeSymbol(typ),
		ElementElementType:       elementElementType,
		Properties:               properties,
		ElementProperties:        elementProperties,
		ElementElementProperties: elementElementProperties,
	}
	if _, primitive := kustolang.DynamicPrimitiveForType(typ); primitive {
		column.ExactType = typ
	} else if _, array := kustolang.DynamicArrayForType(typ); array {
		column.ExactType = typ
	} else if _, bag := kustolang.DynamicBagForType(typ); bag {
		column.ExactType = typ
	}
	return column
}

func typeNameForSchemaCopy(typ string) string {
	if typ == "" {
		return emptyTableTypeName
	}
	return typ
}

func dynamicPropertiesFromTypeSymbol(typ *kustolang.TypeSymbol) (DynamicProperties, DynamicProperties) {
	bag, _ := kustolang.DynamicBagForType(typ)
	elementBag, _ := dynamicArrayElementBagForType(typ)
	return dynamicBagPropertiesFromSymbol(bag), dynamicBagPropertiesFromSymbol(elementBag)
}

func dynamicNestedElementMetadataFromTypeSymbol(typ *kustolang.TypeSymbol) (DynamicProperties, string) {
	outer, ok := kustolang.DynamicArrayForType(typ)
	if !ok || outer == nil {
		return nil, ""
	}
	inner, ok := kustolang.DynamicArrayForType(outer.ElementType)
	if !ok || inner == nil || inner.ElementType == nil {
		return nil, ""
	}
	if bag, ok := kustolang.DynamicBagForType(inner.ElementType); ok {
		return dynamicBagPropertiesFromSymbol(bag), ""
	}
	if _, nested := kustolang.DynamicArrayForType(inner.ElementType); nested {
		return nil, ""
	}
	return nil, publicTypeName(inner.ElementType)
}

func dynamicScalarElementTypeFromTypeSymbol(typ *kustolang.TypeSymbol) string {
	array, ok := kustolang.DynamicArrayForType(typ)
	if !ok || array == nil || array.ElementType == nil {
		return ""
	}
	if _, ok := kustolang.DynamicBagForType(array.ElementType); ok {
		return ""
	}
	if _, ok := kustolang.DynamicArrayForType(array.ElementType); ok {
		return ""
	}
	return publicTypeName(array.ElementType)
}

func dynamicArrayElementBagForType(typ *kustolang.TypeSymbol) (*kustolang.DynamicBagSymbol, bool) {
	array, ok := kustolang.DynamicArrayForType(typ)
	if !ok || array == nil {
		return nil, false
	}
	return kustolang.DynamicBagForType(array.ElementType)
}

func dynamicBagPropertiesFromSymbol(bag *kustolang.DynamicBagSymbol) DynamicProperties {
	return dynamicBagPropertiesFromSymbolSeen(bag, map[*kustolang.DynamicBagSymbol]bool{})
}

func isIdentifierLike(s string) bool {
	s = cleanIdentifier(strings.TrimSpace(s))
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 && !unicode.IsLetter(r) && r != '_' && r != '$' {
			return false
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '$' {
			return false
		}
	}
	return true
}

func cleanIdentifier(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && ((s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"')) {
		return s[1 : len(s)-1]
	}
	if strings.HasPrefix(s, "['") && strings.HasSuffix(s, "']") {
		return s[2 : len(s)-2]
	}
	if strings.HasPrefix(s, "[\"") && strings.HasSuffix(s, "\"]") {
		return s[2 : len(s)-2]
	}
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		inner := strings.TrimSpace(s[1 : len(s)-1])
		if strings.Contains(inner, "*") && !strings.ContainsAny(inner, "[]'\"") {
			return inner
		}
	}
	return s
}

func dynamicBagPropertiesFromSymbolSeen(bag *kustolang.DynamicBagSymbol, seen map[*kustolang.DynamicBagSymbol]bool) DynamicProperties {
	if bag == nil || len(bag.Properties()) == 0 {
		return nil
	}
	if seen[bag] {
		return nil
	}
	seen[bag] = true
	properties := DynamicProperties{}
	for _, property := range bag.Properties() {
		if property == nil || property.SymbolName() == "" {
			continue
		}
		nestedBag, _ := kustolang.DynamicBagForType(property.Type)
		elementBag, _ := dynamicArrayElementBagForType(property.Type)
		properties[property.SymbolName()] = DynamicProperty{
			Type:              publicTypeName(property.Type),
			ElementType:       dynamicScalarElementTypeFromTypeSymbol(property.Type),
			Properties:        dynamicBagPropertiesFromSymbolSeen(nestedBag, seen),
			ElementProperties: dynamicBagPropertiesFromSymbolSeen(elementBag, seen),
		}
	}
	delete(seen, bag)
	return properties
}
