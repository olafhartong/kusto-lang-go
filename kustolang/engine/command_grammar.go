package kustolang

import (
	"slices"
	"strings"
	"sync"
	"sync/atomic"
)

// Ported from the base command-entry and fallback portions of
// src/Kusto.Language/Parser/CommandGrammar.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.
//
// The generated server-specific parser sets are exposed through concrete
// grammar types in server_command_grammars.go. This base retains the common
// unknown/bad-command and command-block behavior.

type CommandGrammar struct {
	Globals                  *GlobalState
	CommandBlock             ResultParser[LexicalToken, *CommandBlock]
	PredefinedRules          *PredefinedRuleParsers
	commandCompletionOnce    sync.Once
	commandCompletionParser  Parser[LexicalToken]
	commandPartialOnce       sync.Once
	commandPartialParsers    []Parser[LexicalToken]
	commandParsers           []commandParserInfo
	commandSymbols           map[string]*CommandSymbol
	allowPartialKnownCommand bool
	serverViewOnce           sync.Once
	serverView               ServerCommandGrammar
}

type recentCommandGrammar struct {
	serverKind string
	grammar    *CommandGrammar
}

type commandParserInfo struct {
	Kind     string
	Tokens   []string
	Elements []commandTokenElement
	Parse    func(*queryParserState) Command
}

type commandTokenElement struct {
	Alternatives         []string
	Rule                 string
	Expected             string
	Forbidden            []string
	Choices              []commandTokenElement
	Children             []commandTokenElement
	List                 []commandTokenElement
	ListSeparator        string
	ListAllowEmpty       bool
	ListMissingElement   bool
	ListMissingSeparator bool
	Name                 string
	Hint                 CompletionHint
	HasHint              bool
	Hidden               bool
	ShapeOptional        bool
	Required             bool
	OptionalGroup        int
	Optional             bool
}

var defaultCommandGrammar atomic.Pointer[CommandGrammar]
var recentCommandGrammarValue atomic.Pointer[recentCommandGrammar]

func CommandGrammarFrom(globals *GlobalState) *CommandGrammar {
	if globals == nil {
		globals = DefaultGlobalState
	}
	if globals.ServerKind == DefaultGlobalState.ServerKind {
		return defaultCommandGrammarInstance()
	}

	var grammar *CommandGrammar
	foundInCache := false
	if globals.Cache != nil {
		grammar, foundInCache = KustoCacheTryGetValue[*CommandGrammar](globals.Cache)
	}

	if grammar == nil {
		recent := recentCommandGrammarValue.Load()
		if recent != nil && recent.serverKind == globals.ServerKind {
			grammar = recent.grammar
		} else {
			grammar = CreateCommandGrammar(globals)
			recentCommandGrammarValue.CompareAndSwap(recent, &recentCommandGrammar{serverKind: globals.ServerKind, grammar: grammar})
		}
	}

	if globals.Cache != nil && !foundInCache {
		grammar = KustoCacheGetOrCreate(globals.Cache, func() *CommandGrammar { return grammar })
	}
	return grammar
}

func defaultCommandGrammarInstance() *CommandGrammar {
	if grammar := defaultCommandGrammar.Load(); grammar != nil {
		return grammar
	}
	grammar := CreateCommandGrammar(DefaultGlobalState)
	if defaultCommandGrammar.CompareAndSwap(nil, grammar) {
		return grammar
	}
	return defaultCommandGrammar.Load()
}

func CreateCommandGrammar(globals *GlobalState) *CommandGrammar {
	return CreateServerCommandGrammar(globals).BaseCommandGrammar()
}

func commandSymbolMap(commands []*CommandSymbol) map[string]*CommandSymbol {
	symbols := make(map[string]*CommandSymbol, len(commands))
	for _, command := range commands {
		if command != nil {
			symbols[command.SymbolName()] = command
		}
	}
	return symbols
}

func NewCommandGrammar(globals *GlobalState) *CommandGrammar {
	return newCommandGrammar(globals)
}

func newCommandGrammar(globals *GlobalState) *CommandGrammar {
	if globals == nil {
		globals = DefaultGlobalState
	}
	grammar := &CommandGrammar{Globals: globals}
	grammar.PredefinedRules = NewPredefinedRuleParsersWithGrammar(QueryGrammarFrom(globals),
		commandQueryInputParser(),
		commandScriptInputParser(grammar),
	)
	grammar.CommandBlock = grammar.commandBlockResultParser()
	return grammar
}

func commandQueryInputParser() ResultParser[LexicalToken, SyntaxElement] {
	return asResultParser[LexicalToken, SyntaxElement](
		MatchConsumerValue(
			func(source Source[LexicalToken], start int) int {
				parser := newQueryParserState(lexicalTokensFromSource(source, start))
				parser.parseQueryBlockStatementList()
				// QueryGrammar.StatementList is a zero-or-more parser in both
				// source revisions, so an empty command query input succeeds
				// without synthesizing a missing expression.
				return parser.pos
			},
			func(source Source[LexicalToken], start int, length int) SyntaxElement {
				tokens := lexicalTokensFromSourceWindow(source, start, length)
				parser := newQueryParserState(tokens)
				return parser.parseQueryBlockStatementList()
			},
		).WithTag("<query-input>"),
	)
}

func commandInputLength(source Source[LexicalToken], start int) int {
	for offset := 0; !source.IsEnd(start + offset); offset++ {
		if source.Peek(start+offset).Kind == SyntaxKindEndOfTextToken {
			return offset
		}
	}
	return -1
}

func commandScriptInputParser(grammar *CommandGrammar) ResultParser[LexicalToken, SyntaxElement] {
	return asResultParser[LexicalToken, SyntaxElement](
		MatchConsumerValue(
			func(source Source[LexicalToken], start int) int {
				length := commandInputLength(source, start)
				if length <= 0 {
					return -1
				}
				return length
			},
			func(source Source[LexicalToken], start int, length int) SyntaxElement {
				tokens := lexicalTokensFromSourceWindow(source, start, length)
				parser := newCommandParserState(tokens, grammar)
				list := parser.parseCommandScriptInputList()
				if list == nil || list.Count() == 0 {
					return nil
				}
				return list
			},
		).WithTag("<script-input>"),
	)
}

func lexicalTokensFromSourceWindow(source Source[LexicalToken], start, length int) []LexicalToken {
	tokens := make([]LexicalToken, 0, length+1)
	for offset := 0; offset < length && !source.IsEnd(start+offset); offset++ {
		token := source.Peek(start + offset)
		if token.Kind == SyntaxKindEndOfTextToken {
			break
		}
		tokens = append(tokens, token)
	}
	if !source.IsEnd(start + length) {
		token := source.Peek(start + length)
		if token.Kind == SyntaxKindEndOfTextToken {
			tokens = append(tokens, token)
			return tokens
		}
	}
	tokens = append(tokens, NewLexicalToken(SyntaxKindEndOfTextToken, "", ""))
	return tokens
}

