package kustolang

import "strings"

const declarationRootPipeOnlyRecoveryPayload = "DeclarationRootPipeOnlyRecovery"

// This file implements query-operator parsing.

func (p *queryParserState) parsePipeExpression() Expression {
	var expr Expression
	if p.scanPossibleQueryOperator() {
		expr = queryOperatorExpression(p.parseUnpipedQueryOperator())
		if expr == nil {
			expr = queryOperatorExpression(p.parsePipedQueryOperator())
		}
		if expr == nil {
			expr = p.parseUnnamedExpression()
		}
	} else {
		expr = p.parseUnnamedExpression()
		if expr == nil {
			expr = queryOperatorExpression(p.parseUnpipedQueryOperator())
		}
		if expr == nil {
			expr = queryOperatorExpression(p.parsePipedQueryOperator())
		}
	}
	if expr == nil {
		return nil
	}
	for p.peekToken().Kind == SyntaxKindBarToken {
		bar := p.parseToken()
		operator := p.parseRequiredQueryOperator()
		expr = MustNewPipeExpressionTyped(expr, bar, operator)
	}
	return expr
}

// parsePipeOnlyRecoveryRootExpression matches recovery at declaration roots.
// A terminal operator keyword is scalar there; the same keyword after a bar
// remains an operator.
func (p *queryParserState) parsePipeOnlyRecoveryRootExpression() Expression {
	if p.peekToken().Kind == SyntaxKindCountKeyword || p.peekToken().Kind == SyntaxKindEvaluateKeyword {
		switch p.peekToken(1).Kind {
		case SyntaxKindSemicolonToken, SyntaxKindCloseBraceToken, SyntaxKindEndOfTextToken, SyntaxKindNone:
			token := p.parseToken()
			reference := MustNewNameReferenceTyped(MustNewTokenNameTyped(token), SymbolMatchNone)
			reference.SyntaxNode.setPayload(declarationRootPipeOnlyRecoveryPayload, true)
			return reference
		}
	}
	return p.parseExpression()
}

func (p *queryParserState) scanPossibleQueryOperator() bool {
	kind := p.peekToken().Kind
	switch kind {
	case SyntaxKindEvaluateKeyword:
		return p.scanName(1) > 0 || p.scanSpecificQueryOperatorParameter(QueryOperatorParameters.EvaluateParameters, false, 1) != nil
	case SyntaxKindFacetKeyword:
		return p.peekToken(1).Kind == SyntaxKindByKeyword
	case SyntaxKindPartitionKeyword:
		return p.peekToken(1).Kind == SyntaxKindByKeyword || p.scanSpecificQueryOperatorParameter(QueryOperatorParameters.PartitionParameters, false, 1) != nil
	case SyntaxKindRangeKeyword:
		nameLen := p.scanName(1)
		return nameLen > 0 && p.peekToken(1+nameLen).Kind == SyntaxKindFromKeyword
	case SyntaxKindInlineExternalTableKeyword:
		return p.peekToken(1).Kind == SyntaxKindOpenParenToken || p.peekToken(1).Kind == SyntaxKindKindKeyword
	case SyntaxKindReduceKeyword:
		return p.peekToken(1).Kind == SyntaxKindByKeyword || p.scanSpecificQueryOperatorParameter(QueryOperatorParameters.ReduceParameters, false, 1) != nil
	case SyntaxKindRenderKeyword:
		return p.scanTokenText(p.renderChartTypes(), 1) || p.scanIdentifierOrKeywordAsIdentifier(1)
	case SyntaxKindForkKeyword, SyntaxKindLookupKeyword, SyntaxKindPartitionByKeyword:
		return false
	case SyntaxKindAsKeyword,
		SyntaxKindAssertSchemaKeyword,
		SyntaxKindConsumeKeyword,
		SyntaxKindCountKeyword,
		SyntaxKindDistinctKeyword,
		SyntaxKindExtendKeyword,
		SyntaxKindExecuteAndCacheKeyword,
		SyntaxKindFindKeyword,
		SyntaxKindFilterKeyword,
		SyntaxKindGetSchemaKeyword,
		SyntaxKindGraphMatchKeyword,
		SyntaxKindGraphToTableKeyword,
		SyntaxKindInvokeKeyword,
		SyntaxKindJoinKeyword,
		SyntaxKindMacroExpandKeyword,
		SyntaxKindMakeGraphKeyword,
		SyntaxKindGraphMarkComponentsKeyword,
		SyntaxKindMakeSeriesKeyword,
		SyntaxKindMvApplyKeyword,
		SyntaxKindMvDashApplyKeyword,
		SyntaxKindMvExpandKeyword,
		SyntaxKindMvDashExpandKeyword,
		SyntaxKindOrderKeyword,
		SyntaxKindParseKeyword,
		SyntaxKindParseKvKeyword,
		SyntaxKindParseWhereKeyword,
		SyntaxKindPrintKeyword,
		SyntaxKindProjectKeyword,
		SyntaxKindProjectAwayKeyword,
		SyntaxKindProjectByNamesKeyword,
		SyntaxKindProjectKeepKeyword,
		SyntaxKindProjectRenameKeyword,
		SyntaxKindProjectReorderKeyword,
		SyntaxKindSampleDistinctKeyword,
		SyntaxKindSampleKeyword,
		SyntaxKindScanKeyword,
		SyntaxKindSearchKeyword,
		SyntaxKindSerializeKeyword,
		SyntaxKindSortKeyword,
		SyntaxKindSummarizeKeyword,
		SyntaxKindTakeKeyword,
		SyntaxKindLimitKeyword,
		SyntaxKindTopHittersKeyword,
		SyntaxKindTopNestedKeyword,
		SyntaxKindTopKeyword,
		SyntaxKindUnionKeyword,
		SyntaxKindWhereKeyword:
		return !p.canBeIdentifier(kind)
	default:
		return false
	}
}

func (p *queryParserState) scanPartitionOperatorStart() bool {
	if p.peekToken(1).Kind == SyntaxKindByKeyword || p.scanSpecificQueryOperatorParameter(QueryOperatorParameters.PartitionParameters, false, 1) != nil {
		return true
	}
	for i := 1; i <= 5; i++ {
		switch p.peekToken(i).Kind {
		case SyntaxKindEqualToken:
			return true
		case SyntaxKindByKeyword, SyntaxKindBarToken, SyntaxKindEndOfTextToken, SyntaxKindNone:
			return false
		}
	}
	return false
}

func (p *queryParserState) parseRequiredQueryOperator() QueryOperator {
	if operator := p.parsePipedQueryOperator(); !isNilSyntaxElement(operator) {
		return operator
	}
	if operator := p.parseUnpipedQueryOperator(); !isNilSyntaxElement(operator) {
		return operator
	}
	if operator := p.parseBadQueryOperator(); !isNilSyntaxElement(operator) {
		return operator
	}
	return p.createMissingQueryOperator()
}

func (p *queryParserState) parseUnpipedQueryOperator() QueryOperator {
	switch p.peekToken().Kind {
	case SyntaxKindEvaluateKeyword:
		return p.parseEvaluateOperator()
	case SyntaxKindFindKeyword:
		return p.parseFindOperator()
	case SyntaxKindMacroExpandKeyword:
		return p.parseMacroExpandOperator()
	case SyntaxKindPrintKeyword:
		return p.parsePrintOperator()
	case SyntaxKindSearchKeyword:
		return p.parseSearchOperator()
	case SyntaxKindUnionKeyword:
		return p.parseUnionOperator()
	default:
		return nil
	}
}

func (p *queryParserState) parsePipedQueryOperator() QueryOperator {
	switch p.peekToken().Kind {
	case SyntaxKindAsKeyword:
		return p.parseAsOperator()
	case SyntaxKindAssertSchemaKeyword:
		return p.parseAssertSchemaOperator()
	case SyntaxKindConsumeKeyword:
		return p.parseConsumeOperator()
	case SyntaxKindCountKeyword:
		if operator := p.parseCountOperator(); operator != nil {
			return operator
		}
		return nil
	case SyntaxKindDistinctKeyword:
		return p.parseDistinctOperator()
	case SyntaxKindExtendKeyword:
		return p.parseExtendOperator()
	case SyntaxKindEvaluateKeyword:
		return p.parseEvaluateOperator()
	case SyntaxKindExecuteAndCacheKeyword:
		return p.parseExecuteAndCacheOperator()
	case SyntaxKindFacetKeyword:
		return p.parseFacetOperator()
	case SyntaxKindForkKeyword:
		return p.parseForkOperator()
	case SyntaxKindFindKeyword:
		return p.parseFindOperator()
	case SyntaxKindFilterKeyword, SyntaxKindWhereKeyword:
		return p.parseFilterOperator()
	case SyntaxKindGetSchemaKeyword:
		return p.parseGetSchemaOperator()
	case SyntaxKindInvokeKeyword:
		return p.parseInvokeOperator()
	case SyntaxKindInlineExternalTableKeyword:
		return p.parseInlineExternalTableExpression()
	case SyntaxKindJoinKeyword:
		return p.parseJoinOperator()
	case SyntaxKindLookupKeyword:
		return p.parseLookupOperator()
	case SyntaxKindMakeGraphKeyword:
		return p.parseMakeGraphOperator()
	case SyntaxKindMakeSeriesKeyword:
		return p.parseMakeSeriesOperator()
	case SyntaxKindGraphMatchKeyword:
		return p.parseGraphMatchOperator()
	case SyntaxKindGraphMarkComponentsKeyword:
		return p.parseGraphMarkComponentsOperator()
	case SyntaxKindGraphShortestPathsKeyword:
		return p.parseGraphShortestPathsOperator()
	case SyntaxKindGraphWhereNodesKeyword:
		return p.parseGraphWhereNodesOperator()
	case SyntaxKindGraphWhereEdgesKeyword:
		return p.parseGraphWhereEdgesOperator()
	case SyntaxKindGraphToTableKeyword:
		return p.parseGraphToTableOperator()
	case SyntaxKindMvDashApplyKeyword, SyntaxKindMvApplyKeyword:
		return p.parseMvApplyOperator()
	case SyntaxKindMvDashExpandKeyword, SyntaxKindMvExpandKeyword:
		return p.parseMvExpandOperator()
	case SyntaxKindParseKeyword:
		return p.parseParseOperator()
	case SyntaxKindParseKvKeyword:
		return p.parseParseKvOperator()
	case SyntaxKindParseWhereKeyword:
		return p.parseParseWhereOperator()
	case SyntaxKindPartitionKeyword:
		return p.parsePartitionOperator()
	case SyntaxKindPartitionByKeyword:
		return p.parsePartitionByOperator()
	case SyntaxKindProjectKeyword:
		return p.parseProjectOperator()
	case SyntaxKindProjectAwayKeyword:
		return p.parseProjectAwayOperator()
	case SyntaxKindProjectByNamesKeyword:
		return p.parseProjectByNamesOperator()
	case SyntaxKindProjectKeepKeyword:
		return p.parseProjectKeepOperator()
	case SyntaxKindProjectRenameKeyword:
		return p.parseProjectRenameOperator()
	case SyntaxKindProjectReorderKeyword:
		return p.parseProjectReorderOperator()
	case SyntaxKindRangeKeyword:
		if operator := p.parseRangeOperator(); operator != nil {
			return operator
		}
		return nil
	case SyntaxKindReduceKeyword:
		return p.parseReduceByOperator()
	case SyntaxKindSampleKeyword:
		return p.parseSampleOperator()
	case SyntaxKindSampleDistinctKeyword:
		return p.parseSampleDistinctOperator()
	case SyntaxKindRenderKeyword:
		return p.parseRenderOperator()
	case SyntaxKindScanKeyword:
		return p.parseScanOperator()
	case SyntaxKindSearchKeyword:
		return p.parseSearchOperator()
	case SyntaxKindSerializeKeyword:
		return p.parseSerializeOperator()
	case SyntaxKindSortKeyword, SyntaxKindOrderKeyword:
		return p.parseSortOperator()
	case SyntaxKindSummarizeKeyword:
		return p.parseSummarizeOperator()
	case SyntaxKindTakeKeyword, SyntaxKindLimitKeyword:
		return p.parseTakeOperator()
	case SyntaxKindTopHittersKeyword:
		return p.parseTopHittersOperator()
	case SyntaxKindTopNestedKeyword:
		return p.parseTopNestedOperator()
	case SyntaxKindTopKeyword:
		return p.parseTopOperator()
	default:
		return nil
	}
}

func (p *queryParserState) parseCountAsIdentifierClause() *CountAsIdentifierClause {
	asKeyword := p.parseToken(SyntaxKindAsKeyword)
	if asKeyword == nil {
		return nil
	}
	identifier := p.parseRequiredToken(SyntaxKindIdentifierToken)
	return MustNewCountAsIdentifierClauseTyped(asKeyword, identifier)
}

func (p *queryParserState) parseCountOperator() *CountOperator {
	if p.peekToken().Kind != SyntaxKindCountKeyword || p.peekToken(1).Kind == SyntaxKindOpenParenToken {
		return nil
	}
	keyword := p.parseToken()
	asIdentifier := p.parseCountAsIdentifierClause()
	return MustNewCountOperatorTyped(keyword, asIdentifier)
}

func (p *queryParserState) parseAssertSchemaOperator() *AssertSchemaOperator {
	keyword := p.parseToken(SyntaxKindAssertSchemaKeyword)
	if keyword == nil {
		return nil
	}
	schema := p.parseRowSchema(false)
	if schema == nil {
		schema = createMissingRowSchema()
	}
	return MustNewAssertSchemaOperatorTyped(keyword, schema)
}

func (p *queryParserState) parseConsumeOperator() *ConsumeOperator {
	keyword := p.parseToken(SyntaxKindConsumeKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.ConsumeParameters, false)
	return MustNewConsumeOperatorTyped(keyword, parameters)
}

func (p *queryParserState) parseExecuteAndCacheOperator() *ExecuteAndCacheOperator {
	keyword := p.parseToken(SyntaxKindExecuteAndCacheKeyword)
	if keyword == nil {
		return nil
	}
	return MustNewExecuteAndCacheOperatorTyped(keyword)
}

