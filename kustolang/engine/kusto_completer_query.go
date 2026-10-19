package kustolang

import "strings"

// Ported from the query-operator completion surface in
// src/Kusto.Language/Editor/Kusto/KustoCompleter.cs, scoped to the native
// QueryParser operator subset currently ported in query_parser_operator.go.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var queryRootOperatorCompletionKinds = []SyntaxKind{
	SyntaxKindEvaluateKeyword,
	SyntaxKindFindKeyword,
	SyntaxKindMacroExpandKeyword,
	SyntaxKindPrintKeyword,
	SyntaxKindRangeKeyword,
	SyntaxKindSearchKeyword,
	SyntaxKindUnionKeyword,
}

var queryPipedOperatorCompletionKinds = []SyntaxKind{
	SyntaxKindAsKeyword,
	SyntaxKindAssertSchemaKeyword,
	SyntaxKindConsumeKeyword,
	SyntaxKindCountKeyword,
	SyntaxKindDistinctKeyword,
	SyntaxKindExtendKeyword,
	SyntaxKindEvaluateKeyword,
	SyntaxKindExecuteAndCacheKeyword,
	SyntaxKindFacetKeyword,
	SyntaxKindForkKeyword,
	SyntaxKindFilterKeyword,
	SyntaxKindWhereKeyword,
	SyntaxKindGetSchemaKeyword,
	SyntaxKindInvokeKeyword,
	SyntaxKindJoinKeyword,
	SyntaxKindLookupKeyword,
	SyntaxKindMakeGraphKeyword,
	SyntaxKindMakeSeriesKeyword,
	SyntaxKindGraphMatchKeyword,
	SyntaxKindGraphMarkComponentsKeyword,
	SyntaxKindGraphShortestPathsKeyword,
	SyntaxKindGraphToTableKeyword,
	SyntaxKindMvDashApplyKeyword,
	SyntaxKindMvApplyKeyword,
	SyntaxKindMvDashExpandKeyword,
	SyntaxKindMvExpandKeyword,
	SyntaxKindParseKeyword,
	SyntaxKindParseKvKeyword,
	SyntaxKindParseWhereKeyword,
	SyntaxKindPartitionKeyword,
	SyntaxKindPartitionByKeyword,
	SyntaxKindProjectKeyword,
	SyntaxKindProjectAwayKeyword,
	SyntaxKindProjectByNamesKeyword,
	SyntaxKindProjectKeepKeyword,
	SyntaxKindProjectRenameKeyword,
	SyntaxKindProjectReorderKeyword,
	SyntaxKindReduceKeyword,
	SyntaxKindSampleKeyword,
	SyntaxKindSampleDistinctKeyword,
	SyntaxKindRenderKeyword,
	SyntaxKindScanKeyword,
	SyntaxKindSearchKeyword,
	SyntaxKindSerializeKeyword,
	SyntaxKindSortKeyword,
	SyntaxKindOrderKeyword,
	SyntaxKindSummarizeKeyword,
	SyntaxKindTakeKeyword,
	SyntaxKindLimitKeyword,
	SyntaxKindTopHittersKeyword,
	SyntaxKindTopNestedKeyword,
	SyntaxKindTopKeyword,
	SyntaxKindUnionKeyword,
}

type queryCompletionContext struct {
	afterPipe   bool
	partialText string
}

func (c *kustoCompleter) getQuerySyntaxCompletions(position int, builder *kustoCompletionBuilder) {
	if c.queryCompletionFollowsIncompleteStatement(position) {
		return
	}
	c.getQuerySyntaxCompletionsFromAnnotations(position, builder)

	context, ok := c.queryCompletionContext(position)
	if !ok {
		return
	}

	kinds := queryRootOperatorCompletionKinds
	if context.afterPipe {
		kinds = queryPipedOperatorCompletionKinds
	}

	for _, kind := range kinds {
		item, ok := querySyntaxCompletionItem(kind)
		if !ok || !queryCompletionItemMatchesPartial(item, context.partialText) {
			continue
		}
		c.addSyntaxCompletionItem(builder, item, CompletionHintQuery, SymbolMatchNone, nil, position)
	}
}

func (c *kustoCompleter) queryCompletionFollowsIncompleteStatement(position int) bool {
	refs, index, ok := kustoServiceTokenLeftRef(c.code, position)
	if !ok || refs[index].token == nil || refs[index].token.Kind() != SyntaxKindSemicolonToken {
		return false
	}

	for i := len(refs[index].ancestors) - 1; i >= 0; i-- {
		separated, ok := WrapGeneratedSyntaxElement(refs[index].ancestors[i].element).(*SeparatedElement)
		if !ok || separated.Element() == nil {
			continue
		}
		statement, ok := WrapGeneratedSyntaxElement(separated.Element()).(Statement)
		return ok && syntaxElementHasMissingChildren(statement)
	}

	return false
}

