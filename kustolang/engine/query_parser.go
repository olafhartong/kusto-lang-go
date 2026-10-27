package kustolang

import "strings"

// QueryParser builds a QueryBlock node, typed PipeExpression syntax, and
// top-level pipe-segment metadata.

type QueryBlock struct {
	*SyntaxNode
	Segments []PipeSegment
}

type PipeSegment struct {
	StartToken int
	EndToken   int
	Text       string
}

func (q *QueryBlock) Clone(includeDiagnostics bool) SyntaxElement {
	return q.CloneTyped(includeDiagnostics)
}

func (q *QueryBlock) CloneTyped(includeDiagnostics bool) *QueryBlock {
	if q == nil {
		return nil
	}
	var clonedNode *SyntaxNode
	if q.SyntaxNode != nil {
		clonedNode, _ = q.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	}
	return &QueryBlock{
		SyntaxNode: clonedNode,
		Segments:   append([]PipeSegment{}, q.Segments...),
	}
}

func ParseQuery(text string) *QueryBlock {
	return ParseQueryWithOptions(text, DefaultParseOptions)
}

func ParseQueryWithOptions(text string, options ParseOptions) *QueryBlock {
	return parseQueryWithParser(newQueryParserStateWithOwnedTokens(LexWithOptions(text, options), options, CancellationToken{}))
}

func ParseQueryTokens(tokens []LexicalToken) *QueryBlock {
	return ParseQueryTokensFromWithOptions(tokens, 0, DefaultParseOptions)
}

func ParseQueryTokensWithOptions(tokens []LexicalToken, options ParseOptions) *QueryBlock {
	return ParseQueryTokensFromWithOptions(tokens, 0, options)
}

func ParseQueryTokensWithOptionsAndCancellation(tokens []LexicalToken, options ParseOptions, cancellationToken CancellationToken) *QueryBlock {
	return parseQueryTokensFromWithOptionsAndCancellation(tokens, 0, options, cancellationToken)
}

func ParseQueryTokensFrom(tokens []LexicalToken, start int) *QueryBlock {
	return ParseQueryTokensFromWithOptions(tokens, start, DefaultParseOptions)
}

func ParseQueryTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) *QueryBlock {
	return parseQueryTokensFromWithOptionsAndCancellation(tokens, start, options, CancellationToken{})
}

func parseQueryTokensFromWithOptionsAndCancellation(tokens []LexicalToken, start int, options ParseOptions, cancellationToken CancellationToken) *QueryBlock {
	cancellationToken.ThrowIfCancellationRequested()
	tokens = lexicalTokensFrom(tokens, start)
	return parseQueryWithParser(newQueryParserStateWithOptionsAndCancellation(tokens, options, cancellationToken))
}

func parseQueryWithParser(parser *queryParserState) *QueryBlock {
	tokens, options, cancellationToken := parser.tokens, parser.options, parser.cancellation
	directives := parser.parseDirectiveList()
	// Upstream ParseQuery consumes directives once, before its statement list.
	// Recovery must not replay them, which duplicates text and shifts locations.
	statementTokens := tokens[parser.pos:]
	cancellationToken.ThrowIfCancellationRequested()
	statements := parser.parseQueryBlockStatementList()
	cancellationToken.ThrowIfCancellationRequested()
	var skipped *SkippedTokens
	if statements.Count() == 0 && len(statementTokens) > 0 {
		statements = shallowQueryStatementListWithCancellation(statementTokens, options, cancellationToken)
	} else if !parser.atEndOfQuery() && !parser.shouldReportQuerySkippedTokens(statements) && !queryStatementsHaveSourceTailBoundaryOperator(statements) {
		statements = shallowQueryStatementListWithCancellation(statementTokens, options, cancellationToken, containedDiagnosticsWithCode(statements, "KS176")...)
	} else {
		skipped = parser.parseQuerySkippedTokens()
	}
	endOfQuery := parser.parseToken(SyntaxKindEndOfTextToken)
	if endOfQuery == nil {
		for _, token := range tokens {
			if token.Kind == SyntaxKindEndOfTextToken {
				endOfQuery = SyntaxTokenFromLexical(token)
				break
			}
		}
	}
	block := MustNewQueryBlockTyped(
		directives,
		AsSeparatedSyntaxListOf[Statement](statements),
		skipped,
		endOfQuery,
	)
	block.Segments = pipeSegmentsWithCancellation(tokens, cancellationToken)
	cancellationToken.ThrowIfCancellationRequested()
	return block
}