func (p *queryParserState) parseExtendOperator() *ExtendOperator {
	keyword := p.parseToken(SyntaxKindExtendKeyword)
	if keyword == nil {
		return nil
	}
	expressions := p.parseNamedExpressionCommaList(true)
	return MustNewExtendOperatorTyped(keyword, expressions)
}

func (p *queryParserState) parseFilterOperator() *FilterOperator {
	keyword := p.parseToken(SyntaxKindWhereKeyword, SyntaxKindFilterKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.FilterParameters, true)
	condition := p.parseNamedExpression()
	if condition == nil {
		condition = p.createMissingExpression()
	}
	return MustNewFilterOperatorTyped(keyword, parameters, condition)
}

func (p *queryParserState) parseGetSchemaOperator() *GetSchemaOperator {
	keyword := p.parseToken(SyntaxKindGetSchemaKeyword)
	if keyword == nil {
		return nil
	}
	kindParameter := p.parseSpecificQueryOperatorParameter(QueryOperatorParameters.GetSchemaKind, true)
	return MustNewGetSchemaOperatorTyped(keyword, kindParameter)
}

func (p *queryParserState) parseProjectOperator() *ProjectOperator {
	keyword := p.parseToken(SyntaxKindProjectKeyword)
	if keyword == nil {
		return nil
	}
	expressions := p.parseNamedExpressionCommaList(false)
	return MustNewProjectOperatorTyped(keyword, expressions)
}

func (p *queryParserState) parseProjectAwayOperator() *ProjectAwayOperator {
	keyword := p.parseToken(SyntaxKindProjectAwayKeyword)
	if keyword == nil {
		return nil
	}
	expressions := p.parseSimpleOrWildcardedNameReferenceCommaList(false)
	return MustNewProjectAwayOperatorTyped(keyword, expressions)
}

func (p *queryParserState) parseProjectKeepOperator() *ProjectKeepOperator {
	keyword := p.parseToken(SyntaxKindProjectKeepKeyword)
	if keyword == nil {
		return nil
	}
	expressions := p.parseSimpleOrWildcardedNameReferenceCommaList(false)
	return MustNewProjectKeepOperatorTyped(keyword, expressions)
}

func (p *queryParserState) parseProjectRenameOperator() *ProjectRenameOperator {
	keyword := p.parseToken(SyntaxKindProjectRenameKeyword)
	if keyword == nil {
		return nil
	}
	expressions := p.parseNamedExpressionCommaList(false)
	return MustNewProjectRenameOperatorTyped(keyword, expressions)
}

func (p *queryParserState) parseProjectReorderOperator() *ProjectReorderOperator {
	keyword := p.parseToken(SyntaxKindProjectReorderKeyword)
	if keyword == nil {
		return nil
	}
	expressions := p.parseProjectReorderExpressionCommaList(false)
	return MustNewProjectReorderOperatorTyped(keyword, expressions)
}

func (p *queryParserState) parseProjectByNamesOperator() *ProjectByNamesOperator {
	keyword := p.parseToken(SyntaxKindProjectByNamesKeyword)
	if keyword == nil {
		return nil
	}
	expressions := p.parseUnnamedExpressionCommaList(false)
	return MustNewProjectByNamesOperatorTyped(keyword, expressions)
}

func (p *queryParserState) parseAsOperator() *AsOperator {
	keyword := p.parseToken(SyntaxKindAsKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.AsParameters, true)
	name := p.parseNameDeclaration()
	if name == nil {
		name = p.createMissingNameDeclaration()
	}
	return MustNewAsOperatorTyped(keyword, parameters, name)
}

func (p *queryParserState) parseSerializeOperator() *SerializeOperator {
	keyword := p.parseToken(SyntaxKindSerializeKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.SerializedParameters, true)
	expressions := p.parseNamedExpressionCommaList(false)
	return MustNewSerializeOperatorTyped(keyword, parameters, expressions)
}

func (p *queryParserState) parseRangeOperator() *RangeOperator {
	if p.peekToken().Kind != SyntaxKindRangeKeyword || p.peekToken(1).Kind == SyntaxKindOpenParenToken {
		return nil
	}
	keyword := p.parseToken()
	var name *NameDeclaration
	if p.peekToken().Kind == SyntaxKindFromKeyword {
		name = createMissingNameDeclaration()
	} else {
		name = p.parseNameDeclaration()
		if name == nil {
			name = p.createMissingNameDeclaration()
		}
	}
	fromKeyword := p.parseRequiredToken(SyntaxKindFromKeyword)
	fromExpr := p.parseUnnamedExpression()
	if fromExpr == nil {
		fromExpr = p.createMissingExpression()
	}
	toKeyword := p.parseRequiredToken(SyntaxKindToKeyword)
	toExpr := p.parseUnnamedExpression()
	if toExpr == nil {
		toExpr = p.createMissingExpression()
	}
	stepKeyword := p.parseRequiredToken(SyntaxKindStepKeyword)
	stepExpr := p.parseUnnamedExpression()
	if stepExpr == nil {
		stepExpr = p.createMissingExpression()
	}
	return MustNewRangeOperatorTyped(keyword, name, fromKeyword, fromExpr, toKeyword, toExpr, stepKeyword, stepExpr)
}

func (p *queryParserState) parseInvalidKeywordAsNameDeclaration() *NameDeclaration {
	if !IsKeywordKind(p.peekToken().Kind) {
		return nil
	}
	token := p.parseToken()
	if token == nil {
		return nil
	}
	return MustNewNameDeclarationTyped(MustNewTokenNameTyped(token), DiagnosticMissingNameWithKeyword(token.RawText()))
}

func (p *queryParserState) parseInvokeOperator() *InvokeOperator {
	keyword := p.parseToken(SyntaxKindInvokeKeyword)
	if keyword == nil {
		return nil
	}
	function := p.parseDotCompositeFunctionCall()
	if isNilSyntaxElement(function) {
		function = p.createMissingExpression()
	}
	return MustNewInvokeOperatorTyped(keyword, function)
}

func (p *queryParserState) parseSampleOperator() *SampleOperator {
	keyword := p.parseToken(SyntaxKindSampleKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.SampleParameters, true)
	expr := p.parseNamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewSampleOperatorTyped(keyword, parameters, expr)
}

func (p *queryParserState) parseSampleDistinctOperator() *SampleDistinctOperator {
	keyword := p.parseToken(SyntaxKindSampleDistinctKeyword)
	if keyword == nil {
		return nil
	}
	parameters := AsSyntaxListOf[*NamedParameter](NewTypedSyntaxList("NamedParameter"))
	if len(QueryOperatorParameters.SampleDistinctParameters) > 0 {
		parameters = p.parseQueryOperatorParameterList(QueryOperatorParameters.SampleDistinctParameters, true)
	}
	expr := p.parseNamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	ofKeyword := p.parseRequiredToken(SyntaxKindOfKeyword)
	ofExpr := p.parseNamedExpression()
	if ofExpr == nil {
		ofExpr = p.createMissingExpression()
	}
	return MustNewSampleDistinctOperatorTyped(keyword, parameters, expr, ofKeyword, ofExpr)
}

func (p *queryParserState) parseEvaluateSchemaClause() *EvaluateSchemaClause {
	colon := p.parseToken(SyntaxKindColonToken)
	if colon == nil {
		return nil
	}
	schema := p.parseEvaluateRowSchema()
	if schema == nil {
		schema = createMissingEvaluateRowSchema()
	}
	return MustNewEvaluateSchemaClauseTyped(colon, schema)
}

func createMissingEvaluateRowSchema() *EvaluateRowSchema {
	return createMissingEvaluateRowSchemaWithDiagnostics()
}

func createMissingEvaluateRowSchemaWithDiagnostics(diagnostics ...Diagnostic) *EvaluateRowSchema {
	diagnostics = append([]Diagnostic{DiagnosticMissingSchemaDeclaration()}, diagnostics...)
	return MustNewEvaluateRowSchemaTyped(
		createOmittedToken(SyntaxKindOpenParenToken),
		nil,
		nil,
		nil,
		AsSeparatedSyntaxListOf[*NameAndTypeDeclaration](NewTypedSyntaxList("SeparatedElement<NameAndTypeDeclaration>")),
		createOmittedToken(SyntaxKindCloseParenToken),
		diagnostics...,
	)
}

func (p *queryParserState) parseEvaluateOperator() *EvaluateOperator {
	keyword := p.parseToken(SyntaxKindEvaluateKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.EvaluateParameters, false)
	functionCall := p.parseRequiredFunctionCallExpression()
	schema := p.parseEvaluateSchemaClause()
	return MustNewEvaluateOperatorTyped(keyword, parameters, functionCall, schema)
}

func (p *queryParserState) parseJoinOnClause() JoinConditionClause {
	keyword := p.parseToken(SyntaxKindOnKeyword)
	if keyword == nil {
		return nil
	}
	expressions := parseCommaList(p, "Expression", p.parseUnnamedExpression, p.createMissingExpression, nil, true, false)
	return MustNewJoinOnClauseTyped(keyword, expressions)
}

func (p *queryParserState) parseJoinWhereClause() JoinConditionClause {
	keyword := p.parseToken(SyntaxKindWhereKeyword)
	if keyword == nil {
		return nil
	}
	expr := p.parseUnnamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewJoinWhereClauseTyped(keyword, expr)
}

func createMissingJoinOnClause() *JoinOnClause {
	return MustNewJoinOnClauseTyped(
		createOmittedToken(SyntaxKindJoinOnClause),
		AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>")),
		DiagnosticMissingJoinOnClause(),
	)
}

func createOmittedJoinOnClause() *JoinOnClause {
	return MustNewJoinOnClauseTyped(
		createOmittedToken(SyntaxKindJoinOnClause),
		AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>")),
	)
}

func (p *queryParserState) parseJoinOperator() *JoinOperator {
	keyword := p.parseToken(SyntaxKindJoinKeyword)
	if keyword == nil {
		return nil
	}
	rejectedParameter := p.atQueryOperatorParameterMissingEquals(QueryOperatorParameters.JoinParameters)
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.JoinParameters, true)
	missingRightAtEndOfQuery := p.atEndOfQuery()
	expr := p.parseUnnamedExpression()
	missingExpression := expr == nil
	if expr == nil {
		expr = p.createMissingExpression()
	}
	var diagnostics []Diagnostic
	condition := p.parseJoinOnClause()
	if condition == nil {
		condition = p.parseJoinWhereClause()
	}
	missingConditionDiagnosticAdded := false
	if condition == nil && parameters.Count() == 0 && missingExpression && missingRightAtEndOfQuery {
		diagnostics = append(diagnostics, DiagnosticMissingJoinOnClause().WithLocationText(keyword.RawText()))
		missingConditionDiagnosticAdded = true
	}
	if condition == nil && rejectedParameter && syntaxElementIsMissing(expr) {
		condition = createOmittedJoinOnClause()
		if !missingConditionDiagnosticAdded {
			diagnostics = append(diagnostics, DiagnosticMissingJoinOnClause().WithLocationText(keyword.RawText()))
		}
	}
	return MustNewJoinOperatorTyped(keyword, parameters, expr, condition, diagnostics...)
}

func (p *queryParserState) parseLookupOperator() *LookupOperator {
	keyword := p.parseToken(SyntaxKindLookupKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.LookupParameters, true)
	expr := p.parseUnnamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	condition := p.parseJoinOnClause()
	if condition == nil {
		condition = createMissingJoinOnClause()
	}
	return MustNewLookupOperatorTyped(keyword, parameters, expr, condition)
}

func (p *queryParserState) parseDefaultExpressionClause() *DefaultExpressionClause {
	keyword := p.parseToken(SyntaxKindDefaultKeyword)
	if keyword == nil {
		return nil
	}
	equal := p.parseRequiredToken(SyntaxKindEqualToken)
	expr := p.parseNamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewDefaultExpressionClauseTyped(keyword, equal, expr)
}

func (p *queryParserState) parseMakeSeriesByClause() *MakeSeriesByClause {
	keyword := p.parseToken(SyntaxKindByKeyword)
	if keyword == nil {
		return nil
	}
	expressions := p.parseNamedExpressionCommaListWithEnd(true, nil)
	return MustNewMakeSeriesByClauseTyped(keyword, expressions)
}

func (p *queryParserState) parseMakeSeriesExpression() *MakeSeriesExpression {
	expr := p.parseNamedExpression()
	if expr == nil {
		return nil
	}
	return MustNewMakeSeriesExpressionTyped(expr, p.parseDefaultExpressionClause())
}

func (p *queryParserState) createMissingMakeSeriesExpression() *MakeSeriesExpression {
	return MustNewMakeSeriesExpressionTyped(p.createMissingExpression(), nil)
}

func (p *queryParserState) scanMakeSeriesExpressionListEnd() bool {
	switch p.peekToken().Kind {
	case SyntaxKindOnKeyword, SyntaxKindFromKeyword, SyntaxKindInKeyword:
		return true
	default:
		return p.scanCommonListEnd()
	}
}

func (p *queryParserState) parseMakeSeriesExpressionList() *SeparatedSyntaxListOf[*MakeSeriesExpression] {
	return parseCommaList(
		p,
		"MakeSeriesExpression",
		p.parseMakeSeriesExpression,
		p.createMissingMakeSeriesExpression,
		p.scanMakeSeriesExpressionListEnd,
		true,
		false,
	)
}

func (p *queryParserState) parseRequiredMakeSeriesInRangeClause() *MakeSeriesInRangeClause {
	inKeyword := p.parseRequiredToken(SyntaxKindInKeyword)
	rangeKeyword := p.parseRequiredToken(SyntaxKindRangeKeyword)
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	expressions := p.parseNamedExpressionCommaList(false)
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewMakeSeriesInRangeClauseTyped(inKeyword, rangeKeyword, MustNewExpressionListTyped(open, expressions, close))
}

func (p *queryParserState) parseMakeSeriesFromClause() *MakeSeriesFromClause {
	keyword := p.parseToken(SyntaxKindFromKeyword)
	if keyword == nil {
		return nil
	}
	expr := p.parseUnnamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewMakeSeriesFromClauseTyped(keyword, expr)
}

func (p *queryParserState) parseMakeSeriesToClause() *MakeSeriesToClause {
	keyword := p.parseToken(SyntaxKindToKeyword)
	if keyword == nil {
		return nil
	}
	expr := p.parseUnnamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewMakeSeriesToClauseTyped(keyword, expr)
}

