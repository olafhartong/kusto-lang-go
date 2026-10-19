package kustolang

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Ported from src/Kusto.Language/Editor/Kusto/KustoCompleter.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type kustoCompleter struct {
	code              *KustoCode
	options           CompletionOptions
	cancellationToken CancellationToken
}

type completionMode int

const (
	completionModeIsolated completionMode = iota
	completionModeCombined
)

type kustoCompletionBuilder struct {
	list     []CompletionItem
	indexMap map[string]int
}

func newKustoCompletionBuilder() *kustoCompletionBuilder {
	return &kustoCompletionBuilder{indexMap: map[string]int{}}
}

func (b *kustoCompletionBuilder) count() int {
	if b == nil {
		return 0
	}
	return len(b.list)
}

func (b *kustoCompletionBuilder) items() []CompletionItem {
	if b == nil {
		return EmptyReadOnlyList[CompletionItem]()
	}
	return ListToReadOnly(b.list)
}

func (b *kustoCompletionBuilder) add(item CompletionItem) {
	if b == nil {
		return
	}
	if index, ok := b.indexMap[item.DisplayText]; ok {
		if kustoCompletionBuilderPrefersOther(b.list[index], item) {
			b.list[index] = item
		}
		return
	}
	b.indexMap[item.DisplayText] = len(b.list)
	b.list = append(b.list, item)
}

func (b *kustoCompletionBuilder) addRange(items []CompletionItem) {
	for _, item := range items {
		b.add(item)
	}
}

func kustoCompletionBuilderPrefersOther(existing CompletionItem, other CompletionItem) bool {
	if existing.Kind == CompletionKindKeyword && other.Kind == CompletionKindQueryPrefix {
		return true
	}
	return false
}

func newKustoCompleter(code *KustoCode, options CompletionOptions, cancellationToken CancellationToken) *kustoCompleter {
	return &kustoCompleter{code: code, options: options, cancellationToken: cancellationToken}
}

func utf8FirstRune(text string) (rune, bool) {
	if text == "" {
		return 0, false
	}
	r, _ := utf8.DecodeRuneInString(text)
	return r, true
}

func utf8LastRune(text string) (rune, bool) {
	if text == "" {
		return 0, false
	}
	r, _ := utf8.DecodeLastRuneInString(text)
	return r, true
}

func (c *kustoCompleter) getCompletionItems(position int) CompletionInfo {
	editStart, editLength := c.getEditRange(position)
	if !c.shouldComplete(position) {
		return NewCompletionInfo(nil, editStart, editLength)
	}

	builder := newKustoCompletionBuilder()
	mode := completionModeCombined
	if mode == completionModeCombined && c.options.IncludeSymbols() {
		mode = c.getSymbolCompletions(editStart, builder)
	}
	if mode == completionModeCombined && c.options.IncludeSyntax() {
		c.getSyntaxCompletions(editStart, builder)
		if c.options.IncludeExtendedSyntax() {
			c.getExtendedSyntaxCompletions(editStart, builder)
		}
	}
	return NewCompletionInfo(c.orderCompletionItems(builder.items()), editStart, editLength)
}

func (c *kustoCompleter) orderCompletionItems(items []CompletionItem) []CompletionItem {
	ordered := make([]CompletionItem, 0, len(items))
	for _, item := range items {
		item = c.retriggered(item)
		item = item.WithOrderText(c.createOrderText(item))
		ordered = append(ordered, item)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].OrderText < ordered[j].OrderText
	})
	return ListToReadOnly(ordered)
}

func (c *kustoCompleter) createOrderText(item CompletionItem) string {
	return fmt.Sprintf("%02d%d_%s", c.getOrderingRank(item), item.priority, strings.ToLower(item.OrderText))
}

func (c *kustoCompleter) retriggered(item CompletionItem) CompletionItem {
	if item.Retrigger {
		return item
	}
	beforeText := item.BeforeText()
	if beforeText == "" {
		return item
	}
	prevChar, _ := utf8LastRune(beforeText)
	if IsWhitespace(prevChar) || prevChar == '(' {
		return item.WithRetrigger(true)
	}
	return item
}