func ParseCommandBlock(text string) *CommandBlock {
	return ParseCommandBlockTokens(Lex(text))
}

func ParseCommandBlockTokens(tokens []LexicalToken) *CommandBlock {
	return CommandGrammarFrom(nil).ParseCommandBlockTokens(tokens)
}

func (g *CommandGrammar) ParseCommandBlock(text string) *CommandBlock {
	return g.ParseCommandBlockTokens(Lex(text))
}

func (g *CommandGrammar) ParseCommandBlockTokens(tokens []LexicalToken) *CommandBlock {
	return g.parseCommandBlockTokensWithCancellation(tokens, CancellationToken{})
}

func newCommandParserState(tokens []LexicalToken, grammar *CommandGrammar) *queryParserState {
	return newCommandParserStateWithCancellation(tokens, grammar, CancellationToken{})
}

func (g *CommandGrammar) parseCommandBlockTokensWithCancellation(tokens []LexicalToken, cancellationToken CancellationToken) *CommandBlock {
	cancellationToken.ThrowIfCancellationRequested()
	block := newCommandParserStateWithCancellation(tokens, g, cancellationToken).parseCommandBlock()
	cancellationToken.ThrowIfCancellationRequested()
	return block
}

func newCommandParserStateWithCancellation(tokens []LexicalToken, grammar *CommandGrammar, cancellationToken CancellationToken) *queryParserState {
	state := newQueryParserStateWithOptionsAndCancellation(tokens, DefaultParseOptions, cancellationToken)
	state.commandGrammar = grammar
	return state
}

func (g *CommandGrammar) commandBlockResultParser() ResultParser[LexicalToken, *CommandBlock] {
	return queryStateSyntaxParser("<command-block>", func(p *queryParserState) *CommandBlock {
		p.commandGrammar = g
		return p.parseCommandBlock()
	})
}

func IsBetterCommandSyntax(command1 SyntaxElement, command2 SyntaxElement) bool {
	diagnostics1 := GetContainedDiagnostics(command1, DiagnosticsIncludeSyntactic)
	diagnostics2 := GetContainedDiagnostics(command2, DiagnosticsIncludeSyntactic)
	if len(diagnostics1) == 0 && len(diagnostics2) == 0 {
		return false
	}
	if len(diagnostics1) > 0 && len(diagnostics2) == 0 {
		return false
	}
	if len(diagnostics1) == 0 && len(diagnostics2) > 0 {
		return true
	}
	if diagnostics1[0].Start > diagnostics2[0].Start {
		return true
	}
	if diagnostics1[0].Start < diagnostics2[0].Start {
		return false
	}
	return false
}

func UnknownCommandTokenParser() ResultParser[LexicalToken, *SyntaxToken] {
	return asResultParser[LexicalToken, *SyntaxToken](
		MatchValue(isUnknownCommandToken, func(token LexicalToken) *SyntaxToken {
			return SyntaxTokenFromLexical(token)
		}),
	)
}

func CommandLimiterParser() Parser[LexicalToken] {
	return MatchConsumer(func(source Source[LexicalToken], start int) int {
		return commandLimiterLength(source, start)
	})
}

func commandLimiterLength(source Source[LexicalToken], start int) int {
	pos := start + 1
	for !source.IsEnd(pos) && !IsPossibleStartOfCommand(source.Peek(pos)) {
		pos++
	}
	return pos - start
}

func IsPossibleStartOfCommand(token LexicalToken, isFirstToken ...bool) bool {
	if token.Kind != SyntaxKindDotToken {
		return false
	}
	first := len(isFirstToken) > 0 && isFirstToken[0]
	if HasLineBreaks(token.Trivia) {
		lastLineStart := LastLineBreakEnd(token.Trivia)
		return lastLineStart >= 0 && IsWhitespaceOnly(token.Trivia, lastLineStart, len(token.Trivia)-lastLineStart)
	}
	if first && (token.Trivia == "" || IsWhitespaceOnly(token.Trivia)) {
		return true
	}
	return false
}

func CreateMissingCommandStatement(Source[LexicalToken], int) Statement {
	return createMissingCommandStatement()
}

func (p *queryParserState) parseCommandBlock() *CommandBlock {
	directives := p.parseDirectiveList()
	statements := p.parseCommandBlockStatementList()
	skipped := p.parseCommandBlockSkippedTokens()
	end := p.parseToken(SyntaxKindEndOfTextToken)
	return MustNewCommandBlockTyped(directives, statements, skipped, end)
}

func (p *queryParserState) parseCommandBlockStatementList() *SeparatedSyntaxListOf[Statement] {
	var elements []SyntaxElement
	first := p.parseCommandStatement()
	var statement Statement
	if first == nil {
		statement = createMissingCommandStatement()
	} else {
		statement = first
	}

	for {
		semicolon := p.parseToken(SyntaxKindSemicolonToken)
		if semicolon == nil {
			elements = append(elements, NewSeparatedElement(statement, nil))
			break
		}

		elements = append(elements, NewSeparatedElement(statement, semicolon))
		if p.atEndOfQuery() {
			break
		}

		next := p.parseQueryBlockStatement()
		if next == nil {
			if p.peekToken().Kind == SyntaxKindSemicolonToken {
				next = createMissingCommandStatement()
			} else {
				break
			}
		}
		statement = next
	}

	return AsSeparatedSyntaxListOf[Statement](NewTypedSyntaxList("SeparatedElement<Statement>", elements...))
}

func (p *queryParserState) parseCommandScriptInputList() *SyntaxList {
	var elements []SyntaxElement
	for !p.atEndOfQuery() {
		start := p.pos
		limit := commandLimiterLength(NewArraySource(p.tokens), start)
		if limit <= 0 {
			break
		}
		segmentTokens := lexicalTokensFromSliceWindow(p.tokens, start, limit)
		segmentParser := newCommandParserState(segmentTokens, p.commandGrammar)
		expr := segmentParser.parseCommandPipeExpression()
		if expr == nil {
			break
		}
		semicolon := segmentParser.parseToken(SyntaxKindSemicolonToken)
		elements = append(elements, NewSeparatedElement(expr, semicolon))
		consumed := segmentParser.pos
		if consumed <= 0 {
			break
		}
		p.pos = start + consumed
	}
	return NewTypedSyntaxList("SeparatedElement<Expression>", elements...)
}

