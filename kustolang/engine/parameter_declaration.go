package kustolang

import "strings"

// Ported from src/Kusto.Language/Symbols/Parameter.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func ParameterDeclaration(parameter *Parameter) string {
	if parameter == nil {
		return ""
	}
	name := BracketNameIfNecessary(parameter.Name)
	if parameter.TypeKind == ParameterTypeDeclared && len(parameter.DeclaredTypes) == 1 {
		if table := parameterDeclaredTable(parameter); table != nil {
			return name + ": " + tableSchemaParameterDeclaration(table)
		}
		typeName := parameterDeclaredTypeName(parameter.DeclaredTypes[0])
		if parameter.DefaultValue != nil {
			return name + ": " + typeName + " = " + parameterDefaultValueText(parameter.DefaultValue)
		}
		return name + ": " + typeName
	}
	if parameter.TypeKind == ParameterTypeTabular {
		return name + ": (*)"
	}
	return name + ": " + ScalarTypes.Dynamic.SymbolName()
}

func ParameterListDeclaration(parameters []*Parameter) string {
	parts := make([]string, 0, len(parameters))
	for _, parameter := range parameters {
		if parameter != nil {
			parts = append(parts, ParameterDeclaration(parameter))
		}
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func parameterDeclaredTable(parameter *Parameter) *TableSymbol {
	if parameter == nil {
		return nil
	}
	return tableSymbolFromDeclaredSymbol(parameter.DeclaredSymbol(0))
}

func tableSymbolFromDeclaredSymbol(symbol Symbol) *TableSymbol {
	switch s := symbol.(type) {
	case *TableSymbol:
		return s
	case *MaterializedViewSymbol:
		if s != nil {
			return &s.TableSymbol
		}
	case *ExternalTableSymbol:
		if s != nil {
			return &s.TableSymbol
		}
	case *StoredQueryResultSymbol:
		if s != nil {
			return &s.TableSymbol
		}
	}
	return nil
}

func tableSchemaParameterDeclaration(table *TableSymbol) string {
	if table == nil {
		return "(*)"
	}
	columns := table.Columns()
	if len(columns) == 0 {
		return "(*)"
	}
	parts := make([]string, 0, len(columns))
	for _, column := range columns {
		if column != nil {
			parts = append(parts, columnParameterDeclaration(column))
		}
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func columnParameterDeclaration(column *ColumnSymbol) string {
	if column == nil {
		return ""
	}
	typeName := ""
	if column.Type != nil {
		typeName = column.Type.SymbolName()
	}
	return BracketNameIfNecessary(column.SymbolName()) + ": " + typeName
}

func parameterDeclaredTypeName(typ *TypeSymbol) string {
	if typ == nil {
		return ScalarTypes.Unknown.SymbolName()
	}
	if text := SchemaDisplayScalarTypeText(typ); text != "" {
		return text
	}
	return typ.SymbolName()
}

func parameterDefaultValueText(value Expression) string {
	if value == nil {
		return ""
	}
	return value.Text(IncludeTriviaInterior)
}
