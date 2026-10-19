package kustolang

import (
	"fmt"
	"strings"
)

// FunctionParameterDiagnosticValues formats allowed values for diagnostics.
func FunctionParameterDiagnosticValues(parameter *Parameter) []string {
	if parameter == nil {
		return nil
	}
	values := make([]string, len(parameter.Values))
	for i, value := range parameter.Values {
		values[i] = literalDiagnosticValue(value)
	}
	return values
}

func MakePrefixedTuple(context CustomReturnTypeContext, parameterName string, baseTuple *TupleSymbol) *TupleSymbol {
	if baseTuple == nil {
		return NewTupleSymbol(nil)
	}
	functionPrefix := ""
	if context.Signature != nil && context.Signature.Symbol != nil {
		functionPrefix = context.Signature.Symbol.SymbolName()
		if fn, ok := context.Signature.Symbol.(*FunctionSymbol); ok && fn.ResultNamePrefix() != "" {
			functionPrefix = fn.ResultNamePrefix()
		}
	}
	argumentPrefix := context.GetResultName(context.GetArgument(parameterName))
	columns := make([]*ColumnSymbol, 0, len(baseTuple.Columns()))
	for _, column := range baseTuple.Columns() {
		if column != nil {
			columns = append(columns, column.WithName(MakeColumnName(functionPrefix, argumentPrefix, column.SymbolName())).WithOriginalColumns(column))
		}
	}
	return NewTupleSymbol(columns)
}

func MakeColumnName(nameParts ...string) string {
	if len(nameParts) == 1 {
		return nameParts[0]
	}
	return strings.Join(nameParts, "_")
}

func AddReferencedColumn(columns []*ColumnSymbol, context CustomReturnTypeContext, parameterName string) []*ColumnSymbol {
	if column, ok := context.GetArgumentReferencedSymbol(parameterName).(*ColumnSymbol); ok {
		return append(columns, column)
	}
	return columns
}

func AddReferencedColumns(columns []*ColumnSymbol, context CustomReturnTypeContext, parameterName string) []*ColumnSymbol {
	for _, symbol := range context.GetArgumentReferencedSymbols(parameterName) {
		if column, ok := symbol.(*ColumnSymbol); ok {
			columns = append(columns, column)
		}
	}
	return columns
}

func GetConstantValue(expr Expression, contexts ...CustomReturnTypeContext) string {
	for _, context := range contexts {
		if value := context.GetConstantValue(expr); value != nil {
			return functionConstantValueText(value)
		}
	}
	if value := ExpressionConstantValue(expr); value != nil {
		return functionConstantValueText(value)
	}
	return ""
}

func functionConstantValueText(value any) string {
	if value, ok := value.(bool); ok {
		if value {
			return "True"
		}
		return "False"
	}
	return fmt.Sprint(value)
}

func GetConstantValues(expressions *SeparatedSyntaxListOf[Expression]) []string {
	values := []string{}
	if expressions == nil {
		return values
	}
	for _, separated := range expressions.Elements() {
		if separated != nil {
			values = append(values, GetConstantValue(separated.Element()))
		}
	}
	return values
}

func GetSummarizeByColumns(arguments []Expression) []*ColumnSymbol {
	if len(arguments) == 0 || arguments[0] == nil {
		return []*ColumnSymbol{}
	}
	summarize := GetFirstAncestorOrSelf[*SummarizeOperator](arguments[0])
	if summarize == nil || summarize.ByClause() == nil || summarize.ByClause().Expressions() == nil {
		return []*ColumnSymbol{}
	}
	columns := []*ColumnSymbol{}
	for _, separated := range summarize.ByClause().Expressions().Elements() {
		if separated != nil {
			if column := GetResultColumn(GetUnderlyingExpression(separated.Element())); column != nil {
				columns = append(columns, column)
			}
		}
	}
	return columns
}

func MakeValidNameFragment(text string) string {
	for _, ch := range text {
		if !IsLetterOrDigit(ch) {
			builder := strings.Builder{}
			for _, candidate := range text {
				if IsLetterOrDigit(candidate) {
					builder.WriteRune(candidate)
				} else {
					builder.WriteRune('_')
				}
			}
			return builder.String()
		}
	}
	return text
}

func IsBoolean(expr Expression, contexts ...CustomReturnTypeContext) bool {
	for _, context := range contexts {
		for i, candidate := range context.Arguments {
			if candidate == expr && i < len(context.ArgumentTypes) {
				return context.ArgumentTypes[i] == ScalarTypes.Bool.Type()
			}
		}
	}
	return sameType(ExpressionResultType(expr), ScalarTypes.Bool.Type()) || expr != nil && expr.Kind() == SyntaxKindBooleanLiteralExpression
}
