package kustolang

// Ported from src/Kusto.Language/Symbols/CombineKind.cs and the column/table
// combine behavior used by ColumnSymbol.cs, TableSymbol.cs and binder helpers.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type CombineKind int

const (
	CombineKindUnifySameNameAndType CombineKind = iota
	CombineKindUnifySameName
	CombineKindUniqueNames
)

func CombineColumns(kind CombineKind, columnSets ...[]*ColumnSymbol) []*ColumnSymbol {
	var columns []*ColumnSymbol
	for _, set := range columnSets {
		columns = append(columns, set...)
	}

	switch kind {
	case CombineKindUnifySameNameAndType:
		return unifyColumnsWithSameNameAndType(columns)
	case CombineKindUnifySameName:
		return unifyColumnsWithSameName(columns)
	case CombineKindUniqueNames:
		return makeColumnNamesUnique(columns)
	default:
		return append([]*ColumnSymbol{}, columns...)
	}
}

func CombineTableSymbolsWithKind(kind CombineKind, tables ...*TableSymbol) *TableSymbol {
	columnSets := make([][]*ColumnSymbol, 0, len(tables))
	for _, table := range tables {
		if table != nil {
			columnSets = append(columnSets, table.Columns())
		}
	}
	return NewTableSymbol("", CombineColumns(kind, columnSets...)...)
}

func unifyColumnsWithSameName(columns []*ColumnSymbol) []*ColumnSymbol {
	nameOrder := orderedColumnNames(columns)
	out := make([]*ColumnSymbol, 0, len(nameOrder))
	for _, name := range nameOrder {
		group := columnsByName(columns, name)
		if len(group) == 1 {
			out = append(out, group[0])
			continue
		}
		commonType := GetCommonColumnType(group)
		if commonType == nil {
			commonType = ScalarTypes.Dynamic.Type()
		}
		if sameColumnTypes(group, commonType) {
			out = append(out, group[0])
			continue
		}
		out = append(out, NewColumnSymbolFull(name, commonType, "", group, nil, nil))
	}
	return out
}

func unifyColumnsWithSameNameAndType(columns []*ColumnSymbol) []*ColumnSymbol {
	columnMap := NewColumnMapFromColumns(columns)
	uniqueNames := NewUniqueNameTable()
	out := make([]*ColumnSymbol, 0, len(columns))
	for _, column := range columns {
		if column == nil {
			continue
		}
		name := column.SymbolName()
		if !columnMap.HasColumns(name) {
			continue
		}
		if columnMap.HasMultipleTypes(name) {
			for _, typ := range columnMap.GetTypes(name) {
				sameTypeColumns := columnMap.GetColumnsByType(name, typ)
				newType := GetCommonColumnType(sameTypeColumns, ConversionNone)
				if newType == nil {
					newType = typ
				}
				suggestedName := name + "_" + typeNameForColumnSuffix(newType)
				newName := uniqueNames.GetOrAddName(suggestedName)
				out = append(out, NewColumnSymbolFull(newName, newType, "", sameTypeColumns, nil, nil))
			}
		} else if columnMap.HasMultipleColumns(name, column.Type) {
			group := columnMap.GetColumnsByType(name, column.Type)
			newType := GetCommonColumnType(group, ConversionNone)
			if newType == nil {
				newType = column.Type
			}
			newColumn := NewColumnSymbolFull(name, newType, "", group, nil, nil)
			out = append(out, getUniqueColumn(newColumn, uniqueNames))
		} else {
			out = append(out, getUniqueColumn(column, uniqueNames))
		}
		columnMap.Remove(name)
	}
	return out
}

func makeColumnNamesUnique(columns []*ColumnSymbol) []*ColumnSymbol {
	firstIndex := map[string]int{}
	orderedNames := []string{}
	for index, column := range columns {
		if column == nil {
			continue
		}
		name := column.SymbolName()
		if _, ok := firstIndex[name]; !ok {
			firstIndex[name] = index
			orderedNames = append(orderedNames, name)
		}
	}

	uniqueNames := NewUniqueNameTable()
	uniqueNames.AddNames(orderedNames)
	out := make([]*ColumnSymbol, 0, len(columns))
	for index, column := range columns {
		if column == nil {
			continue
		}
		if firstIndex[column.SymbolName()] != index {
			out = append(out, getUniqueColumn(column, uniqueNames))
		} else {
			out = append(out, column)
		}
	}
	return out
}

func orderedColumnNames(columns []*ColumnSymbol) []string {
	seen := map[string]struct{}{}
	names := []string{}
	for _, column := range columns {
		if column == nil {
			continue
		}
		name := column.SymbolName()
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	return names
}

func orderedColumnTypes(columns []*ColumnSymbol) []*TypeSymbol {
	var types []*TypeSymbol
	for _, column := range columns {
		if column == nil {
			continue
		}
		found := false
		for _, typ := range types {
			if sameType(typ, column.Type) {
				found = true
				break
			}
		}
		if !found {
			types = append(types, column.Type)
		}
	}
	return types
}

func columnsByName(columns []*ColumnSymbol, name string) []*ColumnSymbol {
	var out []*ColumnSymbol
	for _, column := range columns {
		if column != nil && column.SymbolName() == name {
			out = append(out, column)
		}
	}
	return out
}

func columnsByExactType(columns []*ColumnSymbol, typ *TypeSymbol) []*ColumnSymbol {
	var out []*ColumnSymbol
	for _, column := range columns {
		if column != nil && sameType(column.Type, typ) {
			out = append(out, column)
		}
	}
	return out
}

func sameColumnTypes(columns []*ColumnSymbol, typ *TypeSymbol) bool {
	for _, column := range columns {
		if column == nil {
			continue
		}
		if !sameType(column.Type, typ) {
			return false
		}
	}
	return true
}

func typeNameForColumnSuffix(typ *TypeSymbol) string {
	if typ == nil || typ.SymbolName() == "" {
		return ScalarTypes.Dynamic.SymbolName()
	}
	return typ.SymbolName()
}

func getUniqueColumn(column *ColumnSymbol, uniqueNames *UniqueNameTable) *ColumnSymbol {
	if column == nil {
		return nil
	}
	uniqueName := uniqueNames.GetOrAddName(column.SymbolName())
	if uniqueName != column.SymbolName() {
		return NewColumnSymbolFull(uniqueName, column.Type, column.Description, []*ColumnSymbol{column}, column.Source, column.Examples)
	}
	return column
}