func queryStatementsHaveSourceTailBoundaryOperator(statements SyntaxElement) bool {
	found := false
	WalkSyntax(statements, func(element SyntaxElement) {
		if found {
			return
		}
		switch syntax := WrapGeneratedSyntaxElement(element).(type) {
		case *AsOperator, *ConsumeOperator, *CountOperator, *DistinctOperator, *ExecuteAndCacheOperator, *FilterOperator, *FindOperator, *GetSchemaOperator,
			*GraphMarkComponentsOperator, *GraphMatchOperator, *GraphShortestPathsOperator, *GraphWhereEdgesOperator, *GraphWhereNodesOperator,
			*InvokeOperator, *JoinOperator, *LookupOperator, *MvApplyOperator, *MvExpandOperator, *RangeOperator,
			*MakeSeriesOperator, *PartitionOperator, *RenderOperator,
			*SampleOperator, *SampleDistinctOperator, *ScanOperator, *SearchOperator, *SerializeOperator, *SummarizeOperator, *TakeOperator,
			*TopHittersOperator, *TopNestedOperator,
			*UnionOperator,
			*ProjectAwayOperator, *ProjectKeepOperator, *ProjectRenameOperator, *ProjectReorderOperator,
			*ParseKvOperator, *AssertSchemaOperator, *MaterializedViewCombineExpression, *ToTableExpression:
			found = true
		case *DataTableExpression:
			found = true
		case *ElementExpression:
			if isDatabaseSourceElementBaseSyntax(syntax.Expression()) {
				found = true
			}
		}
		if reduce, ok := WrapGeneratedSyntaxElement(element).(*ReduceByOperator); ok && reduce.With() == nil {
			found = true
		}
	})
	if found {
		return true
	}
	return queryStatementsHaveTopWithTailStatement(statements)
}

func isDatabaseSourceElementBaseSyntax(expression Expression) bool {
	call, ok := expression.(*FunctionCallExpression)
	return ok && call.Name() != nil && ordinalEqualFold(NameReferenceSimpleName(call.Name()), "database")
}

func queryStatementsHaveTopWithTailStatement(statements SyntaxElement) bool {
	hasTop := false
	hasWith := false
	WalkSyntax(statements, func(element SyntaxElement) {
		switch syntax := WrapGeneratedSyntaxElement(element).(type) {
		case *TopOperator:
			hasTop = true
		case *NameReference:
			if name, ok := syntax.Name().(*TokenName); ok && name.Name() != nil {
				hasWith = hasWith || name.Name().RawText() == "with"
			}
		}
	})
	return hasTop && hasWith
}

func (p *queryParserState) parseDirectiveList() *SyntaxListOf[*Directive] {
	var directives []SyntaxElement
	for p.peekToken().Kind == SyntaxKindDirectiveToken {
		token := p.parseToken(SyntaxKindDirectiveToken)
		directives = append(directives, MustNewDirectiveTyped(token))
	}
	return AsSyntaxListOf[*Directive](NewTypedSyntaxList("Directive", directives...))
}