func lexicalTokensFromSliceWindow(tokens []LexicalToken, start, length int) []LexicalToken {
	end := start + length
	if start < 0 {
		start = 0
	}
	if end > len(tokens) {
		end = len(tokens)
	}
	window := append([]LexicalToken{}, tokens[start:end]...)
	if len(window) == 0 || window[len(window)-1].Kind != SyntaxKindEndOfTextToken {
		window = append(window, NewLexicalToken(SyntaxKindEndOfTextToken, "", ""))
	}
	return window
}

func (p *queryParserState) parseCommandStatement() *ExpressionStatement {
	expr := p.parseCommandPipeExpression()
	if expr == nil {
		return nil
	}
	return MustNewExpressionStatementTyped(expr)
}

func (p *queryParserState) parseCommandPipeExpression() Expression {
	command := p.parseCommand()
	if command == nil {
		return nil
	}
	expr, ok := command.(Expression)
	if !ok {
		return nil
	}
	for p.peekToken().Kind == SyntaxKindBarToken {
		bar := p.parseToken()
		operator := p.parseRequiredQueryOperator()
		expr = MustNewPipeExpressionTyped(expr, bar, operator)
	}
	return expr
}

func (p *queryParserState) parseCommand() Command {
	if command := p.parseKnownCommand(); command != nil {
		return command
	}

	dot := p.parseToken(SyntaxKindDotToken)
	if dot == nil {
		return nil
	}

	var parts []SyntaxElement
	for isUnknownCommandToken(p.peekToken()) {
		parts = append(parts, p.parseToken())
	}
	if len(parts) > 0 {
		return MustNewUnknownCommandTyped(dot, AsSyntaxListOf[SyntaxElement](NewTypedSyntaxList("SyntaxElement", parts...)))
	}
	return MustNewBadCommandTyped(dot, DiagnosticMissingCommand())
}

func (p *queryParserState) parseKnownCommand() Command {
	grammar := p.commandGrammar
	if grammar == nil || len(grammar.commandParsers) == 0 {
		return nil
	}

	start := p.pos
	var best Command
	bestEnd := -1
	for _, info := range grammar.commandParsers {
		p.pos = start
		command := info.Parse(p)
		if command != nil && (p.pos > bestEnd || (p.pos == bestEnd && IsBetterCommandSyntax(command, best))) {
			best = command
			bestEnd = p.pos
		}
	}
	if grammar.allowPartialKnownCommand {
		p.pos = start
		partial := p.parsePartialKnownCommand(grammar.commandParsers)
		partialEnd := p.pos
		if partial != nil && partialEnd > bestEnd {
			return p.withCommandSkippedTokens(partial)
		}
	}
	if best != nil {
		p.pos = bestEnd
		return p.withCommandSkippedTokens(best)
	}
	p.pos = start
	return nil
}

type commandPartialMatch struct {
	end       int
	parts     []SyntaxElement
	expected  []string
	recovered bool
}

func (p *queryParserState) parsePartialKnownCommand(infos []commandParserInfo) Command {
	return p.parsePartialKnownCommandWithParsers()
}

func (p *queryParserState) parsePartialKnownCommandWithParsers() Command {
	start := p.pos
	grammar := p.commandGrammar
	if grammar == nil {
		return nil
	}
	parsers := grammar.partialCommandParsers()
	if len(parsers) == 0 {
		return nil
	}

	source := NewArraySource(p.tokens)
	length := ScanPartialBest(parsers, source, start)
	// Upstream lets a lone dot fall through to BadCommand/UnknownCommand.
	if length < 2 {
		p.pos = start
		return nil
	}

	output := make([]any, 0, length+1)
	bestParsers := make([]Parser[LexicalToken], 0, 1)
	consumed := ParsePartialBest(parsers, source, start, &output, 0, &bestParsers, commandPartialOnFailure)
	if consumed < 2 {
		p.pos = start
		return nil
	}

	elements := make([]SyntaxElement, 0, len(output))
	for _, value := range output {
		if element, ok := value.(SyntaxElement); ok && element != nil {
			elements = append(elements, element)
		}
	}
	if len(elements) == 0 {
		p.pos = start
		return nil
	}
	dot, ok := elements[0].(*SyntaxToken)
	if !ok || dot.Kind() != SyntaxKindDotToken {
		p.pos = start
		return nil
	}

	kinds := make([]string, 0, len(bestParsers))
	for _, parser := range bestParsers {
		if parser != nil && parser.Tag() != "" {
			kinds = append(kinds, parser.Tag())
		}
	}
	p.pos = start + consumed
	return MustNewPartialCommandTyped(kinds, dot, AsSyntaxListOf[SyntaxElement](NewTypedSyntaxList("SyntaxElement", elements[1:]...)))
}

func commandPartialOnFailure(parser Parser[LexicalToken], output *[]any) {
	expected := CommandParserExpectedTerms(parser)
	appendOutput(output, NewMissingToken(SyntaxKindIdentifierToken, DiagnosticTermsExpected(expected...)))
}

func (p *queryParserState) parsePartialKnownCommandWithDescriptors(infos []commandParserInfo) Command {
	start := p.pos
	dot := p.parseToken(SyntaxKindDotToken)
	if dot == nil {
		return nil
	}

	var bestParts []SyntaxElement
	var bestKinds []string
	var bestExpected []string
	bestEnd := -1
	for _, info := range infos {
		matches := p.parsePartialCommandElementList(info.Elements, start+1)
		match, ok := bestCommandPartialMatch(matches)
		if !ok || match.end <= start+1 {
			continue
		}
		if match.end > bestEnd {
			bestEnd = match.end
			bestParts = match.parts
			bestKinds = []string{info.Kind}
			bestExpected = append([]string{}, match.expected...)
		} else if match.end == bestEnd {
			bestKinds = append(bestKinds, info.Kind)
			bestExpected = append(bestExpected, match.expected...)
		}
	}

	if bestEnd < 0 {
		p.pos = start
		return nil
	}
	p.pos = bestEnd
	bestExpected = distinctSortedStrings(bestExpected)
	if len(bestExpected) > 0 {
		bestParts = append(bestParts, createMissingCommandPartialToken(bestExpected))
	}
	return MustNewPartialCommandTyped(bestKinds, dot, AsSyntaxListOf[SyntaxElement](NewTypedSyntaxList("SyntaxElement", bestParts...)))
}