func (c *kustoCompleter) getQuerySyntaxCompletionsFromAnnotations(position int, builder *kustoCompletionBuilder) bool {
	if c == nil || c.code == nil || builder == nil {
		return false
	}
	root, startIndex, matchIndex, ok := c.extendedGrammarContext(position)
	if !ok || root == nil || startIndex < 0 || matchIndex < startIndex || matchIndex >= len(c.code.Tokens) {
		return false
	}

	parsers := FindAnnotatedParsers(
		NewArraySource(c.code.Tokens[:matchIndex+1]),
		matchIndex,
		root,
		startIndex,
	)
	if len(parsers) == 0 {
		return false
	}

	hints := c.getCompletionHint(position)
	for _, parser := range parsers {
		hints |= syntaxCompletionHintFromAnnotations(parser.Annotations())
	}
	match := symbolMatchFromCompletionHint(hints)
	left := GetCompleteExpressionLeftOfPosition(c.code, position)
	queryContext, hasQueryContext := c.queryCompletionContext(position)

	for _, parser := range parsers {
		for _, item := range syntaxCompletionItemsFromAnnotations(parser.Annotations()) {
			if hasQueryContext && item.Kind == CompletionKindQueryPrefix && !queryCompletionItemMatchesPartial(item, queryContext.partialText) {
				continue
			}
			c.addSyntaxCompletionItem(builder, item, hints, match, left, position)
		}
	}
	return true
}

func symbolMatchFromCompletionHint(hint CompletionHint) SymbolMatch {
	match := SymbolMatchNone
	if hint&CompletionHintColumn != 0 {
		match |= SymbolMatchColumn
	}
	if hint&CompletionHintTabularFunction != 0 {
		match |= SymbolMatchFunction | SymbolMatchTabular | SymbolMatchLocal
	}
	if hint&(CompletionHintScalarFunction|CompletionHintAggregate) != 0 {
		match |= SymbolMatchFunction | SymbolMatchScalar | SymbolMatchLocal
	}
	if hint&(CompletionHintFunction|CompletionHintDatabaseFunction) != 0 {
		match |= SymbolMatchFunction
	}
	if hint&(CompletionHintScalar|CompletionHintBoolean|CompletionHintNumber) != 0 {
		match |= SymbolMatchScalar | SymbolMatchColumn | SymbolMatchFunction | SymbolMatchLocal
	}
	if hint&CompletionHintNonScalar != 0 {
		match |= SymbolMatchNonScalar | SymbolMatchTable | SymbolMatchExternalTable | SymbolMatchMaterializedView |
			SymbolMatchFunction | SymbolMatchLocal | SymbolMatchEntityGroupElement | SymbolMatchGraph
	}
	if hint&CompletionHintTabular != 0 {
		match |= SymbolMatchTabular | SymbolMatchTable | SymbolMatchExternalTable | SymbolMatchMaterializedView |
			SymbolMatchFunction | SymbolMatchLocal | SymbolMatchEntityGroupElement | SymbolMatchGraph
	}
	if hint&CompletionHintExpression != 0 {
		match |= SymbolMatchScalar | SymbolMatchNonScalar | SymbolMatchColumn | SymbolMatchTable |
			SymbolMatchExternalTable | SymbolMatchMaterializedView | SymbolMatchFunction | SymbolMatchLocal |
			SymbolMatchEntityGroup | SymbolMatchEntityGroupElement | SymbolMatchGraph
	}
	if hint&CompletionHintTable != 0 {
		match |= SymbolMatchTable
	}
	if hint&CompletionHintExternalTable != 0 {
		match |= SymbolMatchExternalTable
	}
	if hint&CompletionHintMaterializedView != 0 {
		match |= SymbolMatchMaterializedView
	}
	if hint&CompletionHintDatabase != 0 {
		match |= SymbolMatchDatabase
	}
	if hint&CompletionHintCluster != 0 {
		match |= SymbolMatchCluster
	}
	if hint&CompletionHintEntityGroup != 0 {
		match |= SymbolMatchEntityGroup
	}
	if hint&CompletionHintOption != 0 {
		match |= SymbolMatchOption
	}
	if hint&CompletionHintGraph != 0 {
		match |= SymbolMatchGraph
	}
	if hint&CompletionHintGraphModel != 0 {
		match |= SymbolMatchGraphModel
	}
	if hint&CompletionHintGraphSnapshot != 0 {
		match |= SymbolMatchGraphSnapshot
	}
	if hint&CompletionHintStoredQueryResult != 0 {
		match |= SymbolMatchStoredQueryResult
	}
	return match
}