func (c *kustoCompleter) getOrderingRank(item CompletionItem) CompletionRank {
	c.cancellationToken.ThrowIfCancellationRequested()
	if item.rank != CompletionRankDefault {
		return item.rank
	}
	switch item.Kind {
	case CompletionKindExample:
		return CompletionRankLiteral
	case CompletionKindQueryPrefix, CompletionKindKeyword:
		return CompletionRankKeyword
	case CompletionKindAggregateFunction:
		return CompletionRankAggregate
	case CompletionKindColumn:
		return CompletionRankColumn
	case CompletionKindTable:
		return CompletionRankTable
	case CompletionKindMaterialiedView, CompletionKindEntityGroup, CompletionKindGraph, CompletionKindStoredQueryResult:
		return CompletionRankEntity
	case CompletionKindVariable, CompletionKindParameter:
		if item.DisplayText == "$left" || item.DisplayText == "$right" {
			return CompletionRankLiteral
		}
		return CompletionRankVariable
	case CompletionKindBuiltInFunction, CompletionKindLocalFunction, CompletionKindDatabaseFunction:
		return CompletionRankFunction
	case CompletionKindScalarInfix:
		if item.DisplayText != "" {
			first, _ := utf8FirstRune(item.DisplayText)
			if IsLetterOrDigit(first) {
				return CompletionRankStringOperator
			}
		}
		return CompletionRankMathOperator
	default:
		return CompletionRankOther
	}
}

func (c *kustoCompleter) getSyntaxCompletions(position int, builder *kustoCompletionBuilder) {
	if c == nil || c.code == nil || builder == nil {
		return
	}
	if kustoCodeIsQueryCompletionPosition(c.code, position) {
		c.getQuerySyntaxCompletions(position, builder)
	} else if c.code.Kind == CodeKindCommand {
		c.getCommandSyntaxCompletions(position, builder)
	}
}

func kustoCodeIsQueryCompletionPosition(code *KustoCode, position int) bool {
	if code == nil || code.Syntax() == nil {
		return false
	}
	if code.Kind == CodeKindQuery {
		return true
	}
	block, ok := WrapGeneratedSyntaxElement(code.Syntax()).(*CommandBlock)
	return ok && commandBlockQueryStatementPosition(code.Syntax(), block, position)
}

func (c *kustoCompleter) getCommandSyntaxCompletions(position int, builder *kustoCompletionBuilder) {
	if c.getCommandSyntaxCompletionsFromAnnotations(position, builder) {
		return
	}

	prefix, ok := c.commandTokenPrefixBefore(position)
	if !ok {
		return
	}
	grammar := CommandGrammarFrom(c.code.Globals)
	if grammar == nil {
		return
	}

	candidates := map[string]struct{}{}
	for _, info := range grammar.commandParsers {
		if len(info.Elements) > 0 {
			collectCommandNextTokens(info.Elements, prefix, candidates, 0)
			continue
		}
		if len(prefix) < len(info.Tokens) && commandTokenPrefixMatches(info.Tokens, prefix) {
			candidates[info.Tokens[len(prefix)]] = struct{}{}
		}
	}

	for candidate := range candidates {
		item, ok := commandSyntaxCompletionItem(candidate)
		if !ok {
			continue
		}
		c.addSyntaxCompletionItem(builder, item, CompletionHintCommand, SymbolMatchNone, nil, position)
	}
}

type commandCompletionContext struct {
	startIndex  int
	findOffset  int
	partialText string
}

func (c *kustoCompleter) getCommandSyntaxCompletionsFromAnnotations(position int, builder *kustoCompletionBuilder) bool {
	context, ok := c.commandCompletionContext(position)
	if !ok {
		return false
	}
	grammar := CommandGrammarFrom(c.code.Globals)
	if grammar == nil {
		return false
	}
	parser := grammar.CommandCompletionParser()
	if parser == nil {
		return false
	}

	parsers := FindAnnotatedParsers(NewArraySource(c.code.Tokens), context.findOffset, parser, context.startIndex)
	if len(parsers) == 0 {
		return false
	}

	before := builder.count()
	for _, parser := range parsers {
		annotations := parser.Annotations()
		hints := syntaxCompletionHintFromAnnotations(annotations)
		if hints == CompletionHintNone {
			hints = CompletionHintCommand
		}
		for _, item := range syntaxCompletionItemsFromAnnotations(annotations) {
			if !commandCompletionItemMatchesPartial(item, context.partialText) {
				continue
			}
			c.addSyntaxCompletionItem(builder, item, hints, SymbolMatchNone, nil, position)
		}
	}
	return builder.count() > before
}

