package kustolang

// Ported from the syntax conversion helpers in
// src/Kusto.Language/Symbols/Parameter.cs plus the declared-type binding shape
// used by src/Kusto.Language/Binder/Binder_Misc.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func NewParameterFromFunctionParameter(declaration *FunctionParameter) *Parameter {
	if declaration == nil || declaration.NameAndType() == nil {
		return NewParameter("", ScalarTypes.Unknown.Type())
	}
	nameAndType := declaration.NameAndType()
	name := parameterNameFromDeclaration(nameAndType.Name())
	symbol := declaredSymbolFromTypeExpression(nameAndType.Type())
	if syntaxElementIsMissing(nameAndType.Type()) {
		symbol = ScalarTypes.Unknown.Type()
	}
	var defaultValue Expression
	if declaration.DefaultValue() != nil {
		defaultValue = declaration.DefaultValue().Value()
	}
	return NewParameterSymbolType(name, symbol, WithParameterDefaultValue(defaultValue))
}

func NewParametersFromFunctionParameters(declaration *FunctionParameters) []*Parameter {
	if declaration == nil || declaration.Parameters() == nil {
		return []*Parameter{}
	}
	parameters := declaration.Parameters()
	out := make([]*Parameter, 0, parameters.Count())
	for _, element := range parameters.Elements() {
		if element != nil {
			out = append(out, NewParameterFromFunctionParameter(element.Element()))
		}
	}
	if out == nil {
		return []*Parameter{}
	}
	return out
}

func ParseParameterList(parameters string) []*Parameter {
	return NewParametersFromFunctionParameters(ParseFunctionParameters(parameters))
}

func parameterNameFromDeclaration(declaration *NameDeclaration) string {
	name, err := rowSchemaColumnName(declaration)
	if err == nil {
		return name
	}
	if declaration != nil {
		return declaration.Text(IncludeTriviaInterior)
	}
	return ""
}

func declaredSymbolFromTypeExpression(typeExpression TypeExpression) Symbol {
	switch expression := typeExpression.(type) {
	case *PrimitiveTypeExpression:
		return declaredSymbolFromPrimitiveTypeExpression(expression)
	case *SchemaTypeExpression:
		return declaredTableSymbolFromSchemaTypeExpression(expression)
	default:
		return Error
	}
}

func declaredSymbolFromPrimitiveTypeExpression(expression *PrimitiveTypeExpression) Symbol {
	if expression == nil || expression.Type() == nil {
		return Error
	}
	scalar := ScalarTypes.Get(expression.Type().ValueText())
	if scalar == nil {
		return Error
	}
	return scalar.Type()
}

func declaredTableSymbolFromSchemaTypeExpression(expression *SchemaTypeExpression) *TableSymbol {
	if expression == nil || expression.Columns() == nil {
		return EmptyTableSymbol
	}
	columns := expression.Columns()
	if columns.Count() == 1 {
		if element := columns.At(0); element != nil {
			if _, ok := element.Element().(*StarExpression); ok {
				return EmptyTableSymbol
			}
		}
	}
	out := make([]*ColumnSymbol, 0, columns.Count())
	for _, element := range columns.Elements() {
		if element == nil || element.Element() == nil {
			continue
		}
		declaration, ok := element.Element().(*NameAndTypeDeclaration)
		if !ok || declaration == nil {
			continue
		}
		name, err := rowSchemaColumnName(declaration.Name())
		if err != nil {
			continue
		}
		columnType := returnTypeFromSymbol(declaredSymbolFromTypeExpression(declaration.Type()))
		if columnType == nil {
			columnType = &Error.TypeSymbol
		}
		out = append(out, NewColumnSymbol(name, columnType))
	}
	if len(out) == 0 {
		return NewTableSymbol("")
	}
	return NewTableSymbol("", out...)
}
