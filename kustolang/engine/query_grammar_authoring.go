package kustolang

// Ported from src/Kusto.Language/Parser/QueryGrammar.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func QueryGrammarCreateMissingNameDeclaration(Source[LexicalToken], int) *NameDeclaration {
	return createMissingNameDeclaration()
}

func QueryGrammarCreateMissingNameDeclarationExpression(Source[LexicalToken], int) Expression {
	return createMissingNameDeclaration()
}

func QueryGrammarCreateMissingNameReference(Source[LexicalToken], int) Expression {
	return createMissingNameReferenceNode()
}

func QueryGrammarCreateMissingNameToken(Source[LexicalToken], int) *SyntaxToken {
	return createOmittedToken(SyntaxKindIdentifierToken)
}

func QueryGrammarCreateMissingExpression(source Source[LexicalToken], start int) Expression {
	return queryGrammarMissingExpression(source, start)
}

func QueryGrammarCreateMissingValue(Source[LexicalToken], int) Expression {
	return createMissingValueExpression()
}

func QueryGrammarCreateMissingType(Source[LexicalToken], int) TypeExpression {
	return createMissingType()
}

func QueryGrammarCreateMissingLongLiteral(Source[LexicalToken], int) Expression {
	return createMissingLongLiteral()
}

func QueryGrammarCreateMissingRealLiteral(Source[LexicalToken], int) Expression {
	return createMissingRealLiteral()
}

func QueryGrammarCreateMissingStringLiteral(Source[LexicalToken], int) Expression {
	return createMissingStringLiteral()
}

func QueryGrammarCreateMissingBooleanLiteral(Source[LexicalToken], int) Expression {
	return createMissingBoolLiteral()
}

func QueryGrammarCreateMissingTypeOfLiteral(Source[LexicalToken], int) Expression {
	return createMissingTypeOfLiteral()
}

func QueryGrammarCreateMissingJSONValue(Source[LexicalToken], int) Expression {
	return queryGrammarMissingJsonValue()
}

func QueryGrammarCreateMissingJSONPair(Source[LexicalToken], int) *JsonPair {
	return queryGrammarMissingJsonPair()
}

func QueryGrammarCreateMissingQueryOperator(Source[LexicalToken], int) QueryOperator {
	return MustNewBadQueryOperatorTyped(createOmittedToken(SyntaxKindIdentifierToken), DiagnosticQueryOperatorExpected())
}

func QueryGrammarCreateMissingQueryOperatorExpression(Source[LexicalToken], int) Expression {
	return MustNewBadQueryOperatorTyped(createOmittedToken(SyntaxKindIdentifierToken), DiagnosticQueryOperatorExpected())
}

func QueryGrammarCreateMissingNamedExpression(Source[LexicalToken], int) NamedExpression {
	return MustNewSimpleNamedExpressionTyped(
		queryGrammarRecoveryNameDeclaration(),
		createOmittedToken(SyntaxKindEqualToken),
		queryGrammarRecoveryNameReference(),
		DiagnosticMissingName(),
	)
}

func QueryGrammarCreateMissingScanAssignment(Source[LexicalToken], int) *ScanAssignment {
	return MustNewScanAssignmentTyped(
		queryGrammarRecoveryNameReference(),
		createOmittedToken(SyntaxKindEqualEqualToken),
		queryGrammarRecoveryNameReference(),
		DiagnosticMissingName(),
	)
}

func QueryGrammarCreateMissingJoinOnClause(Source[LexicalToken], int) JoinConditionClause {
	return MustNewJoinOnClauseTyped(
		createOmittedToken(SyntaxKindJoinOnClause),
		queryGrammarRecoverySeparatedList[Expression]("SeparatedElement<Expression>"),
		DiagnosticMissingJoinOnClause(),
	)
}

func QueryGrammarCreateMissingFunctionCall(Source[LexicalToken], int) *FunctionCallExpression {
	return queryGrammarRecoveryFunctionCall()
}

func QueryGrammarCreateMissingFunctionCallExpression(Source[LexicalToken], int) Expression {
	return queryGrammarRecoveryFunctionCall()
}

func QueryGrammarCreateMissingSchemaType(Source[LexicalToken], int) *SchemaTypeExpression {
	return MustNewSchemaTypeExpressionTyped(
		createOmittedToken(SyntaxKindOpenParenToken),
		queryGrammarRecoverySeparatedList[Expression]("SeparatedElement<Expression>"),
		createOmittedToken(SyntaxKindCloseParenToken),
		DiagnosticMissingSchemaDeclaration(),
	)
}