func (p *queryParserState) parsePartialCommandElementList(elements []commandTokenElement, start int) []commandPartialMatch {
	type state struct {
		end       int
		parts     []SyntaxElement
		recovered bool
	}
	states := []state{{end: start}}
	var failures []commandPartialMatch
	for i := 0; i < len(elements); {
		element := elements[i]
		if element.OptionalGroup != 0 {
			groupEnd := i + 1
			for groupEnd < len(elements) && elements[groupEnd].OptionalGroup == element.OptionalGroup {
				groupEnd++
			}
			children := cloneCommandTokenElements(elements[i:groupEnd])
			for index := range children {
				children[index].OptionalGroup = 0
			}
			element = commandTokenElement{Children: children, Optional: true}
			i = groupEnd
		} else {
			i++
		}

		var next []state
		for _, current := range states {
			completed, partial := p.parsePartialCommandElement(element, current.end)
			maxCompletedEnd := -1
			for _, match := range completed {
				if match.end > maxCompletedEnd {
					maxCompletedEnd = match.end
				}
			}
			for _, match := range partial {
				match.parts = appendCommandPartialParts(current.parts, match.parts)
				match.recovered = match.recovered || current.recovered
				failures = append(failures, match)
				if element.Optional && match.end > current.end && len(match.expected) > 0 && match.end > maxCompletedEnd {
					parts := appendCommandPartialParts(match.parts, []SyntaxElement{createMissingCommandPartialToken(distinctSortedStrings(append([]string{}, match.expected...)))})
					next = append(next, state{end: match.end, parts: parts, recovered: true})
				}
			}
			for _, match := range completed {
				next = append(next, state{
					end:       match.end,
					parts:     appendCommandPartialParts(current.parts, match.parts),
					recovered: current.recovered || match.recovered,
				})
			}
		}
		states = next
		if len(states) == 0 {
			break
		}
	}
	for _, current := range states {
		if current.recovered {
			failures = append(failures, commandPartialMatch{end: current.end, parts: current.parts, recovered: true})
		}
	}
	return failures
}

func (p *queryParserState) parsePartialCommandElement(element commandTokenElement, start int) (completed, failures []commandPartialMatch) {
	if len(element.Forbidden) > 0 {
		old := p.pos
		p.pos = start
		forbidden := p.matchesForbiddenCommandToken(element.Forbidden)
		p.pos = old
		if forbidden {
			if element.Optional {
				return []commandPartialMatch{{end: start}}, nil
			}
			return nil, []commandPartialMatch{{end: start, expected: p.commandElementExpectedTerms(element)}}
		}
	}

	base := element
	base.Optional = false
	switch {
	case len(base.Choices) > 0:
		for _, choice := range base.Choices {
			choiceCompleted, choiceFailures := p.parsePartialCommandElement(choice, start)
			completed = append(completed, choiceCompleted...)
			failures = append(failures, choiceFailures...)
		}
	case len(base.Children) > 0:
		for _, match := range p.parsePartialCommandElementList(base.Children, start) {
			if match.recovered && len(match.expected) == 0 {
				completed = append(completed, match)
			} else {
				failures = append(failures, match)
			}
		}
		old := p.pos
		p.pos = start
		child := p.parseCommandNestedCustomElement(base.Children)
		if child != nil {
			completed = append(completed, commandPartialMatch{end: p.pos, parts: flattenCommandPartialElement(child)})
		}
		p.pos = old
	case len(base.List) > 0:
		completed, failures = p.parsePartialCommandList(base, start)
	case len(base.Alternatives) > 0:
		old := p.pos
		p.pos = start
		if token := p.parseCommandTokenAlternatives(base.Alternatives); token != nil {
			completed = append(completed, commandPartialMatch{end: p.pos, parts: []SyntaxElement{token}})
		} else {
			failures = append(failures, commandPartialMatch{end: start, expected: p.commandElementExpectedTerms(base)})
		}
		p.pos = old
	case base.Rule != "":
		old := p.pos
		p.pos = start
		if child := p.parseCommandRuleElement(base.Rule); child != nil {
			completed = append(completed, commandPartialMatch{end: p.pos, parts: []SyntaxElement{child}})
		} else {
			failures = append(failures, commandPartialMatch{end: start, expected: p.commandElementExpectedTerms(base)})
		}
		p.pos = old
	}
	if element.Optional {
		completed = append(completed, commandPartialMatch{end: start})
	}
	return completed, failures
}

func (p *queryParserState) parsePartialCommandList(element commandTokenElement, start int) (completed, failures []commandPartialMatch) {
	if len(element.List) != 1 {
		return nil, []commandPartialMatch{{end: start, expected: []string{"<token>"}}}
	}
	item := element.List[0]
	first, firstFailures := p.parsePartialCommandElement(item, start)
	failures = append(failures, firstFailures...)
	if element.ListAllowEmpty {
		completed = append(completed, commandPartialMatch{end: start})
	}
	for _, match := range first {
		tailCompleted, tailFailures := p.parsePartialCommandListTail(element, item, match)
		completed = append(completed, tailCompleted...)
		failures = append(failures, tailFailures...)
	}
	return completed, failures
}

func (p *queryParserState) parsePartialCommandListTail(list, item commandTokenElement, current commandPartialMatch) (completed, failures []commandPartialMatch) {
	if list.ListSeparator == "" {
		next, nextFailures := p.parsePartialCommandElement(item, current.end)
		if len(next) == 0 {
			return []commandPartialMatch{current}, nextFailures
		}
		for _, match := range next {
			if match.end == current.end {
				completed = append(completed, current)
				continue
			}
			match.parts = appendCommandPartialParts(current.parts, match.parts)
			tailCompleted, tailFailures := p.parsePartialCommandListTail(list, item, match)
			completed = append(completed, tailCompleted...)
			failures = append(failures, tailFailures...)
		}
		return completed, failures
	}

	old := p.pos
	p.pos = current.end
	separator := p.parseCommandTokenAlternatives([]string{list.ListSeparator})
	separatorEnd := p.pos
	p.pos = old
	if separator == nil {
		return []commandPartialMatch{current}, nil
	}
	next, nextFailures := p.parsePartialCommandElement(item, separatorEnd)
	prefix := appendCommandPartialParts(current.parts, []SyntaxElement{separator})
	for _, failure := range nextFailures {
		failure.parts = appendCommandPartialParts(prefix, failure.parts)
		failures = append(failures, failure)
	}
	if len(next) == 0 {
		return nil, failures
	}
	for _, match := range next {
		if match.end == separatorEnd {
			continue
		}
		match.parts = appendCommandPartialParts(prefix, match.parts)
		tailCompleted, tailFailures := p.parsePartialCommandListTail(list, item, match)
		completed = append(completed, tailCompleted...)
		failures = append(failures, tailFailures...)
	}
	if len(completed) == 0 && len(failures) == 0 {
		return []commandPartialMatch{current}, nil
	}
	return completed, failures
}