func (p *queryParserState) shouldReportQuerySkippedTokens(statements SyntaxElement) bool {
	if queryStatementsHaveQueryParametersStatement(statements) {
		return true
	}
	if p.peekToken().Kind == SyntaxKindToKeyword && queryStatementsStartWithNameExpression(statements, "restrict") {
		return true
	}
	if queryStatementsStartWithMalformedQueryParametersDeclaration(statements) {
		return true
	}
	if queryStatementsStartWithDetachedOperatorName(statements) {
		// Current QueryParser keeps the name statements recovered from detached
		// distinct/parse operators and assigns only their remaining tail to
		// SkippedTokens. The opaque fallback loses the intervening diagnostics.
		return true
	}
	if p.shouldPreserveCurrentRangeElementRecovery(statements) {
		return true
	}
	if queryStatementsHaveMissingTypeOf(statements) && startsLikelyTypeName(p.peekToken().Kind) {
		return true
	}
	if queryStatementsHaveMissingSchemaDeclaration(statements) && p.startsLikelySchemaDeclaration(p.peekToken().Kind) {
		return true
	}
	if queryStatementsHaveMissingFindInClauseClose(statements) && p.peekToken().Kind == SyntaxKindBarToken {
		return true
	}
	if queryStatementsHaveMissingMakeSeriesInRangeClause(statements) {
		return true
	}
	if p.peekToken().Kind == SyntaxKindToKeyword && queryStatementsHaveMakeSeriesOperator(statements) {
		return true
	}
	if p.peekToken().Kind == SyntaxKindToKeyword && queryStatementsStartWithPrintOperator(statements) {
		return true
	}
	if queryStatementsHaveMissingMakeGraphDirection(statements) {
		return true
	}
	if p.peekToken().Kind == SyntaxKindDynamicKeyword && queryStatementsHaveDataTableOrSearchRoot(statements) {
		return true
	}
	if isExpressionInfixKeyword(p.peekToken().Kind) {
		return true
	}
	switch p.peekToken().Kind {
	case SyntaxKindCommaToken:
		if queryStatementsStartWithSyntacticallyInvalidLetStatement(statements) {
			return true
		}
		if queryStatementsHaveRestrictStatement(statements) {
			return true
		}
		switch p.peekToken(1).Kind {
		case SyntaxKindEndOfTextToken, SyntaxKindNone, SyntaxKindBarToken, SyntaxKindSemicolonToken:
			return true
		default:
			return queryStatementsHaveSingleExpressionCommaTailOperator(statements) ||
				queryStatementsHaveOperatorOwnedCommaTail(statements) ||
				queryStatementsHaveSourceExpressionCommaTail(statements) ||
				queryStatementsHaveEvaluateOperator(statements) ||
				queryStatementsHaveMissingDotDot(statements) ||
				queryStatementsHaveMissingExpressionCoupleClose(statements) ||
				queryStatementsHaveMissingMaterializedViewCombineNameClose(statements) ||
				queryStatementsHaveInlineExternalTableExpression(statements) ||
				queryStatementsHaveDeprecatedRenderPropertyCommaTail(statements)
		}
	case SyntaxKindEqualToken,
		SyntaxKindEqualEqualToken,
		SyntaxKindEqualTildeToken,
		SyntaxKindBangEqualToken,
		SyntaxKindBangTildeToken,
		SyntaxKindLessThanToken,
		SyntaxKindLessThanOrEqualToken,
		SyntaxKindGreaterThanToken,
		SyntaxKindGreaterThanOrEqualToken,
		SyntaxKindLessThanGreaterThanToken,
		SyntaxKindFatArrowToken,
		SyntaxKindBangToken,
		SyntaxKindColonToken,
		SyntaxKindDotToken,
		SyntaxKindDotDotToken,
		SyntaxKindAsteriskToken,
		SyntaxKindSlashToken,
		SyntaxKindPercentToken,
		SyntaxKindQuestionToken,
		SyntaxKindOpenBraceToken,
		SyntaxKindCloseParenToken,
		SyntaxKindCloseBracketToken,
		SyntaxKindCloseBraceToken,
		SyntaxKindDirectiveToken,
		SyntaxKindAscKeyword,
		SyntaxKindByKeyword,
		SyntaxKindDescKeyword,
		SyntaxKindFirstKeyword,
		SyntaxKindGrannyAscKeyword,
		SyntaxKindGrannyDescKeyword,
		SyntaxKindLastKeyword,
		SyntaxKindNullsKeyword,
		SyntaxKindOfKeyword,
		SyntaxKindHintDotConcurrencyKeyword,
		SyntaxKindHintDotDistributionKeyword,
		SyntaxKindHintDotMaterializedKeyword,
		SyntaxKindHintDotNumPartitions,
		SyntaxKindHintDotProgressiveTopKeyword,
		SyntaxKindHintDotRemoteKeyword,
		SyntaxKindHintDotShuffleKeyKeyword,
		SyntaxKindHintDotSpreadKeyword,
		SyntaxKindHintDotStrategyKeyword,
		SyntaxKindBadToken:
		return true
	default:
		return false
	}
}

func queryStatementsHaveQueryParametersStatement(statements SyntaxElement) bool {
	found := false
	WalkSyntax(statements, func(element SyntaxElement) {
		if !found {
			_, found = WrapGeneratedSyntaxElement(element).(*QueryParametersStatement)
		}
	})
	return found
}

func queryStatementsStartWithNameExpression(statements SyntaxElement, name string) bool {
	list, ok := statements.(*SyntaxList)
	if !ok || list.Count() == 0 {
		return false
	}
	separated, ok := list.At(0).(*SeparatedElement)
	if !ok || separated.Element() == nil {
		return false
	}
	statement, ok := WrapGeneratedSyntaxElement(separated.Element()).(*ExpressionStatement)
	if !ok {
		return false
	}
	reference, ok := statement.Expression().(*NameReference)
	return ok && ordinalEqualFold(NameReferenceSimpleName(reference), name)
}