func QueryGrammarCreateMissingRowSchema(Source[LexicalToken], int) *RowSchema {
	return MustNewRowSchemaTyped(
		createOmittedToken(SyntaxKindOpenParenToken),
		nil,
		queryGrammarRecoverySeparatedList[*NameAndTypeDeclaration]("SeparatedElement<NameAndTypeDeclaration>"),
		createOmittedToken(SyntaxKindCloseParenToken),
		DiagnosticMissingSchemaDeclaration(),
	)
}

func QueryGrammarCreateMissingInlineExternalTableKindClause(Source[LexicalToken], int) *InlineExternalTableKindClause {
	return MustNewInlineExternalTableKindClauseTyped(
		createOmittedToken(SyntaxKindKindKeyword),
		createOmittedToken(SyntaxKindEqualToken),
		createOmittedToken(SyntaxKindStringLiteralToken),
		DiagnosticMissingExternalTableKind(),
	)
}

func QueryGrammarCreateMissingInlineExternalTableDataFormatClause(Source[LexicalToken], int) *InlineExternalTableDataFormatClause {
	return MustNewInlineExternalTableDataFormatClauseTyped(
		createOmittedToken(SyntaxKindDataFormatKeyword),
		createOmittedToken(SyntaxKindEqualToken),
		createOmittedToken(SyntaxKindStringLiteralToken),
		DiagnosticMissingDataFormat(),
	)
}

func QueryGrammarCreateMissingInlineExternalTableConnectionStringsClause(Source[LexicalToken], int) *InlineExternalTableConnectionStringsClause {
	return MustNewInlineExternalTableConnectionStringsClauseTyped(
		createOmittedToken(SyntaxKindOpenParenToken),
		queryGrammarRecoverySeparatedList[Expression]("SeparatedElement<Expression>"),
		createOmittedToken(SyntaxKindCloseParenToken),
		DiagnosticMissingConnectionStrings(),
	)
}

func QueryGrammarCreateMissingPathFormatTokens(Source[LexicalToken], int) *InlineExternalTablePathFormatPartitionColumnReference {
	return MustNewInlineExternalTablePathFormatPartitionColumnReferenceTyped(
		queryGrammarMissingExpression(nil, 0),
		MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, createOmittedToken(SyntaxKindStringLiteralToken)),
		DiagnosticMissingPathFormatTokens(),
	)
}

func QueryGrammarCreateMissingPartitionColumnDeclaration(source Source[LexicalToken], start int) *PartitionColumnDeclaration {
	return MustNewPartitionColumnDeclarationTyped(
		queryGrammarRecoveryNameDeclaration(),
		createOmittedToken(SyntaxKindColonToken),
		MustNewPrimitiveTypeExpressionTyped(createOmittedToken(SyntaxKindStringKeyword)),
		createOmittedToken(SyntaxKindEqualToken),
		queryGrammarMissingExpression(source, start),
		DiagnosticMissingPartitionColumnDeclaration(),
	)
}

func QueryGrammarCreateMissingEvaluateRowSchema(Source[LexicalToken], int) *EvaluateRowSchema {
	return MustNewEvaluateRowSchemaTyped(
		createOmittedToken(SyntaxKindOpenParenToken),
		nil,
		nil,
		nil,
		queryGrammarRecoverySeparatedList[*NameAndTypeDeclaration]("SeparatedElement<NameAndTypeDeclaration>"),
		createOmittedToken(SyntaxKindCloseParenToken),
		DiagnosticMissingSchemaDeclaration(),
	)
}

func QueryGrammarCreateMissingMakeSeriesExpression(source Source[LexicalToken], start int) *MakeSeriesExpression {
	return MustNewMakeSeriesExpressionTyped(queryGrammarMissingExpression(source, start), nil)
}

func QueryGrammarCreateMissingMvExpandExpression(source Source[LexicalToken], start int) *MvExpandExpression {
	return MustNewMvExpandExpressionTyped(queryGrammarMissingExpression(source, start), nil)
}

func QueryGrammarCreateMissingMvApplyExpression(source Source[LexicalToken], start int) *MvApplyExpression {
	return MustNewMvApplyExpressionTyped(queryGrammarMissingExpression(source, start), nil)
}

func QueryGrammarCreateMissingMvApplySubqueryExpression(source Source[LexicalToken], start int) *MvApplySubqueryExpression {
	return MustNewMvApplySubqueryExpressionTyped(
		createOmittedToken(SyntaxKindOpenParenToken),
		queryGrammarMissingExpression(source, start),
		createOmittedToken(SyntaxKindCloseParenToken),
	)
}