func (p *queryParserState) parseRequiredMakeSeriesStepClause() *MakeSeriesStepClause {
	keyword := p.parseRequiredToken(SyntaxKindStepKeyword)
	expr := p.parseUnnamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewMakeSeriesStepClauseTyped(keyword, expr)
}

func (p *queryParserState) parseMakeSeriesFromToStepClause() *MakeSeriesFromToStepClause {
	switch p.peekToken().Kind {
	case SyntaxKindFromKeyword, SyntaxKindToKeyword, SyntaxKindStepKeyword:
	default:
		return nil
	}
	fromClause := p.parseMakeSeriesFromClause()
	toClause := p.parseMakeSeriesToClause()
	stepClause := p.parseRequiredMakeSeriesStepClause()
	return MustNewMakeSeriesFromToStepClauseTyped(fromClause, toClause, stepClause)
}

func (p *queryParserState) parseMakeSeriesOperator() *MakeSeriesOperator {
	keyword := p.parseToken(SyntaxKindMakeSeriesKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.MakeSeriesParameters, true)
	expressions := p.parseMakeSeriesExpressionList()
	onKeyword := p.parseRequiredToken(SyntaxKindOnKeyword)
	onExpression := p.parseNamedExpression()
	if onExpression == nil {
		onExpression = p.createMissingExpression()
	}
	onClause := MustNewMakeSeriesOnClauseTyped(onKeyword, onExpression)
	var rangeClause MakeSeriesRangeClause
	if fromToStep := p.parseMakeSeriesFromToStepClause(); fromToStep != nil {
		rangeClause = fromToStep
	} else {
		rangeClause = p.parseRequiredMakeSeriesInRangeClause()
	}
	byClause := p.parseMakeSeriesByClause()
	return MustNewMakeSeriesOperatorTyped(keyword, parameters, expressions, onClause, rangeClause, byClause)
}

func (p *queryParserState) parseToTypeOfClause() *ToTypeOfClause {
	keyword := p.parseToken(SyntaxKindToKeyword)
	if keyword == nil {
		return nil
	}
	return MustNewToTypeOfClauseTyped(keyword, p.parseTypeOfLiteralOrMissing())
}

func (p *queryParserState) parseMvExpandExpression() *MvExpandExpression {
	if p.peekToken().Kind == SyntaxKindToKeyword {
		return MustNewMvExpandExpressionTyped(p.createMissingExpression(), p.parseToTypeOfClause())
	}
	expr := p.parseNamedExpression()
	if expr == nil {
		return nil
	}
	return MustNewMvExpandExpressionTyped(expr, p.parseToTypeOfClause())
}

func (p *queryParserState) createMissingMvExpandExpression() *MvExpandExpression {
	return MustNewMvExpandExpressionTyped(p.createMissingExpression(), nil)
}

func createMissingExpressionNode() Expression {
	return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone, DiagnosticMissingExpression())
}

func (p *queryParserState) scanMvExpandExpressionListEnd() bool {
	return p.peekToken().Kind == SyntaxKindLimitKeyword || p.scanCommonListEnd()
}

func (p *queryParserState) parseMvExpandExpressionList() *SeparatedSyntaxListOf[*MvExpandExpression] {
	return parseCommaList(
		p,
		"MvExpandExpression",
		p.parseMvExpandExpression,
		p.createMissingMvExpandExpression,
		p.scanMvExpandExpressionListEnd,
		true,
		false,
	)
}

func (p *queryParserState) parseMvExpandRowLimitClause() *MvExpandRowLimitClause {
	keyword := p.parseToken(SyntaxKindLimitKeyword)
	if keyword == nil {
		return nil
	}
	expr := p.parseUnnamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewMvExpandRowLimitClauseTyped(keyword, expr)
}

func (p *queryParserState) parseMvExpandOperator() *MvExpandOperator {
	keyword := p.parseToken(SyntaxKindMvExpandKeyword, SyntaxKindMvDashExpandKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.MvExpandParameters, true)
	expressions := p.parseMvExpandExpressionList()
	rowLimit := p.parseMvExpandRowLimitClause()
	return MustNewMvExpandOperatorTyped(keyword, parameters, expressions, rowLimit)
}

func (p *queryParserState) parseMvApplyExpression() *MvApplyExpression {
	if p.peekToken().Kind == SyntaxKindToKeyword {
		return MustNewMvApplyExpressionTyped(p.createMissingExpression(), p.parseToTypeOfClause())
	}
	expr := p.parseNamedExpression()
	if expr == nil {
		return nil
	}
	return MustNewMvApplyExpressionTyped(expr, p.parseToTypeOfClause())
}

func (p *queryParserState) createMissingMvApplyExpression() *MvApplyExpression {
	return MustNewMvApplyExpressionTyped(p.createMissingExpression(), nil)
}

func (p *queryParserState) scanMvApplyExpressionListEnd() bool {
	switch p.peekToken().Kind {
	case SyntaxKindLimitKeyword, SyntaxKindIdKeyword, SyntaxKindOnKeyword:
		next := p.peekToken(1).Kind
		return next != SyntaxKindCommaToken && next != SyntaxKindToKeyword && next != SyntaxKindEqualToken
	default:
		return p.scanCommonListEnd()
	}
}

func (p *queryParserState) parseMvApplyExpressionList() *SeparatedSyntaxListOf[*MvApplyExpression] {
	if p.peekToken().Kind == SyntaxKindToKeyword {
		save := p.pos
		clause := p.parseToTypeOfClause()
		if p.peekToken().Kind != SyntaxKindCommaToken {
			return AsSeparatedSyntaxListOf[*MvApplyExpression](NewTypedSyntaxList(
				"SeparatedElement<MvApplyExpression>",
				NewSeparatedElement(MustNewMvApplyExpressionTyped(nil, clause), nil),
			))
		}
		p.pos = save
	}

	return parseCommaList(
		p,
		"MvApplyExpression",
		p.parseMvApplyExpression,
		p.createMissingMvApplyExpression,
		p.scanMvApplyExpressionListEnd,
		true,
		false,
	)
}

func diagnosticsHaveCodeMessage(diagnostics []Diagnostic, code string, message string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code && diagnostic.Message == message {
			return true
		}
	}
	return false
}

func (p *queryParserState) parseMvApplyRowLimitClause() *MvApplyRowLimitClause {
	keyword := p.parseToken(SyntaxKindLimitKeyword)
	if keyword == nil {
		return nil
	}
	expr := p.parseUnnamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewMvApplyRowLimitClauseTyped(keyword, expr)
}

func (p *queryParserState) parseMvApplyContextIdClause() *MvApplyContextIdClause {
	keyword := p.parseToken(SyntaxKindIdKeyword)
	if keyword == nil {
		return nil
	}
	expr := p.parseUnnamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewMvApplyContextIdClauseTyped(keyword, expr)
}

func (p *queryParserState) parseMvApplySubqueryExpression() *MvApplySubqueryExpression {
	open := p.parseToken(SyntaxKindOpenParenToken)
	if open == nil {
		return nil
	}
	expr := p.parseContextualSubExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewMvApplySubqueryExpressionTyped(open, expr, close)
}

func (p *queryParserState) createMissingMvApplySubqueryExpression(diagnostics ...Diagnostic) *MvApplySubqueryExpression {
	return MustNewMvApplySubqueryExpressionTyped(
		createMissingToken(SyntaxKindOpenParenToken),
		p.createMissingExpression(),
		createMissingToken(SyntaxKindCloseParenToken),
		diagnostics...,
	)
}

func (p *queryParserState) parseMvApplyOperator() *MvApplyOperator {
	keyword := p.parseToken(SyntaxKindMvApplyKeyword, SyntaxKindMvDashApplyKeyword)
	if keyword == nil {
		return nil
	}
	rejectedParameter := p.atQueryOperatorParameterMissingEquals(QueryOperatorParameters.MvApplyParameters)
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.MvApplyParameters, true)
	expressions := p.parseMvApplyExpressionList()
	rowLimit := p.parseMvApplyRowLimitClause()
	idClause := p.parseMvApplyContextIdClause()
	onKeyword := p.parseRequiredToken(SyntaxKindOnKeyword)
	subquery := p.parseMvApplySubqueryExpression()
	if subquery == nil {
		if rejectedParameter {
			subquery = p.createMissingMvApplySubqueryExpression(DiagnosticQueryOperatorExpected())
		} else {
			subquery = p.createMissingMvApplySubqueryExpression()
		}
	}
	return MustNewMvApplyOperatorTyped(keyword, parameters, expressions, rowLimit, idClause, onKeyword, subquery)
}

func (p *queryParserState) parseParseKvWithClause() *ParseKvWithClause {
	keyword := p.parseToken(SyntaxKindWithKeyword)
	if keyword == nil {
		return nil
	}
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	properties := p.parseQueryOperatorParameterCommaList(QueryOperatorParameters.ParseKvWithProperties, queryOperatorParameterNameDeclaredOnly)
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewParseKvWithClauseTyped(keyword, open, properties, close)
}

func (p *queryParserState) parseParseKvOperator() *ParseKvOperator {
	keyword := p.parseToken(SyntaxKindParseKvKeyword)
	if keyword == nil {
		return nil
	}
	expr := p.parseUnnamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	asKeyword := p.parseRequiredToken(SyntaxKindAsKeyword)
	keys := p.parseRowSchema(false)
	if keys == nil {
		keys = createMissingRowSchema()
	}
	withClause := p.parseParseKvWithClause()
	return MustNewParseKvOperatorTyped(keyword, expr, asKeyword, keys, withClause)
}

func (p *queryParserState) parsePartitionQuery() *PartitionQuery {
	open := p.parseToken(SyntaxKindOpenBraceToken)
	if open == nil {
		return nil
	}
	query := p.parseExpression()
	if query == nil {
		query = p.createMissingExpression()
	}
	close := p.parseRequiredToken(SyntaxKindCloseBraceToken)
	return MustNewPartitionQueryTyped(open, query, close)
}

func (p *queryParserState) parsePartitionSubquery() *PartitionSubquery {
	open := p.parseToken(SyntaxKindOpenParenToken)
	if open == nil {
		return nil
	}
	subquery := p.parsePartitionSubExpression()
	if subquery == nil {
		subquery = p.createMissingExpression()
	}
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewPartitionSubqueryTyped(open, subquery, close)
}

func (p *queryParserState) parsePartitionSubExpression() Expression {
	expr := queryOperatorExpression(p.parseRequiredQueryOperator())
	for expr != nil && p.peekToken().Kind == SyntaxKindBarToken {
		bar := p.parseToken()
		operator := p.parseRequiredQueryOperator()
		expr = MustNewPipeExpressionTyped(expr, bar, operator)
	}
	return expr
}

func (p *queryParserState) parseContextualSubExpression() Expression {
	var expr Expression
	if contextual := p.parseContextualDataTableExpression(); !isNilSyntaxElement(contextual) {
		expr = contextual
	} else {
		expr = queryOperatorExpression(p.parseRequiredQueryOperator())
	}
	for expr != nil && p.peekToken().Kind == SyntaxKindBarToken {
		bar := p.parseToken()
		operator := p.parseRequiredQueryOperator()
		expr = MustNewPipeExpressionTyped(expr, bar, operator)
	}
	return expr
}

func (p *queryParserState) parsePartitionScope() *PartitionScope {
	keyword := p.parseToken(SyntaxKindInKeyword)
	if keyword == nil {
		return nil
	}
	var expr Expression
	if function := p.parseFunctionCallExpression(); !isNilSyntaxElement(function) {
		expr = function
	} else if dynamic := p.parseDynamicLiteral(); !isNilSyntaxElement(dynamic) {
		expr = dynamic
	}
	if isNilSyntaxElement(expr) {
		expr = p.createMissingExpression()
	}
	return MustNewPartitionScopeTyped(keyword, expr)
}

func (p *queryParserState) parsePartitionOperand() PartitionOperand {
	switch p.peekToken().Kind {
	case SyntaxKindOpenBraceToken:
		return p.parsePartitionQuery()
	case SyntaxKindOpenParenToken:
		return p.parsePartitionSubquery()
	default:
		return nil
	}
}

func (p *queryParserState) createMissingPartitionOperand() PartitionOperand {
	return MustNewPartitionSubqueryTyped(
		createMissingToken(SyntaxKindOpenParenToken),
		p.createMissingExpression(),
		createMissingToken(SyntaxKindCloseParenToken),
	)
}

func (p *queryParserState) parsePartitionOperator() *PartitionOperator {
	keyword := p.parseToken(SyntaxKindPartitionKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.PartitionParameters, false)
	byKeyword := p.parseRequiredToken(SyntaxKindByKeyword)
	byExpr := p.parseSimplePathExpression()
	if byExpr == nil {
		byExpr = p.createMissingNameReference()
	}
	scope := p.parsePartitionScope()
	operand := p.parsePartitionOperand()
	if operand == nil {
		operand = p.createMissingPartitionOperand()
	}
	return MustNewPartitionOperatorTyped(keyword, parameters, byKeyword, byExpr, scope, operand)
}

func (p *queryParserState) consumeUnexpectedPartitionExpressionTail() bool {
	consumed := false
	for startsLikelyExtraPartitionExpressionTail(p.peekToken().Kind) {
		if !p.consumeUnexpectedPartitionTailExpression() {
			break
		}
		consumed = true
	}
	if p.consumeUnexpectedPartitionCommaExpressionTail() {
		consumed = true
	}
	return consumed
}

func (p *queryParserState) consumeUnexpectedPartitionCommaExpressionTail() bool {
	consumed := false
	for p.peekToken().Kind == SyntaxKindCommaToken {
		p.parseToken(SyntaxKindCommaToken)
		consumed = true
		for {
			switch p.peekToken().Kind {
			case SyntaxKindInKeyword,
				SyntaxKindIdKeyword,
				SyntaxKindOpenParenToken,
				SyntaxKindOpenBraceToken,
				SyntaxKindBarToken,
				SyntaxKindSemicolonToken,
				SyntaxKindEndOfTextToken,
				SyntaxKindNone:
				return consumed
			case SyntaxKindCommaToken:
				break
			default:
				if !p.consumeUnexpectedPartitionTailExpression() {
					p.parseToken()
				}
				continue
			}
			break
		}
	}
	return consumed
}