func (c *kustoCompleter) commandCompletionContext(position int) (commandCompletionContext, bool) {
	if c == nil || c.code == nil || c.code.Kind != CodeKindCommand {
		return commandCompletionContext{}, false
	}
	tokens := c.code.Tokens
	starts := c.code.TokenStarts
	if len(tokens) == 0 || len(starts) < len(tokens) {
		return commandCompletionContext{}, false
	}

	segmentStart := 0
	for i, token := range tokens {
		textStart := kustoCompleterLexicalTextStart(starts[i], token)
		if textStart >= position || token.Kind == SyntaxKindEndOfTextToken {
			break
		}
		if token.Kind == SyntaxKindSemicolonToken || token.Kind == SyntaxKindDirectiveToken {
			segmentStart = i + 1
		}
	}
	if segmentStart >= len(tokens) ||
		kustoCompleterLexicalTextStart(starts[segmentStart], tokens[segmentStart]) >= position ||
		tokens[segmentStart].Kind != SyntaxKindDotToken {
		return commandCompletionContext{}, false
	}

	for i := segmentStart + 1; i < len(tokens); i++ {
		token := tokens[i]
		textStart := kustoCompleterLexicalTextStart(starts[i], token)
		if textStart >= position || token.Kind == SyntaxKindEndOfTextToken || token.Kind == SyntaxKindSemicolonToken {
			break
		}
		if token.Kind == SyntaxKindBarToken {
			return commandCompletionContext{}, false
		}
	}

	findOffset := c.code.GetTokenIndex(position)
	if findOffset < segmentStart || findOffset >= len(tokens) || findOffset >= len(starts) {
		return commandCompletionContext{}, false
	}

	partialText := ""
	token := tokens[findOffset]
	textStart := kustoCompleterLexicalTextStart(starts[findOffset], token)
	if token.Kind != SyntaxKindEndOfTextToken && token.Text != "" && position >= textStart {
		textEnd := textStart + len(token.Text)
		switch {
		case position == textStart:
			partialText = token.Text
		case position < textEnd:
			partialText = token.Text[:position-textStart]
		case position == textEnd:
			partialText = token.Text
		}
	}

	return commandCompletionContext{
		startIndex:  segmentStart,
		findOffset:  findOffset,
		partialText: partialText,
	}, true
}

func commandCompletionItemMatchesPartial(item CompletionItem, partial string) bool {
	if partial == "" {
		return true
	}
	lowerPartial := strings.ToLower(partial)
	return strings.HasPrefix(strings.ToLower(item.MatchText), lowerPartial) ||
		strings.HasPrefix(strings.ToLower(item.DisplayText), lowerPartial)
}

func commandTokenPrefixMatches(tokens []string, prefix []string) bool {
	if len(prefix) > len(tokens) {
		return false
	}
	for i, token := range prefix {
		if !strings.EqualFold(tokens[i], token) {
			return false
		}
	}
	return true
}

func (c *kustoCompleter) commandTokenPrefixBefore(position int) ([]string, bool) {
	if c == nil || c.code == nil || c.code.Kind != CodeKindCommand {
		return nil, false
	}
	tokens := c.code.Tokens
	starts := c.code.TokenStarts
	if len(tokens) == 0 || len(starts) < len(tokens) {
		return nil, false
	}

	segmentStart := 0
	for i, token := range tokens {
		if kustoCompleterLexicalTextStart(starts[i], token) >= position || token.Kind == SyntaxKindEndOfTextToken {
			break
		}
		if token.Kind == SyntaxKindSemicolonToken {
			segmentStart = i + 1
		}
	}

	index := segmentStart
	for index < len(tokens) && kustoCompleterLexicalTextStart(starts[index], tokens[index]) < position && tokens[index].Kind == SyntaxKindEndOfTextToken {
		index++
	}
	if index >= len(tokens) || kustoCompleterLexicalTextStart(starts[index], tokens[index]) >= position || tokens[index].Kind != SyntaxKindDotToken {
		return nil, false
	}

	prefix := []string{}
	for i := index + 1; i < len(tokens); i++ {
		token := tokens[i]
		if kustoCompleterLexicalTextStart(starts[i], token) >= position || token.Kind == SyntaxKindEndOfTextToken || token.Kind == SyntaxKindSemicolonToken {
			break
		}
		if token.Kind == SyntaxKindBarToken {
			return nil, false
		}
		if token.Text != "" {
			prefix = append(prefix, strings.ToLower(token.Text))
		}
	}
	return prefix, true
}