func appendCommandPartialParts(left, right []SyntaxElement) []SyntaxElement {
	parts := make([]SyntaxElement, 0, len(left)+len(right))
	parts = append(parts, left...)
	parts = append(parts, right...)
	return parts
}

func flattenCommandPartialElement(element SyntaxElement) []SyntaxElement {
	custom, ok := element.(*CustomNode)
	if !ok {
		return []SyntaxElement{element}
	}
	var parts []SyntaxElement
	for i := 0; i < custom.ChildCount(); i++ {
		if child := custom.GetChild(i); child != nil {
			parts = append(parts, flattenCommandPartialElement(cloneSyntaxElementPreservingGeneratedWrapper(child, true))...)
		}
	}
	return parts
}

func bestCommandPartialMatch(matches []commandPartialMatch) (commandPartialMatch, bool) {
	best := commandPartialMatch{end: -1}
	for _, match := range matches {
		if match.end > best.end || (match.end == best.end && isBetterCommandPartialMatch(match, best)) {
			best = match
		} else if match.end == best.end && len(match.expected) > 0 && len(best.expected) > 0 && match.recovered == best.recovered {
			best.expected = append(best.expected, match.expected...)
		}
	}
	return best, best.end >= 0
}

func isBetterCommandPartialMatch(match, best commandPartialMatch) bool {
	if len(match.expected) > 0 && len(best.expected) == 0 {
		return true
	}
	if len(match.expected) == 0 && len(best.expected) > 0 {
		return false
	}
	if match.recovered != best.recovered {
		return !match.recovered
	}
	return false
}

func commandElementExpectedTerms(element commandTokenElement) []string {
	if len(element.Alternatives) > 0 {
		return append([]string{}, element.Alternatives...)
	}
	if element.Rule != "" {
		return []string{commandRuleExpectedTerm(element)}
	}
	return []string{"<token>"}
}

func (p *queryParserState) commandElementExpectedTerms(element commandTokenElement) []string {
	if len(element.Alternatives) > 0 {
		return CommandParserExpectedTerms(commandPartialTokenTextsParser(element.Alternatives))
	}
	if element.Rule != "" && p.commandGrammar != nil {
		if parser := p.commandRuleParser(element.Rule); parser != nil {
			return CommandParserExpectedTerms(parser)
		}
	}
	return commandElementExpectedTerms(element)
}

func commandRuleExpectedTerm(element commandTokenElement) string {
	if element.Expected != "" {
		return element.Expected
	}
	terms := map[string]string{
		"NameDeclaration": "<name>", "TableNameReference": "<table>",
		"MaterializedViewNameReference": "<materializedview>", "FunctionNameReference": "<function>",
		"QueryInput": "<query-input>", "ScriptInput": "<script-input>", "StringLiteral": "<string-literal>",
	}
	if term := terms[element.Rule]; term != "" {
		return term
	}
	return "<" + strings.ToLower(element.Rule) + ">"
}

func distinctSortedStrings(values []string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		if value != "" {
			seen[value] = true
		}
	}
	values = values[:0]
	for value := range seen {
		values = append(values, value)
	}
	slices.Sort(values)
	return values
}

func createMissingCommandPartialToken(expected []string) *SyntaxToken {
	if len(expected) == 0 {
		expected = []string{"<token>"}
	}
	if len(expected) == 1 && isCommandPartialPunctuation(expected[0]) {
		expected = []string{"'" + expected[0] + "'"}
	}
	return NewMissingToken(SyntaxKindIdentifierToken, DiagnosticTermsExpected(expected...))
}

func isCommandPartialPunctuation(term string) bool {
	if term == "" || strings.HasPrefix(term, "<") {
		return false
	}
	if len(term) >= 2 && strings.HasPrefix(term, "'") && strings.HasSuffix(term, "'") {
		return false
	}
	for _, ch := range term {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			return false
		}
	}
	return true
}

func (p *queryParserState) withCommandSkippedTokens(command Command) Command {
	var tokens []SyntaxElement
	for isUnknownCommandToken(p.peekToken()) {
		tokens = append(tokens, p.parseToken())
	}
	if len(tokens) == 0 {
		return command
	}
	var diagnostics []Diagnostic
	first, _ := tokens[0].(*SyntaxToken)
	if p.commandSkippedTokensContinueLongerProduction(command, first) {
		diagnostics = append(diagnostics, DiagnosticIncompleteFragment())
	}
	return MustNewCommandAndSkippedTokensTyped(
		command,
		MustNewSkippedTokensTyped(AsSyntaxListOf[*SyntaxToken](NewTypedSyntaxList("SyntaxToken", tokens...)), diagnostics...),
	)
}

func (p *queryParserState) commandSkippedTokensContinueLongerProduction(command Command, first *SyntaxToken) bool {
	custom, ok := command.(*CustomCommand)
	if !ok || custom == nil || first == nil || p.commandGrammar == nil {
		return false
	}
	var base []commandTokenElement
	for _, info := range p.commandGrammar.commandParsers {
		if info.Kind == custom.CommandKind() {
			base = info.Elements
			break
		}
	}
	if len(base) == 0 {
		return false
	}
	firstText := first.ValueText()
	for _, candidate := range p.commandGrammar.commandParsers {
		if len(candidate.Elements) <= len(base) || !commandElementPrefixMatches(base, candidate.Elements) {
			continue
		}
		for _, alternative := range candidate.Elements[len(base)].Alternatives {
			if ordinalEqualFold(firstText, alternative) {
				return true
			}
		}
	}
	return false
}

func commandElementPrefixMatches(prefix, candidate []commandTokenElement) bool {
	if len(prefix) > len(candidate) {
		return false
	}
	for i := range prefix {
		if prefix[i].Rule != candidate[i].Rule || len(prefix[i].Alternatives) != len(candidate[i].Alternatives) {
			return false
		}
		for j := range prefix[i].Alternatives {
			if prefix[i].Alternatives[j] != candidate[i].Alternatives[j] {
				return false
			}
		}
	}
	return true
}

func showVersionCommandParserInfo() commandParserInfo {
	return fixedTokenCommandParserInfo("ShowVersion", []string{"show", "version"})
}

func fixedTokenCommandParserInfo(kind string, tokens []string) commandParserInfo {
	elements := make([]commandTokenElement, 0, len(tokens))
	for _, token := range tokens {
		elements = append(elements, commandTokenElement{Alternatives: []string{token}})
	}
	return tokenCommandParserInfo(kind, elements)
}