func (p *queryParserState) consumeUnexpectedPartitionTailExpression() bool {
	switch p.peekToken().Kind {
	case SyntaxKindIdentifierToken, SyntaxKindOpenBracketToken:
		return p.parseSimplePathExpression() != nil
	case SyntaxKindDynamicKeyword:
		return p.parseDynamicLiteral() != nil
	default:
		return p.parseLiteral() != nil
	}
}

func startsLikelyExtraPartitionExpressionTail(kind SyntaxKind) bool {
	switch kind {
	case SyntaxKindIdentifierToken,
		SyntaxKindOpenBracketToken,
		SyntaxKindStringLiteralToken,
		SyntaxKindBooleanLiteralToken,
		SyntaxKindLongLiteralToken,
		SyntaxKindRealLiteralToken,
		SyntaxKindDecimalLiteralToken,
		SyntaxKindIntLiteralToken,
		SyntaxKindGuidLiteralToken,
		SyntaxKindRawGuidLiteralToken,
		SyntaxKindDateTimeLiteralToken,
		SyntaxKindTimespanLiteralToken,
		SyntaxKindNullKeyword,
		SyntaxKindDynamicKeyword:
		return true
	default:
		return false
	}
}

func (p *queryParserState) parsePartitionByIdClause() *PartitionByIdClause {
	keyword := p.parseToken(SyntaxKindIdKeyword)
	if keyword == nil {
		return nil
	}
	value := p.parseLiteral()
	if value == nil {
		value = p.createMissingValue()
	}
	return MustNewPartitionByIdClauseTyped(keyword, value)
}

func (p *queryParserState) parsePartitionByOperator() *PartitionByOperator {
	keyword := p.parseToken(SyntaxKindPartitionByKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.PartitionByParameters, false)
	entity := p.parseSimplePathExpression()
	if entity == nil {
		entity = p.createMissingNameReference()
	}
	idClause := p.parsePartitionByIdClause()
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	subquery := p.parseContextualSubExpression()
	if subquery == nil {
		subquery = p.createMissingExpression()
	}
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewPartitionByOperatorTyped(keyword, parameters, entity, idClause, open, subquery, close)
}

func (p *queryParserState) parseSimplePathExpression() Expression {
	expr := p.parsePathElementSelector()
	for expr != nil {
		switch p.peekToken().Kind {
		case SyntaxKindDotToken:
			dot := p.parseToken()
			selector := p.parsePathElementSelector()
			if selector == nil {
				selector = p.createMissingNameReference()
			}
			expr = MustNewPathExpressionTyped(expr, dot, selector)
		case SyntaxKindOpenBracketToken:
			selector := p.parsePathElementSelector()
			if selector == nil {
				selector = p.createMissingExpression()
			}
			expr = MustNewElementExpressionTyped(expr, selector)
		default:
			return expr
		}
	}
	return expr
}

func (p *queryParserState) parseParseWithExpression() SyntaxNodeElement {
	if p.peekToken().Kind == SyntaxKindAsteriskToken {
		return MustNewStarExpressionTyped(p.parseToken())
	}
	if expr := p.parseStringOrCompoundStringLiteral(); expr != nil {
		return expr
	}
	if expr := p.parseNameAndOptionalTypeDeclaration(); expr != nil {
		if element, ok := expr.(SyntaxNodeElement); ok {
			return element
		}
	}
	return nil
}

func (p *queryParserState) parseParseWithExpressionList() *SyntaxListOf[SyntaxNodeElement] {
	var elements []SyntaxElement
	for {
		if p.peekToken().Kind == SyntaxKindEndOfTextToken || p.peekToken().Kind == SyntaxKindNone || p.scanCommonListEnd() {
			break
		}
		expr := p.parseParseWithExpression()
		if expr == nil {
			break
		}
		elements = append(elements, expr)
	}
	return AsSyntaxListOf[SyntaxNodeElement](NewTypedSyntaxList("SyntaxNode", elements...))
}

func (p *queryParserState) parseParseOperator() *ParseOperator {
	keyword := p.parseToken(SyntaxKindParseKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.ParseParameters, true)
	expr := p.parseUnnamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	withKeyword := p.parseRequiredToken(SyntaxKindWithKeyword)
	patterns := p.parseParseWithExpressionList()
	return MustNewParseOperatorTyped(keyword, parameters, expr, withKeyword, patterns)
}

func (p *queryParserState) parseParseWhereOperator() *ParseWhereOperator {
	keyword := p.parseToken(SyntaxKindParseWhereKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.ParseParameters, true)
	expr := p.parseUnnamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	withKeyword := p.parseRequiredToken(SyntaxKindWithKeyword)
	patterns := p.parseParseWithExpressionList()
	return MustNewParseWhereOperatorTyped(keyword, parameters, expr, withKeyword, patterns)
}

func (p *queryParserState) parseRenderWithClause() *RenderWithClause {
	keyword := p.parseToken(SyntaxKindWithKeyword)
	if keyword == nil {
		return nil
	}
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	leadingComma := p.parseToken(SyntaxKindCommaToken)
	properties := p.parseQueryOperatorParameterCommaList(QueryOperatorParameters.RenderWithProperties, queryOperatorParameterNameDeclaredOnly)
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewRenderWithClauseTyped(keyword, open, leadingComma, properties, close)
}

func (p *queryParserState) parseDeprecatedRenderProperty() *NamedParameter {
	if parameter := p.scanSpecificQueryOperatorParameter([]*QueryOperatorParameter{
		QueryOperatorParameters.RenderKind,
		QueryOperatorParameters.RenderTitle,
		QueryOperatorParameters.RenderAccumulate,
	}, true, 0); parameter != nil {
		return p.parseQueryOperatorParameter(parameter)
	}
	if p.peekToken().Kind == SyntaxKindWithKeyword && p.peekToken(1).Kind != SyntaxKindOpenParenToken {
		return p.parseQueryOperatorParameter(QueryOperatorParameters.RenderWithDeprecated)
	}
	if p.scanSpecificQueryOperatorParameter([]*QueryOperatorParameter{QueryOperatorParameters.RenderByDeprecated}, false, 0) != nil {
		return p.parseQueryOperatorParameterWithEnd(QueryOperatorParameters.RenderByDeprecated, p.scanDeprecatedRenderParameterNameListEnd)
	}
	return nil
}

func (p *queryParserState) scanDeprecatedRenderParameterNameListEnd() bool {
	names := []*QueryOperatorParameter{
		QueryOperatorParameters.RenderKind,
		QueryOperatorParameters.RenderTitle,
		QueryOperatorParameters.RenderAccumulate,
		QueryOperatorParameters.RenderWithDeprecated,
		QueryOperatorParameters.RenderByDeprecated,
	}
	return p.scanQueryOperatorParameterNameListEnd(names)
}

func (p *queryParserState) parseDeprecatedRenderPropertyList() *SyntaxListOf[*NamedParameter] {
	var elements []SyntaxElement
	for {
		property := p.parseDeprecatedRenderProperty()
		if property == nil {
			break
		}
		elements = append(elements, property)
	}
	return AsSyntaxListOf[*NamedParameter](NewTypedSyntaxList("NamedParameter", elements...))
}

func (p *queryParserState) parseRenderOperator() *RenderOperator {
	keyword := p.parseToken(SyntaxKindRenderKeyword)
	if keyword == nil {
		return nil
	}
	chartTypes := p.renderChartTypes()
	chartType := p.parseTokenText(chartTypes)
	if chartType == nil {
		chartType = p.parseToken(SyntaxKindIdentifierToken)
	}
	if chartType == nil {
		chartType = CreateMissingSyntaxTokenForTexts(chartTypes)
	}
	parameters := p.parseDeprecatedRenderPropertyList()
	withClause := p.parseRenderWithClause()
	return MustNewRenderOperatorTyped(keyword, chartType, parameters, withClause)
}

func (p *queryParserState) renderChartTypes() []string {
	return KustoFacts.ChartTypes
}

func (p *queryParserState) parseDataScopeClause() *DataScopeClause {
	keyword := p.parseToken(SyntaxKindDataScopeKeyword)
	if keyword == nil {
		return nil
	}
	equal := p.parseRequiredToken(SyntaxKindEqualToken)
	value := p.parseTokenTextOrDiagnostic(KustoFacts.DataScopeValues)
	if value == nil {
		value = NewMissingToken(SyntaxKindIdentifierToken, DiagnosticTermsExpected(KustoFacts.DataScopeValues...))
	}
	return MustNewDataScopeClauseTyped(keyword, equal, value)
}

func (p *queryParserState) parseFindOperandNameWithOptionalAsOperator() Expression {
	expr := p.parseBracketedEntityNamePathElementSelector()
	if expr == nil {
		expr = p.parseBarePathElementSelector()
	}
	if expr == nil {
		return nil
	}
	bar := p.parseToken(SyntaxKindBarToken)
	if bar == nil {
		return expr
	}
	var operator QueryOperator
	if asOperator := p.parseAsOperator(); asOperator != nil {
		operator = asOperator
	} else {
		operator = p.createMissingQueryOperator()
	}
	return MustNewPipeExpressionTyped(expr, bar, operator)
}

func (p *queryParserState) parseFindInClause() *FindInClause {
	keyword := p.parseToken(SyntaxKindInKeyword)
	if keyword == nil {
		return nil
	}
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	expressions := p.parseFindOperandCommaList(true)
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewFindInClauseTyped(keyword, open, expressions, close)
}

func (p *queryParserState) parseTypedColumnNameReference() Expression {
	name := p.parseNameReference()
	if name == nil {
		return nil
	}
	if p.peekToken().Kind != SyntaxKindColonToken {
		return name
	}
	colon := p.parseToken()
	typ := p.parseParamTypeExtended()
	if typ == nil {
		typ = p.parseInvalidParamType()
	}
	if typ == nil {
		typ = createMissingType()
	}
	return MustNewTypedColumnReferenceTyped(name, colon, typ)
}

func (p *queryParserState) parsePackExpression() *PackExpression {
	keyword := p.parseToken(SyntaxKindPackKeyword)
	if keyword == nil {
		return nil
	}
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	asterisk := p.parseRequiredToken(SyntaxKindAsteriskToken)
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewPackExpressionTyped(keyword, open, asterisk, close)
}

func (p *queryParserState) parseFindProjectColumn() Expression {
	if pack := p.parsePackExpression(); pack != nil {
		return pack
	}
	if p.peekToken().Kind == SyntaxKindAsteriskToken {
		return MustNewStarExpressionTyped(p.parseToken())
	}
	return p.parseTypedColumnNameReference()
}

func (p *queryParserState) parseFindProjectColumnCommaList(oneOrMore bool) *SeparatedSyntaxListOf[Expression] {
	return p.parseExpressionCommaListWithEnd(oneOrMore, nil, p.parseFindProjectColumn)
}

func emptyFindProjectColumns() *SeparatedSyntaxListOf[Expression] {
	return AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>"))
}

func (p *queryParserState) parseFindProjectClause() *FindProjectClause {
	switch p.peekToken().Kind {
	case SyntaxKindProjectKeyword:
		return MustNewFindProjectClauseTyped(p.parseToken(), p.parseFindProjectColumnCommaList(true))
	case SyntaxKindProjectSmartKeyword:
		return MustNewFindProjectClauseTyped(p.parseToken(), emptyFindProjectColumns())
	default:
		return nil
	}
}

func (p *queryParserState) parseFindProjectAwayClause() *FindProjectClause {
	keyword := p.parseToken(SyntaxKind_ProjectAwayKeyword)
	if keyword == nil {
		return nil
	}
	return MustNewFindProjectClauseTyped(keyword, p.parseFindProjectColumnCommaList(true))
}

func (p *queryParserState) parseFindOperator() *FindOperator {
	keyword := p.parseToken(SyntaxKindFindKeyword)
	if keyword == nil {
		return nil
	}
	dataScope := p.parseDataScopeClause()
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.FindParameters, false)
	inClause := p.parseFindInClause()
	var where *SyntaxToken
	if parameters.Count() > 0 || inClause != nil {
		where = p.parseRequiredToken(SyntaxKindWhereKeyword)
	} else {
		where = p.parseToken(SyntaxKindWhereKeyword)
	}
	condition := p.parseUnnamedExpression()
	if condition == nil {
		condition = p.createMissingExpression()
	}
	project := p.parseFindProjectClause()
	projectAway := p.parseFindProjectAwayClause()
	return MustNewFindOperatorTyped(keyword, dataScope, parameters, inClause, where, condition, project, projectAway)
}

func (p *queryParserState) parseSearchPredicate() Expression {
	if expr := p.parseUnnamedExpression(); expr != nil {
		return expr
	}
	if p.peekToken().Kind != SyntaxKindAsteriskToken {
		return nil
	}
	expr := MustNewStarExpressionTyped(p.parseToken())
	if andKeyword := p.parseToken(SyntaxKindAndKeyword); andKeyword != nil {
		right := p.parseUnnamedExpression()
		if right == nil {
			right = p.createMissingExpression()
		}
		return MustNewBinaryExpressionTyped(SyntaxKindAndExpression, expr, andKeyword, right)
	}
	return expr
}

func (p *queryParserState) parseSearchOperator() *SearchOperator {
	keyword := p.parseToken(SyntaxKindSearchKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.SearchParameters, true)
	dataScope := p.parseDataScopeClause()
	inClause := p.parseFindInClause()
	condition := p.parseSearchPredicate()
	if condition == nil {
		condition = p.createMissingExpression()
	}
	return MustNewSearchOperatorTyped(keyword, parameters, dataScope, inClause, condition)
}