func QueryGrammarCreateMissingForkExpression(source Source[LexicalToken], start int) *ForkExpression {
	return MustNewForkExpressionTyped(
		nil,
		createOmittedToken(SyntaxKindOpenParenToken),
		queryGrammarMissingExpression(source, start),
		createOmittedToken(SyntaxKindCloseParenToken),
	)
}

func QueryGrammarCreateMissingPartitionOperand(source Source[LexicalToken], start int) PartitionOperand {
	return MustNewPartitionSubqueryTyped(
		createOmittedToken(SyntaxKindOpenParenToken),
		queryGrammarMissingExpression(source, start),
		createOmittedToken(SyntaxKindCloseParenToken),
	)
}

func QueryGrammarCreateMissingStatement(Source[LexicalToken], int) Statement {
	return MustNewExpressionStatementTyped(queryGrammarRecoveryNameReference(), DiagnosticMissingStatement())
}

func QueryGrammarCreateMissingStatementElement(source Source[LexicalToken], start int) *SeparatedElementOf[Statement] {
	return AsSeparatedElementOf[Statement](NewSeparatedElement(QueryGrammarCreateMissingStatement(source, start), nil))
}

func QueryGrammarCreateMissingNameAndTypeDeclaration(Source[LexicalToken], int) *NameAndTypeDeclaration {
	return MustNewNameAndTypeDeclarationTyped(
		queryGrammarRecoveryNameDeclaration(),
		createOmittedToken(SyntaxKindColonToken),
		MustNewPrimitiveTypeExpressionTyped(createOmittedToken(SyntaxKindStringKeyword)),
		DiagnosticMissingParameter(),
	)
}

func QueryGrammarCreateMissingFunctionParameter(Source[LexicalToken], int) *FunctionParameter {
	return MustNewFunctionParameterTyped(
		MustNewNameAndTypeDeclarationTyped(
			queryGrammarRecoveryNameDeclaration(),
			createOmittedToken(SyntaxKindColonToken),
			MustNewPrimitiveTypeExpressionTyped(createOmittedToken(SyntaxKindStringKeyword)),
		),
		nil,
		DiagnosticMissingParameter(),
	)
}

func QueryGrammarCreateMissingNamedParameter(Source[LexicalToken], int) *NamedParameter {
	return MustNewNamedParameterTyped(
		queryGrammarRecoveryNameDeclaration(),
		createOmittedToken(SyntaxKindEqualToken),
		queryGrammarRecoveryNameReference(),
		CompletionHintNone,
		DiagnosticMissingParameter(),
	)
}

func QueryGrammarCreateMissingFunctionDeclaration(Source[LexicalToken], int) *FunctionDeclaration {
	return MustNewFunctionDeclarationTyped(
		nil,
		MustNewFunctionParametersTyped(
			createOmittedToken(SyntaxKindOpenParenToken),
			queryGrammarRecoverySeparatedList[*FunctionParameter]("SeparatedElement<FunctionParameter>"),
			createOmittedToken(SyntaxKindCloseParenToken),
		),
		MustNewFunctionBodyTyped(
			createOmittedToken(SyntaxKindOpenBraceToken),
			queryGrammarRecoverySeparatedList[Statement]("SeparatedElement<Statement>"),
			nil,
			nil,
			createOmittedToken(SyntaxKindCloseBraceToken),
		),
		DiagnosticMissingFunctionDeclaration(),
	)
}

func QueryGrammarCreateMissingTokenLiteralFromList(tokens []string) func(Source[LexicalToken], int) Expression {
	diagnostic := DiagnosticTokenExpected(tokens...)
	return func(Source[LexicalToken], int) Expression {
		return MustNewLiteralExpressionTyped(
			SyntaxKindTokenLiteralExpression,
			createOmittedToken(SyntaxKindIdentifierToken),
			diagnostic,
		)
	}
}

func QueryGrammarCreateMissingTokenLiteral(tokens ...string) func(Source[LexicalToken], int) Expression {
	return QueryGrammarCreateMissingTokenLiteralFromList(tokens)
}

func QueryGrammarCreateMissingMaterializedViewCombineBaseClause(source Source[LexicalToken], start int) *MaterializedViewCombineClause {
	return queryGrammarRecoveryMaterializedViewCombineClause("base", source, start)
}

