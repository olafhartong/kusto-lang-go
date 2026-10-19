package kustolang

// Ported from src/Kusto.Language/Parser/QueryGrammar.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// CreateMissingJsonValue preserves the exact pinned C# factory spelling.
func CreateMissingJsonValue(source Source[LexicalToken], start int) Expression {
	return internal.QueryGrammarCreateMissingJSONValue(source, start)
}

// CreateMissingJsonPair preserves the exact pinned C# factory spelling.
func CreateMissingJsonPair(source Source[LexicalToken], start int) *JsonPair {
	return internal.QueryGrammarCreateMissingJSONPair(source, start)
}

func CreateMissingNamedExpression(source Source[LexicalToken], start int) NamedExpression {
	return internal.QueryGrammarCreateMissingNamedExpression(source, start)
}

func CreateMissingScanAssignment(source Source[LexicalToken], start int) *ScanAssignment {
	return internal.QueryGrammarCreateMissingScanAssignment(source, start)
}

func CreateMissingJoinOnClause(source Source[LexicalToken], start int) JoinConditionClause {
	return internal.QueryGrammarCreateMissingJoinOnClause(source, start)
}

func CreateMissingFunctionCall(source Source[LexicalToken], start int) *FunctionCallExpression {
	return internal.QueryGrammarCreateMissingFunctionCall(source, start)
}

func CreateMissingFunctionCallExpression(source Source[LexicalToken], start int) Expression {
	return internal.QueryGrammarCreateMissingFunctionCallExpression(source, start)
}

func CreateMissingSchemaType(source Source[LexicalToken], start int) *SchemaTypeExpression {
	return internal.QueryGrammarCreateMissingSchemaType(source, start)
}

func CreateMissingRowSchema(source Source[LexicalToken], start int) *RowSchema {
	return internal.QueryGrammarCreateMissingRowSchema(source, start)
}

func CreateMissingInlineExternalTableKindClause(source Source[LexicalToken], start int) *InlineExternalTableKindClause {
	return internal.QueryGrammarCreateMissingInlineExternalTableKindClause(source, start)
}

func CreateMissingInlineExternalTableDataFormatClause(source Source[LexicalToken], start int) *InlineExternalTableDataFormatClause {
	return internal.QueryGrammarCreateMissingInlineExternalTableDataFormatClause(source, start)
}

func CreateMissingInlineExternalTableConnectionStringsClause(source Source[LexicalToken], start int) *InlineExternalTableConnectionStringsClause {
	return internal.QueryGrammarCreateMissingInlineExternalTableConnectionStringsClause(source, start)
}

func CreateMissingPathFormatTokens(source Source[LexicalToken], start int) *InlineExternalTablePathFormatPartitionColumnReference {
	return internal.QueryGrammarCreateMissingPathFormatTokens(source, start)
}

func CreateMissingPartitionColumnDeclaration(source Source[LexicalToken], start int) *PartitionColumnDeclaration {
	return internal.QueryGrammarCreateMissingPartitionColumnDeclaration(source, start)
}

func CreateMissingEvaluateRowSchema(source Source[LexicalToken], start int) *EvaluateRowSchema {
	return internal.QueryGrammarCreateMissingEvaluateRowSchema(source, start)
}

func CreateMissingMakeSeriesExpression(source Source[LexicalToken], start int) *MakeSeriesExpression {
	return internal.QueryGrammarCreateMissingMakeSeriesExpression(source, start)
}

func CreateMissingMvExpandExpression(source Source[LexicalToken], start int) *MvExpandExpression {
	return internal.QueryGrammarCreateMissingMvExpandExpression(source, start)
}

func CreateMissingMvApplyExpression(source Source[LexicalToken], start int) *MvApplyExpression {
	return internal.QueryGrammarCreateMissingMvApplyExpression(source, start)
}

func CreateMissingMvApplySubqueryExpression(source Source[LexicalToken], start int) *MvApplySubqueryExpression {
	return internal.QueryGrammarCreateMissingMvApplySubqueryExpression(source, start)
}

func CreateMissingForkExpression(source Source[LexicalToken], start int) *ForkExpression {
	return internal.QueryGrammarCreateMissingForkExpression(source, start)
}

func CreateMissingPartitionOperand(source Source[LexicalToken], start int) PartitionOperand {
	return internal.QueryGrammarCreateMissingPartitionOperand(source, start)
}

func CreateMissingStatement(source Source[LexicalToken], start int) Statement {
	return internal.QueryGrammarCreateMissingStatement(source, start)
}

func CreateMissingStatementElement(source Source[LexicalToken], start int) *SeparatedElementOf[Statement] {
	return internal.QueryGrammarCreateMissingStatementElement(source, start)
}

func CreateMissingNameAndTypeDeclaration(source Source[LexicalToken], start int) *NameAndTypeDeclaration {
	return internal.QueryGrammarCreateMissingNameAndTypeDeclaration(source, start)
}

func CreateMissingFunctionParameter(source Source[LexicalToken], start int) *FunctionParameter {
	return internal.QueryGrammarCreateMissingFunctionParameter(source, start)
}

func CreateMissingNamedParameter(source Source[LexicalToken], start int) *NamedParameter {
	return internal.QueryGrammarCreateMissingNamedParameter(source, start)
}

func CreateMissingFunctionDeclaration(source Source[LexicalToken], start int) *FunctionDeclaration {
	return internal.QueryGrammarCreateMissingFunctionDeclaration(source, start)
}

// CreateMissingTokenLiteral preserves the params-string overload.
func CreateMissingTokenLiteral(tokens ...string) func(Source[LexicalToken], int) Expression {
	return internal.QueryGrammarCreateMissingTokenLiteral(tokens...)
}

// CreateMissingTokenLiteralFromList preserves the IReadOnlyList<string> overload.
func CreateMissingTokenLiteralFromList(tokens []string) func(Source[LexicalToken], int) Expression {
	return internal.QueryGrammarCreateMissingTokenLiteralFromList(tokens)
}

func CreateMissingMaterializedViewCombineBaseClause(source Source[LexicalToken], start int) *MaterializedViewCombineClause {
	return internal.QueryGrammarCreateMissingMaterializedViewCombineBaseClause(source, start)
}

func CreateMissingMaterializedViewCombineDeltaClause(source Source[LexicalToken], start int) *MaterializedViewCombineClause {
	return internal.QueryGrammarCreateMissingMaterializedViewCombineDeltaClause(source, start)
}

func CreateMissingMaterializedViewCombineAggregatesClause(source Source[LexicalToken], start int) *MaterializedViewCombineClause {
	return internal.QueryGrammarCreateMissingMaterializedViewCombineAggregatesClause(source, start)
}

func CreateMissingMaterializedViewCombineNameClause(source Source[LexicalToken], start int) *MaterializedViewCombineNameClause {
	return internal.QueryGrammarCreateMissingMaterializedViewCombineNameClause(source, start)
}