func kustoCompleterLexicalTextStart(tokenStart int, token LexicalToken) int {
	return tokenStart + len(token.Trivia)
}

func commandSyntaxCompletionItem(text string) (CompletionItem, bool) {
	if strings.TrimSpace(text) == "" || strings.HasPrefix(text, "_") {
		return CompletionItem{}, false
	}
	kind := CompletionKindCommandPrefix
	if syntaxCompletionKindForText(text, CompletionKindCommandPrefix) == CompletionKindPunctuation {
		kind = CompletionKindPunctuation
	}
	return NewCompletionItem(kind, text), true
}

const maxCommandCompletionDepth = 80

func collectCommandNextTokens(elements []commandTokenElement, prefix []string, out map[string]struct{}, depth int) {
	if depth > maxCommandCompletionDepth || len(elements) == 0 {
		return
	}
	first := elements[0]
	rest := elements[1:]

	if first.Hidden {
		if len(prefix) == 0 {
			if len(first.Choices) > 0 || len(first.Children) > 0 || len(first.List) > 0 || first.Rule != "" {
				collectCommandNextTokens(rest, prefix, out, depth+1)
			}
			return
		}
		unhidden := first
		unhidden.Hidden = false
		collectCommandNextTokens(appendCommandElements([]commandTokenElement{unhidden}, rest), prefix, out, depth+1)
		if commandElementCanSkip(first) {
			collectCommandNextTokens(rest, prefix, out, depth+1)
		}
		return
	}

	if len(first.Choices) > 0 {
		for _, choice := range first.Choices {
			if choice.Hidden && len(prefix) == 0 {
				continue
			}
			collectCommandNextTokens(appendCommandElements([]commandTokenElement{choice}, rest), prefix, out, depth+1)
		}
		if commandElementCanSkip(first) {
			collectCommandNextTokens(rest, prefix, out, depth+1)
		}
		return
	}
	if len(first.Children) > 0 {
		collectCommandNextTokens(appendCommandElements(first.Children, rest), prefix, out, depth+1)
		if commandElementCanSkip(first) {
			collectCommandNextTokens(rest, prefix, out, depth+1)
		}
		return
	}
	if len(first.List) > 0 {
		collectCommandNextTokens(appendCommandElements(first.List, rest), prefix, out, depth+1)
		if commandElementCanSkip(first) || first.ListAllowEmpty {
			collectCommandNextTokens(rest, prefix, out, depth+1)
		}
		return
	}

	if len(prefix) == 0 {
		for _, alternative := range first.Alternatives {
			out[alternative] = struct{}{}
		}
		if commandElementCanSkip(first) {
			collectCommandNextTokens(rest, prefix, out, depth+1)
		}
		return
	}

	if len(first.Alternatives) > 0 {
		for _, alternative := range first.Alternatives {
			if strings.EqualFold(alternative, prefix[0]) {
				collectCommandNextTokens(rest, prefix[1:], out, depth+1)
				break
			}
		}
		if commandElementCanSkip(first) {
			collectCommandNextTokens(rest, prefix, out, depth+1)
		}
		return
	}

	if first.Rule != "" {
		collectCommandNextTokens(rest, prefix[1:], out, depth+1)
	}
	if commandElementCanSkip(first) {
		collectCommandNextTokens(rest, prefix, out, depth+1)
	}
}

func appendCommandElements(first []commandTokenElement, rest []commandTokenElement) []commandTokenElement {
	combined := make([]commandTokenElement, 0, len(first)+len(rest))
	combined = append(combined, first...)
	combined = append(combined, rest...)
	return combined
}

func commandElementCanSkip(element commandTokenElement) bool {
	return element.Optional || element.ShapeOptional
}

