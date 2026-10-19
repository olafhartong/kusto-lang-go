package kustolang

// Ported from src/Kusto.Language/FunctionHelpers.cs. See
// third_party/kusto/LICENSE and third_party/kusto/UPSTREAM.json.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

func NewTupleSymbol(columns []*ColumnSymbol, relatedTable ...*TableSymbol) *TupleSymbol {
	return internal.NewTupleSymbol(columns, relatedTable...)
}

func NewTupleSymbolOf(columns ...*ColumnSymbol) *TupleSymbol {
	return internal.NewTupleSymbolOf(columns...)
}

func TupleSymbolFromSchema(schema string) (*TupleSymbol, error) {
	return internal.TupleSymbolFromSchema(schema)
}

func MakePrefixedTuple(context CustomReturnTypeContext, parameterName string, baseTuple *TupleSymbol) *TupleSymbol {
	return internal.MakePrefixedTuple(context, parameterName, baseTuple)
}

func MakeColumnName(nameParts ...string) string {
	return internal.MakeColumnName(nameParts...)
}

func AddReferencedColumn(columns []*ColumnSymbol, context CustomReturnTypeContext, parameterName string) []*ColumnSymbol {
	return internal.AddReferencedColumn(columns, context, parameterName)
}

func AddReferencedColumns(columns []*ColumnSymbol, context CustomReturnTypeContext, parameterName string) []*ColumnSymbol {
	return internal.AddReferencedColumns(columns, context, parameterName)
}

func GetConstantValue(expression Expression, contexts ...CustomReturnTypeContext) string {
	return internal.GetConstantValue(expression, contexts...)
}

func GetConstantValues(expressions *SeparatedSyntaxListOf[Expression]) []string {
	return internal.GetConstantValues(expressions)
}

func GetSummarizeByColumns(arguments []Expression) []*ColumnSymbol {
	return internal.GetSummarizeByColumns(arguments)
}

func MakeValidNameFragment(text string) string {
	return internal.MakeValidNameFragment(text)
}

func IsBoolean(expression Expression, contexts ...CustomReturnTypeContext) bool {
	return internal.IsBoolean(expression, contexts...)
}