func queryStatementsStartWithMalformedQueryParametersDeclaration(statements SyntaxElement) bool {
	list, ok := statements.(*SyntaxList)
	if !ok || list.Count() != 2 || !queryStatementsStartWithNameExpression(statements, "declare") {
		return false
	}
	separated, ok := list.At(1).(*SeparatedElement)
	if !ok || separated.Element() == nil {
		return false
	}
	statement, ok := WrapGeneratedSyntaxElement(separated.Element()).(*ExpressionStatement)
	if !ok {
		return false
	}
	found := false
	WalkSyntax(statement.Expression(), func(element SyntaxElement) {
		if found {
			return
		}
		reference, ok := WrapGeneratedSyntaxElement(element).(*NameReference)
		found = ok && strings.HasPrefix(strings.ToLower(NameReferenceSimpleName(reference)), "query_parameters")
	})
	return found
}

func queryStatementsStartWithSyntacticallyInvalidLetStatement(statements SyntaxElement) bool {
	list, ok := statements.(*SyntaxList)
	if !ok || list.Count() == 0 {
		return false
	}
	separated, ok := list.At(0).(*SeparatedElement)
	if !ok || separated.Element() == nil {
		return false
	}
	_, ok = WrapGeneratedSyntaxElement(separated.Element()).(*LetStatement)
	return ok && separated.Element().ContainsSyntaxDiagnostics()
}

func queryStatementsStartWithDetachedOperatorName(statements SyntaxElement) bool {
	list, ok := WrapGeneratedSyntaxElement(statements).(*SyntaxList)
	if !ok || list.Count() < 2 {
		return false
	}
	for i := 0; i < list.Count(); i++ {
		separated, ok := list.At(i).(*SeparatedElement)
		if !ok || separated.Element() == nil {
			continue
		}
		statement, ok := WrapGeneratedSyntaxElement(separated.Element()).(*ExpressionStatement)
		if !ok {
			continue
		}
		reference, ok := statement.Expression().(*NameReference)
		if !ok {
			continue
		}
		name, ok := reference.Name().(*TokenName)
		if !ok || name.Name() == nil {
			continue
		}
		switch name.Name().Kind() {
		case SyntaxKindDistinctKeyword, SyntaxKindParseKeyword:
			return true
		}
	}
	return false
}

func queryStatementsHaveDataTableOrSearchRoot(statements SyntaxElement) bool {
	list, ok := statements.(*SyntaxList)
	if !ok || list.Count() != 1 {
		return false
	}
	separated, ok := list.At(0).(*SeparatedElement)
	if !ok || separated.Element() == nil {
		return false
	}
	statement, ok := WrapGeneratedSyntaxElement(separated.Element()).(*ExpressionStatement)
	if !ok || statement.Expression() == nil {
		return false
	}
	switch expression := statement.Expression().(type) {
	case *DataTableExpression:
		return true
	case *BinaryExpression:
		return expression.Kind() == SyntaxKindSearchExpression
	default:
		return false
	}
}

func (p *queryParserState) shouldPreserveCurrentRangeElementRecovery(statements SyntaxElement) bool {
	if p == nil || p.peekToken().Kind != SyntaxKindToKeyword {
		return false
	}
	list, ok := statements.(*SyntaxList)
	if !ok || list.Count() != 3 {
		return false
	}

	expressionAt := func(index int) (Expression, *SeparatedElement, bool) {
		separated, ok := list.At(index).(*SeparatedElement)
		if !ok || separated.Element() == nil {
			return nil, nil, false
		}
		statement, ok := WrapGeneratedSyntaxElement(separated.Element()).(*ExpressionStatement)
		if !ok || statement.Expression() == nil {
			return nil, nil, false
		}
		return statement.Expression(), separated, true
	}

	first, firstSeparated, ok := expressionAt(0)
	if !ok {
		return false
	}
	element, ok := first.(*ElementExpression)
	if !ok {
		return false
	}
	rangeName, ok := element.Expression().(*NameReference)
	if !ok || !ordinalEqualFold(NameReferenceSimpleName(rangeName), "range") {
		return false
	}
	selector, ok := element.Selector().(*BracketedExpression)
	if !ok || selector.CloseBracket() == nil || selector.CloseBracket().RawText() != "" ||
		selector.Expression() == nil || selector.Expression().Kind() != SyntaxKindStringLiteralExpression {
		return false
	}
	if separator := firstSeparated.Separator(); separator == nil || separator.Kind() != SyntaxKindSemicolonToken || separator.RawText() != "" {
		return false
	}

	second, secondSeparated, ok := expressionAt(1)
	if !ok {
		return false
	}
	fromName, ok := second.(*NameReference)
	if !ok || !ordinalEqualFold(NameReferenceSimpleName(fromName), "from") {
		return false
	}
	if separator := secondSeparated.Separator(); separator == nil || separator.Kind() != SyntaxKindSemicolonToken || separator.RawText() != "" {
		return false
	}

	third, thirdSeparated, ok := expressionAt(2)
	return ok && third.Kind() == SyntaxKindLongLiteralExpression && thirdSeparated.Separator() == nil
}