func (c *kustoCompleter) addSyntaxCompletionItem(builder *kustoCompletionBuilder, item CompletionItem, hints CompletionHint, match SymbolMatch, left Expression, position int) {
	if builder == nil || !c.includeSyntax(item, hints, match, left, position) {
		return
	}
	if c.shouldAugmentSyntaxCompletionItem(item) {
		item = addSpaceToCompletionItem(item)
	}
	builder.add(item)
}

func (c *kustoCompleter) includeSyntax(item CompletionItem, hints CompletionHint, match SymbolMatch, left Expression, position int) bool {
	if c != nil && !c.options.IncludePunctuationOnlySyntax() && item.Kind == CompletionKindPunctuation {
		return false
	}
	switch item.Kind {
	case CompletionKindQueryPrefix:
		return hints&(CompletionHintQuery|CompletionHintKeyword) != 0 ||
			match.Has(SymbolMatchTabular) || match.Has(SymbolMatchNonScalar)
	case CompletionKindTabularPrefix:
		return match.Has(SymbolMatchTabular) || match.Has(SymbolMatchNonScalar)
	case CompletionKindTabularSuffix:
		return left != nil && ExpressionResultType(left) != nil && IsTabular(ExpressionResultType(left))
	case CompletionKindScalarPrefix, CompletionKindExample:
		return match.Has(SymbolMatchScalar)
	case CompletionKindScalarInfix:
		return c.anyInfixMatches(left, item.MatchText, position)
	default:
		return true
	}
}

func addSpaceToCompletionItem(item CompletionItem) CompletionItem {
	beforeText := item.BeforeText()
	if beforeText != "" {
		last, _ := utf8LastRune(beforeText)
		if !IsWhitespace(last) {
			return item.WithBeforeText(beforeText + " ")
		}
	}
	return item
}

var punctuationWithoutSpace = map[string]struct{}{
	"(": {},
	"[": {},
	"{": {},
	"*": {},
	"@": {},
	"$": {},
}

func (c *kustoCompleter) shouldAugmentSyntaxCompletionItem(item CompletionItem) bool {
	if c == nil || !c.options.AutoAppendWhitespace() {
		return false
	}
	if item.AfterText() != "" {
		return false
	}

	beforeText := item.BeforeText()
	if strings.HasSuffix(beforeText, "=") && len(beforeText) > 1 {
		first, _ := utf8FirstRune(beforeText)
		beforeEqual := beforeText[:len(beforeText)-1]
		prev, _ := utf8LastRune(beforeEqual)
		if IsLetterOrDigit(first) && !IsWhitespace(prev) {
			return false
		}
	}

	switch item.Kind {
	case CompletionKindColumn,
		CompletionKindBuiltInFunction,
		CompletionKindLocalFunction,
		CompletionKindDatabaseFunction,
		CompletionKindAggregateFunction,
		CompletionKindParameter,
		CompletionKindVariable,
		CompletionKindTable,
		CompletionKindDatabase,
		CompletionKindCluster:
		return false
	case CompletionKindPunctuation:
		_, ok := punctuationWithoutSpace[item.DisplayText]
		return ok
	default:
		return true
	}
}

func (c *kustoCompleter) shouldAutoComplete(position int, key rune) bool {
	if c == nil || c.code == nil || key == '\x00' {
		return false
	}
	if !c.shouldComplete(position) {
		return false
	}
	if unicode.IsControl(key) && key != '\b' {
		return false
	}
	if key == '\r' || key == '\n' {
		return false
	}

	token := kustoCompleterTokenLeftOfPosition(c.code, position)
	if token == nil {
		return false
	}

	if unicode.IsSpace(key) && position == token.End()+1 {
		if kustoCompleterAllowsWhitespaceAfter(token.Kind()) {
			return true
		}
		if kustoCompleterIsSyntaxOrExpressionStart(token.Kind()) {
			return true
		}
		return GetCompleteExpressionLeftOfPosition(c.code, position) != nil
	}

	if token.End() != position {
		return false
	}
	if key == '\b' {
		return SyntaxCategoryOf(token.Kind()) == SyntaxCategoryIdentifier
	}
	if kustoCompleterAllowsImmediateFollower(token.Kind()) {
		return true
	}
	return kustoCompleterIsSyntaxOrExpressionStart(token.Kind())
}