func tokenCommandParserInfo(kind string, elements []commandTokenElement) commandParserInfo {
	copiedElements := cloneCommandTokenElements(elements)
	tokens := make([]string, 0, len(copiedElements))
	for _, element := range copiedElements {
		if len(element.Alternatives) > 0 {
			tokens = append(tokens, element.Alternatives[0])
		}
	}
	copiedTokens := append([]string{}, tokens...)
	return commandParserInfo{
		Kind:     kind,
		Tokens:   copiedTokens,
		Elements: cloneCommandTokenElements(copiedElements),
		Parse: func(p *queryParserState) Command {
			return p.parseTokenCommand(kind, copiedElements)
		},
	}
}

func cloneCommandTokenElements(elements []commandTokenElement) []commandTokenElement {
	copied := make([]commandTokenElement, len(elements))
	for i, element := range elements {
		copied[i] = commandTokenElement{
			Alternatives:         append([]string{}, element.Alternatives...),
			Rule:                 element.Rule,
			Expected:             element.Expected,
			Forbidden:            append([]string{}, element.Forbidden...),
			Choices:              cloneCommandTokenElements(element.Choices),
			Children:             cloneCommandTokenElements(element.Children),
			List:                 cloneCommandTokenElements(element.List),
			ListSeparator:        element.ListSeparator,
			ListAllowEmpty:       element.ListAllowEmpty,
			ListMissingElement:   element.ListMissingElement,
			ListMissingSeparator: element.ListMissingSeparator,
			Name:                 element.Name,
			Hint:                 element.Hint,
			HasHint:              element.HasHint,
			Hidden:               element.Hidden,
			ShapeOptional:        element.ShapeOptional,
			Required:             element.Required,
			OptionalGroup:        element.OptionalGroup,
			Optional:             element.Optional,
		}
	}
	return copied
}

func (p *queryParserState) parseTokenCommand(kind string, elements []commandTokenElement) Command {
	start := p.pos
	fail := func() Command {
		p.pos = start
		return nil
	}
	dot := p.parseToken(SyntaxKindDotToken)
	if dot == nil {
		return fail()
	}
	children, shape, ok := p.parseCommandElementList(elements)
	if !ok {
		return fail()
	}
	custom, err := NewCustomNodeWithShape(shape, children)
	if err != nil {
		return fail()
	}
	command := MustNewCustomCommandTyped(kind, dot, custom)
	p.attachCustomCommandSemanticInfo(command)
	return command
}

func (p *queryParserState) parseCommandElementList(elements []commandTokenElement) ([]SyntaxElement, []CustomElementDescriptor, bool) {
	start := p.pos
	children := make([]SyntaxElement, 0, len(elements))
	shape := CustomElementDescriptorDefaultShape(len(elements))
	for i := 0; i < len(elements); {
		element := elements[i]
		shape[i] = commandElementDescriptor(element)
		if element.OptionalGroup != 0 {
			groupEnd := i + 1
			for groupEnd < len(elements) && elements[groupEnd].OptionalGroup == element.OptionalGroup {
				groupEnd++
			}
			for j := i; j < groupEnd; j++ {
				shape[j] = commandElementDescriptor(elements[j])
			}
			groupChildren, ok := p.parseCommandElementGroup(elements[i:groupEnd])
			if !ok {
				for j := i; j < groupEnd; j++ {
					if !elements[j].Optional && !elements[j].ShapeOptional {
						goto fail
					}
					shape[j] = commandElementDescriptor(elements[j])
					shape[j].Optional = true
					children = append(children, nil)
				}
			} else {
				children = append(children, groupChildren...)
			}
			i = groupEnd
			continue
		}
		child := p.parseCommandElement(element)
		if child == nil {
			if !element.Optional {
				if element.Required && p.pos == start {
					goto fail
				}
				if recovered := p.recoverMissingCommandElement(element); recovered != nil {
					children = append(children, recovered)
					i++
					continue
				}
				goto fail
			}
			shape[i].Optional = true
			children = append(children, nil)
			i++
			continue
		}
		children = append(children, child)
		i++
	}
	return children, shape, true

fail:
	p.pos = start
	return nil, nil, false
}

func (p *queryParserState) recoverMissingCommandElement(element commandTokenElement) SyntaxElement {
	if !element.Required {
		if !p.atEndOfQuery() {
			return nil
		}
		for _, text := range element.Alternatives {
			if text == "]" {
				return CreateMissingSyntaxTokenForText(text, DiagnosticTermsExpected(text))
			}
		}
		return nil
	}
	if element.Rule == "QueryInput" {
		return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone, DiagnosticMissingExpression())
	}
	if element.Rule == "ScriptInput" {
		return MustNewExpressionStatementTyped(
			MustNewBadCommandTyped(createOmittedToken(SyntaxKindDotToken)),
			DiagnosticMissingStatement(),
		)
	}
	if element.Rule != "" {
		return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone, DiagnosticMissingName())
	}
	if len(element.Alternatives) > 0 {
		text := element.Alternatives[0]
		return CreateMissingSyntaxTokenForText(text, DiagnosticTermsExpected(text))
	}
	if len(element.Choices) > 0 {
		choice := element.Choices[0]
		choice.Required = true
		return p.recoverMissingCommandElement(choice)
	}
	if len(element.Children) > 0 {
		children := make([]SyntaxElement, 0, len(element.Children))
		shape := CustomElementDescriptorDefaultShape(len(element.Children))
		for i, child := range element.Children {
			shape[i] = commandElementDescriptor(child)
			child.Required = true
			children = append(children, p.recoverMissingCommandElement(child))
		}
		custom, err := NewCustomNodeWithShape(shape, children)
		if err == nil {
			return custom
		}
	}
	return nil
}

func commandElementDescriptor(element commandTokenElement) CustomElementDescriptor {
	descriptor := DefaultCustomElementDescriptor
	if element.Name != "" {
		descriptor.Name = element.Name
	}
	if element.HasHint {
		descriptor.Hint = element.Hint
	}
	if element.ShapeOptional {
		descriptor.Optional = true
	}
	return descriptor
}

func (p *queryParserState) attachCustomCommandSemanticInfo(command *CustomCommand) {
	if p == nil || p.commandGrammar == nil || p.commandGrammar.Globals == nil || command == nil {
		return
	}
	var symbol *CommandSymbol
	if p.commandGrammar.commandSymbols != nil {
		symbol = p.commandGrammar.commandSymbols[command.CommandKind()]
	} else {
		symbol = p.commandGrammar.Globals.GetCommand(command.CommandKind())
	}
	if symbol == nil {
		return
	}
	var result *TypeSymbol
	if table := symbol.ResultType(); table != nil {
		result = &table.TypeSymbol
	}
	command.SetSemanticInfo(NewSemanticInfoForSymbol(symbol, result))
}