func queryStatementsStartWithPrintOperator(statements SyntaxElement) bool {
	list, ok := statements.(*SyntaxList)
	if !ok || list.Count() == 0 {
		return false
	}
	separated, ok := list.At(0).(*SeparatedElement)
	if !ok || separated.Element() == nil {
		return false
	}
	statement, ok := WrapGeneratedSyntaxElement(separated.Element()).(*ExpressionStatement)
	if !ok || statement.Expression() == nil {
		return false
	}
	_, ok = statement.Expression().(*PrintOperator)
	return ok
}

func queryStatementsHaveMakeSeriesOperator(statements SyntaxElement) bool {
	found := false
	WalkSyntax(statements, func(element SyntaxElement) {
		if !found {
			_, found = WrapGeneratedSyntaxElement(element).(*MakeSeriesOperator)
		}
	})
	return found
}

func queryStatementsHaveRestrictStatement(statements SyntaxElement) bool {
	found := false
	WalkSyntax(statements, func(element SyntaxElement) {
		if !found {
			_, found = WrapGeneratedSyntaxElement(element).(*RestrictStatement)
		}
	})
	return found
}

func queryStatementsHaveOperatorOwnedCommaTail(statements SyntaxElement) bool {
	found := false
	WalkSyntax(statements, func(element SyntaxElement) {
		if found {
			return
		}
		switch operator := WrapGeneratedSyntaxElement(element).(type) {
		case *AsOperator, *ConsumeOperator, *CountOperator, *ExecuteAndCacheOperator, *FilterOperator, *GraphMarkComponentsOperator,
			*FindOperator, *GetSchemaOperator, *JoinOperator, *MvApplyOperator,
			*MacroExpandOperator, *PartitionOperator, *PartitionByOperator, *ProjectOperator, *TakeOperator:
			found = true
		case *RenderOperator:
			found = operator.WithClause() != nil && operator.WithClause().CloseParen() != nil && operator.WithClause().CloseParen().RawText() != ""
		case *ReduceByOperator:
			found = operator.With() == nil
		}
	})
	return found
}

func queryStatementsHaveEvaluateOperator(statements SyntaxElement) bool {
	found := false
	WalkSyntax(statements, func(element SyntaxElement) {
		if found {
			return
		}
		_, found = WrapGeneratedSyntaxElement(element).(*EvaluateOperator)
	})
	return found
}

func queryStatementsHaveMissingMakeGraphDirection(statements SyntaxElement) bool {
	found := false
	WalkSyntax(statements, func(element SyntaxElement) {
		if found {
			return
		}
		makeGraph, ok := WrapGeneratedSyntaxElement(element).(*MakeGraphOperator)
		found = ok && makeGraph.DirectionToken() != nil && makeGraph.DirectionToken().ByteWidth() == 0
	})
	return found
}

func (p *queryParserState) startsLikelySchemaDeclaration(kind SyntaxKind) bool {
	return kind == SyntaxKindAsteriskToken ||
		kind == SyntaxKindIdentifierToken ||
		kind == SyntaxKindOpenBracketToken ||
		kind == SyntaxKindOpenBraceToken ||
		p.isExtendedKeywordAsIdentifier(kind)
}

func isExpressionInfixKeyword(kind SyntaxKind) bool {
	switch kind {
	case SyntaxKindAndKeyword,
		SyntaxKindOrKeyword,
		SyntaxKindHasKeyword,
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
		SyntaxKindInKeyword,
		SyntaxKindInCsKeyword,
		SyntaxKindNotInKeyword,
		SyntaxKindNotInCsKeyword,
		SyntaxKindHasAnyKeyword,
		SyntaxKindHasAllKeyword,
		SyntaxKindBetweenKeyword,
		SyntaxKindNotBetweenKeyword:
		return true
	default:
		return false
	}
}

func queryStatementsHaveSingleExpressionCommaTailOperator(statements SyntaxElement) bool {
	found := false
	WalkSyntax(statements, func(element SyntaxElement) {
		if found {
			return
		}
		switch WrapGeneratedSyntaxElement(element).(type) {
		case *CountOperator, *FacetOperator, *ForkOperator, *GraphMatchOperator, *GraphShortestPathsOperator, *GraphWhereEdgesOperator, *GraphWhereNodesOperator, *InvokeOperator, *MakeSeriesOperator, *ParseKvOperator, *ParseOperator, *ParseWhereOperator, *RangeOperator, *SampleOperator, *SampleDistinctOperator, *ScanOperator, *SearchOperator, *TakeOperator, *TopOperator, *TopHittersOperator:
			found = true
		}
	})
	return found
}