func (c *kustoCompleter) shouldComplete(position int) bool {
	if c == nil || c.code == nil || c.code.Syntax() == nil {
		return false
	}
	token, previous := kustoCompleterTokenAt(c.code, position)
	if token == nil {
		return false
	}

	affinity := GetTokenWithAffinity(c.code, position)
	if affinity == nil {
		affinity = token
	}

	leftOrSurrounding := token
	if position == token.TriviaStart() && previous != nil {
		leftOrSurrounding = previous
	}
	if kustoCompleterIsInsideLiteral(leftOrSurrounding, position) ||
		(kustoCompleterIsInsideDirective(leftOrSurrounding, position) &&
			kustoCompleterLexicallyInsideDirective(c.code, position)) {
		return false
	}

	if position <= affinity.TextStart() && affinity.TriviaWidth() > 0 {
		return !KustoCompleterIsInsideComment(token.Trivia(), position-token.TriviaStart())
	}
	return true
}

func (c *kustoCompleter) getEditRange(position int) (editStart int, editLength int) {
	// Use the lexical stream for the editable token range because incomplete
	// dash-joined names can be recovered out of the syntax tree as skipped text.
	index, ok := kustoCompleterLexicalTokenIndexWithAffinity(c.code, position)
	if !ok {
		return position, 0
	}

	curr := index
	for prev := curr - 1; prev >= 0; prev = curr - 1 {
		prevToken := kustoCompleterLexicalPositionedToken(c.code, prev)
		currToken := kustoCompleterLexicalPositionedToken(c.code, curr)
		if !kustoCompleterIsKeywordOrIdentifierOrDash(prevToken.Kind()) ||
			!kustoCompleterIsKeywordOrIdentifierOrDash(currToken.Kind()) ||
			prevToken.End() != currToken.TextStart() {
			break
		}
		curr = prev
	}

	first := curr
	var builder strings.Builder
	for next := curr + 1; next < len(c.code.Tokens); next = curr + 1 {
		currToken := kustoCompleterLexicalPositionedToken(c.code, curr)
		nextToken := kustoCompleterLexicalPositionedToken(c.code, next)
		if !kustoCompleterIsKeywordOrIdentifierOrDash(nextToken.Kind()) ||
			!kustoCompleterIsKeywordOrIdentifierOrDash(currToken.Kind()) ||
			currToken.End() != nextToken.TextStart() {
			break
		}
		builder.WriteString(currToken.RawText())
		curr = next
	}

	last := curr
	firstToken := kustoCompleterLexicalPositionedToken(c.code, first)
	if first != last && builder.Len() > 0 && kustoCompleterIsStartOfKnownNameWithDashes(builder.String()) {
		lastToken := kustoCompleterLexicalPositionedToken(c.code, last)
		return firstToken.TextStart(), lastToken.End() - firstToken.TextStart()
	}
	return firstToken.TextStart(), firstToken.Width()
}

func KustoCompleterIsInsideComment(trivia string, triviaOffset int) bool {
	start, length, ok := TryGetCommentSpan(trivia, triviaOffset)
	if !ok {
		return false
	}
	if triviaOffset == start {
		return false
	}
	if triviaOffset == start+length && HasLineBreaks(trivia) {
		return false
	}
	return true
}

func kustoCompleterIsInsideLiteral(token *SyntaxToken, position int) bool {
	return token != nil &&
		SyntaxCategoryOf(token.Kind()) == SyntaxCategoryLiteral &&
		position >= token.TextStart() &&
		position <= token.End()
}

func kustoCompleterIsInsideDirective(token *SyntaxToken, position int) bool {
	return token != nil &&
		token.Kind() == SyntaxKindDirectiveToken &&
		position >= token.TextStart() &&
		position <= token.End()
}

func kustoCompleterLexicallyInsideDirective(code *KustoCode, position int) bool {
	if code == nil {
		return false
	}
	for index, token := range code.Tokens {
		if index >= len(code.TokenStarts) || token.Kind != SyntaxKindDirectiveToken {
			continue
		}
		textStart := kustoCompleterLexicalTextStart(code.TokenStarts[index], token)
		if position >= textStart && position <= textStart+len(token.Text) {
			return true
		}
	}
	return false
}