func (p *queryParserState) parseCommandElement(element commandTokenElement) SyntaxElement {
	if len(element.Forbidden) > 0 && p.matchesForbiddenCommandToken(element.Forbidden) {
		return nil
	}
	if len(element.Choices) > 0 {
		return p.parseCommandChoiceElement(element.Choices)
	}
	if len(element.Children) > 0 {
		return p.parseCommandNestedCustomElement(element.Children)
	}
	if len(element.List) > 0 {
		return p.parseCommandListElement(element)
	}
	if len(element.Alternatives) > 0 {
		token := p.parseCommandTokenAlternatives(element.Alternatives)
		if token == nil {
			return nil
		}
		return token
	}
	if element.Rule != "" {
		return p.parseCommandRuleElement(element.Rule)
	}
	return nil
}

func (p *queryParserState) parseCommandChoiceElement(choices []commandTokenElement) SyntaxElement {
	start := p.pos
	var best SyntaxElement
	bestEnd := -1
	for _, choice := range choices {
		p.pos = start
		child := p.parseCommandElement(choice)
		if child != nil && p.pos > bestEnd {
			best = child
			bestEnd = p.pos
		}
	}
	if best == nil {
		p.pos = start
		return nil
	}
	p.pos = bestEnd
	return best
}

func (p *queryParserState) parseCommandNestedCustomElement(elements []commandTokenElement) SyntaxElement {
	start := p.pos
	children, shape, ok := p.parseCommandElementList(elements)
	if !ok {
		p.pos = start
		return nil
	}
	custom, err := NewCustomNodeWithShape(shape, children)
	if err != nil {
		panic(err)
	}
	return custom
}

func (p *queryParserState) parseCommandListElement(list commandTokenElement) SyntaxElement {
	if len(list.List) != 1 {
		return nil
	}
	if list.ListSeparator == "" {
		return p.parseCommandUnseparatedListElement(list.List[0], list.ListAllowEmpty)
	}
	return p.parseCommandSeparatedListElement(list.List[0], list.ListSeparator, list.ListAllowEmpty, list.ListMissingElement, list.ListMissingSeparator)
}

func (p *queryParserState) parseCommandUnseparatedListElement(element commandTokenElement, allowEmpty bool) SyntaxElement {
	start := p.pos
	first := p.parseCommandElement(element)
	if first == nil {
		p.pos = start
		if allowEmpty {
			return NewTypedSyntaxList("SyntaxElement")
		}
		return nil
	}

	items := []SyntaxElement{first}
	for {
		itemStart := p.pos
		next := p.parseCommandElement(element)
		if next == nil || p.pos == itemStart {
			p.pos = itemStart
			break
		}
		items = append(items, next)
	}
	return NewTypedSyntaxList("SyntaxElement", items...)
}

func (p *queryParserState) parseCommandSeparatedListElement(element commandTokenElement, separator string, allowEmpty, missingElement, missingSeparator bool) SyntaxElement {
	start := p.pos
	first := p.parseCommandElement(element)
	if first == nil {
		p.pos = start
		if allowEmpty {
			return NewTypedSyntaxList("SeparatedElement<SyntaxElement>")
		}
		return nil
	}

	separated := make([]SyntaxElement, 0, 1)
	current := first
	for {
		separatorStart := p.pos
		separatorToken := p.parseCommandTokenAlternatives([]string{separator})
		if separatorToken == nil {
			if !missingSeparator {
				separated = append(separated, NewSeparatedElement(current, nil))
				break
			}
			itemStart := p.pos
			next := p.parseCommandElement(element)
			if next == nil || p.pos == itemStart {
				p.pos = itemStart
				separated = append(separated, NewSeparatedElement(current, nil))
				break
			}
			separated = append(separated, NewSeparatedElement(current, CreateMissingSyntaxTokenForText(separator)))
			current = next
			continue
		}
		next := p.parseCommandElement(element)
		if next == nil {
			if missingElement {
				separated = append(separated, NewSeparatedElement(current, separatorToken))
				current = createMissingCommandListElement(element)
				continue
			}
			p.pos = separatorStart
			separated = append(separated, NewSeparatedElement(current, nil))
			break
		}
		separated = append(separated, NewSeparatedElement(current, separatorToken))
		current = next
	}
	return NewTypedSyntaxList("SeparatedElement<SyntaxElement>", separated...)
}

func createMissingCommandListElement(element commandTokenElement) SyntaxElement {
	if len(element.Children) > 0 {
		children := make([]SyntaxElement, 0, len(element.Children))
		shape := CustomElementDescriptorDefaultShape(len(element.Children))
		for i, child := range element.Children {
			shape[i] = commandElementDescriptor(child)
			children = append(children, createMissingCommandListElement(child))
		}
		custom, err := NewCustomNodeWithShape(shape, children)
		if err != nil {
			return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone, DiagnosticMissingParameter())
		}
		return custom
	}
	if len(element.Choices) > 0 {
		return createMissingCommandListElement(element.Choices[0])
	}
	if len(element.List) > 0 {
		return NewTypedSyntaxListWithDiagnostics("SeparatedElement<SyntaxElement>", nil, DiagnosticMissingParameter())
	}
	if len(element.Alternatives) > 0 {
		return CreateMissingSyntaxTokenForTexts(element.Alternatives)
	}
	if element.Rule != "" {
		switch element.Rule {
		case "Value":
			return createMissingValueExpression()
		case "StringLiteral", "BracketedStringLiteral", "InputText", "BracketedInputText":
			return createMissingStringLiteral()
		case "NameDeclaration", "QualifiedNameDeclaration", "WildcardedNameDeclaration", "QualifiedWildcardedNameDeclaration":
			return createMissingNameDeclaration()
		case "TableNameReference", "ExternalTableNameReference", "MaterializedViewNameReference", "FunctionNameReference", "EntityGroupNameReference", "EntityGroups", "DatabaseNameReference", "ClusterNameReference", "GraphModelNameReference", "GraphSnapshotNameReference", "GraphModelSnapshotNameReference", "DatabaseOrTableNameReference", "DatabaseOrTableOrColumnNameReference", "TableOrColumnNameReference", "DatabaseTableNameReference", "DatabaseTableColumnNameReference", "TableColumnNameReference", "DatabaseExternalTableNameReference", "DatabaseMaterializedViewNameReference", "DatabaseFunctionNameReference", "DatabaseEntityGroupNameReference", "ColumnNameReference":
			return createMissingNameReferenceNode()
		case "Type":
			return createMissingType()
		default:
			return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone, DiagnosticMissingParameter())
		}
	}
	return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone, DiagnosticMissingParameter())
}