func queryStatementsHaveSourceExpressionCommaTail(statements SyntaxElement) bool {
	list, ok := statements.(*SyntaxList)
	if !ok || list.Count() != 1 {
		return false
	}
	separated, ok := list.At(0).(*SeparatedElement)
	if !ok || separated.Element() == nil {
		return false
	}
	statement, ok := WrapGeneratedSyntaxElement(separated.Element()).(*ExpressionStatement)
	if !ok || statement.Expression() == nil {
		return false
	}
	switch expression := statement.Expression().(type) {
	case *NameReference, *PathExpression, *FunctionCallExpression, *DataTableExpression, *ExternalDataExpression, *ContextualDataTableExpression:
		return true
	case *BinaryExpression:
		return expression.Kind() == SyntaxKindSearchExpression
	default:
		return false
	}
}

func queryStatementsHaveInlineExternalTableExpression(statements SyntaxElement) bool {
	found := false
	WalkSyntax(statements, func(element SyntaxElement) {
		if found {
			return
		}
		_, found = WrapGeneratedSyntaxElement(element).(*InlineExternalTableExpression)
	})
	return found
}

func queryStatementsHaveDeprecatedRenderPropertyCommaTail(statements SyntaxElement) bool {
	found := false
	WalkSyntax(statements, func(element SyntaxElement) {
		if found {
			return
		}
		render, ok := WrapGeneratedSyntaxElement(element).(*RenderOperator)
		found = ok && render.WithClause() == nil
	})
	return found
}

func queryStatementsHaveMissingDotDot(statements SyntaxElement) bool {
	for _, diagnostic := range containedDiagnosticsWithCode(statements, "KS005") {
		if diagnostic.Message == "Expected: .." {
			return true
		}
	}
	return false
}

func queryStatementsHaveMissingExpressionCoupleClose(statements SyntaxElement) (found bool) {
	defer func() {
		if recover() != nil {
			found = false
		}
	}()
	WalkSyntax(statements, func(element SyntaxElement) {
		if found {
			return
		}
		couple, ok := WrapGeneratedSyntaxElement(element).(*ExpressionCouple)
		if !ok || couple.CloseParen() == nil {
			return
		}
		found = couple.CloseParen().Kind() == SyntaxKindCloseParenToken && couple.CloseParen().RawText() == ""
	})
	return found
}

func queryStatementsHaveMissingMaterializedViewCombineNameClose(statements SyntaxElement) (found bool) {
	defer func() {
		if recover() != nil {
			found = false
		}
	}()
	WalkSyntax(statements, func(element SyntaxElement) {
		if found {
			return
		}
		nameClause, ok := WrapGeneratedSyntaxElement(element).(*MaterializedViewCombineNameClause)
		if !ok || nameClause.CloseParen() == nil {
			return
		}
		found = nameClause.CloseParen().Kind() == SyntaxKindCloseParenToken && nameClause.CloseParen().RawText() == ""
	})
	return found
}

func queryStatementsHaveMissingSchemaDeclaration(statements SyntaxElement) bool {
	for _, diagnostic := range containedDiagnosticsWithCode(statements, "KS006") {
		if diagnostic.Message == "Missing schema declaration" {
			return true
		}
	}
	return false
}

func queryStatementsHaveMissingMakeSeriesInRangeClause(statements SyntaxElement) bool {
	hasMissingIn := false
	hasMissingRange := false
	hasMissingOpenParen := false
	for _, diagnostic := range containedDiagnosticsWithCode(statements, "KS005") {
		switch diagnostic.Message {
		case "Expected: in":
			hasMissingIn = true
		case "Expected: range":
			hasMissingRange = true
		case "Expected: (":
			hasMissingOpenParen = true
		}
	}
	return hasMissingIn && hasMissingRange && hasMissingOpenParen
}

func queryStatementsHaveMissingFindInClauseClose(statements SyntaxElement) (found bool) {
	defer func() {
		if recover() != nil {
			found = false
		}
	}()
	WalkSyntax(statements, func(element SyntaxElement) {
		if found {
			return
		}
		clause, ok := WrapGeneratedSyntaxElement(element).(*FindInClause)
		if !ok || clause.CloseParen() == nil {
			return
		}
		found = clause.CloseParen().Kind() == SyntaxKindCloseParenToken && clause.CloseParen().RawText() == ""
	})
	return found
}