func (p *queryParserState) parseReduceByWithClause() *ReduceByWithClause {
	keyword := p.parseToken(SyntaxKindWithKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterCommaList(QueryOperatorParameters.ReduceWithParameters, queryOperatorParameterNameDeclaredOrKnown)
	if parameters == nil {
		parameters = AsSeparatedSyntaxListOf[*NamedParameter](NewTypedSyntaxList("SeparatedElement<NamedParameter>"))
	}
	return MustNewReduceByWithClauseTyped(keyword, parameters)
}

func (p *queryParserState) parseReduceByOperator() *ReduceByOperator {
	keyword := p.parseToken(SyntaxKindReduceKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseReduceLeadingParameterList()
	byKeyword := p.parseRequiredToken(SyntaxKindByKeyword)
	expr := p.parseNamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	withClause := p.parseReduceByWithClause()
	return MustNewReduceByOperatorTyped(keyword, parameters, byKeyword, expr, withClause)
}

func (p *queryParserState) parseReduceLeadingParameterList() *SyntaxListOf[*NamedParameter] {
	var elements []SyntaxElement
	for {
		nameText, nameLen, parameter := p.scanQueryOperatorParameterNameWithKind(QueryOperatorParameters.ReduceLeadingParameters, queryOperatorParameterNameDeclaredOrKnown, false, 0)
		if parameter == nil {
			break
		}
		if !parameter.HasNoEquals() &&
			p.peekToken(nameLen).Kind == SyntaxKindEqualToken &&
			p.peekToken(nameLen+1).Kind == SyntaxKindByKeyword {
			nameToken := p.parseQueryOperatorParameterNameTokenWithText(nameText, nameLen)
			if nameToken == nil {
				break
			}
			name := MustNewNameDeclarationTyped(MustNewTokenNameTyped(nameToken))
			equal := p.parseRequiredToken(SyntaxKindEqualToken)
			elements = append(elements, MustNewNamedParameterTyped(name, equal, p.createMissingValue(), queryOperatorParameterExpressionHint(parameter)))
			continue
		}
		named := p.parseQueryOperatorParameter(parameter)
		if named == nil {
			break
		}
		elements = append(elements, named)
	}
	return AsSyntaxListOf[*NamedParameter](NewTypedSyntaxList("NamedParameter", elements...))
}

func (p *queryParserState) scanNameEqualsClause(offset ...int) int {
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	length := p.scanName(delta)
	if length > 0 && p.peekToken(delta+length).Kind == SyntaxKindEqualToken {
		return length + 1
	}
	return -1
}

func (p *queryParserState) parseNameEqualsClause() *NameEqualsClause {
	if p.scanNameEqualsClause() <= 0 {
		return nil
	}
	name := p.parseNameDeclaration()
	equal := p.parseRequiredToken(SyntaxKindEqualToken)
	return MustNewNameEqualsClauseTyped(name, equal)
}

func (p *queryParserState) parseForkPipeExpression() Expression {
	var expr Expression
	if operator := p.parsePipedQueryOperator(); operator != nil {
		expr = queryOperatorExpression(operator)
	} else if operator := p.parseUnpipedQueryOperator(); operator != nil {
		expr = queryOperatorExpression(operator)
	}
	for expr != nil && p.peekToken().Kind == SyntaxKindBarToken {
		bar := p.parseToken()
		operator := p.parseRequiredQueryOperator()
		expr = MustNewPipeExpressionTyped(expr, bar, operator)
	}
	return expr
}

func (p *queryParserState) parseForkExpression() *ForkExpression {
	var nameEquals *NameEqualsClause
	if p.scanNameEqualsClause() > 0 {
		nameEquals = p.parseNameEqualsClause()
	}
	if nameEquals == nil && p.peekToken().Kind != SyntaxKindOpenParenToken {
		return nil
	}
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	expr := p.parseForkPipeExpression()
	if expr == nil {
		expr = p.parseExpression()
	}
	if expr == nil {
		expr = p.createMissingExpression()
	}
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewForkExpressionTyped(nameEquals, open, expr, close)
}

func (p *queryParserState) createMissingForkExpression() *ForkExpression {
	return MustNewForkExpressionTyped(
		nil,
		createMissingToken(SyntaxKindOpenParenToken),
		p.createMissingExpression(),
		createMissingToken(SyntaxKindCloseParenToken),
	)
}

func (p *queryParserState) parseForkExpressionList() *SyntaxListOf[*ForkExpression] {
	var elements []SyntaxElement
	for !p.scanCommonListEnd() {
		expr := p.parseForkExpression()
		if expr == nil {
			break
		}
		elements = append(elements, expr)
	}
	if len(elements) == 0 {
		elements = append(elements, p.createMissingForkExpression())
	}
	return AsSyntaxListOf[*ForkExpression](NewTypedSyntaxList("ForkExpression", elements...))
}

func (p *queryParserState) parseMakeGraphOperator() *MakeGraphOperator {
	keyword := p.parseToken(SyntaxKindMakeGraphKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.GraphMakeParameters, true)
	sourceColumn := p.parseNameReference()
	if sourceColumn == nil {
		sourceColumn = p.createMissingNameReferenceNode()
	}
	direction := p.parseFirstTokenTextAsKind(
		struct {
			text string
			kind SyntaxKind
		}{"-->", SyntaxKindDashDashGreaterThanToken},
		struct {
			text string
			kind SyntaxKind
		}{"--", SyntaxKindDashDashToken},
	)
	if direction == nil {
		direction = CreateMissingSyntaxTokenForKinds([]SyntaxKind{SyntaxKindDashDashGreaterThanToken, SyntaxKindDashDashToken})
	}
	targetColumn := p.parseNameReference()
	if targetColumn == nil {
		targetColumn = p.createMissingNameReferenceNode()
	}
	var withClause MakeGraphWithClause
	if implicitWithClause := p.parseMakeGraphWithImplicitIDClause(); implicitWithClause != nil {
		withClause = implicitWithClause
	} else {
		withClause = p.parseMakeGraphWithTablesAndKeysClause()
	}
	partitionedBy := p.parseMakeGraphPartitionedByClause()
	return MustNewMakeGraphOperatorTyped(keyword, parameters, sourceColumn, direction, targetColumn, withClause, partitionedBy)
}

func (p *queryParserState) parseMakeGraphWithImplicitIDClause() *MakeGraphWithImplicitIdClause {
	keyword := p.parseToken(SyntaxKindWithNodeIdKeyword)
	if keyword == nil {
		return nil
	}
	equal := p.parseRequiredToken(SyntaxKindEqualToken)
	name := p.parseNameDeclaration()
	if name == nil {
		name = p.createMissingNameDeclaration()
	}
	return MustNewMakeGraphWithImplicitIdClauseTyped(keyword, equal, name)
}

func (p *queryParserState) parseMakeGraphWithTablesAndKeysClause() *MakeGraphWithTablesAndKeysClause {
	keyword := p.parseToken(SyntaxKindWithKeyword)
	if keyword == nil {
		return nil
	}
	return MustNewMakeGraphWithTablesAndKeysClauseTyped(keyword, p.parseMakeGraphTableAndKeyClauseCommaList())
}

func (p *queryParserState) parseMakeGraphTableAndKeyClauseCommaList() *SeparatedSyntaxListOf[*MakeGraphTableAndKeyClause] {
	var elements []SyntaxElement
	element := p.parseMakeGraphTableAndKeyClause()
	if element == nil && p.peekToken().Kind == SyntaxKindCommaToken {
		element = createMissingMakeGraphTableAndKeyClause()
	}
	for element != nil {
		if p.peekToken().Kind == SyntaxKindEndOfTextToken || p.peekToken().Kind == SyntaxKindNone || p.scanMakeGraphTableAndKeyClauseListEnd() {
			elements = append(elements, NewSeparatedElement(element, nil))
			return AsSeparatedSyntaxListOf[*MakeGraphTableAndKeyClause](NewTypedSyntaxList("SeparatedElement<MakeGraphTableAndKeyClause>", elements...))
		}
		if p.peekToken().Kind == SyntaxKindCommaToken {
			comma := p.parseToken()
			elements = append(elements, NewSeparatedElement(element, comma))
			if p.scanMakeGraphTableAndKeyClauseListEnd() {
				element = createMissingMakeGraphTableAndKeyClause()
				continue
			}
			next := p.parseMakeGraphTableAndKeyClause()
			if next != nil {
				element = next
				continue
			}
			element = createMissingMakeGraphTableAndKeyClause()
			continue
		}
		next := p.parseMakeGraphTableAndKeyClause()
		if next == nil {
			elements = append(elements, NewSeparatedElement(element, nil))
			return AsSeparatedSyntaxListOf[*MakeGraphTableAndKeyClause](NewTypedSyntaxList("SeparatedElement<MakeGraphTableAndKeyClause>", elements...))
		}
		elements = append(elements, NewSeparatedElement(element, createMissingToken(SyntaxKindCommaToken)))
		element = next
	}
	if len(elements) == 0 {
		elements = append(elements, NewSeparatedElement(createMissingMakeGraphTableAndKeyClause(), nil))
	}
	return AsSeparatedSyntaxListOf[*MakeGraphTableAndKeyClause](NewTypedSyntaxList("SeparatedElement<MakeGraphTableAndKeyClause>", elements...))
}

func (p *queryParserState) scanMakeGraphTableAndKeyClauseListEnd() bool {
	return p.scanCommonListEnd() || p.peekToken().Kind == SyntaxKindPartitionedByKeyword
}

func createMissingMakeGraphTableAndKeyClause() *MakeGraphTableAndKeyClause {
	missingReference := func() *NameReference {
		return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone)
	}
	return MustNewMakeGraphTableAndKeyClauseTyped(
		missingReference(),
		createOmittedToken(SyntaxKindOnKeyword),
		missingReference(),
		DiagnosticMissingExpression(),
	)
}

func (p *queryParserState) parseMakeGraphTableAndKeyClause() *MakeGraphTableAndKeyClause {
	table := p.parseInvocationExpression()
	if table == nil {
		return nil
	}
	onKeyword := p.parseRequiredToken(SyntaxKindOnKeyword)
	column := p.parseNameReference()
	if column == nil {
		column = p.createMissingNameReferenceNode()
	}
	return MustNewMakeGraphTableAndKeyClauseTyped(table, onKeyword, column)
}

func (p *queryParserState) parseMakeGraphPartitionedByClause() *MakeGraphPartitionedByClause {
	keyword := p.parseToken(SyntaxKindPartitionedByKeyword)
	if keyword == nil {
		return nil
	}
	entity := p.parseNameReference()
	if entity == nil {
		entity = p.createMissingNameReferenceNode()
	}
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	subquery := p.parseContextualSubExpression()
	if subquery == nil {
		subquery = p.createMissingExpression()
	}
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewMakeGraphPartitionedByClauseTyped(keyword, entity, open, subquery, close)
}

func (p *queryParserState) parseGraphMarkComponentsOperator() *GraphMarkComponentsOperator {
	keyword := p.parseToken(SyntaxKindGraphMarkComponentsKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.GraphMarkComponentsParameters, true)
	return MustNewGraphMarkComponentsOperatorTyped(keyword, parameters)
}

func (p *queryParserState) parseGraphWhereNodesOperator() *GraphWhereNodesOperator {
	keyword := p.parseToken(SyntaxKindGraphWhereNodesKeyword)
	if keyword == nil {
		return nil
	}
	condition := p.parseUnnamedExpression()
	if condition == nil {
		condition = p.createMissingExpression()
	}
	return MustNewGraphWhereNodesOperatorTyped(keyword, condition)
}

func (p *queryParserState) parseGraphWhereEdgesOperator() *GraphWhereEdgesOperator {
	keyword := p.parseToken(SyntaxKindGraphWhereEdgesKeyword)
	if keyword == nil {
		return nil
	}
	condition := p.parseUnnamedExpression()
	if condition == nil {
		condition = p.createMissingExpression()
	}
	return MustNewGraphWhereEdgesOperatorTyped(keyword, condition)
}

func (p *queryParserState) parseGraphMatchOperator() *GraphMatchOperator {
	keyword := p.parseToken(SyntaxKindGraphMatchKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.GraphMatchParameters, true)
	return MustNewGraphMatchOperatorTyped(keyword, parameters, p.parseGraphMatchPatternCommaList(), p.parseGraphMatchWhereClause(), p.parseProjectClause())
}

func (p *queryParserState) parseGraphShortestPathsOperator() *GraphShortestPathsOperator {
	keyword := p.parseToken(SyntaxKindGraphShortestPathsKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.GraphShortestPathsParameters, true)
	return MustNewGraphShortestPathsOperatorTyped(keyword, parameters, p.parseGraphMatchPatternCommaList(), p.parseGraphMatchWhereClause(), p.parseProjectClause())
}

func (p *queryParserState) parseGraphMatchWhereClause() *WhereClause {
	keyword := p.parseToken(SyntaxKindWhereKeyword)
	if keyword == nil {
		return nil
	}
	condition := p.parseExpression()
	if condition == nil {
		condition = p.createMissingExpression()
	}
	return MustNewWhereClauseTyped(keyword, condition)
}

func (p *queryParserState) parseGraphMatchPatternCommaList() *SeparatedSyntaxListOf[*GraphMatchPattern] {
	var elements []SyntaxElement
	element := p.parseGraphMatchPattern()
	if element == nil && p.peekToken().Kind == SyntaxKindCommaToken {
		element = createMissingGraphMatchPattern()
	}
	for element != nil {
		if p.peekToken().Kind == SyntaxKindEndOfTextToken || p.peekToken().Kind == SyntaxKindNone || p.scanGraphMatchPatternListEnd() {
			elements = append(elements, NewSeparatedElement(element, nil))
			return AsSeparatedSyntaxListOf[*GraphMatchPattern](NewTypedSyntaxList("SeparatedElement<GraphMatchPattern>", elements...))
		}
		if p.peekToken().Kind == SyntaxKindCommaToken {
			comma := p.parseToken()
			elements = append(elements, NewSeparatedElement(element, comma))
			if p.scanGraphMatchPatternListEnd() {
				elements = append(elements, NewSeparatedElement(createMissingGraphMatchPattern(), nil))
				break
			}
			next := p.parseGraphMatchPattern()
			if next != nil {
				element = next
				continue
			}
			element = createMissingGraphMatchPattern()
			continue
		}
		next := p.parseGraphMatchPattern()
		if next == nil {
			elements = append(elements, NewSeparatedElement(element, nil))
			return AsSeparatedSyntaxListOf[*GraphMatchPattern](NewTypedSyntaxList("SeparatedElement<GraphMatchPattern>", elements...))
		}
		elements = append(elements, NewSeparatedElement(element, createMissingToken(SyntaxKindCommaToken)))
		element = next
	}
	if len(elements) == 0 {
		elements = append(elements, NewSeparatedElement(createMissingGraphMatchPattern(), nil))
	}
	return AsSeparatedSyntaxListOf[*GraphMatchPattern](NewTypedSyntaxList("SeparatedElement<GraphMatchPattern>", elements...))
}

func (p *queryParserState) scanGraphMatchPatternListEnd() bool {
	switch p.peekToken().Kind {
	case SyntaxKindWhereKeyword, SyntaxKindProjectKeyword:
		return true
	default:
		return p.scanCommonListEnd()
	}
}

func createMissingGraphMatchPattern() *GraphMatchPattern {
	return MustNewGraphMatchPatternTyped(
		AsSyntaxListOf[GraphMatchPatternNotation](NewTypedSyntaxList("GraphMatchPatternNotation", createMissingGraphMatchPatternNotation())),
		DiagnosticMissingGraphMatchPattern(),
	)
}

func createMissingGraphMatchPatternNotation() GraphMatchPatternNotation {
	return MustNewGraphMatchPatternNodeTyped(
		createOmittedToken(SyntaxKindOpenParenToken),
		MustNewNameDeclarationTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken))),
		createOmittedToken(SyntaxKindCloseParenToken),
	)
}