func kustoCompleterTokenAt(code *KustoCode, position int) (*SyntaxToken, *SyntaxToken) {
	if code == nil || code.Syntax() == nil {
		return nil, nil
	}
	refs := kustoServiceTokenRefs(code.Syntax())
	index := kustoCodeServiceTokenRefIndexAt(refs, position)
	if index < 0 {
		return nil, nil
	}
	token := refs[index].positionedToken()
	var previous *SyntaxToken
	if index > 0 {
		previous = refs[index-1].positionedToken()
	}
	return token, previous
}

func kustoCompleterLexicalTokenIndexWithAffinity(code *KustoCode, position int) (int, bool) {
	if code == nil || len(code.Tokens) == 0 {
		return 0, false
	}
	index := code.GetTokenIndex(position)
	if index < 0 || index >= len(code.Tokens) {
		return 0, false
	}
	token := kustoCompleterLexicalPositionedToken(code, index)
	if token != nil && token.Kind() == SyntaxKindEndOfTextToken && index > 0 {
		previous := kustoCompleterLexicalPositionedToken(code, index-1)
		if previous != nil && previous.End() == position && kustoCompleterIsKeywordOrIdentifierOrDash(previous.Kind()) {
			return index - 1, true
		}
	}
	if HasAffinity(token, position) {
		return index, true
	}
	if index > 0 {
		previous := kustoCompleterLexicalPositionedToken(code, index-1)
		if HasAffinity(previous, position) {
			return index - 1, true
		}
	}
	return 0, false
}

func kustoCompleterLexicalPositionedToken(code *KustoCode, index int) *SyntaxToken {
	if code == nil || index < 0 || index >= len(code.Tokens) || index >= len(code.TokenStarts) {
		return nil
	}
	return SyntaxTokenFromLexical(code.Tokens[index]).withTriviaStart(code.TokenStarts[index])
}

func kustoCompleterTokenLeftOfPosition(code *KustoCode, position int) *SyntaxToken {
	refs, index, ok := kustoServiceTokenLeftRef(code, position)
	if !ok {
		return nil
	}
	for index >= 0 {
		token := refs[index].positionedToken()
		if !syntaxElementIsMissing(token) {
			return token
		}
		index--
	}
	return nil
}

func kustoCompleterIsKeywordOrIdentifierOrDash(kind SyntaxKind) bool {
	return kind == SyntaxKindIdentifierToken ||
		kind == SyntaxKindMinusToken ||
		SyntaxCategoryOf(kind) == SyntaxCategoryKeyword
}

func kustoCompleterIsStartOfKnownNameWithDashes(text string) bool {
	for _, data := range syntaxFacts {
		name := data.Text
		if strings.Contains(name, "-") && strings.HasPrefix(name, text) {
			return true
		}
	}
	return false
}

func kustoCompleterAllowsWhitespaceAfter(kind SyntaxKind) bool {
	switch kind {
	case SyntaxKindOpenParenToken,
		SyntaxKindOpenBracketToken,
		SyntaxKindOpenBraceToken,
		SyntaxKindCommaToken,
		SyntaxKindColonToken,
		SyntaxKindBarToken,
		SyntaxKindEqualToken,
		SyntaxKindFatArrowToken,
		SyntaxKindCloseParenToken,
		SyntaxKindDashDashToken,
		SyntaxKindDashDashGreaterThanToken,
		SyntaxKindLessThanDashDashToken:
		return true
	default:
		return false
	}
}

func kustoCompleterAllowsImmediateFollower(kind SyntaxKind) bool {
	switch kind {
	case SyntaxKindOpenParenToken,
		SyntaxKindOpenBracketToken,
		SyntaxKindOpenBraceToken,
		SyntaxKindDotToken,
		SyntaxKindColonToken,
		SyntaxKindEqualToken,
		SyntaxKindBangToken:
		return true
	default:
		return false
	}
}

func kustoCompleterIsSyntaxOrExpressionStart(kind SyntaxKind) bool {
	switch SyntaxCategoryOf(kind) {
	case SyntaxCategoryIdentifier, SyntaxCategoryOperator, SyntaxCategoryKeyword:
		return true
	default:
		return false
	}
}