func queryStatementsHaveMissingTypeOf(statements SyntaxElement) bool {
	for _, diagnostic := range containedDiagnosticsWithCode(statements, "KS006") {
		if diagnostic.Message == "Missing typeof" {
			return true
		}
	}
	return false
}

func startsLikelyTypeName(kind SyntaxKind) bool {
	switch kind {
	case SyntaxKindBoolKeyword,
		SyntaxKindBooleanKeyword,
		SyntaxKindByteKeyword,
		SyntaxKindCharKeyword,
		SyntaxKindDateKeyword,
		SyntaxKindDateTimeKeyword,
		SyntaxKindDecimalKeyword,
		SyntaxKindDoubleKeyword,
		SyntaxKindDynamicKeyword,
		SyntaxKindFloatKeyword,
		SyntaxKindGuidKeyword,
		SyntaxKindInt8Keyword,
		SyntaxKindInt16Keyword,
		SyntaxKindIntKeyword,
		SyntaxKindInt32Keyword,
		SyntaxKindInt64Keyword,
		SyntaxKindLongKeyword,
		SyntaxKindRealKeyword,
		SyntaxKindStringKeyword,
		SyntaxKindTimeKeyword,
		SyntaxKindTimespanKeyword,
		SyntaxKindUInt8Keyword,
		SyntaxKindUInt16Keyword,
		SyntaxKindUIntKeyword,
		SyntaxKindUInt32Keyword,
		SyntaxKindUInt64Keyword,
		SyntaxKindULongKeyword,
		SyntaxKindUniqueIdKeyword:
		return true
	default:
		return false
	}
}

func (p *queryParserState) parseQuerySkippedTokens() *SkippedTokens {
	var tokens []SyntaxElement
	for {
		kind := p.peekToken().Kind
		if kind == SyntaxKindEndOfTextToken || kind == SyntaxKindNone {
			break
		}
		token := p.peekToken()
		p.pos++
		parts := []LexicalToken{token}
		if expanded, ok := expandGraphPunctuationToken(token); ok {
			parts = expanded
		}
		for _, part := range parts {
			if len(tokens) == 0 {
				tokens = append(tokens, SyntaxTokenFromLexical(part, DiagnosticIncompleteFragment()))
			} else {
				tokens = append(tokens, SyntaxTokenFromLexical(part))
			}
		}
	}
	if len(tokens) == 0 {
		return nil
	}
	return MustNewSkippedTokensTyped(AsSyntaxListOf[*SyntaxToken](NewTypedSyntaxList("SyntaxToken", tokens...)))
}

func expandGraphPunctuationToken(token LexicalToken) ([]LexicalToken, bool) {
	var kinds []SyntaxKind
	switch token.Kind {
	case SyntaxKindDashDashToken:
		kinds = []SyntaxKind{SyntaxKindMinusToken, SyntaxKindMinusToken}
	case SyntaxKindDashDashGreaterThanToken:
		kinds = []SyntaxKind{SyntaxKindMinusToken, SyntaxKindMinusToken, SyntaxKindGreaterThanToken}
	case SyntaxKindLessThanDashDashToken:
		kinds = []SyntaxKind{SyntaxKindLessThanToken, SyntaxKindMinusToken, SyntaxKindMinusToken}
	case SyntaxKindBracketDashToken:
		kinds = []SyntaxKind{SyntaxKindCloseBracketToken, SyntaxKindMinusToken}
	case SyntaxKindBracketDashGreaterThanToken:
		kinds = []SyntaxKind{SyntaxKindCloseBracketToken, SyntaxKindMinusToken, SyntaxKindGreaterThanToken}
	case SyntaxKindDashBracketToken:
		kinds = []SyntaxKind{SyntaxKindMinusToken, SyntaxKindOpenBracketToken}
	case SyntaxKindLessThanDashBracketToken:
		kinds = []SyntaxKind{SyntaxKindLessThanToken, SyntaxKindMinusToken, SyntaxKindOpenBracketToken}
	default:
		return nil, false
	}

	parts := make([]LexicalToken, 0, len(kinds))
	for i, kind := range kinds {
		trivia := ""
		var diagnostics []Diagnostic
		if i == 0 {
			trivia = token.Trivia
			diagnostics = token.Diagnostics
		}
		parts = append(parts, NewLexicalToken(kind, trivia, SyntaxText(kind), diagnostics...))
	}
	return parts, true
}