func (c *kustoCompleter) queryCompletionContext(position int) (queryCompletionContext, bool) {
	if c == nil || c.code == nil || !kustoCodeIsQueryCompletionPosition(c.code, position) {
		return queryCompletionContext{}, false
	}
	tokens := c.code.Tokens
	starts := c.code.TokenStarts
	if len(tokens) == 0 || len(starts) < len(tokens) {
		return queryCompletionContext{}, false
	}

	segmentStart := 0
	afterPipe := false
	depth := 0
	for i, token := range tokens {
		textStart := kustoCompleterLexicalTextStart(starts[i], token)
		if textStart >= position || token.Kind == SyntaxKindEndOfTextToken {
			break
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
				segmentStart = i + 1
				afterPipe = true
			}
		case SyntaxKindSemicolonToken:
			if depth == 0 {
				segmentStart = i + 1
				afterPipe = false
			}
		}
	}

	findOffset := c.code.GetTokenIndex(position)
	if findOffset < segmentStart || findOffset >= len(tokens) || findOffset >= len(starts) {
		return queryCompletionContext{}, false
	}
	if !querySegmentIsAtOperatorPosition(tokens, starts, segmentStart, findOffset, position) {
		return queryCompletionContext{}, false
	}

	return queryCompletionContext{
		afterPipe:   afterPipe,
		partialText: c.queryCompletionPartialText(findOffset, position),
	}, true
}

func querySegmentIsAtOperatorPosition(tokens []LexicalToken, starts []int, segmentStart int, findOffset int, position int) bool {
	for i := segmentStart; i < findOffset && i < len(tokens) && i < len(starts); i++ {
		token := tokens[i]
		if token.Kind == SyntaxKindEndOfTextToken || token.Text == "" {
			continue
		}
		textStart := kustoCompleterLexicalTextStart(starts[i], token)
		if textStart >= position {
			break
		}
		return false
	}
	return true
}

func (c *kustoCompleter) queryCompletionPartialText(findOffset int, position int) string {
	if c == nil || c.code == nil || findOffset < 0 || findOffset >= len(c.code.Tokens) || findOffset >= len(c.code.TokenStarts) {
		return ""
	}
	token := c.code.Tokens[findOffset]
	textStart := kustoCompleterLexicalTextStart(c.code.TokenStarts[findOffset], token)
	if token.Kind == SyntaxKindEndOfTextToken || token.Text == "" || position < textStart {
		return ""
	}
	textEnd := textStart + len(token.Text)
	if position == textStart {
		return c.queryCompletionContiguousNameText(findOffset)
	}
	if position < textEnd {
		return token.Text[:position-textStart]
	}
	return token.Text
}

func (c *kustoCompleter) queryCompletionContiguousNameText(start int) string {
	tokens := c.code.Tokens
	starts := c.code.TokenStarts
	if start < 0 || start >= len(tokens) || start >= len(starts) {
		return ""
	}
	if !kustoCompleterIsKeywordOrIdentifierOrDash(tokens[start].Kind) {
		return tokens[start].Text
	}

	var builder strings.Builder
	for i := start; i < len(tokens) && i < len(starts); i++ {
		token := tokens[i]
		if !kustoCompleterIsKeywordOrIdentifierOrDash(token.Kind) {
			break
		}
		if i > start {
			prev := tokens[i-1]
			prevEnd := kustoCompleterLexicalTextStart(starts[i-1], prev) + len(prev.Text)
			currStart := kustoCompleterLexicalTextStart(starts[i], token)
			if prevEnd != currStart {
				break
			}
		}
		builder.WriteString(token.Text)
	}
	return builder.String()
}

func querySyntaxCompletionItem(kind SyntaxKind) (CompletionItem, bool) {
	return CreateSyntaxCompletionItemForKind(kind, CompletionKindQueryPrefix, CompletionPriorityNormal)
}

func queryCompletionItemMatchesPartial(item CompletionItem, partial string) bool {
	if partial == "" {
		return true
	}
	lowerPartial := strings.ToLower(partial)
	return strings.HasPrefix(strings.ToLower(item.MatchText), lowerPartial) ||
		strings.HasPrefix(strings.ToLower(item.DisplayText), lowerPartial)
}