func (p *queryParserState) parseGraphMatchPattern() *GraphMatchPattern {
	var elements []SyntaxElement
	for {
		if p.peekToken().Kind == SyntaxKindEndOfTextToken || p.peekToken().Kind == SyntaxKindNone || p.scanGraphMatchPatternListEnd() || p.peekToken().Kind == SyntaxKindCommaToken {
			break
		}
		element := p.parseGraphMatchPatternNode()
		if element == nil {
			element = p.parseGraphMatchPatternEdge()
		}
		if element == nil {
			break
		}
		elements = append(elements, element)
	}
	if len(elements) == 0 {
		return nil
	}
	return MustNewGraphMatchPatternTyped(AsSyntaxListOf[GraphMatchPatternNotation](NewTypedSyntaxList("GraphMatchPatternNotation", elements...)))
}

func (p *queryParserState) parseGraphMatchPatternNode() GraphMatchPatternNotation {
	open := p.parseToken(SyntaxKindOpenParenToken)
	if open == nil {
		return nil
	}
	name := p.parseNameDeclaration()
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewGraphMatchPatternNodeTyped(open, name, close)
}

func (p *queryParserState) parseGraphMatchPatternEdge() GraphMatchPatternNotation {
	firstToken := p.parseFirstTokenTextAsKind(
		struct {
			text string
			kind SyntaxKind
		}{"-->", SyntaxKindDashDashGreaterThanToken},
		struct {
			text string
			kind SyntaxKind
		}{"<--", SyntaxKindLessThanDashDashToken},
		struct {
			text string
			kind SyntaxKind
		}{"--", SyntaxKindDashDashToken},
	)
	if firstToken != nil {
		return MustNewGraphMatchPatternEdgeTyped(firstToken, nil, nil, nil)
	}

	firstToken = p.parseFirstTokenTextAsKind(
		struct {
			text string
			kind SyntaxKind
		}{"-[", SyntaxKindDashBracketToken},
		struct {
			text string
			kind SyntaxKind
		}{"<-[", SyntaxKindLessThanDashBracketToken},
	)
	if firstToken == nil {
		return nil
	}
	name := p.parseNameDeclaration()
	rangeChild := p.parseGraphMatchPatternEdgeRange()
	lastToken := p.parseFirstTokenTextAsKind(
		struct {
			text string
			kind SyntaxKind
		}{"]->", SyntaxKindBracketDashGreaterThanToken},
		struct {
			text string
			kind SyntaxKind
		}{"]-", SyntaxKindBracketDashToken},
	)
	if lastToken == nil {
		lastToken = CreateMissingSyntaxTokenForKinds([]SyntaxKind{SyntaxKindBracketDashGreaterThanToken, SyntaxKindBracketDashToken})
	}
	return MustNewGraphMatchPatternEdgeTyped(firstToken, name, rangeChild, lastToken)
}

func (p *queryParserState) parseGraphMatchPatternEdgeRange() *GraphMatchPatternEdgeRange {
	asterisk := p.parseToken(SyntaxKindAsteriskToken)
	if asterisk == nil {
		return nil
	}
	rangeStart := p.parseInvocationExpression()
	if rangeStart == nil {
		rangeStart = p.createMissingValue()
	}
	dotDot := p.parseRequiredToken(SyntaxKindDotDotToken)
	rangeEnd := p.parseInvocationExpression()
	if rangeEnd == nil {
		rangeEnd = p.createMissingExpression()
	}
	return MustNewGraphMatchPatternEdgeRangeTyped(asterisk, rangeStart, dotDot, rangeEnd)
}

func (p *queryParserState) parseWhereClause() *WhereClause {
	keyword := p.parseToken(SyntaxKindWhereKeyword)
	if keyword == nil {
		return nil
	}
	condition := p.parseExpression()
	if condition == nil {
		condition = p.createMissingExpression()
	}
	var diagnostics []Diagnostic
	if p.consumeUnexpectedCommaExpressionTail() {
		diagnostics = append(diagnostics, DiagnosticIncompleteFragment())
	}
	return MustNewWhereClauseTyped(keyword, condition, diagnostics...)
}

func (p *queryParserState) parseProjectClause() *ProjectClause {
	keyword := p.parseToken(SyntaxKindProjectKeyword)
	if keyword == nil {
		return nil
	}
	return MustNewProjectClauseTyped(keyword, p.parseNamedExpressionCommaList(true))
}

func (p *queryParserState) parseGraphToTableOperator() *GraphToTableOperator {
	keyword := p.parseToken(SyntaxKindGraphToTableKeyword)
	if keyword == nil {
		return nil
	}
	return MustNewGraphToTableOperatorTyped(keyword, p.parseGraphToTableOutputClauseCommaList())
}

func (p *queryParserState) parseGraphToTableOutputClauseCommaList() *SeparatedSyntaxListOf[*GraphToTableOutputClause] {
	var elements []SyntaxElement
	element := p.parseGraphToTableOutputClause()
	if element == nil && p.peekToken().Kind == SyntaxKindCommaToken {
		element = createMissingGraphToTableOutputClause()
	}
	for element != nil {
		if p.peekToken().Kind == SyntaxKindEndOfTextToken || p.peekToken().Kind == SyntaxKindNone || p.scanCommonListEnd() {
			elements = append(elements, NewSeparatedElement(element, nil))
			return AsSeparatedSyntaxListOf[*GraphToTableOutputClause](NewTypedSyntaxList("SeparatedElement<GraphToTableOutputClause>", elements...))
		}
		if p.peekToken().Kind == SyntaxKindCommaToken {
			comma := p.parseToken()
			elements = append(elements, NewSeparatedElement(element, comma))
			if p.scanCommonListEnd() {
				elements = append(elements, NewSeparatedElement(createMissingGraphToTableOutputClause(), nil))
				break
			}
			next := p.parseGraphToTableOutputClause()
			if next != nil {
				element = next
				continue
			}
			element = createMissingGraphToTableOutputClause()
			continue
		}
		elements = append(elements, NewSeparatedElement(element, nil))
		return AsSeparatedSyntaxListOf[*GraphToTableOutputClause](NewTypedSyntaxList("SeparatedElement<GraphToTableOutputClause>", elements...))

	}
	if len(elements) == 0 {
		elements = append(elements, NewSeparatedElement(createMissingGraphToTableOutputClause(), nil))
	}
	return AsSeparatedSyntaxListOf[*GraphToTableOutputClause](NewTypedSyntaxList("SeparatedElement<GraphToTableOutputClause>", elements...))
}

func createMissingGraphToTableOutputClause() *GraphToTableOutputClause {
	return MustNewGraphToTableOutputClauseTyped(
		createOmittedToken(SyntaxKindOpenParenToken),
		MustNewGraphToTableAsClauseTyped(
			createOmittedToken(SyntaxKindOpenParenToken),
			MustNewNameDeclarationTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken))),
		),
		AsSyntaxListOf[*NamedParameter](NewTypedSyntaxList("NamedParameter")),
		DiagnosticIncorrectNumberOfOutputGraphToTableEntities(),
	)
}

func (p *queryParserState) parseGraphToTableOutputClause() *GraphToTableOutputClause {
	keyword := p.parseToken(SyntaxKindNodesKeyword, SyntaxKindGraphEdgesKeyword)
	if keyword == nil {
		return nil
	}
	asClause := p.parseGraphToTableAsClause()
	parameters := QueryOperatorParameters.GraphToTableNodesParameters
	if keyword.Kind() == SyntaxKindGraphEdgesKeyword {
		parameters = QueryOperatorParameters.GraphToTableEdgesParameters
	}
	return MustNewGraphToTableOutputClauseTyped(keyword, asClause, p.parseQueryOperatorParameterList(parameters, true))
}

func (p *queryParserState) parseGraphToTableAsClause() *GraphToTableAsClause {
	keyword := p.parseToken(SyntaxKindAsKeyword)
	if keyword == nil {
		return nil
	}
	name := p.parseNameDeclaration()
	if name == nil {
		name = p.createMissingNameDeclaration()
	}
	return MustNewGraphToTableAsClauseTyped(keyword, name)
}

func (p *queryParserState) parseForkOperator() *ForkOperator {
	keyword := p.parseToken(SyntaxKindForkKeyword)
	if keyword == nil {
		return nil
	}
	return MustNewForkOperatorTyped(keyword, p.parseForkExpressionList())
}

func (p *queryParserState) parseFacetWithClause() FacetWithClause {
	keyword := p.parseToken(SyntaxKindWithKeyword)
	if keyword == nil {
		return nil
	}
	open := p.parseToken(SyntaxKindOpenParenToken)
	if open != nil {
		expr := p.parseForkPipeExpression()
		if expr == nil {
			expr = p.createMissingQueryOperatorExpression()
		}
		close := p.parseRequiredToken(SyntaxKindCloseParenToken)
		return MustNewFacetWithExpressionClauseTyped(keyword, open, expr, close)
	}
	operator := p.parsePipedQueryOperator()
	if isNilSyntaxElement(operator) {
		operator = p.parseUnpipedQueryOperator()
	}
	if isNilSyntaxElement(operator) {
		operator = p.createMissingQueryOperator()
	}
	return MustNewFacetWithOperatorClauseTyped(keyword, operator)
}

func (p *queryParserState) scanFacetExpressionListEnd() bool {
	return p.peekToken().Kind == SyntaxKindWithKeyword || p.scanCommonListEnd()
}

func (p *queryParserState) parseFacetExpressionList() *SeparatedSyntaxListOf[Expression] {
	var elements []SyntaxElement
	var element Expression
	if !p.scanFacetExpressionListEnd() {
		element = p.parseSimplePathExpression()
	}
	if isNilSyntaxElement(element) && p.peekToken().Kind == SyntaxKindCommaToken {
		element = p.createMissingNameReference()
	}
	for !isNilSyntaxElement(element) {
		if p.peekToken().Kind == SyntaxKindEndOfTextToken || p.peekToken().Kind == SyntaxKindNone || p.scanFacetExpressionListEnd() {
			elements = append(elements, NewSeparatedElement(element, nil))
			return AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>", elements...))
		}
		if p.peekToken().Kind == SyntaxKindCommaToken {
			comma := p.parseToken()
			elements = append(elements, NewSeparatedElement(element, comma))
			if p.scanFacetExpressionListEnd() {
				elements = append(elements, NewSeparatedElement(p.createMissingNameReference(), nil))
				break
			}
			next := p.parseSimplePathExpression()
			if !isNilSyntaxElement(next) {
				element = next
				continue
			}
			element = p.createMissingNameReference()
			continue
		}
		start := p.pos
		next := p.parseSimplePathExpression()
		if !isNilSyntaxElement(next) {
			if p.pos == start {
				elements = append(elements, NewSeparatedElement(element, nil))
				return AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>", elements...))
			}
			elements = append(elements, NewSeparatedElement(element, createMissingToken(SyntaxKindCommaToken)))
			element = next
			continue
		}
		elements = append(elements, NewSeparatedElement(element, nil))
		return AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>", elements...))
	}
	if len(elements) == 0 {
		elements = append(elements, NewSeparatedElement(p.createMissingNameReference(), nil))
	}
	return AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>", elements...))
}

func (p *queryParserState) parseFacetOperator() *FacetOperator {
	keyword := p.parseToken(SyntaxKindFacetKeyword)
	if keyword == nil {
		return nil
	}
	byKeyword := p.parseRequiredToken(SyntaxKindByKeyword)
	expressions := p.parseFacetExpressionList()
	withClause := p.parseFacetWithClause()
	return MustNewFacetOperatorTyped(keyword, byKeyword, expressions, withClause)
}

func (p *queryParserState) parseSummarizeByClause() *SummarizeByClause {
	keyword := p.parseToken(SyntaxKindByKeyword)
	if keyword == nil {
		return nil
	}
	expressions := p.parseNamedExpressionCommaListWithEnd(true, func() bool {
		return p.scanCommonListEnd() || p.scanSummarizeByBinClause()
	})
	binClause := p.parseSummarizeByBinClause()
	return MustNewSummarizeByClauseTyped(keyword, expressions, binClause)
}

func (p *queryParserState) scanSummarizeByBinClause() bool {
	return p.peekToken().Kind == SyntaxKindBinKeyword && p.peekToken(1).Kind == SyntaxKindEqualToken && SyntaxCategoryOf(p.peekToken(2).Kind) == SyntaxCategoryLiteral
}

func (p *queryParserState) parseSummarizeByBinClause() NamedExpression {
	if p.peekToken().Kind != SyntaxKindBinKeyword || p.peekToken(1).Kind != SyntaxKindEqualToken {
		return nil
	}
	keyword := p.parseToken()
	equal := p.parseToken(SyntaxKindEqualToken)
	value := p.parseUnnamedExpression()
	if value == nil {
		value = p.createMissingExpression()
	}
	name := MustNewNameDeclarationTyped(MustNewTokenNameTyped(keyword))
	return MustNewSimpleNamedExpressionTyped(name, equal, value)
}