func (p *queryParserState) parseCommandElementGroup(elements []commandTokenElement) ([]SyntaxElement, bool) {
	start := p.pos
	children := make([]SyntaxElement, 0, len(elements))
	consumed := false
	for _, element := range elements {
		elementStart := p.pos
		child := p.parseCommandElement(element)
		if child == nil {
			if element.Optional {
				children = append(children, nil)
				continue
			}
			p.pos = start
			return nil, false
		}
		if p.pos != elementStart {
			consumed = true
		}
		children = append(children, child)
	}
	if !consumed {
		p.pos = start
		return nil, false
	}
	return children, true
}

func (p *queryParserState) matchesForbiddenCommandToken(forbidden []string) bool {
	source := NewArraySource(p.tokens)
	for _, text := range forbidden {
		if matchesLexicalTextIgnoreCase(source, p.pos, text) > 0 {
			return true
		}
	}
	return false
}

func (p *queryParserState) parseCommandTokenAlternatives(alternatives []string) *SyntaxToken {
	start := p.pos
	for _, text := range alternatives {
		p.pos = start
		if token := p.parseCommandTokenText(text); token != nil {
			return token
		}
	}
	p.pos = start
	return nil
}

func (p *queryParserState) parseCommandRuleElement(rule string) SyntaxElement {
	parser := p.commandRuleParser(rule)
	if parser == nil {
		return nil
	}
	result := parser.ParseValue(NewArraySource(p.tokens), p.pos)
	if result.Length < 0 || isNilSyntaxElement(result.Value) {
		return nil
	}
	p.pos += result.Length
	return result.Value
}

func (p *queryParserState) commandRuleParser(rule string) ResultParser[LexicalToken, SyntaxElement] {
	if p == nil || p.commandGrammar == nil || p.commandGrammar.PredefinedRules == nil {
		return nil
	}
	rules := p.commandGrammar.PredefinedRules
	switch rule {
	case "Value":
		return rules.Value
	case "StringLiteral":
		return rules.StringLiteral
	case "BracketedStringLiteral":
		return rules.BracketedStringLiteral
	case "RawGuidLiteral":
		return rules.RawGuidLiteral
	case "GuidLiteral":
		return rules.GuidLiteral
	case "AnyGuidLiteralOrString":
		return rules.AnyGuidLiteralOrString
	case "Type":
		return rules.Type
	case "NameDeclaration":
		return rules.NameDeclaration
	case "QualifiedNameDeclaration":
		return rules.QualifiedNameDeclaration
	case "WildcardedNameDeclaration":
		return rules.WildcardedNameDeclaration
	case "QualifiedWildcardedNameDeclaration":
		return rules.QualifiedWildcardedNameDeclaration
	case "ColumnNameReference":
		return rules.ColumnNameReference
	case "TableNameReference":
		return rules.TableNameReference
	case "ExternalTableNameReference":
		return rules.ExternalTableNameReference
	case "MaterializedViewNameReference":
		return rules.MaterializedViewNameReference
	case "FunctionNameReference":
		return rules.FunctionNameReference
	case "EntityGroupNameReference", "EntityGroups":
		return rules.EntityGroupNameReference
	case "DatabaseNameReference":
		return rules.DatabaseNameReference
	case "ClusterNameReference":
		return rules.ClusterNameReference
	case "GraphModelNameReference":
		return rules.GraphModelNameReference
	case "GraphSnapshotNameReference":
		return rules.GraphSnapshotNameReference
	case "GraphModelSnapshotNameReference":
		return rules.GraphModelSnapshotNameReference
	case "DatabaseOrTableNameReference":
		return rules.DatabaseOrTableNameReference
	case "DatabaseOrTableOrColumnNameReference":
		return rules.DatabaseOrTableOrColumnNameReference
	case "TableOrColumnNameReference":
		return rules.TableOrColumnNameReference
	case "DatabaseTableNameReference":
		return rules.DatabaseTableNameReference
	case "DatabaseTableColumnNameReference":
		return rules.DatabaseTableColumnNameReference
	case "TableColumnNameReference":
		return rules.TableColumnNameReference
	case "DatabaseExternalTableNameReference":
		return rules.DatabaseExternalTableNameReference
	case "DatabaseMaterializedViewNameReference":
		return rules.DatabaseMaterializedViewNameReference
	case "DatabaseFunctionNameReference":
		return rules.DatabaseFunctionNameReference
	case "DatabaseEntityGroupNameReference":
		return rules.DatabaseEntityGroupNameReference
	case "QueryInput":
		return rules.QueryInput
	case "ScriptInput":
		return rules.ScriptInput
	case "FunctionDeclaration":
		return rules.FunctionDeclaration
	case "FunctionBody":
		return rules.FunctionBody
	case "InputText":
		return rules.InputText
	case "BracketedInputText":
		return rules.BracketedInputText
	default:
		return nil
	}
}

func (p *queryParserState) parseCommandTokenText(text string) *SyntaxToken {
	length := matchesLexicalTextIgnoreCase(NewArraySource(p.tokens), p.pos, text)
	if length <= 0 {
		return nil
	}
	token := ProduceSyntaxTokenFromText(NewArraySource(p.tokens), p.pos, length, text)
	p.pos += length
	return token
}

func matchesLexicalTextIgnoreCase(source Source[LexicalToken], start int, text string) int {
	lowerText := strings.ToLower(text)
	textOffset := 0
	i := 0
	for ; !source.IsEnd(start + i); i++ {
		token := source.Peek(start + i)
		if i > 0 && token.Trivia != "" {
			break
		}
		if len(token.Text) > len(lowerText)-textOffset {
			break
		}
		if !strings.HasPrefix(lowerText[textOffset:], strings.ToLower(token.Text)) {
			break
		}
		textOffset += len(token.Text)
		if textOffset == len(lowerText) {
			return i + 1
		}
	}
	return -(i + 1)
}

func (p *queryParserState) parseCommandBlockSkippedTokens() *SkippedTokens {
	var tokens []SyntaxElement
	for !p.atEndOfQuery() {
		token := p.parseToken()
		if token == nil {
			break
		}
		tokens = append(tokens, token)
	}
	if len(tokens) == 0 {
		return nil
	}
	return MustNewSkippedTokensTyped(AsSyntaxListOf[*SyntaxToken](NewTypedSyntaxList("SyntaxToken", tokens...)))
}

func createMissingCommandStatement() Statement {
	return MustNewExpressionStatementTyped(
		MustNewBadCommandTyped(createMissingToken(SyntaxKindDotToken), DiagnosticMissingCommand()),
	)
}

func isUnknownCommandToken(token LexicalToken) bool {
	switch token.Kind {
	case SyntaxKindNone,
		SyntaxKindBarToken,
		SyntaxKindLessThanBarToken,
		SyntaxKindEndOfTextToken,
		SyntaxKindSemicolonToken:
		return false
	default:
		return true
	}
}