func QueryGrammarCreateMissingMaterializedViewCombineDeltaClause(source Source[LexicalToken], start int) *MaterializedViewCombineClause {
	return queryGrammarRecoveryMaterializedViewCombineClause("delta", source, start)
}

func QueryGrammarCreateMissingMaterializedViewCombineAggregatesClause(source Source[LexicalToken], start int) *MaterializedViewCombineClause {
	return queryGrammarRecoveryMaterializedViewCombineClause("aggregates", source, start)
}

func QueryGrammarCreateMissingMaterializedViewCombineNameClause(source Source[LexicalToken], start int) *MaterializedViewCombineNameClause {
	return MustNewMaterializedViewCombineNameClauseTyped(
		createOmittedToken(SyntaxKindOpenParenToken),
		queryGrammarMissingExpression(source, start),
		createOmittedToken(SyntaxKindCloseParenToken),
	)
}

func queryGrammarRecoveryNameDeclaration() *NameDeclaration {
	return MustNewNameDeclarationTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)))
}

func queryGrammarRecoveryNameReference() *NameReference {
	return MustNewNameReferenceTyped(
		MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)),
		SymbolMatchNone,
	)
}

func queryGrammarRecoverySeparatedList[T SyntaxElement](elementType string) *SeparatedSyntaxListOf[T] {
	return AsSeparatedSyntaxListOf[T](NewTypedSyntaxList(elementType))
}

func queryGrammarRecoveryFunctionCall() *FunctionCallExpression {
	return MustNewFunctionCallExpressionTyped(
		queryGrammarRecoveryNameReference(),
		MustNewExpressionListTyped(
			createOmittedToken(SyntaxKindOpenParenToken),
			queryGrammarRecoverySeparatedList[Expression]("SeparatedElement<Expression>"),
			createOmittedToken(SyntaxKindCloseParenToken),
		),
		DiagnosticMissingFunctionCall(),
	)
}

func queryGrammarRecoveryMaterializedViewCombineClause(name string, source Source[LexicalToken], start int) *MaterializedViewCombineClause {
	return MustNewMaterializedViewCombineClauseTyped(
		createOmittedToken(SyntaxKindMaterializedViewCombineClause),
		createOmittedToken(SyntaxKindOpenParenToken),
		queryGrammarMissingExpression(source, start),
		createOmittedToken(SyntaxKindCloseParenToken),
		DiagnosticMissingMaterializedViewCombineClause(name),
	)
}

var queryGrammarStringOperatorKinds = [...]SyntaxKind{
	SyntaxKindEqualTildeToken,
	SyntaxKindBangTildeToken,
	SyntaxKindHasKeyword,
	SyntaxKindColonToken,
	SyntaxKindNotHasKeyword,
	SyntaxKindHasCsKeyword,
	SyntaxKindNotHasCsKeyword,
	SyntaxKindHasPrefixKeyword,
	SyntaxKindNotHasPrefixKeyword,
	SyntaxKindHasPrefixCsKeyword,
	SyntaxKindNotHasPrefixCsKeyword,
	SyntaxKindHasSuffixKeyword,
	SyntaxKindNotHasSuffixKeyword,
	SyntaxKindHasSuffixCsKeyword,
	SyntaxKindNotHasSuffixCsKeyword,
	SyntaxKindLikeKeyword,
	SyntaxKindNotLikeKeyword,
	SyntaxKindLikeCsKeyword,
	SyntaxKindNotLikeCsKeyword,
	SyntaxKindContainsKeyword,
	SyntaxKindNotContainsKeyword,
	SyntaxKindNotBangContainsKeyword,
	SyntaxKindContainsCsKeyword,
	SyntaxKindContains_CsKeyword,
	SyntaxKindNotContainsCsKeyword,
	SyntaxKindNotBangContainsCsKeyword,
	SyntaxKindStartsWithKeyword,
	SyntaxKindNotStartsWithKeyword,
	SyntaxKindStartsWithCsKeyword,
	SyntaxKindNotStartsWithCsKeyword,
	SyntaxKindEndsWithKeyword,
	SyntaxKindNotEndsWithKeyword,
	SyntaxKindEndsWithCsKeyword,
	SyntaxKindNotEndsWithCsKeyword,
	SyntaxKindMatchesRegexKeyword,
}

func QueryGrammarStringOperatorMap() map[SyntaxKind]SyntaxKind {
	operators := make(map[SyntaxKind]SyntaxKind, len(queryGrammarStringOperatorKinds))
	for _, kind := range queryGrammarStringOperatorKinds {
		operators[kind] = stringOperationKind(kind)
	}
	return operators
}