func (p *queryParserState) parseSummarizeOperator() *SummarizeOperator {
	keyword := p.parseToken(SyntaxKindSummarizeKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.SummarizeParameters, true)
	aggregates := p.parseNamedExpressionCommaListWithEnd(false, func() bool {
		return p.peekToken().Kind == SyntaxKindByKeyword || p.scanCommonListEnd()
	})
	byClause := p.parseSummarizeByClause()
	return MustNewSummarizeOperatorTyped(keyword, parameters, aggregates, byClause)
}

func (p *queryParserState) parseDistinctOperator() *DistinctOperator {
	keyword := p.parseToken(SyntaxKindDistinctKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.DistinctParameters, true)
	expressions := p.parseDistinctExpressionCommaList(true)
	return MustNewDistinctOperatorTyped(keyword, parameters, expressions)
}

func (p *queryParserState) parseTakeOperator() *TakeOperator {
	keyword := p.parseToken(SyntaxKindTakeKeyword, SyntaxKindLimitKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.TakeParameters, true)
	expr := p.parseNamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewTakeOperatorTyped(keyword, parameters, expr)
}

func (p *queryParserState) consumeUnexpectedSingleExpressionTail() bool {
	consumed := false
	for startsLikelyExtraTopOrderedExpression(p.peekToken().Kind) {
		if p.parseNamedExpression() == nil {
			break
		}
		consumed = true
	}
	return consumed
}

func (p *queryParserState) consumeUnexpectedCommaExpressionTail() bool {
	consumed := false
	for p.peekToken().Kind == SyntaxKindCommaToken {
		p.parseToken(SyntaxKindCommaToken)
		consumed = true
		for startsLikelyExtraTopOrderedExpression(p.peekToken().Kind) {
			if p.parseNamedExpression() == nil {
				break
			}
			consumed = true
		}
	}
	return consumed
}

func (p *queryParserState) parseSortOperator() *SortOperator {
	keyword := p.parseToken(SyntaxKindOrderKeyword, SyntaxKindSortKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.SortParameters, false)
	byKeyword := p.parseRequiredToken(SyntaxKindByKeyword)
	expressions := p.parseOrderedExpressionCommaList(true)
	return MustNewSortOperatorTyped(keyword, parameters, byKeyword, expressions)
}

func (p *queryParserState) parseScanAssignment() *ScanAssignment {
	name := p.parseExtendedNameReference()
	if name == nil {
		return nil
	}
	equal := p.parseRequiredToken(SyntaxKindEqualToken)
	expr := p.parseUnnamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewScanAssignmentTyped(name, equal, expr)
}

func createMissingScanAssignment() *ScanAssignment {
	return MustNewScanAssignmentTyped(
		MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone),
		createOmittedToken(SyntaxKindEqualEqualToken),
		MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone),
		DiagnosticMissingName(),
	)
}

func (p *queryParserState) parseScanAssignmentCommaList() *SeparatedSyntaxListOf[*ScanAssignment] {
	return parseCommaList(p, "ScanAssignment", p.parseScanAssignment, createMissingScanAssignment, p.atCommonListEnd, true, false)
}

func (p *queryParserState) parseScanComputationClause() *ScanComputationClause {
	arrow := p.parseToken(SyntaxKindFatArrowToken)
	if arrow == nil {
		return nil
	}
	return MustNewScanComputationClauseTyped(arrow, p.parseScanAssignmentCommaList())
}

func (p *queryParserState) parseRequiredScanStepOutputKind() *SyntaxToken {
	if token := p.parseTokenText(KustoFacts.ScanStepOutputValues); token != nil {
		return token
	}
	token := p.peekToken()
	if token.Kind == SyntaxKindIdentifierToken || IsKeywordKind(token.Kind) {
		p.pos++
		return SyntaxTokenFromLexical(token, DiagnosticTermsExpected(KustoFacts.ScanStepOutputValues...))
	}
	return CreateMissingSyntaxTokenForTexts(KustoFacts.ScanStepOutputValues)
}

func (p *queryParserState) parseScanStepOutput() *ScanStepOutput {
	output := p.parseToken(SyntaxKindOutputKeyword)
	if output == nil {
		return nil
	}
	equality := p.parseRequiredToken(SyntaxKindEqualToken)
	outputKind := p.parseRequiredScanStepOutputKind()
	return MustNewScanStepOutputTyped(output, equality, outputKind)
}

func (p *queryParserState) parseScanStep() *ScanStep {
	keyword := p.parseToken(SyntaxKindStepKeyword)
	if keyword == nil {
		return nil
	}
	name := p.parseExtendedNameDeclaration()
	if name == nil {
		name = p.createMissingNameDeclaration()
	}
	optional := p.parseToken(SyntaxKindOptionalKeyword)
	output := p.parseScanStepOutput()
	colon := p.parseRequiredToken(SyntaxKindColonToken)
	expr := p.parseUnnamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	computation := p.parseScanComputationClause()
	semicolon := p.parseRequiredToken(SyntaxKindSemicolonToken)
	return MustNewScanStepTyped(keyword, name, optional, output, colon, expr, computation, semicolon)
}

func (p *queryParserState) scanScanListEnd() bool {
	switch p.peekToken().Kind {
	case SyntaxKindPartitionKeyword, SyntaxKindOrderKeyword, SyntaxKindByKeyword, SyntaxKindDeclareKeyword, SyntaxKindWithKeyword:
		return true
	default:
		return p.scanCommonListEnd()
	}
}

func (p *queryParserState) parseScanOrderByClause() *ScanOrderByClause {
	keyword := p.parseToken(SyntaxKindOrderKeyword)
	if keyword == nil {
		return nil
	}
	byKeyword := p.parseRequiredToken(SyntaxKindByKeyword)
	expressions := p.parseExpressionCommaListWithEnd(true, p.scanScanListEnd, p.parseOrderedExpression)
	return MustNewScanOrderByClauseTyped(keyword, byKeyword, expressions)
}

func (p *queryParserState) parseScanPartitionByClause() *ScanPartitionByClause {
	keyword := p.parseToken(SyntaxKindPartitionKeyword)
	if keyword == nil {
		return nil
	}
	byKeyword := p.parseRequiredToken(SyntaxKindByKeyword)
	expressions := p.parseExpressionCommaListWithEnd(true, p.scanScanListEnd, p.parseUnnamedExpression)
	return MustNewScanPartitionByClauseTyped(keyword, byKeyword, expressions)
}

func (p *queryParserState) parseScanDeclareClause() *ScanDeclareClause {
	keyword := p.parseToken(SyntaxKindDeclareKeyword)
	if keyword == nil {
		return nil
	}
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	declarations := p.parseFunctionParameterCommaList(func() bool { return p.scanCommonListEnd() })
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewScanDeclareClauseTyped(keyword, open, declarations, close)
}

func (p *queryParserState) parseScanStepList() *SyntaxListOf[*ScanStep] {
	var elements []SyntaxElement
	for !p.scanCommonListEnd() {
		step := p.parseScanStep()
		if step == nil {
			break
		}
		elements = append(elements, step)
	}
	return AsSyntaxListOf[*ScanStep](NewTypedSyntaxList("ScanStep", elements...))
}

func (p *queryParserState) parseScanOperator() *ScanOperator {
	keyword := p.parseToken(SyntaxKindScanKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.ScanParameters, false)
	order := p.parseScanOrderByClause()
	partition := p.parseScanPartitionByClause()
	declare := p.parseScanDeclareClause()
	with := p.parseRequiredToken(SyntaxKindWithKeyword)
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	steps := p.parseScanStepList()
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewScanOperatorTyped(keyword, parameters, order, partition, declare, with, open, steps, close)
}

func (p *queryParserState) parseTopOperator() *TopOperator {
	keyword := p.parseToken(SyntaxKindTopKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.TopParameters, true)
	expr := p.parseNamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	byKeyword := p.parseRequiredToken(SyntaxKindByKeyword)
	byExpr := p.parseOrderedExpression()
	if byExpr == nil {
		byExpr = p.createMissingExpression()
	}
	return MustNewTopOperatorTyped(keyword, parameters, expr, byKeyword, byExpr)
}

func startsLikelyExtraTopOrderedExpression(kind SyntaxKind) bool {
	switch kind {
	case SyntaxKindIdentifierToken,
		SyntaxKindOpenBracketToken,
		SyntaxKindOpenParenToken,
		SyntaxKindStringLiteralToken,
		SyntaxKindBooleanLiteralToken,
		SyntaxKindLongLiteralToken,
		SyntaxKindRealLiteralToken,
		SyntaxKindDecimalLiteralToken,
		SyntaxKindIntLiteralToken,
		SyntaxKindGuidLiteralToken,
		SyntaxKindRawGuidLiteralToken,
		SyntaxKindDateTimeLiteralToken,
		SyntaxKindTimespanLiteralToken,
		SyntaxKindNullKeyword,
		SyntaxKindDynamicKeyword:
		return true
	default:
		return false
	}
}

func (p *queryParserState) parseTopHittersByClause() *TopHittersByClause {
	keyword := p.parseToken(SyntaxKindByKeyword)
	if keyword == nil {
		return nil
	}
	expr := p.parseNamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewTopHittersByClauseTyped(keyword, expr)
}

func (p *queryParserState) parseTopHittersOperator() *TopHittersOperator {
	keyword := p.parseToken(SyntaxKindTopHittersKeyword)
	if keyword == nil {
		return nil
	}
	expr := p.parseNamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	ofKeyword := p.parseRequiredToken(SyntaxKindOfKeyword)
	ofExpr := p.parseNamedExpression()
	if ofExpr == nil {
		ofExpr = p.createMissingExpression()
	}
	byClause := p.parseTopHittersByClause()
	return MustNewTopHittersOperatorTyped(keyword, expr, ofKeyword, ofExpr, byClause)
}

func (p *queryParserState) parseTopNestedWithOthersClause() *TopNestedWithOthersClause {
	keyword := p.parseToken(SyntaxKindWithKeyword)
	if keyword == nil {
		return nil
	}
	others := p.parseRequiredToken(SyntaxKindOthersKeyword)
	equal := p.parseRequiredToken(SyntaxKindEqualToken)
	expr := p.parseLiteral()
	if expr == nil {
		expr = p.parseUnnamedExpression()
	}
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewTopNestedWithOthersClauseTyped(keyword, others, equal, expr)
}

func (p *queryParserState) parseTopNestedClause() *TopNestedClause {
	keyword := p.parseToken(SyntaxKindTopNestedKeyword)
	if keyword == nil {
		return nil
	}
	expr := p.parseNamedExpression()
	ofKeyword := p.parseRequiredToken(SyntaxKindOfKeyword)
	ofExpr := p.parseNamedExpression()
	if ofExpr == nil {
		ofExpr = p.createMissingExpression()
	}
	withClause := p.parseTopNestedWithOthersClause()
	byKeyword := p.parseRequiredToken(SyntaxKindByKeyword)
	byExpr := p.parseOrderedExpression()
	if byExpr == nil {
		byExpr = p.createMissingExpression()
	}
	return MustNewTopNestedClauseTyped(keyword, expr, ofKeyword, ofExpr, withClause, byKeyword, byExpr)
}

func (p *queryParserState) createMissingTopNestedClause(diagnostics ...Diagnostic) *TopNestedClause {
	missingName := func() *NameReference {
		return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone)
	}
	allDiagnostics := append([]Diagnostic{DiagnosticMissingClause()}, diagnostics...)
	return MustNewTopNestedClauseTyped(
		createOmittedToken(SyntaxKindTopNestedKeyword),
		nil,
		createOmittedToken(SyntaxKindOfKeyword),
		missingName(),
		nil,
		createOmittedToken(SyntaxKindByKeyword),
		missingName(),
		allDiagnostics...,
	)
}

func (p *queryParserState) parseTopNestedClauseList() *SeparatedSyntaxListOf[*TopNestedClause] {
	return parseCommaList(
		p,
		"TopNestedClause",
		p.parseTopNestedClause,
		func() *TopNestedClause { return p.createMissingTopNestedClause() },
		p.atCommonListEnd,
		true,
		false,
	)
}

func (p *queryParserState) parseTopNestedOperator() *TopNestedOperator {
	if p.peekToken().Kind != SyntaxKindTopNestedKeyword {
		return nil
	}
	return MustNewTopNestedOperatorTyped(p.parseTopNestedClauseList())
}

func (p *queryParserState) parseUnionOperator() *UnionOperator {
	keyword := p.parseToken(SyntaxKindUnionKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.UnionParameters, true)
	expressions := p.parseUnionExpressionCommaList(true)
	return MustNewUnionOperatorTyped(keyword, parameters, expressions)
}

func (p *queryParserState) parseEntityGroupReference() Expression {
	if expr := p.parseEntityGroup(); expr != nil {
		return expr
	}
	if p.scanMacroExpandEntityPathStart() {
		return p.parseEntityPathExpression()
	}
	if reference := p.parseNameReference(); reference != nil {
		return reference
	}
	return nil
}

func (p *queryParserState) scanMacroExpandEntityPathStart() bool {
	token := p.peekToken()
	if token.Kind == SyntaxKindDatabaseKeyword {
		return p.scanFunctionCallStart()
	}
	if token.Kind == SyntaxKindIdentifierToken && strings.EqualFold(token.Text, "cluster") {
		return p.scanFunctionCallStart()
	}
	return false
}

func (p *queryParserState) parseMacroExpandEntityGroupReference() Expression {
	if expr := p.parseEntityGroupReference(); expr != nil {
		return expr
	}
	return p.createMissingExpression()
}

func (p *queryParserState) parseMacroExpandScopeReferenceName() *MacroExpandScopeReferenceName {
	asKeyword := p.parseToken(SyntaxKindAsKeyword)
	if asKeyword == nil {
		return nil
	}
	name := p.parseNameDeclaration()
	if name == nil {
		name = p.createMissingNameDeclaration()
	}
	return MustNewMacroExpandScopeReferenceNameTyped(asKeyword, name)
}

func (p *queryParserState) parseMacroExpandOperator() *MacroExpandOperator {
	keyword := p.parseToken(SyntaxKindMacroExpandKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.UnionParameters, true)
	entityGroup := p.parseMacroExpandEntityGroupReference()
	scopeReferenceName := p.parseMacroExpandScopeReferenceName()
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	statementList := AsSeparatedSyntaxListOf[Statement](p.parseQueryBlockStatementList())
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewMacroExpandOperatorTyped(keyword, parameters, entityGroup, scopeReferenceName, open, statementList, close)
}