func (p *queryParserState) atEndOfQuery() bool {
	return p.peekToken().Kind == SyntaxKindEndOfTextToken || p.peekToken().Kind == SyntaxKindNone
}

func containedDiagnosticsWithCode(element SyntaxElement, code string) (out []Diagnostic) {
	defer func() {
		if recover() != nil {
			out = nil
		}
	}()
	if element == nil {
		return nil
	}
	return diagnosticsWithCode(element.ContainedDiagnostics(), code)
}

func diagnosticsWithCode(diagnostics []Diagnostic, code string) []Diagnostic {
	var out []Diagnostic
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			out = append(out, diagnostic)
		}
	}
	return out
}

func shallowQueryStatementList(tokens []LexicalToken, options ParseOptions, diagnostics ...Diagnostic) *SyntaxList {
	return shallowQueryStatementListWithCancellation(tokens, options, CancellationToken{}, diagnostics...)
}

func shallowQueryStatementListWithCancellation(tokens []LexicalToken, options ParseOptions, cancellationToken CancellationToken, diagnostics ...Diagnostic) *SyntaxList {
	parser := newQueryParserStateWithOptionsAndCancellation(tokens, options, cancellationToken)
	expression := parser.parseExpression()
	var statement *ExpressionStatement
	if expression == nil || !parser.atEndOfQuery() {
		statement = MustNewExpressionStatement([]SyntaxElement{shallowQueryExpressionNodeWithCancellation(tokens, cancellationToken, diagnostics...)})
	} else {
		statement = MustNewExpressionStatementTyped(expression)
	}
	return NewTypedSyntaxList("SeparatedElement<Statement>", NewSeparatedElement(statement, nil))
}

func shallowQueryExpressionNode(tokens []LexicalToken, diagnostics ...Diagnostic) *SyntaxNode {
	return shallowQueryExpressionNodeWithCancellation(tokens, CancellationToken{}, diagnostics...)
}

func shallowQueryExpressionNodeWithCancellation(tokens []LexicalToken, cancellationToken CancellationToken, diagnostics ...Diagnostic) *SyntaxNode {
	queryTokens := make([]SyntaxElement, 0, len(tokens))
	for i, token := range tokens {
		if i&255 == 0 {
			cancellationToken.ThrowIfCancellationRequested()
		}
		if token.Kind == SyntaxKindEndOfTextToken {
			break
		}
		queryTokens = append(queryTokens, SyntaxTokenFromLexical(token))
	}
	return NewSyntaxNode(SyntaxKindNone, queryTokens, diagnostics...)
}

func pipeSegments(tokens []LexicalToken) []PipeSegment {
	return pipeSegmentsWithCancellation(tokens, CancellationToken{})
}

func pipeSegmentsWithCancellation(tokens []LexicalToken, cancellationToken CancellationToken) []PipeSegment {
	var segments []PipeSegment
	start := 0
	depth := 0
	for i, token := range tokens {
		if i&255 == 0 {
			cancellationToken.ThrowIfCancellationRequested()
		}
		switch token.Kind {
		case SyntaxKindOpenParenToken, SyntaxKindOpenBracketToken, SyntaxKindOpenBraceToken:
			depth++
		case SyntaxKindCloseParenToken, SyntaxKindCloseBracketToken, SyntaxKindCloseBraceToken:
			if depth > 0 {
				depth--
			}
		case SyntaxKindBarToken:
			if depth == 0 {
				segments = append(segments, makePipeSegmentWithCancellation(tokens, start, i, cancellationToken))
				start = i + 1
			}
		case SyntaxKindEndOfTextToken:
			if start < i {
				segments = append(segments, makePipeSegmentWithCancellation(tokens, start, i, cancellationToken))
			}
			return segments
		}
	}
	if start < len(tokens) {
		segments = append(segments, makePipeSegmentWithCancellation(tokens, start, len(tokens), cancellationToken))
	}
	return segments
}

func makePipeSegment(tokens []LexicalToken, start, end int) PipeSegment {
	return makePipeSegmentWithCancellation(tokens, start, end, CancellationToken{})
}

func makePipeSegmentWithCancellation(tokens []LexicalToken, start, end int, cancellationToken CancellationToken) PipeSegment {
	var text strings.Builder
	for i := start; i < end && i < len(tokens); i++ {
		if (i-start)&255 == 0 {
			cancellationToken.ThrowIfCancellationRequested()
		}
		text.WriteString(tokens[i].Trivia)
		text.WriteString(tokens[i].Text)
	}
	return PipeSegment{StartToken: start, EndToken: end, Text: text.String()}
}