func (p *queryParserState) parsePrintOperator() *PrintOperator {
	keyword := p.parseToken(SyntaxKindPrintKeyword)
	if keyword == nil {
		return nil
	}
	expressions := p.parseNamedExpressionCommaList(true)
	return MustNewPrintOperatorTyped(keyword, expressions)
}

func (p *queryParserState) parseBadQueryOperator() *BadQueryOperator {
	keyword := p.parseToken(SyntaxKindIdentifierToken)
	if keyword == nil {
		return nil
	}
	return MustNewBadQueryOperatorTyped(keyword, DiagnosticQueryOperatorExpected())
}

func (p *queryParserState) createMissingQueryOperator() QueryOperator {
	return MustNewBadQueryOperatorTyped(createOmittedToken(SyntaxKindIdentifierToken), DiagnosticQueryOperatorExpected())
}

func (p *queryParserState) createMissingQueryOperatorExpression() Expression {
	return MustNewBadQueryOperatorTyped(createOmittedToken(SyntaxKindIdentifierToken), DiagnosticQueryOperatorExpected())
}

func (p *queryParserState) emptyNamedParameterList() *SyntaxListOf[*NamedParameter] {
	return AsSyntaxListOf[*NamedParameter](NewTypedSyntaxList("NamedParameter"))
}

func (p *queryParserState) parseOrderingNullsClause() *OrderingNullsClause {
	keyword := p.parseToken(SyntaxKindNullsKeyword)
	if keyword == nil {
		return nil
	}
	firstOrLast := p.parseToken(SyntaxKindFirstKeyword, SyntaxKindLastKeyword)
	if firstOrLast == nil {
		firstOrLast = NewMissingToken(SyntaxKindFirstKeyword, DiagnosticMissingFirstOrLast())
	}
	return MustNewOrderingNullsClauseTyped(keyword, firstOrLast)
}

func (p *queryParserState) parseRequiredOrderingClause() *OrderingClause {
	keyword := p.parseToken(SyntaxKindAscKeyword, SyntaxKindDescKeyword, SyntaxKindGrannyAscKeyword, SyntaxKindGrannyDescKeyword)
	nulls := p.parseOrderingNullsClause()
	return MustNewOrderingClauseTyped(keyword, nulls)
}

func (p *queryParserState) parseRequiredOrderingNoNullsClause() *OrderingClause {
	keyword := p.parseToken(SyntaxKindAscKeyword, SyntaxKindDescKeyword, SyntaxKindGrannyAscKeyword, SyntaxKindGrannyDescKeyword)
	return MustNewOrderingClauseTyped(keyword, nil)
}

func (p *queryParserState) parseOrderedExpression() Expression {
	expr := p.parseNamedExpression()
	if expr == nil {
		return nil
	}
	switch p.peekToken().Kind {
	case SyntaxKindAscKeyword, SyntaxKindDescKeyword, SyntaxKindGrannyAscKeyword, SyntaxKindGrannyDescKeyword, SyntaxKindNullsKeyword:
		return MustNewOrderedExpressionTyped(expr, p.parseRequiredOrderingClause())
	default:
		return expr
	}
}

func queryOperatorExpression(operator QueryOperator) Expression {
	if isNilSyntaxElement(operator) {
		return nil
	}
	if expr, ok := operator.(Expression); ok {
		return expr
	}
	return nil
}

func (p *queryParserState) parseNamedExpressionCommaList(oneOrMore bool) *SeparatedSyntaxListOf[Expression] {
	return p.parseNamedExpressionCommaListWithEnd(oneOrMore, p.atCommonListEnd)
}

func (p *queryParserState) parseNamedExpressionCommaListWithEnd(oneOrMore bool, atEnd func() bool) *SeparatedSyntaxListOf[Expression] {
	return p.parseExpressionCommaListWithEnd(oneOrMore, atEnd, p.parseNamedExpression)
}

func (p *queryParserState) parseDistinctExpressionCommaList(oneOrMore bool) *SeparatedSyntaxListOf[Expression] {
	return p.parseExpressionCommaListWithEnd(oneOrMore, p.atCommonListEnd, p.parseDistinctExpression)
}

func (p *queryParserState) parseOrderedExpressionCommaList(oneOrMore bool) *SeparatedSyntaxListOf[Expression] {
	return p.parseExpressionCommaListWithEnd(oneOrMore, p.atCommonListEnd, p.parseOrderedExpression)
}

func (p *queryParserState) parseProjectReorderExpressionCommaList(oneOrMore bool) *SeparatedSyntaxListOf[Expression] {
	return p.parseExpressionCommaListWithEnd(oneOrMore, p.atCommonListEnd, p.parseProjectReorderExpression)
}

func (p *queryParserState) parseSimpleOrWildcardedNameReferenceCommaList(oneOrMore bool) *SeparatedSyntaxListOf[Expression] {
	return p.parseExpressionCommaListWithEnd(oneOrMore, p.atCommonListEnd, p.parseSimpleOrWildcardedNameReferenceExpression)
}

func (p *queryParserState) parseUnionExpressionCommaList(oneOrMore bool) *SeparatedSyntaxListOf[Expression] {
	return p.parseExpressionCommaListWithEnd(oneOrMore, p.atCommonListEnd, p.parseUnionExpression)
}

func (p *queryParserState) parseFindOperandCommaList(oneOrMore bool) *SeparatedSyntaxListOf[Expression] {
	return p.parseExpressionCommaListWithEnd(oneOrMore, p.atCommonListEnd, p.parseFindOperand)
}

func (p *queryParserState) atCommonListEnd() bool {
	return p.scanCommonListEnd()
}

func (p *queryParserState) parseExpressionCommaListWithEnd(oneOrMore bool, atEnd func() bool, parseElement func() Expression) *SeparatedSyntaxListOf[Expression] {
	return parseCommaList(p, "Expression", parseElement, p.createMissingExpression, atEnd, oneOrMore, false)
}

func (p *queryParserState) parseDistinctExpression() Expression {
	if p.peekToken().Kind == SyntaxKindAsteriskToken {
		return MustNewStarExpressionTyped(p.parseToken())
	}
	return p.parseNamedExpression()
}

func (p *queryParserState) parseProjectReorderExpression() Expression {
	expr := p.parseSimpleOrWildcardedNameReferenceExpression()
	if expr == nil {
		return nil
	}
	switch p.peekToken().Kind {
	case SyntaxKindAscKeyword, SyntaxKindDescKeyword, SyntaxKindGrannyAscKeyword, SyntaxKindGrannyDescKeyword:
		return MustNewOrderedExpressionTyped(expr, p.parseRequiredOrderingNoNullsClause())
	default:
		return expr
	}
}

func (p *queryParserState) parseSimpleOrWildcardedNameReferenceExpression() Expression {
	wildcard := p.parseBareWildcardedNameReference()
	if wildcard != nil {
		return wildcard
	}
	reference := p.parseNameReference()
	if reference == nil {
		return nil
	}
	return reference
}

func (p *queryParserState) parseUnionExpression() Expression {
	if expr := p.parseParenthesizedExpression(); expr != nil {
		return expr
	}
	return p.parseUnionWildcardedEntityExpression()
}

func (p *queryParserState) parseFindOperand() Expression {
	start := p.pos

	aliasExpr := p.parseFindOperandNameWithOptionalAsOperator()
	aliasEnd := p.pos

	p.pos = start
	entityExpr := p.parseWildcardedEntityExpression()
	entityEnd := p.pos

	if aliasExpr == nil && entityExpr == nil {
		p.pos = start
		return nil
	}
	if entityExpr != nil && (aliasExpr == nil || entityEnd > aliasEnd) {
		p.pos = entityEnd
		return entityExpr
	}
	p.pos = aliasEnd
	return aliasExpr
}

func (p *queryParserState) parseWildcardedEntityExpression() Expression {
	if p.scanWildcardedName() > 0 {
		return p.parseWildcardedNameReference()
	}
	expr := p.parsePathElementSelectorOrFunctionCall()
	for expr != nil {
		switch p.peekToken().Kind {
		case SyntaxKindDotToken:
			dot := p.parseToken()
			var selector Expression
			if p.scanWildcardedName() > 0 {
				selector = p.parseWildcardedNameReference()
			} else {
				selector = p.parsePathElementSelectorOrFunctionCall()
			}
			if selector == nil {
				selector = p.createMissingNameReference()
			}
			expr = MustNewPathExpressionTyped(expr, dot, selector)
			if _, ok := selector.(*NameReference); ok && selector.Kind() == SyntaxKindNameReference {
				if nameRef, ok := selector.(*NameReference); ok && nameRef.Name().Kind() == SyntaxKindWildcardedName {
					return expr
				}
			}
		case SyntaxKindOpenBracketToken:
			selector := p.parseBracketedEntityNamePathElementSelector()
			if selector == nil {
				selector = p.parseBracketedExpression()
			}
			if selector == nil {
				return expr
			}
			expr = MustNewElementExpressionTyped(expr, selector)
		default:
			return expr
		}
	}
	return expr
}

func (p *queryParserState) parseUnionWildcardedEntityExpression() Expression {
	if p.scanWildcardedName() > 0 {
		return p.parseWildcardedNameReference()
	}
	expr := p.parseUnionPathElementSelectorOrFunctionCall()
	for expr != nil {
		switch p.peekToken().Kind {
		case SyntaxKindDotToken:
			dot := p.parseToken()
			var selector Expression
			if p.scanWildcardedName() > 0 {
				selector = p.parseWildcardedNameReference()
			} else {
				selector = p.parseUnionPathElementSelectorOrFunctionCall()
			}
			if selector == nil {
				selector = p.createMissingNameReference()
			}
			expr = MustNewPathExpressionTyped(expr, dot, selector)
			if _, ok := selector.(*NameReference); ok && selector.Kind() == SyntaxKindNameReference {
				if nameRef, ok := selector.(*NameReference); ok && nameRef.Name().Kind() == SyntaxKindWildcardedName {
					return expr
				}
			}
		case SyntaxKindOpenBracketToken:
			selector := p.parseBracketedEntityNamePathElementSelector()
			if selector == nil {
				selector = p.parseBracketedExpression()
			}
			if selector == nil {
				return expr
			}
			expr = MustNewElementExpressionTyped(expr, selector)
		default:
			return expr
		}
	}
	return expr
}

func (p *queryParserState) parseUnionPathElementSelectorOrFunctionCall() Expression {
	if p.scanFunctionCallStart() {
		return p.parseFunctionCallExpression()
	}
	return p.parsePathElementSelector()
}

func (p *queryParserState) parseWildcardedNameReference() *NameReference {
	if reference := p.parseBracketedWildcardedNameReference(); reference != nil {
		return reference
	}
	return p.parseBareWildcardedNameReference()
}

func (p *queryParserState) parseBracketedWildcardedNameReference() *NameReference {
	if p.peekToken().Kind != SyntaxKindOpenBracketToken || p.scanWildcardedName(1) <= 0 {
		return nil
	}
	open := p.parseToken()
	pattern := p.parseWildcardedIdentifier()
	close := p.parseRequiredToken(SyntaxKindCloseBracketToken)
	return MustNewNameReferenceTyped(MustNewBracketedWildcardedNameTyped(open, pattern, close), SymbolMatchNone)
}

func (p *queryParserState) parseBareWildcardedNameReference() *NameReference {
	pattern := p.parseWildcardedIdentifier()
	if pattern == nil {
		return nil
	}
	return MustNewNameReferenceTyped(MustNewWildcardedNameTyped(pattern), SymbolMatchNone)
}

func (p *queryParserState) parseWildcardedIdentifier() *SyntaxToken {
	length := p.scanWildcardedName()
	if length <= 0 {
		return nil
	}
	trivia := p.peekToken().Trivia
	rawText := ""
	valueText := ""
	for i := 0; i < length; i++ {
		token := p.peekToken(i)
		if i > 0 && p.options.AllowNonAdjacentWildcardParts {
			rawText += token.Trivia
		}
		rawText += token.Text
		valueText += token.Text
	}
	p.pos += length
	return &SyntaxToken{kind: SyntaxKindIdentifierToken, trivia: trivia, text: rawText, valueText: valueText}
}

func (p *queryParserState) scanTokenText(values []string, offset int) bool {
	token := p.peekToken(offset)
	for _, value := range values {
		if token.Text == value {
			return true
		}
	}
	return false
}

func (p *queryParserState) parseTokenText(values []string) *SyntaxToken {
	if !p.scanTokenText(values, 0) {
		return nil
	}
	return p.parseToken()
}

func (p *queryParserState) parseTokenTextOrDiagnostic(values []string) *SyntaxToken {
	if token := p.parseTokenText(values); token != nil {
		return token
	}
	token := p.peekToken()
	if token.Kind == SyntaxKindNone || token.Kind == SyntaxKindEndOfTextToken {
		return nil
	}
	p.pos++
	return SyntaxTokenFromLexical(token, DiagnosticTermsExpected(values...))
}

func (p *queryParserState) scanWildcardedName(offset ...int) int {
	start := 0
	if len(offset) > 0 {
		start = offset[0]
	}
	pos := start
	token := p.peekToken(pos)
	switch {
	case token.Kind == SyntaxKindAsteriskToken:
		pos++
	case p.isWildcardedNamePartToken(pos) &&
		p.peekToken(pos+1).Kind == SyntaxKindAsteriskToken &&
		(p.peekToken(pos+1).Trivia == "" || p.options.AllowNonAdjacentWildcardParts):
		pos += 2
	default:
		return -1
	}
	for {
		token = p.peekToken(pos)
		if token.Trivia != "" && !p.options.AllowNonAdjacentWildcardParts {
			break
		}
		if p.isWildcardedNamePartToken(pos) || token.Kind == SyntaxKindLongLiteralToken || token.Kind == SyntaxKindAsteriskToken {
			pos++
			continue
		}
		break
	}
	if pos > start {
		return pos - start
	}
	return -1
}

func (p *queryParserState) isWildcardedNamePartToken(offset int) bool {
	kind := p.peekToken(offset).Kind
	return kind == SyntaxKindIdentifierToken || p.isExtendedKeywordAsIdentifier(kind)
}
