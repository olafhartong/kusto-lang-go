package kustolang

import (
	"strings"
	"sync"
)

// Ported from the public shape of src/Kusto.Language/KustoCode.cs.

type CodeKind string

const (
	CodeKindUnknown   CodeKind = "Unknown"
	CodeKindQuery     CodeKind = "Query"
	CodeKindCommand   CodeKind = "Command"
	CodeKindDirective CodeKind = "Directive"
)

type IncludeFunctionKind uint32

const (
	IncludeFunctionKindBuiltInFunctions  IncludeFunctionKind = 1
	IncludeFunctionKindDatabaseFunctions                     = IncludeFunctionKindBuiltInFunctions << 1
	IncludeFunctionKindLocalFunctions                        = IncludeFunctionKindDatabaseFunctions << 1
	IncludeFunctionKindLocalViews                            = IncludeFunctionKindLocalFunctions << 1

	IncludeFunctionKindNone IncludeFunctionKind = 0
	IncludeFunctionKindAll                      = IncludeFunctionKindBuiltInFunctions |
		IncludeFunctionKindDatabaseFunctions |
		IncludeFunctionKindLocalFunctions |
		IncludeFunctionKindLocalViews
)

type kustoCodeAnalysisState uint8

const (
	kustoCodeAnalysisNotRequested kustoCodeAnalysisState = iota
	kustoCodeAnalysisPerformed
	kustoCodeAnalysisNotSafe
)

type KustoCode struct {
	Text                         string
	Kind                         CodeKind
	Globals                      *GlobalState
	Tree                         *SyntaxTree
	ResultType                   *TypeSymbol
	Tokens                       []LexicalToken
	TokenStarts                  []int
	Query                        *QueryBlock
	preserveDuplicateDiagnostics bool
	parseLimitExceeded           bool
	analysisState                kustoCodeAnalysisState
	lineStartsCache              *kustoCodeLineStartsCache
	diagnosticsCache             *kustoCodeDiagnosticsCache
}

type kustoCodeDiagnosticsCache struct {
	mu        sync.Mutex
	all       []Diagnostic
	allSet    bool
	syntax    []Diagnostic
	syntaxSet bool
}

type kustoCodeLineStartsCache struct {
	once   sync.Once
	starts []int
}

func newKustoCodeLineStartsCache() *kustoCodeLineStartsCache {
	return &kustoCodeLineStartsCache{}
}

func (c *kustoCodeLineStartsCache) get(text string) []int {
	if c == nil {
		return LineStarts(text)
	}
	c.once.Do(func() {
		c.starts = LineStarts(text)
	})
	return c.starts
}

func Parse(text string) *KustoCode {
	return ParseWithGlobals(text, nil)
}

func ParseWithGlobals(text string, globals *GlobalState) *KustoCode {
	return parseWithGlobalsAndTokens(text, globals, lexForGlobalState(text, globals))
}

func ParseWithGlobalsAndCancellation(text string, globals *GlobalState, cancellationToken CancellationToken) *KustoCode {
	cancellationToken.ThrowIfCancellationRequested()
	tokens := lexForGlobalStateWithCancellation(text, globals, cancellationToken)
	return createKustoCodeWithCancellation(text, globals, tokens, false, cancellationToken)
}

func ParseAndAnalyze(text string, globals ...*GlobalState) *KustoCode {
	var state *GlobalState
	if len(globals) > 0 {
		state = globals[0]
	}
	return ParseAndAnalyzeWithCancellation(text, state, CancellationToken{})
}

func ParseAndAnalyzeWithCancellation(text string, globals *GlobalState, cancellationToken CancellationToken) *KustoCode {
	cancellationToken.ThrowIfCancellationRequested()
	tokens := lexForGlobalStateWithCancellation(text, globals, cancellationToken)
	cancellationToken.ThrowIfCancellationRequested()
	return createKustoCodeWithCancellation(text, globals, tokens, true, cancellationToken)
}

func lexForGlobalState(text string, globals *GlobalState) []LexicalToken {
	return lexForGlobalStateWithCancellation(text, globals, CancellationToken{})
}

func lexForGlobalStateWithCancellation(text string, globals *GlobalState, cancellationToken CancellationToken) []LexicalToken {
	options := DefaultParseOptions
	if globals != nil {
		options = globals.ParseOptions
	}
	return LexWithOptionsAndCancellation(text, options.WithAlwaysProduceEndTokens(true), cancellationToken)
}

func parseWithGlobalsAndTokens(text string, globals *GlobalState, tokens []LexicalToken) *KustoCode {
	return createKustoCode(text, globals, tokens, false)
}

func createKustoCode(text string, globals *GlobalState, tokens []LexicalToken, analyze bool) *KustoCode {
	return createKustoCodeWithCancellation(text, globals, tokens, analyze, CancellationToken{})
}

func createKustoCodeWithCancellation(text string, globals *GlobalState, tokens []LexicalToken, analyze bool, cancellationToken CancellationToken) *KustoCode {
	if globals == nil {
		globals = DefaultGlobalState
	}
	globals = globals.WithParseOptions(globals.ParseOptions.WithAlwaysProduceEndTokens(true))
	parseQuery := func(tokens []LexicalToken) *QueryBlock {
		return ParseQueryTokensWithOptionsAndCancellation(tokens, globals.ParseOptions, cancellationToken)
	}
	if globals.ParseOptions.ParserKind == ParserKindGrammar {
		grammar := QueryGrammarFrom(globals)
		parseQuery = func(tokens []LexicalToken) *QueryBlock {
			return parseFirstSyntaxTokensWithCancellation(grammar.QueryBlock, tokens, cancellationToken)
		}
	}
	return createKustoCodeWithQueryParserAndCancellation(text, globals, tokens, parseQuery, analyze, cancellationToken)
}

func createKustoCodeWithQueryParserAndCancellation(text string, globals *GlobalState, tokens []LexicalToken, parseQuery func([]LexicalToken) *QueryBlock, analyze bool, cancellationToken CancellationToken) *KustoCode {
	return createKustoCodeWithParsersAndCancellation(text, globals, tokens, parseQuery, nil, analyze, cancellationToken)
}

func createKustoCodeWithParsersAndCancellation(
	text string,
	globals *GlobalState,
	tokens []LexicalToken,
	parseQuery func([]LexicalToken) *QueryBlock,
	parseCommand func([]LexicalToken) *CommandBlock,
	analyze bool,
	cancellationToken CancellationToken,
) *KustoCode {
	cancellationToken.ThrowIfCancellationRequested()
	if globals == nil {
		globals = DefaultGlobalState
	}
	if exceedsSafeParserShape(tokens, GlobalStateGetProperty(globals, Properties.MaxAnalysisDepth), cancellationToken) {
		return newParseLimitExceededKustoCode(text, globals, tokens, cancellationToken)
	}
	code := parseWithGlobalsAndTokensAndParsersWithCancellation(text, globals, tokens, parseQuery, parseCommand, cancellationToken)
	cancellationToken.ThrowIfCancellationRequested()
	if analyze && code != nil && code.Syntax() != nil {
		cancellationToken.ThrowIfCancellationRequested()
		if !code.Tree.IsSafeToRecurse(code.Globals) {
			code.analysisState = kustoCodeAnalysisNotSafe
		} else if NewBinder(code.Globals).TryBindWithCancellation(code.Syntax(), cancellationToken) {
			code.analysisState = kustoCodeAnalysisPerformed
			code.ResultType = determineKustoCodeResultType(code)
		}
	}
	return code
}

func parseWithGlobalsAndTokensAndParsersWithCancellation(
	text string,
	globals *GlobalState,
	tokens []LexicalToken,
	parseQuery func([]LexicalToken) *QueryBlock,
	parseCommand func([]LexicalToken) *CommandBlock,
	cancellationToken CancellationToken,
) *KustoCode {
	if globals == nil {
		globals = DefaultGlobalState
	}
	globals = globals.WithParseOptions(globals.ParseOptions.WithAlwaysProduceEndTokens(true))
	starts := tokenStartsWithCancellation(tokens, cancellationToken)
	kind := codeKindOfWithCancellation(text, cancellationToken)
	var root *SyntaxNode
	var query *QueryBlock
	if kind == CodeKindCommand {
		if parseCommand != nil {
			root = parseCommand(tokens).SyntaxNode
		} else {
			root = CommandGrammarFrom(globals).parseCommandBlockTokensWithCancellation(tokens, cancellationToken).SyntaxNode
		}
	} else {
		query = parseQuery(tokens)
		root = query.SyntaxNode
	}
	return &KustoCode{
		Text:             text,
		Kind:             kind,
		Globals:          globals,
		Tree:             NewSyntaxTree(root, nil, 0),
		Tokens:           tokens,
		TokenStarts:      starts,
		Query:            query,
		lineStartsCache:  newKustoCodeLineStartsCache(),
		diagnosticsCache: &kustoCodeDiagnosticsCache{},
	}
}

const absoluteMaxParserShapeDepth = 4096

func exceedsSafeParserShape(tokens []LexicalToken, configuredLimit int, cancellationToken CancellationToken) bool {
	limit := configuredLimit
	if limit < 1 || limit > absoluteMaxParserShapeDepth {
		limit = absoluteMaxParserShapeDepth
	}
	delimiterDepth := 0
	shapeDepth := 0
	for i, token := range tokens {
		if i&255 == 0 {
			cancellationToken.ThrowIfCancellationRequested()
		}
		switch token.Kind {
		case SyntaxKindOpenParenToken, SyntaxKindOpenBracketToken, SyntaxKindOpenBraceToken:
			delimiterDepth++
			shapeDepth++
		case SyntaxKindCloseParenToken, SyntaxKindCloseBracketToken, SyntaxKindCloseBraceToken:
			if delimiterDepth > 0 {
				delimiterDepth--
			}
		case SyntaxKindSemicolonToken, SyntaxKindCommaToken:
			shapeDepth = delimiterDepth
		default:
			if parserShapeToken(token.Kind) {
				shapeDepth++
			}
		}
		if delimiterDepth > limit || shapeDepth > limit {
			return true
		}
	}
	cancellationToken.ThrowIfCancellationRequested()
	return false
}

func parserShapeToken(kind SyntaxKind) bool {
	if additiveExpressionKind(kind) != SyntaxKindNone ||
		multiplicativeExpressionKind(kind) != SyntaxKindNone ||
		relationalExpressionKind(kind) != SyntaxKindNone ||
		stringOperationKind(kind) != SyntaxKindNone {
		return true
	}
	switch kind {
	case SyntaxKindBarToken, SyntaxKindDotToken,
		SyntaxKindAndKeyword, SyntaxKindOrKeyword,
		SyntaxKindEqualEqualToken, SyntaxKindBangEqualToken, SyntaxKindLessThanGreaterThanToken:
		return true
	default:
		return false
	}
}

func newParseLimitExceededKustoCode(text string, globals *GlobalState, tokens []LexicalToken, cancellationToken CancellationToken) *KustoCode {
	cancellationToken.ThrowIfCancellationRequested()
	kind := codeKindOfWithCancellation(text, cancellationToken)
	rootKind := SyntaxKindQueryBlock
	if kind == CodeKindCommand {
		rootKind = SyntaxKindCommandBlock
	}
	return &KustoCode{
		Text:               text,
		Kind:               kind,
		Globals:            globals.WithParseOptions(globals.ParseOptions.WithAlwaysProduceEndTokens(true)),
		Tree:               NewSyntaxTree(NewSyntaxNode(rootKind, nil), nil, 0),
		Tokens:             tokens,
		TokenStarts:        tokenStartsWithCancellation(tokens, cancellationToken),
		parseLimitExceeded: true,
		analysisState:      kustoCodeAnalysisNotSafe,
		lineStartsCache:    newKustoCodeLineStartsCache(),
		diagnosticsCache:   &kustoCodeDiagnosticsCache{},
	}
}

func (c *KustoCode) Syntax() *SyntaxNode {
	if c == nil || c.Tree == nil {
		return nil
	}
	return c.Tree.Root
}

// ParseLimitExceeded reports that parsing was intentionally skipped because
// the token shape exceeded the configured safe analysis depth. Callers should
// stop syntax processing and obtain the structured KS245 from GetDiagnostics.
func (c *KustoCode) ParseLimitExceeded() bool {
	return c != nil && c.parseLimitExceeded
}

func (c *KustoCode) GetDiagnostics() []Diagnostic {
	return c.GetDiagnosticsWithCancellation(CancellationToken{})
}

func (c *KustoCode) GetDiagnosticsWithCancellation(cancellationToken CancellationToken) []Diagnostic {
	if c == nil || c.Tree == nil {
		return []Diagnostic{}
	}
	cache := c.diagnosticsCache
	if cache == nil {
		cache = &kustoCodeDiagnosticsCache{}
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.allSet {
		return append([]Diagnostic(nil), cache.all...)
	}
	cancellationToken.ThrowIfCancellationRequested()
	var diagnostics []Diagnostic
	if c.analysisState == kustoCodeAnalysisNotSafe {
		diagnostics = GetContainedDiagnosticsWithCancellation(c.Tree.Root, DiagnosticsIncludeSyntactic, cancellationToken)
		diagnostics = append(diagnostics, DiagnosticQuerySyntaxDepthExceeded())
	} else if c.preserveDuplicateDiagnostics {
		include := DiagnosticsIncludeSyntactic
		if c.HasSemantics() {
			include = DiagnosticsIncludeDefault
		}
		diagnostics = GetContainedDiagnosticsPreserveDuplicatesWithCancellation(c.Tree.Root, include, cancellationToken)
	} else if c.HasSemantics() {
		diagnostics = GetContainedDiagnosticsWithCancellation(c.Tree.Root, DiagnosticsIncludeDefault, cancellationToken)
	} else {
		diagnostics = GetContainedDiagnosticsWithCancellation(c.Tree.Root, DiagnosticsIncludeSyntactic, cancellationToken)
	}
	cache.all = append([]Diagnostic(nil), diagnostics...)
	cache.allSet = true
	return append([]Diagnostic(nil), diagnostics...)
}

func (c *KustoCode) GetSyntaxDiagnostics() []Diagnostic {
	return c.GetSyntaxDiagnosticsWithCancellation(CancellationToken{})
}

func (c *KustoCode) GetSyntaxDiagnosticsWithCancellation(cancellationToken CancellationToken) []Diagnostic {
	if c == nil || c.Tree == nil {
		return []Diagnostic{}
	}
	cache := c.diagnosticsCache
	if cache == nil {
		cache = &kustoCodeDiagnosticsCache{}
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.syntaxSet {
		return append([]Diagnostic(nil), cache.syntax...)
	}
	cancellationToken.ThrowIfCancellationRequested()
	var diagnostics []Diagnostic
	if c.preserveDuplicateDiagnostics {
		diagnostics = GetContainedDiagnosticsPreserveDuplicatesWithCancellation(c.Tree.Root, DiagnosticsIncludeSyntactic, cancellationToken)
	} else {
		diagnostics = GetContainedDiagnosticsWithCancellation(c.Tree.Root, DiagnosticsIncludeSyntactic, cancellationToken)
	}
	cache.syntax = append([]Diagnostic(nil), diagnostics...)
	cache.syntaxSet = true
	return append([]Diagnostic(nil), diagnostics...)
}

func (c *KustoCode) HasSemantics() bool {
	return c != nil && c.analysisState == kustoCodeAnalysisPerformed
}

// Analyze returns this instance when it is already analyzed with the requested
// globals. Otherwise it reuses the lexical tokens to create and bind a new tree.
func (c *KustoCode) Analyze(globals ...*GlobalState) *KustoCode {
	var state *GlobalState
	if len(globals) > 0 {
		state = globals[0]
	}
	return c.AnalyzeWithCancellation(state, CancellationToken{})
}

func (c *KustoCode) AnalyzeWithCancellation(globals *GlobalState, cancellationToken CancellationToken) *KustoCode {
	if c == nil {
		return nil
	}
	state := c.Globals
	if globals != nil {
		state = globals
	}
	if c.HasSemantics() && c.Globals == state {
		return c
	}
	cancellationToken.ThrowIfCancellationRequested()
	tokens := c.tokensForGlobalsWithCancellation(state, cancellationToken)
	return createKustoCodeWithCancellation(c.Text, state, tokens, true, cancellationToken)
}

// WithGlobals returns this instance for the same GlobalState. A changed state
// preserves whether semantic analysis had already been requested.
func (c *KustoCode) WithGlobals(globals *GlobalState) *KustoCode {
	return c.WithGlobalsWithCancellation(globals, CancellationToken{})
}

func (c *KustoCode) WithGlobalsWithCancellation(globals *GlobalState, cancellationToken CancellationToken) *KustoCode {
	if c == nil {
		return nil
	}
	if globals == nil {
		globals = DefaultGlobalState
	}
	if c.Globals == globals {
		return c
	}
	cancellationToken.ThrowIfCancellationRequested()
	tokens := c.tokensForGlobalsWithCancellation(globals, cancellationToken)
	analyze := c.analysisState != kustoCodeAnalysisNotRequested
	return createKustoCodeWithCancellation(c.Text, globals, tokens, analyze, cancellationToken)
}

// WithText reparses replacement text with the receiver's globals, preserving
// whether semantic analysis had already been requested.
func (c *KustoCode) WithText(text string) *KustoCode {
	return c.WithTextWithCancellation(text, CancellationToken{})
}

func (c *KustoCode) WithTextWithCancellation(text string, cancellationToken CancellationToken) *KustoCode {
	if c == nil {
		return nil
	}
	if c.Text == text {
		return c
	}
	cancellationToken.ThrowIfCancellationRequested()
	options := DefaultParseOptions
	if c.Globals != nil {
		options = c.Globals.ParseOptions
	}
	options = options.WithAlwaysProduceEndTokens(true)
	analyze := c.analysisState != kustoCodeAnalysisNotRequested
	tokens := LexWithOptionsAndCancellation(text, options, cancellationToken)
	return createKustoCodeWithCancellation(text, c.Globals, tokens, analyze, cancellationToken)
}

func (c *KustoCode) tokensForGlobalsWithCancellation(globals *GlobalState, cancellationToken CancellationToken) []LexicalToken {
	if c == nil || globals == nil || c.Globals == nil || c.Globals.ParseOptions.EqualExceptForParseKind(globals.ParseOptions) {
		cancellationToken.ThrowIfCancellationRequested()
		return c.Tokens
	}
	options := globals.ParseOptions.WithAlwaysProduceEndTokens(true)
	return LexWithOptionsAndCancellation(c.Text, options, cancellationToken)
}

func determineKustoCodeResultType(code *KustoCode) *TypeSymbol {
	if code == nil {
		return nil
	}
	var statements *SeparatedSyntaxListOf[Statement]
	if code.Query != nil {
		statements = code.Query.Statements()
	} else if node, ok := WrapGeneratedSyntaxElement(code.Syntax()).(*CommandBlock); ok {
		statements = node.Statements()
	}
	if statements == nil || len(statements.Elements()) == 0 {
		return nil
	}
	last := statements.At(len(statements.Elements()) - 1)
	if last == nil {
		return nil
	}
	expressionStatement, ok := WrapGeneratedSyntaxElement(last.Element()).(*ExpressionStatement)
	if !ok || expressionStatement.Expression() == nil {
		return nil
	}
	expression := expressionStatement.Expression()
	if code.Query == nil {
		// Upstream returns the command expression's ResultType directly.
		// Recovery wrappers such as PartialCommand and
		// CommandAndSkippedTokens are deliberately left unbound.
		return ExpressionResultType(expression)
	}
	if node := simpleBinderSyntaxNode(expression); node != nil {
		if info, bound := node.GetSemanticInfo(); bound && info.ResultType == nil && info.ReferencedSymbol() == nil {
			return nil
		}
	}
	result := ExpressionResultType(expression)
	if result == nil {
		if bracketed, ok := expression.(*BracketedExpression); ok && syntaxElementHasMissingChildren(bracketed) {
			return &Error.TypeSymbol
		}
		result = returnTypeFromSymbol(expressionReferencedSymbol(expression))
	}
	return result
}

func (c *KustoCode) GetSymbolsInScope(position int, match SymbolMatch, include IncludeFunctionKind) []Symbol {
	return c.GetSymbolsInScopeWithCancellation(position, match, include, CancellationToken{})
}

func (c *KustoCode) GetSymbolsInScopeWithCancellation(position int, match SymbolMatch, include IncludeFunctionKind, cancellationToken CancellationToken) []Symbol {
	cancellationToken.ThrowIfCancellationRequested()
	if c == nil || !c.HasSemantics() || c.Syntax() == nil || position < 0 || position > len(c.Text) || match == SymbolMatchNone {
		return []Symbol{}
	}

	collector := newKustoCodeScopeCollector(match, include, cancellationToken)
	if path := c.getPathScopeContextWithCancellation(position, cancellationToken); path.bounded {
		c.addPathSymbolsInScopeWithCancellation(position, path.scope, collector, cancellationToken)
		return collector.symbols
	}
	scopeKind := c.getScopeKindContextWithCancellation(position, cancellationToken)
	switch scopeKind {
	case ScopeKindAggregate:
		c.addAggregateSymbolsInScope(position, c.globalsInScopeWithCancellation(position, cancellationToken), collector)
	case ScopeKindPlugIn:
		c.addPlugInSymbolsInScope(collector)
	case ScopeKindOption:
		c.addOptionSymbolsInScope(collector)
	default:
		context := c.getRowScopeContextWithCancellation(position, cancellationToken)
		if context.rowScope != nil && context.rightRowScope != nil && match.Has(SymbolMatchColumn) {
			collector.addContextSymbol(NewVariableSymbol("$left", NewTupleSymbol(context.rowScope.Columns(), context.rowScope), false, nil, nil))
			collector.addContextSymbol(NewVariableSymbol("$right", NewTupleSymbol(context.rightRowScope.Columns(), context.rightRowScope), false, nil, nil))
			for _, column := range kustoCodeCommonColumnsWithCancellation(context.rowScope, context.rightRowScope, cancellationToken) {
				collector.add(column, kustoCodeFunctionSourceNone)
			}
		} else if context.rowScope != nil {
			for _, column := range context.rowScope.Columns() {
				collector.add(column, kustoCodeFunctionSourceNone)
			}
		}
		for _, symbol := range c.localSymbolsInScopeWithCancellation(position, cancellationToken) {
			collector.add(symbol, kustoCodeFunctionSourceLocal)
		}
		globals := c.globalsInScopeWithCancellation(position, cancellationToken)
		if globals != nil {
			// Binder.TryBind places ambient symbols in the default outer local
			// scope. Client symbols are resolved only by client-parameter syntax.
			for _, symbol := range globals.AmbientSymbols {
				collector.add(symbol, kustoCodeFunctionSourceLocal)
			}
		}
		c.addNormalGlobalSymbolsInScope(globals, collector)
	}
	cancellationToken.ThrowIfCancellationRequested()
	return collector.symbols
}

// addPathSymbolsInScope ports the path-scope branch of Binder.GetSymbolsInContext.
func (c *KustoCode) addPathSymbolsInScope(position int, scope Symbol, collector *kustoCodeScopeCollector) {
	c.addPathSymbolsInScopeWithCancellation(position, scope, collector, CancellationToken{})
}

func (c *KustoCode) addPathSymbolsInScopeWithCancellation(position int, scope Symbol, collector *kustoCodeScopeCollector, cancellationToken CancellationToken) {
	cancellationToken.ThrowIfCancellationRequested()
	if c == nil || scope == nil || collector == nil {
		return
	}
	if group, ok := scope.(*GroupSymbol); ok && kustoCodeIsPassThroughGroupWithCancellation(group, cancellationToken) {
		for _, member := range group.Members() {
			cancellationToken.ThrowIfCancellationRequested()
			c.addPathSymbolsInScopeWithCancellation(position, member, collector, cancellationToken)
		}
		return
	}

	insideControlCommand := c.isInsideControlCommandProperWithCancellation(position, cancellationToken)
	memberMatch := collector.match
	if !insideControlCommand {
		memberMatch &= SymbolMatchColumn | SymbolMatchTable | SymbolMatchMaterializedView | SymbolMatchEntityGroup | SymbolMatchFunction
	}
	if _, ok := scope.(*DatabaseSymbol); ok {
		memberMatch &^= SymbolMatchExternalTable
	}
	if _, ok := tableSymbolFromSymbol(scope); ok && !insideControlCommand {
		memberMatch &^= SymbolMatchColumn
	}
	if memberMatch != SymbolMatchNone {
		for _, member := range FindMembersByMatch(scope, memberMatch) {
			collector.add(member, kustoCodeFunctionSourceNone)
		}
	}

	if c.hasMacroExpandScopeAtWithCancellation(position, cancellationToken) {
		for _, member := range EntityGroupElementSpecialMembers() {
			collector.addContextSymbol(member)
		}
	}
	if collector.match.Has(SymbolMatchFunction) {
		for _, function := range kustoCodeSpecialPathFunctionsWithCancellation(scope, cancellationToken) {
			collector.add(function, kustoCodeFunctionSourceNone)
		}
	}
}

func kustoCodeIsPassThroughGroupWithCancellation(group *GroupSymbol, cancellationToken CancellationToken) bool {
	if group == nil {
		return false
	}
	for _, member := range group.Members() {
		cancellationToken.ThrowIfCancellationRequested()
		switch member.(type) {
		case *DatabaseSymbol, *ClusterSymbol:
			return true
		}
	}
	return false
}

func kustoCodeSpecialPathFunctionsWithCancellation(scope Symbol, cancellationToken CancellationToken) []*FunctionSymbol {
	cancellationToken.ThrowIfCancellationRequested()
	if group, ok := scope.(*GroupSymbol); ok && kustoCodeIsPassThroughGroupWithCancellation(group, cancellationToken) {
		members := group.Members()
		if len(members) == 0 {
			return nil
		}
		return kustoCodeSpecialPathFunctionsWithCancellation(members[0], cancellationToken)
	}
	switch scope.(type) {
	case *ClusterSymbol:
		return []*FunctionSymbol{Functions.Database}
	case *DatabaseSymbol:
		return []*FunctionSymbol{
			Functions.Table,
			Functions.ExternalTable,
			Functions.MaterializedView,
			Functions.EntityGroup,
			Functions.StoredQueryResult,
			Functions.Graph,
		}
	default:
		return nil
	}
}

func (c *KustoCode) hasMacroExpandScopeAtWithCancellation(position int, cancellationToken CancellationToken) bool {
	for _, symbol := range c.localSymbolsInScopeWithCancellation(position, cancellationToken) {
		cancellationToken.ThrowIfCancellationRequested()
		if _, ok := symbol.(*EntityGroupElementSymbol); ok {
			return true
		}
	}
	return false
}

func (c *KustoCode) isInsideControlCommandProperWithCancellation(position int, cancellationToken CancellationToken) bool {
	cancellationToken.ThrowIfCancellationRequested()
	if c == nil || c.Kind != CodeKindCommand || c.Syntax() == nil {
		return false
	}
	token := c.Syntax().GetTokenAt(position)
	if token == nil && position > 0 {
		token = c.Syntax().GetTokenAt(position - 1)
	}
	if token == nil {
		return false
	}
	insideCommand := false
	for element := SyntaxElement(token); element != nil; element = element.Parent() {
		cancellationToken.ThrowIfCancellationRequested()
		wrapped := WrapGeneratedSyntaxElement(element)
		if _, ok := wrapped.(Command); ok {
			insideCommand = true
		}
		if _, ok := wrapped.(*FunctionBody); ok {
			return false
		}
		if list, ok := element.(*SyntaxList); ok && list.Parent() != nil {
			if _, custom := list.Parent().(*CustomNode); custom {
				return false
			}
		}
	}
	return insideCommand
}

// globalsInScope mirrors ContextBuilder.ApplyDirectives without mutating the
// KustoCode's original GlobalState. Only directives ending at or before the
// queried position affect editor and speculative-binding context.
func (c *KustoCode) globalsInScope(position int) *GlobalState {
	return c.globalsInScopeWithCancellation(position, CancellationToken{})
}

func (c *KustoCode) globalsInScopeWithCancellation(position int, cancellationToken CancellationToken) *GlobalState {
	cancellationToken.ThrowIfCancellationRequested()
	if c == nil || c.Globals == nil || c.Syntax() == nil {
		if c == nil {
			return nil
		}
		return c.Globals
	}
	globals := c.Globals
	walkSyntaxWithTriviaStartWithCancellation(c.Syntax(), 0, cancellationToken, func(element SyntaxElement, triviaStart int) {
		directive, ok := WrapGeneratedSyntaxElement(element).(*Directive)
		if !ok || directive.Token() == nil || position < elementEnd(directive, triviaStart) {
			return
		}
		parsed, valid := TryParseClientDirective(NewEditString(directive.Token().RawText()))
		if !valid {
			return
		}
		clusterName, databaseName, valid := TryGetDirectiveClusterAndDatabase(parsed)
		if !valid {
			return
		}
		if clusterName != nil {
			globals = globals.WithCluster(globals.GetCluster(*clusterName))
		}
		if databaseName != nil {
			var database *DatabaseSymbol
			if globals.Cluster != nil {
				database = globals.Cluster.GetDatabase(*databaseName)
			}
			globals = globals.WithDatabase(database)
		}
	})
	return globals
}

func (c *KustoCode) getScopeKindContext(position int) ScopeKind {
	return c.getScopeKindContextWithCancellation(position, CancellationToken{})
}

func (c *KustoCode) getScopeKindContextWithCancellation(position int, cancellationToken CancellationToken) ScopeKind {
	cancellationToken.ThrowIfCancellationRequested()
	if c == nil || c.Syntax() == nil || position < 0 || position > len(c.Text) {
		return ScopeKindNormal
	}
	scopeKind := ScopeKindNormal
	walkSyntaxWithTriviaStartWithCancellation(c.Syntax(), 0, cancellationToken, func(element SyntaxElement, triviaStart int) {
		wrapped := WrapGeneratedSyntaxElement(element)
		if !kustoCodeContextElementContains(c, wrapped, triviaStart, position) {
			return
		}
		switch node := wrapped.(type) {
		case *EvaluateOperator:
			parameters := node.Parameters()
			keyword := node.EvaluateKeyword()
			if parameters == nil || keyword == nil {
				return
			}
			parametersEnd := triviaStart + keyword.FullWidth() + parameters.FullWidth()
			if position >= parametersEnd {
				scopeKind = ScopeKindPlugIn
			}
		case *SummarizeOperator:
			by := node.ByClause()
			if by == nil {
				scopeKind = ScopeKindAggregate
				return
			}
			start, ok := syntaxElementTriviaStartOf(node, by, triviaStart)
			if ok && position < elementTextStart(by, start) {
				scopeKind = ScopeKindAggregate
			}
		case *NamedParameter:
			equal := node.EqualToken()
			if equal == nil || node.Expression() == nil {
				return
			}
			equalStart, equalOK := syntaxElementTriviaStartOf(node, equal, triviaStart)
			expressionStart, expressionOK := syntaxElementTriviaStartOf(node, node.Expression(), triviaStart)
			if equalOK && expressionOK && position >= elementEnd(equal, equalStart) &&
				(syntaxElementIsMissing(node.Expression()) || position < elementEnd(node.Expression(), expressionStart)) {
				scopeKind = ScopeKindNormal
			}
		case *MakeSeriesOperator:
			on := node.OnClause()
			if on == nil {
				return
			}
			onStart, ok := syntaxElementTriviaStartOf(node, on, triviaStart)
			if !ok {
				return
			}
			nodeEnd := triviaStart + node.FullWidth()
			onEnd := elementEnd(on, onStart)
			if position < elementTextStart(on, onStart) || (syntaxElementIsMissing(on) && onEnd == nodeEnd && position >= nodeEnd) {
				scopeKind = ScopeKindAggregate
			}
		case *TopNestedClause:
			by := node.ByKeyword()
			if by == nil || by.Width() == 0 {
				return
			}
			start, ok := syntaxElementTriviaStartOf(node, by, triviaStart)
			if ok && position > elementEnd(by, start) {
				scopeKind = ScopeKindAggregate
			}
		case *SetOptionStatement:
			setKeyword := node.SetKeyword()
			if setKeyword == nil {
				return
			}
			setStart, ok := syntaxElementTriviaStartOf(node, setKeyword, triviaStart)
			if !ok || position < elementEnd(setKeyword, setStart) {
				return
			}
			value := node.ValueClause()
			if value == nil {
				scopeKind = ScopeKindOption
				return
			}
			valueStart, found := syntaxElementTriviaStartOf(node, value, triviaStart)
			if found && position <= elementTextStart(value, valueStart) {
				scopeKind = ScopeKindOption
			}
		case *FunctionCallExpression:
			name := node.Name()
			if name == nil || name.Width() == 0 || NameReferenceSimpleName(name) == "" || node.ArgumentList() == nil {
				return
			}
			nameStart, ok := syntaxElementTriviaStartOf(node, name, triviaStart)
			if !ok || position <= elementEnd(name, nameStart) {
				return
			}
			function := kustoCodeContextFunctionWithCancellation(c, node, scopeKind, position, cancellationToken)
			scopeKind = ArgumentScopeForFunctionCallPosition(c.Globals, function, node.ReferencedSignature(), node, position, scopeKind)
		}
	})
	return scopeKind
}

func kustoCodeContextElementContains(code *KustoCode, element SyntaxElement, triviaStart int, position int) bool {
	if code == nil || element == nil || position < triviaStart {
		return false
	}
	end := triviaStart + element.FullWidth()
	return position <= end || (position <= len(code.Text) && strings.TrimSpace(code.Text[end:position]) == "")
}

func kustoCodeContextFunctionWithCancellation(code *KustoCode, call *FunctionCallExpression, scopeKind ScopeKind, position int, cancellationToken CancellationToken) *FunctionSymbol {
	cancellationToken.ThrowIfCancellationRequested()
	if call == nil {
		return nil
	}
	if function, ok := call.ReferencedSymbol().(*FunctionSymbol); ok {
		return function
	}
	if call.Name() != nil {
		if function, ok := call.Name().ReferencedSymbol().(*FunctionSymbol); ok {
			return function
		}
	}
	if code == nil || code.Globals == nil || call.Name() == nil {
		return nil
	}
	globals := code.globalsInScopeWithCancellation(position, cancellationToken)
	name := NameReferenceSimpleName(call.Name())
	if scopeKind == ScopeKindAggregate {
		if function := globals.GetAggregate(name); function != nil {
			return function
		}
	}
	if scopeKind == ScopeKindPlugIn {
		return globals.GetPlugIn(name)
	}
	for _, symbol := range code.localSymbolsInScopeWithCancellation(position, cancellationToken) {
		cancellationToken.ThrowIfCancellationRequested()
		if symbol.SymbolName() != name {
			continue
		}
		if function, ok := symbol.(*FunctionSymbol); ok {
			return function
		}
		if variable, ok := symbol.(*VariableSymbol); ok {
			function, _ := variable.Type.(*FunctionSymbol)
			return function
		}
	}
	return globals.GetFunction(name)
}

type kustoCodePathScopeContext struct {
	scope   Symbol
	bounded bool
}

func (c *KustoCode) getPathScopeContext(position int) kustoCodePathScopeContext {
	return c.getPathScopeContextWithCancellation(position, CancellationToken{})
}

func (c *KustoCode) getPathScopeContextWithCancellation(position int, cancellationToken CancellationToken) kustoCodePathScopeContext {
	cancellationToken.ThrowIfCancellationRequested()
	if c == nil || !c.HasSemantics() || c.Syntax() == nil || position < 0 || position > len(c.Text) {
		return kustoCodePathScopeContext{}
	}

	var context kustoCodePathScopeContext
	bestWidth := int(^uint(0) >> 1)
	clearWidth := int(^uint(0) >> 1)
	walkSyntaxWithTriviaStartWithCancellation(c.Syntax(), 0, cancellationToken, func(element SyntaxElement, triviaStart int) {
		wrapped := WrapGeneratedSyntaxElement(element)
		switch node := wrapped.(type) {
		case *ParenthesizedExpression:
			if kustoCodeContextElementContains(c, element, triviaStart, position) && element.FullWidth() < clearWidth {
				clearWidth = element.FullWidth()
				if bestWidth >= clearWidth {
					context = kustoCodePathScopeContext{}
					bestWidth = int(^uint(0) >> 1)
				}
			}
		case *FunctionCallExpression:
			if node.Name() != nil && kustoCodeContextElementContains(c, element, triviaStart, position) {
				nameStart, found := syntaxElementTriviaStartOf(node, node.Name(), triviaStart)
				if found && position > elementEnd(node.Name(), nameStart) && element.FullWidth() < clearWidth {
					clearWidth = element.FullWidth()
					if bestWidth >= clearWidth {
						context = kustoCodePathScopeContext{}
						bestWidth = int(^uint(0) >> 1)
					}
				}
			}
		}

		var expression Expression
		var selector Expression
		switch path := wrapped.(type) {
		case *PathExpression:
			expression, selector = path.Expression(), path.Selector()
		case *ElementExpression:
			expression, selector = path.Expression(), path.Selector()
		default:
			return
		}
		if expression == nil || selector == nil {
			return
		}
		selectorStart, found := syntaxElementTriviaStartOf(element, selector, triviaStart)
		if !found {
			return
		}
		pathEnd := triviaStart + element.FullWidth()
		inTrailingTrivia := position > pathEnd && position <= len(c.Text) && strings.TrimSpace(c.Text[pathEnd:position]) == ""
		if position < selectorStart || (position > pathEnd && !inTrailingTrivia) {
			return
		}
		width := pathEnd - selectorStart
		if width >= bestWidth || width >= clearWidth {
			return
		}
		scope := c.editorExpressionResultWithCancellation(expression, position, cancellationToken)
		if scope == nil {
			if reference, ok := expression.(*NameReference); ok {
				scope = c.editorPathQualifierResultWithCancellation(NameReferenceSimpleName(reference), position, cancellationToken)
			}
		}
		context = kustoCodePathScopeContext{scope: scope, bounded: true}
		bestWidth = width
	})
	if context.scope == nil {
		if qualifier := c.incompletePathQualifierWithCancellation(position, cancellationToken); qualifier != "" {
			if scope := c.editorPathQualifierResultWithCancellation(qualifier, position, cancellationToken); scope != nil {
				context = kustoCodePathScopeContext{scope: scope, bounded: true}
			}
		}
	}
	return context
}

func kustoCodeExpressionResultSymbol(expression Expression) Symbol {
	if expression == nil {
		return nil
	}
	switch expression := expression.(type) {
	case *FunctionCallExpression:
		if result, found := expression.SyntaxNode.Payload(functionCallResultSymbolPayload); found {
			if symbol, ok := result.(Symbol); ok {
				return symbol
			}
		}
	case *NameReference:
		if result, found := expression.SyntaxNode.Payload(functionCallResultSymbolPayload); found {
			if symbol, ok := result.(Symbol); ok {
				return symbol
			}
		}
	}
	return GetResultType(expressionReferencedSymbol(expression))
}

func (c *KustoCode) editorExpressionResultWithCancellation(expression Expression, position int, cancellationToken CancellationToken) Symbol {
	cancellationToken.ThrowIfCancellationRequested()
	if result := kustoCodeExpressionResultSymbol(expression); result != nil {
		return result
	}
	switch expression := expression.(type) {
	case *PathExpression:
		receiver := c.editorExpressionResultWithCancellation(expression.Expression(), position, cancellationToken)
		return c.editorSelectorResultWithCancellation(expression.Selector(), position, receiver, cancellationToken)
	case *FunctionCallExpression:
		return c.editorIntrinsicCallResultWithCancellation(expression, position, nil, cancellationToken)
	default:
		return nil
	}
}

func (c *KustoCode) editorSelectorResultWithCancellation(selector Expression, position int, receiver Symbol, cancellationToken CancellationToken) Symbol {
	cancellationToken.ThrowIfCancellationRequested()
	if selector == nil || receiver == nil {
		return nil
	}
	if result := kustoCodeExpressionResultSymbol(selector); result != nil {
		return result
	}
	if call, ok := selector.(*FunctionCallExpression); ok {
		return c.editorIntrinsicCallResultWithCancellation(call, position, receiver, cancellationToken)
	}
	if reference, ok := selector.(*NameReference); ok {
		members := FindMembers(receiver, NameReferenceSimpleName(reference), SymbolMatchAny, false)
		if len(members) == 1 {
			return GetResultType(members[0])
		}
	}
	return nil
}

func (c *KustoCode) editorIntrinsicCallResultWithCancellation(expression Expression, position int, receiver Symbol, cancellationToken CancellationToken) Symbol {
	cancellationToken.ThrowIfCancellationRequested()
	call, ok := expression.(*FunctionCallExpression)
	if !ok || call.Name() == nil || call.ArgumentList() == nil || call.ArgumentList().Expressions() == nil {
		return nil
	}
	globals := c.globalsInScopeWithCancellation(position, cancellationToken)
	if globals == nil {
		return nil
	}
	arguments := call.ArgumentList().Expressions()
	switch strings.ToLower(NameReferenceSimpleName(call.Name())) {
	case "cluster":
		if arguments.Count() != 1 {
			return nil
		}
		name, ok := ExpressionLiteralValue(arguments.At(0).Element()).(string)
		if !ok {
			return nil
		}
		cluster := globals.GetCluster(name)
		if cluster == nil {
			return nil
		}
		return cluster
	case "database":
		if arguments.Count() == 0 {
			return globals.Database
		}
		if arguments.Count() != 1 {
			return nil
		}
		name, ok := ExpressionLiteralValue(arguments.At(0).Element()).(string)
		if !ok {
			return nil
		}
		cluster := globals.Cluster
		if scoped, ok := receiver.(*ClusterSymbol); ok {
			cluster = scoped
		}
		if cluster == nil {
			return nil
		}
		database := cluster.GetDatabase(name)
		if database == nil {
			return nil
		}
		return database
	case "table", "external_table", "materialized_view", "entity_group", "stored_query_result", "graph":
		if arguments.Count() < 1 {
			return nil
		}
		name, ok := ExpressionLiteralValue(arguments.At(0).Element()).(string)
		if !ok {
			return nil
		}
		database, ok := receiver.(*DatabaseSymbol)
		if !ok {
			database = globals.Database
		}
		if database == nil {
			return nil
		}
		switch strings.ToLower(NameReferenceSimpleName(call.Name())) {
		case "table":
			if table := database.GetTable(name); table != nil {
				return table
			}
		case "external_table":
			if table := database.GetExternalTable(name); table != nil {
				return table
			}
		case "materialized_view":
			if view := database.GetMaterializedView(name); view != nil {
				return view
			}
		case "entity_group":
			if group := database.GetEntityGroup(name); group != nil {
				return group
			}
		case "stored_query_result":
			if result := database.GetStoredQueryResult(name); result != nil {
				return result
			}
		case "graph":
			if graph := database.GetGraphModel(name); graph != nil {
				return graph
			}
		}
		return nil
	default:
		return nil
	}
}

func (c *KustoCode) editorPathQualifierResultWithCancellation(name string, position int, cancellationToken CancellationToken) Symbol {
	rows := c.getRowScopeContextWithCancellation(position, cancellationToken)
	switch name {
	case "$left":
		if rows.rowScope != nil && rows.rightRowScope != nil {
			return NewTupleSymbol(rows.rowScope.Columns(), rows.rowScope)
		}
	case "$right":
		if rows.rightRowScope != nil {
			return NewTupleSymbol(rows.rightRowScope.Columns(), rows.rightRowScope)
		}
	default:
		for _, symbol := range c.localSymbolsInScopeWithCancellation(position, cancellationToken) {
			cancellationToken.ThrowIfCancellationRequested()
			if symbol != nil && symbol.SymbolName() == name {
				return GetResultType(symbol)
			}
		}
	}
	return nil
}

func (c *KustoCode) incompletePathQualifierWithCancellation(position int, cancellationToken CancellationToken) string {
	cancellationToken.ThrowIfCancellationRequested()
	if c == nil || position < 0 || position > len(c.Text) {
		return ""
	}
	last := -1
	for i, token := range c.Tokens {
		cancellationToken.ThrowIfCancellationRequested()
		if i >= len(c.TokenStarts) || token.Kind == SyntaxKindEndOfTextToken {
			continue
		}
		textStart := c.TokenStarts[i] + len(token.Trivia)
		textEnd := textStart + len(token.Text)
		if textEnd <= position {
			last = i
		}
	}
	if last <= 0 || c.Tokens[last].Kind != SyntaxKindDotToken {
		return ""
	}
	dotEnd := c.TokenStarts[last] + len(c.Tokens[last].Trivia) + len(c.Tokens[last].Text)
	if strings.TrimSpace(c.Text[dotEnd:position]) != "" {
		return ""
	}
	return c.Tokens[last-1].Text
}

func (c *KustoCode) GetSpeculativeReferencedSymbol(position int, name string, match SymbolMatch) Symbol {
	return c.GetSpeculativeReferencedSymbolWithCancellation(position, name, match, CancellationToken{})
}

func (c *KustoCode) GetSpeculativeReferencedSymbolWithCancellation(position int, name string, match SymbolMatch, cancellationToken CancellationToken) Symbol {
	cancellationToken.ThrowIfCancellationRequested()
	if c == nil || !c.HasSemantics() || name == "" || match == SymbolMatchNone {
		return nil
	}
	for _, symbol := range c.GetSymbolsInScopeWithCancellation(position, match, IncludeFunctionKindAll, cancellationToken) {
		cancellationToken.ThrowIfCancellationRequested()
		if SymbolMatchesName(symbol, name, match, false) {
			return symbol
		}
	}
	return c.getAttachedSpeculativeReferencedSymbolWithCancellation(position, name, match, cancellationToken)
}

func (c *KustoCode) getAttachedSpeculativeReferencedSymbol(position int, name string, match SymbolMatch) Symbol {
	return c.getAttachedSpeculativeReferencedSymbolWithCancellation(position, name, match, CancellationToken{})
}

func (c *KustoCode) getAttachedSpeculativeReferencedSymbolWithCancellation(position int, name string, match SymbolMatch, cancellationToken CancellationToken) Symbol {
	cancellationToken.ThrowIfCancellationRequested()
	if c == nil || c.Syntax() == nil || name == "" || match == SymbolMatchNone {
		return nil
	}
	var best Symbol
	bestStart := -1
	walkSyntaxWithTriviaStartAndAncestors(c.Syntax(), 0, func(element SyntaxElement, triviaStart int, ancestors []SyntaxElement) {
		cancellationToken.ThrowIfCancellationRequested()
		declaration, ok := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
		if !ok || !kustoCodeSpeculativeDeclarationInScope(ancestors) {
			return
		}
		start := elementTextStart(declaration, triviaStart)
		scopeStart := kustoCodeSpeculativeDeclarationScopeStart(c.Syntax(), declaration, ancestors, triviaStart, position)
		if start <= bestStart || scopeStart > position || NameDeclarationSimpleName(declaration) != name {
			return
		}
		symbol := declaration.ReferencedSymbol()
		if SymbolMatchesName(symbol, name, match, false) {
			best = symbol
			bestStart = start
		}
	})
	return best
}

func kustoCodeSpeculativeDeclarationScopeStart(root SyntaxElement, declaration *NameDeclaration, ancestors []SyntaxElement, declarationTriviaStart int, position int) int {
	var statement Statement
	var separated *SeparatedElement
	for i := len(ancestors) - 1; i >= 0; i-- {
		switch node := WrapGeneratedSyntaxElement(ancestors[i]).(type) {
		case *LetStatement:
			statement = node
		case *QueryParametersStatement:
			statement = node
		}
		if statement != nil {
			if candidate, ok := ancestors[i].(*SeparatedElement); ok {
				separated = candidate
			}
		}
	}
	if statement != nil {
		if kustoCodeIsInTriviaAfter(root, statement, position) &&
			(kustoCodeStatementIsIncomplete(statement) || kustoCodeStatementCanHoldMore(statement)) {
			return int(^uint(0) >> 1)
		}
		if separated != nil {
			if start, found := syntaxElementTriviaStartOf(root, separated, 0); found {
				return elementEnd(separated, start)
			}
		}
		if start, ok := syntaxElementTriviaStartOf(root, statement, 0); ok {
			return elementEnd(statement, start)
		}
	}
	return elementEnd(declaration, declarationTriviaStart)
}

func syntaxElementTriviaStartOf(root, target SyntaxElement, triviaStart int) (int, bool) {
	if root == nil || target == nil {
		return 0, false
	}
	if root == target {
		return triviaStart, true
	}
	childStart := triviaStart
	for _, child := range root.Children() {
		if child != nil {
			if start, ok := syntaxElementTriviaStartOf(child, target, childStart); ok {
				return start, true
			}
			childStart += child.FullWidth()
		}
	}
	return 0, false
}

func kustoCodeSpeculativeDeclarationInScope(ancestors []SyntaxElement) bool {
	for _, ancestor := range ancestors {
		if _, ok := WrapGeneratedSyntaxElement(ancestor).(*FunctionDeclaration); ok {
			return false
		}
	}
	return true
}

func (c *KustoCode) GetColumnsInScope(position int) *TableSymbol {
	return c.GetColumnsInScopeWithCancellation(position, CancellationToken{})
}

func (c *KustoCode) GetColumnsInScopeWithCancellation(position int, cancellationToken CancellationToken) *TableSymbol {
	cancellationToken.ThrowIfCancellationRequested()
	return c.getRowScopeContextWithCancellation(position, cancellationToken).rowScope
}

type kustoCodeRowScopeContext struct {
	rowScope      *TableSymbol
	rightRowScope *TableSymbol
	bounded       bool
}

func (c *KustoCode) getRowScopeContext(position int) kustoCodeRowScopeContext {
	return c.getRowScopeContextWithCancellation(position, CancellationToken{})
}

func (c *KustoCode) getRowScopeContextWithCancellation(position int, cancellationToken CancellationToken) kustoCodeRowScopeContext {
	cancellationToken.ThrowIfCancellationRequested()
	if c == nil || !c.HasSemantics() || c.Syntax() == nil || position < 0 || position > len(c.Text) {
		return kustoCodeRowScopeContext{}
	}

	var context kustoCodeRowScopeContext
	bestWidth := int(^uint(0) >> 1)
	walkSyntaxWithTriviaStartWithCancellation(c.Syntax(), 0, cancellationToken, func(element SyntaxElement, triviaStart int) {
		wrapped := WrapGeneratedSyntaxElement(element)
		switch node := wrapped.(type) {
		case *ToScalarExpression, *ToTableExpression:
			end := triviaStart + element.FullWidth()
			inTrailingTrivia := position > end && position <= len(c.Text) && strings.TrimSpace(c.Text[end:position]) == ""
			width := end - triviaStart
			if position >= triviaStart && (position <= end || inTrailingTrivia) && width < bestWidth {
				context = kustoCodeRowScopeContext{bounded: true}
				bestWidth = width
			}
			return
		case *UnionOperator:
			if kustoCodeOperatorHasPipeParent(element) {
				break
			}
			if kustoCodeContextElementContains(c, element, triviaStart, position) && element.FullWidth() < bestWidth {
				context = kustoCodeRowScopeContext{bounded: true}
				bestWidth = element.FullWidth()
			}
			return
		case *FindOperator:
			if kustoCodeOperatorHasPipeParent(element) {
				break
			}
			if kustoCodeContextElementContains(c, element, triviaStart, position) && element.FullWidth() < bestWidth {
				context = c.findRowScopeContextWithCancellation(node, nil, position, triviaStart, cancellationToken)
				bestWidth = element.FullWidth()
			}
			return
		case *SearchOperator:
			if kustoCodeOperatorHasPipeParent(element) {
				break
			}
			if kustoCodeContextElementContains(c, element, triviaStart, position) && element.FullWidth() < bestWidth {
				context = c.searchRowScopeContextWithCancellation(node, nil, position, triviaStart, cancellationToken)
				bestWidth = element.FullWidth()
			}
			return
		case *InvokeOperator:
			if kustoCodeOperatorHasPipeParent(element) {
				break
			}
			if kustoCodeContextElementContains(c, element, triviaStart, position) && element.FullWidth() < bestWidth {
				context = invokeRowScopeContext(node, nil, position, triviaStart)
				bestWidth = element.FullWidth()
			}
			return
		case *MaterializedViewCombineExpression:
			clause := node.AggregationsClause()
			if clause == nil || clause.OpenParen() == nil || !kustoCodeContextElementContains(c, element, triviaStart, position) {
				return
			}
			openStart, found := syntaxElementTriviaStartOf(node, clause.OpenParen(), triviaStart)
			if !found || position <= elementTextStart(clause.OpenParen(), openStart) || element.FullWidth() >= bestWidth {
				return
			}
			context = kustoCodeRowScopeContext{
				rowScope: c.expressionTableResult(node.DeltaClause().Expression()),
				bounded:  true,
			}
			bestWidth = element.FullWidth()
			return
		case *FunctionCallExpression:
			if node.Name() == nil || node.ArgumentList() == nil || !kustoCodeContextElementContains(c, element, triviaStart, position) {
				return
			}
			nameStart, found := syntaxElementTriviaStartOf(node, node.Name(), triviaStart)
			if !found || position <= elementEnd(node.Name(), nameStart) || element.FullWidth() >= bestWidth {
				return
			}
			if rowScope, changed := c.functionArgumentRowScopeWithCancellation(node, position, context.rowScope, cancellationToken); changed {
				context = kustoCodeRowScopeContext{rowScope: rowScope, bounded: true}
				bestWidth = element.FullWidth()
			}
			return
		case *CustomCommand, *PartialCommand:
			if !kustoCodeContextElementContains(c, element, triviaStart, position) || element.FullWidth() >= bestWidth {
				return
			}
			if table := c.commandContextTableWithCancellation(element, triviaStart, position, cancellationToken); table != nil {
				context = kustoCodeRowScopeContext{rowScope: table, bounded: true}
				bestWidth = element.FullWidth()
			}
			return
		}

		if clause, ok := wrapped.(*MakeGraphTableAndKeyClause); ok && clause.Table() != nil && clause.OnKeyword() != nil {
			onStart := triviaStart + clause.Table().FullWidth()
			onTextStart := elementTextStart(clause.OnKeyword(), onStart)
			clauseEnd := triviaStart + clause.FullWidth()
			width := clauseEnd - onTextStart
			inTrailingTrivia := position > clauseEnd && position <= len(c.Text) && strings.TrimSpace(c.Text[clauseEnd:position]) == ""
			if position >= onTextStart && (position <= clauseEnd || inTrailingTrivia) && width < bestWidth {
				// ContextBuilder switches to the selected node table at the on keyword,
				// including when that table cannot be resolved and the scope is nil.
				context = kustoCodeRowScopeContext{rowScope: c.expressionTableResult(clause.Table()), bounded: true}
				bestWidth = width
			}
			return
		}

		pipe, ok := WrapGeneratedSyntaxElement(element).(*PipeExpression)
		if !ok || pipe.Expression() == nil || pipe.Bar() == nil || pipe.Operator() == nil {
			return
		}
		barStart := triviaStart + pipe.Expression().FullWidth()
		operatorStart := barStart + pipe.Bar().FullWidth()
		operatorEnd := operatorStart + pipe.Operator().FullWidth()
		inTrailingTrivia := position > operatorEnd && position <= len(c.Text) && strings.TrimSpace(c.Text[operatorEnd:position]) == ""
		if position < operatorStart || (position > operatorEnd && !inTrailingTrivia) {
			return
		}
		width := operatorEnd - operatorStart
		if width >= bestWidth {
			return
		}
		left := c.expressionTableResult(pipe.Expression())
		context = kustoCodeRowScopeContext{rowScope: left, bounded: true}
		switch operator := WrapGeneratedSyntaxElement(pipe.Operator()).(type) {
		case *UnionOperator:
			context.rowScope = nil
		case *FindOperator:
			context = c.findRowScopeContextWithCancellation(operator, left, position, operatorStart, cancellationToken)
		case *SearchOperator:
			context = c.searchRowScopeContextWithCancellation(operator, left, position, operatorStart, cancellationToken)
		case *InvokeOperator:
			context = invokeRowScopeContext(operator, left, position, operatorStart)
		case *GraphMergeOperator:
			context.rowScope = nil
			if clause := operator.OnClause(); clause != nil {
				clauseStart, found := syntaxElementTriviaStartOf(operator, clause, operatorStart)
				if found && position >= elementTextStart(clause, clauseStart) {
					context.rowScope = graphEdgeShape(expressionGraphResult(pipe.Expression()))
					context.rightRowScope = graphEdgeShape(expressionGraphResult(operator.Graph()))
				}
			}
		case *JoinOperator:
			context.rowScope = nil
			if clause := operator.ConditionClause(); clause != nil {
				clauseStart, found := syntaxElementTriviaStartOf(operator, clause, operatorStart)
				if found && position >= elementTextStart(clause, clauseStart) {
					context.rowScope = left
					context.rightRowScope = c.expressionTableResult(operator.Expression())
				}
			}
		case *LookupOperator:
			context.rowScope = nil
			if clause := operator.LookupClause(); clause != nil && !syntaxElementIsMissing(clause) {
				clauseStart, found := syntaxElementTriviaStartOf(operator, clause, operatorStart)
				if found && position >= elementTextStart(clause, clauseStart) {
					context.rowScope = left
					context.rightRowScope = c.expressionTableResult(operator.Expression())
				}
			}
		case *MvApplyOperator:
			if operator.OnKeyword() != nil && !syntaxElementIsMissing(operator.OnKeyword()) {
				onStart, found := syntaxElementTriviaStartOf(operator, operator.OnKeyword(), operatorStart)
				if found && position >= elementTextStart(operator.OnKeyword(), onStart) {
					context.rowScope = mvApplyIntermediateTable(operator, left)
				}
			}
		}
		bestWidth = width
	})
	return context
}

func (c *KustoCode) commandContextTableWithCancellation(command SyntaxElement, commandStart int, position int, cancellationToken CancellationToken) *TableSymbol {
	cancellationToken.ThrowIfCancellationRequested()
	if c == nil || command == nil {
		return nil
	}
	var nearest *TableSymbol
	walkSyntaxWithTriviaStartWithCancellation(command, commandStart, cancellationToken, func(element SyntaxElement, triviaStart int) {
		reference, ok := WrapGeneratedSyntaxElement(element).(*NameReference)
		if !ok || elementEnd(reference, triviaStart) > position {
			return
		}
		if table, ok := reference.ReferencedSymbol().(*TableSymbol); ok {
			nearest = table
			return
		}
		globals := c.globalsInScopeWithCancellation(position, cancellationToken)
		if reference.Match().Has(SymbolMatchTable) && globals != nil && globals.Database != nil {
			if table := globals.Database.GetTable(NameReferenceSimpleName(reference)); table != nil {
				nearest = table
			}
		}
	})
	return nearest
}

func (c *KustoCode) functionArgumentRowScopeWithCancellation(call *FunctionCallExpression, position int, defaultRowScope *TableSymbol, cancellationToken CancellationToken) (*TableSymbol, bool) {
	cancellationToken.ThrowIfCancellationRequested()
	if c == nil || call == nil || call.Name() == nil || call.ArgumentList() == nil {
		return defaultRowScope, false
	}
	function := functionCallReferencedFunction(call)
	globals := c.globalsInScopeWithCancellation(position, cancellationToken)
	if function == nil && globals != nil {
		name := NameReferenceSimpleName(call.Name())
		function = globals.GetFunction(name)
		if function == nil {
			function = globals.GetAggregate(name)
		}
		if function == nil {
			function = globals.GetPlugIn(name)
		}
		if function == nil && globals.Database != nil {
			function = globals.Database.GetFunction(name)
		}
	}
	if function == nil {
		return defaultRowScope, false
	}

	arguments := functionCallCompletionArguments(call)
	argumentIndex := functionCallCompletionArgumentIndex(call, position)
	signatures := function.Signatures()
	if signature := call.ReferencedSignature(); signature != nil {
		signatures = []*Signature{signature}
	}
	possible := functionPossibleParameters(functionArgumentCompletionContext{
		call:          call,
		function:      function,
		signatures:    signatures,
		arguments:     arguments,
		argumentIndex: argumentIndex,
	}, globals)

	var parameter0, parameter0Common, parameter0Element bool
	for _, candidate := range possible {
		cancellationToken.ThrowIfCancellationRequested()
		if candidate.parameter == nil {
			continue
		}
		switch candidate.parameter.ArgumentKind {
		case ArgumentKindColumnParameter0:
			parameter0 = true
		case ArgumentKindColumnParameter0Common:
			parameter0Common = true
		case ArgumentKindExpressionParameter0Element:
			parameter0Element = true
		}
	}
	if len(arguments) == 0 {
		return defaultRowScope, false
	}
	if parameter0Element {
		if tuple, ok := TupleForType(ExpressionResultType(arguments[0])); ok {
			return NewTableSymbol("", tuple.Columns()...), true
		}
	}
	if parameter0 || parameter0Common {
		if table := c.expressionTableResult(arguments[0]); table != nil {
			if parameter0 {
				return table, true
			}
			return NewTableSymbol("", kustoCodeCommonColumnsWithCancellation(defaultRowScope, table, cancellationToken)...), true
		}
	}
	return defaultRowScope, false
}

func functionCallReferencedFunction(call *FunctionCallExpression) *FunctionSymbol {
	if call == nil {
		return nil
	}
	if function, ok := call.ReferencedSymbol().(*FunctionSymbol); ok {
		return function
	}
	if call.Name() != nil {
		if function, ok := call.Name().ReferencedSymbol().(*FunctionSymbol); ok {
			return function
		}
	}
	return nil
}

func kustoCodeOperatorHasPipeParent(element SyntaxElement) bool {
	if element == nil || element.Parent() == nil {
		return false
	}
	_, ok := WrapGeneratedSyntaxElement(element.Parent()).(*PipeExpression)
	return ok
}

func (c *KustoCode) findRowScopeContextWithCancellation(node *FindOperator, incoming *TableSymbol, position int, operatorStart int, cancellationToken CancellationToken) kustoCodeRowScopeContext {
	cancellationToken.ThrowIfCancellationRequested()
	context := kustoCodeRowScopeContext{rowScope: incoming, bounded: true}
	if node == nil {
		return context
	}
	clause := node.InClause()
	if clause == nil {
		context.rowScope = c.editorFindColumnsTableWithCancellation(node, position, cancellationToken)
		return context
	}
	if start, ok := syntaxElementTriviaStartOf(node, clause, operatorStart); ok && position >= elementEnd(clause, start) {
		context.rowScope = c.editorFindColumnsTableWithCancellation(node, position, cancellationToken)
	}
	return context
}

func (c *KustoCode) searchRowScopeContextWithCancellation(node *SearchOperator, incoming *TableSymbol, position int, operatorStart int, cancellationToken CancellationToken) kustoCodeRowScopeContext {
	cancellationToken.ThrowIfCancellationRequested()
	context := kustoCodeRowScopeContext{rowScope: incoming, bounded: true}
	if node == nil {
		return context
	}
	if condition := node.Condition(); condition != nil {
		if start, ok := syntaxElementTriviaStartOf(node, condition, operatorStart); ok && position >= elementTextStart(condition, start) {
			context.rowScope = c.editorSearchColumnsTableWithCancellation(node, incoming, position, cancellationToken)
		}
	}
	if clause := node.InClause(); clause != nil {
		if start, ok := syntaxElementTriviaStartOf(node, clause, operatorStart); ok && position >= elementTextStart(clause, start) {
			context.rowScope = nil
		}
	}
	return context
}

func invokeRowScopeContext(node *InvokeOperator, incoming *TableSymbol, position int, operatorStart int) kustoCodeRowScopeContext {
	context := kustoCodeRowScopeContext{rowScope: incoming, bounded: true}
	if node == nil || node.Function() == nil || syntaxElementIsMissing(node.Function()) {
		return context
	}
	if start, ok := syntaxElementTriviaStartOf(node, node.Function(), operatorStart); ok && position > elementTextStart(node.Function(), start) {
		context.rowScope = nil
	}
	return context
}

func (c *KustoCode) editorFindColumnsTableWithCancellation(node *FindOperator, position int, cancellationToken CancellationToken) *TableSymbol {
	binder := c.editorContextBinderWithCancellation(position, nil, cancellationToken)
	return binder.findColumnsTable(node)
}

func (c *KustoCode) editorSearchColumnsTableWithCancellation(node *SearchOperator, incoming *TableSymbol, position int, cancellationToken CancellationToken) *TableSymbol {
	binder := c.editorContextBinderWithCancellation(position, incoming, cancellationToken)
	return binder.searchColumnsTable(node)
}

func (c *KustoCode) editorContextBinderWithCancellation(position int, rowScope *TableSymbol, cancellationToken CancellationToken) *Binder {
	cancellationToken.ThrowIfCancellationRequested()
	binder := NewBinder(c.globalsInScopeWithCancellation(position, cancellationToken))
	binder.rowScope = rowScope
	for _, symbol := range c.lexicalSymbolsInScopeWithCancellation(position, cancellationToken) {
		cancellationToken.ThrowIfCancellationRequested()
		binder.LocalScope.AddSymbol(symbol)
	}
	return binder
}

func (c *KustoCode) expressionTableResult(expression Expression) *TableSymbol {
	if table := expressionTableResult(expression); table != nil {
		return table
	}
	if c == nil || c.Globals == nil || c.Globals.Database == nil || expression == nil {
		return nil
	}
	switch source := expression.(type) {
	case *NameReference:
		return c.Globals.Database.GetTable(NameReferenceSimpleName(source))
	case *ParenthesizedExpression:
		return c.expressionTableResult(source.Expression())
	default:
		return nil
	}
}

func kustoCodeCommonColumns(left, right *TableSymbol) []*ColumnSymbol {
	return kustoCodeCommonColumnsWithCancellation(left, right, CancellationToken{})
}

func kustoCodeCommonColumnsWithCancellation(left, right *TableSymbol, cancellationToken CancellationToken) []*ColumnSymbol {
	cancellationToken.ThrowIfCancellationRequested()
	if left == nil || right == nil {
		return nil
	}
	rightNames := make(map[string]struct{}, len(right.Columns()))
	for _, column := range right.Columns() {
		cancellationToken.ThrowIfCancellationRequested()
		if column != nil {
			rightNames[column.SymbolName()] = struct{}{}
		}
	}
	leftNames := make(map[string]struct{}, len(left.Columns()))
	var common []*ColumnSymbol
	for _, column := range left.Columns() {
		cancellationToken.ThrowIfCancellationRequested()
		if column == nil {
			continue
		}
		leftNames[column.SymbolName()] = struct{}{}
		if _, ok := rightNames[column.SymbolName()]; ok {
			common = append(common, column)
		}
	}
	for _, column := range right.Columns() {
		cancellationToken.ThrowIfCancellationRequested()
		if column != nil {
			if _, ok := leftNames[column.SymbolName()]; ok {
				common = append(common, column)
			}
		}
	}
	return unifyColumnsWithSameName(common)
}

type kustoCodeFunctionSource uint8

const (
	kustoCodeFunctionSourceNone kustoCodeFunctionSource = iota
	kustoCodeFunctionSourceGlobal
	kustoCodeFunctionSourceBuiltIn
	kustoCodeFunctionSourceDatabase
	kustoCodeFunctionSourceLocal
)

type kustoCodeScopeCollector struct {
	match             SymbolMatch
	include           IncludeFunctionKind
	symbols           []Symbol
	cancellationToken CancellationToken
}

func newKustoCodeScopeCollector(match SymbolMatch, include IncludeFunctionKind, tokens ...CancellationToken) *kustoCodeScopeCollector {
	var cancellationToken CancellationToken
	if len(tokens) > 0 {
		cancellationToken = tokens[0]
	}
	return &kustoCodeScopeCollector{
		match:             match,
		include:           include,
		cancellationToken: cancellationToken,
	}
}

func (s *kustoCodeScopeCollector) add(symbol Symbol, source kustoCodeFunctionSource) {
	if s != nil {
		s.cancellationToken.ThrowIfCancellationRequested()
	}
	if s == nil || symbol == nil || symbol.SymbolName() == "" || !SymbolMatches(symbol, s.match) {
		return
	}
	if flag := kustoCodeScopeFunctionInclusionFlag(symbol, source); flag != 0 && s.include&flag == 0 {
		return
	}
	s.symbols = append(s.symbols, symbol)
}

func (s *kustoCodeScopeCollector) addContextSymbol(symbol Symbol) {
	if s != nil {
		s.cancellationToken.ThrowIfCancellationRequested()
	}
	if s == nil || symbol == nil || symbol.SymbolName() == "" {
		return
	}
	s.symbols = append(s.symbols, symbol)
}

func kustoCodeScopeFunctionInclusionFlag(symbol Symbol, source kustoCodeFunctionSource) IncludeFunctionKind {
	var function *FunctionSymbol
	switch typed := symbol.(type) {
	case *FunctionSymbol:
		function = typed
	case *VariableSymbol:
		function, _ = typed.Type.(*FunctionSymbol)
	}
	if function == nil {
		return IncludeFunctionKindNone
	}
	switch source {
	case kustoCodeFunctionSourceBuiltIn:
		return IncludeFunctionKindBuiltInFunctions
	case kustoCodeFunctionSourceDatabase:
		return IncludeFunctionKindDatabaseFunctions
	case kustoCodeFunctionSourceLocal:
		if function.IsView() {
			return IncludeFunctionKindLocalViews
		}
		return IncludeFunctionKindLocalFunctions
	default:
		return IncludeFunctionKindNone
	}
}

func (c *KustoCode) localSymbolsInScope(position int) []Symbol {
	return c.localSymbolsInScopeWithCancellation(position, CancellationToken{})
}

func (c *KustoCode) localSymbolsInScopeWithCancellation(position int, cancellationToken CancellationToken) []Symbol {
	return c.collectLocalSymbolsInScopeWithCancellation(position, true, cancellationToken)
}

func (c *KustoCode) lexicalSymbolsInScopeWithCancellation(position int, cancellationToken CancellationToken) []Symbol {
	return c.collectLocalSymbolsInScopeWithCancellation(position, false, cancellationToken)
}

func (c *KustoCode) collectLocalSymbolsInScopeWithCancellation(position int, includeOperatorLocals bool, cancellationToken CancellationToken) []Symbol {
	cancellationToken.ThrowIfCancellationRequested()
	if c == nil || c.Syntax() == nil {
		return nil
	}

	ordered := newKustoCodeLocalSymbols(cancellationToken)
	if c.Query != nil {
		addCompletedStatementDeclarations(c.Syntax(), c.Query.Statements(), position, ordered)
	} else {
		root, ok := WrapGeneratedSyntaxElement(c.Syntax()).(*CommandBlock)
		if !ok {
			return nil
		}
		addCommandBlockDeclarations(c.Syntax(), root, position, ordered)
	}

	type functionScope struct {
		declaration *FunctionDeclaration
		body        *FunctionBody
		bodyStart   int
	}
	var scopes []functionScope
	walkSyntaxWithTriviaStartAndAncestors(c.Syntax(), 0, func(element SyntaxElement, triviaStart int, ancestors []SyntaxElement) {
		cancellationToken.ThrowIfCancellationRequested()
		body, ok := WrapGeneratedSyntaxElement(element).(*FunctionBody)
		if !ok {
			return
		}
		bodyTextStart := elementTextStart(body, triviaStart)
		bodyEnd := elementEnd(body, triviaStart)
		closeBrace := body.CloseBrace()
		closeBraceMissing := closeBrace == nil || closeBrace.Width() == 0
		if position < bodyTextStart || (position >= bodyEnd && !(closeBraceMissing && position <= bodyEnd+1)) {
			return
		}
		var declaration *FunctionDeclaration
		for i := len(ancestors) - 1; i >= 0; i-- {
			if candidate, ok := WrapGeneratedSyntaxElement(ancestors[i]).(*FunctionDeclaration); ok {
				declaration = candidate
				break
			}
		}
		if declaration != nil {
			scopes = append(scopes, functionScope{declaration: declaration, body: body, bodyStart: triviaStart})
		}
	})
	for _, scope := range scopes {
		cancellationToken.ThrowIfCancellationRequested()
		addFunctionParameters(scope.declaration, ordered)
		closeBrace := scope.body.CloseBrace()
		closeBraceMissing := closeBrace == nil || closeBrace.Width() == 0
		closeBraceTextStart := elementEnd(scope.body, scope.bodyStart)
		if closeBrace != nil {
			if closeBraceStart, found := syntaxElementTriviaStartOf(scope.body, closeBrace, scope.bodyStart); found {
				closeBraceTextStart = elementTextStart(closeBrace, closeBraceStart)
			}
		}
		if closeBraceMissing || position < closeBraceTextStart {
			addCompletedStatementDeclarations(c.Syntax(), scope.body.Statements(), position, ordered)
		}
	}
	addPatternParametersInScope(c.Syntax(), position, ordered)
	if includeOperatorLocals {
		addScanDeclarationsInScope(c, position, ordered)
		addGraphPatternDeclarationsInScope(c, position, ordered)
		addPartitionKeyInScope(c, position, ordered)
		addInlineExternalTableDeclarationsInScope(c, position, ordered)
		addMacroExpandScopeInScope(c, position, ordered)
	}
	return ordered.values()
}

func addPartitionKeyInScope(code *KustoCode, position int, symbols *kustoCodeLocalSymbols) {
	if code == nil || code.Syntax() == nil || symbols == nil {
		return
	}
	walkSyntaxWithTriviaStartWithCancellation(code.Syntax(), 0, symbols.cancellationToken, func(element SyntaxElement, triviaStart int) {
		node, ok := WrapGeneratedSyntaxElement(element).(*PartitionOperator)
		if !ok || node.Operand() == nil || !kustoCodeContextElementContains(code, element, triviaStart, position) {
			return
		}
		operandStart, found := syntaxElementTriviaStartOf(node, node.Operand(), triviaStart)
		if !found || position < elementTextStart(node.Operand(), operandStart) {
			return
		}
		column, _ := expressionReferencedSymbol(node.ByExpression()).(*ColumnSymbol)
		if column == nil {
			name, nameOK := node.ByExpression().(*NameReference)
			pipe, pipeOK := WrapGeneratedSyntaxElement(node.Parent()).(*PipeExpression)
			if nameOK && pipeOK {
				if table := code.expressionTableResult(pipe.Expression()); table != nil {
					column = table.GetColumn(NameReferenceSimpleName(name))
				}
			}
		}
		symbols.add(column)
	})
}

func addInlineExternalTableDeclarationsInScope(code *KustoCode, position int, symbols *kustoCodeLocalSymbols) {
	if code == nil || code.Syntax() == nil || symbols == nil {
		return
	}
	walkSyntaxWithTriviaStartWithCancellation(code.Syntax(), 0, symbols.cancellationToken, func(element SyntaxElement, triviaStart int) {
		node, ok := WrapGeneratedSyntaxElement(element).(*InlineExternalTableExpression)
		if !ok || !kustoCodeContextElementContains(code, element, triviaStart, position) {
			return
		}
		if path := node.PathFormat(); path != nil {
			pathStart, found := syntaxElementTriviaStartOf(node, path, triviaStart)
			if found && position > elementTextStart(path, pathStart) {
				if partition := node.PartitionClause(); partition != nil {
					for _, separated := range partition.PartitionColumns().Elements() {
						if separated == nil || separated.Element() == nil {
							continue
						}
						declaration := separated.Element()
						column, _ := declaration.Name().ReferencedSymbol().(*ColumnSymbol)
						if column == nil {
							column = NewColumnSymbolWithSource(NameDeclarationSimpleName(declaration.Name()), typeFromBoundExpression(declaration.Type()), declaration.Name())
						}
						symbols.add(column)
					}
				}
				return
			}
		}
		if schema := node.Schema(); schema != nil {
			for _, separated := range schema.Columns().Elements() {
				if separated == nil || separated.Element() == nil {
					continue
				}
				declaration := separated.Element()
				column, _ := declaration.Name().ReferencedSymbol().(*ColumnSymbol)
				if column == nil {
					column = scanDeclaredColumn(declaration)
				}
				symbols.add(column)
			}
		}
	})
}

func addMacroExpandScopeInScope(code *KustoCode, position int, symbols *kustoCodeLocalSymbols) {
	if code == nil || code.Syntax() == nil || symbols == nil {
		return
	}
	walkSyntaxWithTriviaStartWithCancellation(code.Syntax(), 0, symbols.cancellationToken, func(element SyntaxElement, triviaStart int) {
		node, ok := WrapGeneratedSyntaxElement(element).(*MacroExpandOperator)
		if !ok || node.OpenParen() == nil || !kustoCodeContextElementContains(code, element, triviaStart, position) {
			return
		}
		openStart, found := syntaxElementTriviaStartOf(node, node.OpenParen(), triviaStart)
		if !found || position < elementEnd(node.OpenParen(), openStart) {
			return
		}

		var group *EntityGroupSymbol
		switch entity := expressionReferencedSymbol(node.EntityGroup()).(type) {
		case *EntityGroupSymbol:
			group = entity
		case *VariableSymbol:
			group, _ = entity.Type.(*EntityGroupSymbol)
		}
		name := ""
		if clause := node.ScopeReferenceName(); clause != nil && clause.EntityGroupReferenceName() != nil {
			declaration := clause.EntityGroupReferenceName()
			if symbol := declaration.ReferencedSymbol(); symbol != nil {
				symbols.add(symbol)
				return
			}
			name = NameDeclarationSimpleName(declaration)
		} else if reference, ok := node.EntityGroup().(*NameReference); ok {
			name = NameReferenceSimpleName(reference)
		}
		if name != "" {
			symbols.add(NewEntityGroupElementSymbol(name, group, nil))
		}
	})
}

func addGraphPatternDeclarationsInScope(code *KustoCode, position int, symbols *kustoCodeLocalSymbols) {
	if code == nil || code.Syntax() == nil || symbols == nil {
		return
	}
	walkSyntaxWithTriviaStartWithCancellation(code.Syntax(), 0, symbols.cancellationToken, func(element SyntaxElement, triviaStart int) {
		pipe, ok := WrapGeneratedSyntaxElement(element).(*PipeExpression)
		if !ok || pipe.Expression() == nil || pipe.Operator() == nil {
			return
		}

		var patterns *SeparatedSyntaxListOf[*GraphMatchPattern]
		var where *WhereClause
		var project *ProjectClause
		switch operator := WrapGeneratedSyntaxElement(pipe.Operator()).(type) {
		case *GraphMatchOperator:
			patterns, where, project = operator.Patterns(), operator.WhereClause(), operator.ProjectClause()
		case *GraphShortestPathsOperator:
			patterns, where, project = operator.Patterns(), operator.WhereClause(), operator.ProjectClause()
		default:
			return
		}

		operatorStart, found := syntaxElementTriviaStartOf(pipe, pipe.Operator(), triviaStart)
		if !found {
			return
		}
		operatorEnd := operatorStart + pipe.Operator().FullWidth()
		inTrailingTrivia := position > operatorEnd && position <= len(code.Text) && strings.TrimSpace(code.Text[operatorEnd:position]) == ""
		if position < operatorStart || (position > operatorEnd && !inTrailingTrivia) {
			return
		}
		active := false
		for _, clause := range []SyntaxElement{where, project} {
			if clause == nil {
				continue
			}
			clauseStart, ok := syntaxElementTriviaStartOf(pipe.Operator(), clause, operatorStart)
			if ok && position >= elementTextStart(clause, clauseStart) {
				active = true
				break
			}
		}
		if !active {
			return
		}

		graph := expressionGraphResult(pipe.Expression())
		edgeTuple, nodeTuple := NewTupleSymbol(nil), NewTupleSymbol(nil)
		if graph != nil {
			edgeTuple = NewTupleSymbol(graph.EdgeShape.Columns(), graph.EdgeShape)
			if graph.NodeShape != nil {
				nodeTuple = NewTupleSymbol(graph.NodeShape.Columns(), graph.NodeShape)
			}
		}
		for _, separated := range patterns.Elements() {
			if separated == nil || separated.Element() == nil || separated.Element().PatternElements() == nil {
				continue
			}
			for _, notation := range separated.Element().PatternElements().Elements() {
				var declaration *NameDeclaration
				var tuple *TupleSymbol
				switch notation := notation.(type) {
				case *GraphMatchPatternNode:
					declaration, tuple = notation.Name(), nodeTuple
				case *GraphMatchPatternEdge:
					declaration, tuple = notation.Name(), edgeTuple
				}
				if declaration == nil || NameDeclarationSimpleName(declaration) == "" {
					continue
				}
				variable, _ := declaration.ReferencedSymbol().(*VariableSymbol)
				if variable == nil {
					variable = NewVariableSymbol(NameDeclarationSimpleName(declaration), tuple, false, nil, nil)
				}
				symbols.add(variable)
			}
		}
	})
}

func addScanDeclarationsInScope(code *KustoCode, position int, symbols *kustoCodeLocalSymbols) {
	if code == nil || code.Syntax() == nil || symbols == nil {
		return
	}
	walkSyntaxWithTriviaStartWithCancellation(code.Syntax(), 0, symbols.cancellationToken, func(element SyntaxElement, triviaStart int) {
		scan, ok := WrapGeneratedSyntaxElement(element).(*ScanOperator)
		if !ok || scan.DeclareClause() == nil || scan.WithKeyword() == nil {
			return
		}
		scanEnd := triviaStart + scan.FullWidth()
		inTrailingTrivia := position > scanEnd && position <= len(code.Text) && strings.TrimSpace(code.Text[scanEnd:position]) == ""
		if position < triviaStart || (position > scanEnd && !inTrailingTrivia) {
			return
		}
		withStart, found := syntaxElementTriviaStartOf(scan, scan.WithKeyword(), triviaStart)
		if !found || position <= elementTextStart(scan.WithKeyword(), withStart) {
			return
		}

		var declared []*ColumnSymbol
		for _, separated := range scan.DeclareClause().Declarations().Elements() {
			if separated == nil || separated.Element() == nil || separated.Element().NameAndType() == nil {
				continue
			}
			declaration := separated.Element().NameAndType()
			column, _ := declaration.Name().ReferencedSymbol().(*ColumnSymbol)
			if column == nil {
				column = scanDeclaredColumn(declaration)
			}
			declared = append(declared, column)
			symbols.add(column)
		}

		rowScope := code.getRowScopeContextWithCancellation(position, symbols.cancellationToken).rowScope
		tupleColumns := make([]*ColumnSymbol, 0, len(declared))
		if rowScope != nil {
			tupleColumns = append(tupleColumns, rowScope.Columns()...)
		}
		tupleColumns = append(tupleColumns, declared...)
		stepTuple := NewTupleSymbol(tupleColumns, rowScope)
		for _, step := range scan.Steps().Elements() {
			if step == nil || step.Name() == nil {
				continue
			}
			variable, _ := step.Name().ReferencedSymbol().(*VariableSymbol)
			if variable == nil {
				variable = NewVariableSymbol(NameDeclarationSimpleName(step.Name()), stepTuple, false, nil, nil)
			}
			symbols.add(variable)
		}
	})
}

func addPatternParametersInScope(root SyntaxElement, position int, symbols *kustoCodeLocalSymbols) {
	if root == nil || symbols == nil {
		return
	}
	walkSyntaxWithTriviaStartWithCancellation(root, 0, symbols.cancellationToken, func(element SyntaxElement, triviaStart int) {
		declaration, ok := WrapGeneratedSyntaxElement(element).(*PatternDeclaration)
		if !ok || position < triviaStart || position > elementEnd(declaration, triviaStart) || declaration.Patterns() == nil {
			return
		}
		patternsStart, found := syntaxElementTriviaStartOf(root, declaration.Patterns().SyntaxList, 0)
		if !found || position <= elementTextStart(declaration.Patterns(), patternsStart) {
			return
		}
		if declaration.Parameters() != nil {
			for _, separated := range declaration.Parameters().Elements() {
				if separated == nil || separated.Element() == nil || separated.Element().Name() == nil {
					continue
				}
				symbols.add(separated.Element().Name().ReferencedSymbol())
			}
		}
		if declaration.PathParameter() != nil && declaration.PathParameter().Parameter() != nil && declaration.PathParameter().Parameter().Name() != nil {
			symbols.add(declaration.PathParameter().Parameter().Name().ReferencedSymbol())
		}
	})
}

func addCommandBlockDeclarations(root SyntaxElement, block *CommandBlock, position int, symbols *kustoCodeLocalSymbols) {
	if root == nil || block == nil || block.Statements() == nil || block.Statements().Count() == 0 || symbols == nil {
		return
	}
	first := block.Statements().At(0)
	if first == nil || first.Element() == nil || !commandBlockQueryStatementPosition(root, block, position) {
		addCompletedStatementDeclarations(root, block.Statements(), position, symbols)
		return
	}
	command := firstCommandDescendant(first.Element())
	commandExpression, ok := command.(Expression)
	if !ok {
		return
	}
	resultType := ExpressionResultType(commandExpression)
	if resultType == nil {
		resultType = &Error.TypeSymbol
	}
	symbols.add(NewVariableSymbol("$command_results", resultType, false, nil, commandExpression))
}

func commandBlockQueryStatementPosition(root SyntaxElement, block *CommandBlock, position int) bool {
	if root == nil || block == nil || block.Statements() == nil || block.Statements().Count() == 0 {
		return false
	}
	first := block.Statements().At(0)
	if first == nil || first.Element() == nil || first.Separator() == nil {
		return false
	}
	start, ok := syntaxElementTriviaStartOf(root, first.SeparatedElement, 0)
	return ok && position > elementEnd(first.SeparatedElement, start)
}

type kustoCodeLocalSymbols struct {
	symbols           []Symbol
	index             map[string]int
	cancellationToken CancellationToken
}

func newKustoCodeLocalSymbols(tokens ...CancellationToken) *kustoCodeLocalSymbols {
	var cancellationToken CancellationToken
	if len(tokens) > 0 {
		cancellationToken = tokens[0]
	}
	return &kustoCodeLocalSymbols{index: map[string]int{}, cancellationToken: cancellationToken}
}

func (s *kustoCodeLocalSymbols) add(symbol Symbol) {
	if s != nil {
		s.cancellationToken.ThrowIfCancellationRequested()
	}
	if s == nil || symbol == nil || symbol.SymbolName() == "" {
		return
	}
	if index, exists := s.index[symbol.SymbolName()]; exists {
		s.symbols[index] = symbol
		return
	}
	s.index[symbol.SymbolName()] = len(s.symbols)
	s.symbols = append(s.symbols, symbol)
}

func (s *kustoCodeLocalSymbols) values() []Symbol {
	if s == nil {
		return nil
	}
	return append([]Symbol{}, s.symbols...)
}

func addCompletedStatementDeclarations(root SyntaxElement, statements *SeparatedSyntaxListOf[Statement], position int, symbols *kustoCodeLocalSymbols) {
	if root == nil || statements == nil || symbols == nil {
		return
	}
	for _, separated := range statements.Elements() {
		symbols.cancellationToken.ThrowIfCancellationRequested()
		if separated == nil || separated.Element() == nil {
			continue
		}
		statement := separated.Element()
		separatedStart, ok := syntaxElementTriviaStartOf(root, separated.SeparatedElement, 0)
		if !ok {
			continue
		}
		if position < elementEnd(separated.SeparatedElement, separatedStart) ||
			(kustoCodeIsInTriviaAfter(root, statement, position) &&
				(kustoCodeStatementIsIncomplete(statement) || kustoCodeStatementCanHoldMore(statement))) {
			break
		}
		addStatementDeclarations(statement, symbols)
	}
}

func kustoCodeStatementIsIncomplete(statement Statement) bool {
	if statement == nil {
		return false
	}
	last := statement.GetLastToken(true)
	return last != nil && last.Width() == 0
}

func kustoCodeIsInTriviaAfter(root SyntaxElement, node SyntaxElement, position int) bool {
	if root == nil || node == nil {
		return false
	}
	last := node.GetLastToken()
	if last == nil {
		return false
	}
	next := last.GetNextToken()
	if next == nil || next.Kind() == SyntaxKindEndOfTextToken {
		return true
	}
	nextStart, found := syntaxElementTriviaStartOf(root, next, 0)
	return found && position < elementTextStart(next, nextStart)
}

func kustoCodeStatementCanHoldMore(statement Statement) bool {
	if statement == nil {
		return false
	}
	last := statement.GetLastToken(true)
	if last == nil {
		return false
	}
	statementTextStart := elementTextStart(statement, 0)
	statementNode := simpleBinderSyntaxNode(statement)
	for subNode := last.Parent(); subNode != nil && subNode != statement && subNode != statementNode; subNode = subNode.Parent() {
		subStart, found := syntaxElementTriviaStartOf(statement, subNode, 0)
		if !found || elementTextStart(subNode, subStart) <= statementTextStart {
			continue
		}
		for index := subNode.ChildCount() - 1; index >= 0; index-- {
			child := subNode.GetChild(index)
			if child == nil && subNode.IsOptional(index) {
				return true
			}
			if child != nil && child.Kind() == SyntaxKindList {
				return true
			}
			if child != nil && child.Width() > 0 {
				break
			}
		}
	}
	return false
}

func addStatementDeclarations(statement Statement, symbols *kustoCodeLocalSymbols) {
	switch statement := statement.(type) {
	case *LetStatement:
		if statement.Name() != nil {
			symbols.add(statement.Name().ReferencedSymbol())
		}
	case *QueryParametersStatement:
		if statement.Parameters() == nil {
			return
		}
		for _, separated := range statement.Parameters().Elements() {
			if separated == nil || separated.Element() == nil || separated.Element().NameAndType() == nil || separated.Element().NameAndType().Name() == nil {
				continue
			}
			symbols.add(separated.Element().NameAndType().Name().ReferencedSymbol())
		}
	case *PatternStatement:
		if statement.Name() != nil {
			symbols.add(statement.Name().ReferencedSymbol())
		}
	}
}

func addFunctionParameters(declaration *FunctionDeclaration, symbols *kustoCodeLocalSymbols) {
	if declaration == nil || declaration.Parameters() == nil || declaration.Parameters().Parameters() == nil {
		return
	}
	for _, separated := range declaration.Parameters().Parameters().Elements() {
		if separated == nil || separated.Element() == nil || separated.Element().NameAndType() == nil || separated.Element().NameAndType().Name() == nil {
			continue
		}
		symbols.add(separated.Element().NameAndType().Name().ReferencedSymbol())
	}
}

func (c *KustoCode) addNormalGlobalSymbolsInScope(globals *GlobalState, collector *kustoCodeScopeCollector) {
	if c == nil || globals == nil || collector == nil {
		return
	}
	// Binder.GetSymbolsInContext enumerates built-ins before database members.
	for _, function := range globals.Functions {
		collector.add(function, kustoCodeFunctionSourceBuiltIn)
	}
	if globals.Database != nil {
		for _, member := range globals.Database.Members() {
			if _, ok := member.(*FunctionSymbol); ok {
				collector.add(member, kustoCodeFunctionSourceDatabase)
			} else {
				collector.add(member, kustoCodeFunctionSourceGlobal)
			}
		}
	}
	if globals.Cluster != nil {
		for _, database := range globals.Cluster.Databases() {
			collector.add(database, kustoCodeFunctionSourceGlobal)
		}
	}
	for _, cluster := range globals.Clusters {
		collector.add(cluster, kustoCodeFunctionSourceGlobal)
	}
}

func (c *KustoCode) addAggregateSymbolsInScope(position int, globals *GlobalState, collector *kustoCodeScopeCollector) {
	if c == nil || globals == nil || collector == nil {
		return
	}
	for _, function := range globals.Aggregates {
		collector.add(function, kustoCodeFunctionSourceBuiltIn)
	}
	for _, function := range globals.Functions {
		collector.add(function, kustoCodeFunctionSourceBuiltIn)
	}
	for _, symbol := range c.localSymbolsInScope(position) {
		switch symbol := symbol.(type) {
		case *FunctionSymbol, *PatternSymbol:
			collector.add(symbol, kustoCodeFunctionSourceLocal)
		case *VariableSymbol:
			if _, ok := symbol.Type.(*FunctionSymbol); ok {
				collector.add(symbol, kustoCodeFunctionSourceLocal)
			}
		}
	}
	if globals.Database != nil {
		for _, function := range globals.Database.Functions() {
			collector.add(function, kustoCodeFunctionSourceDatabase)
		}
	}
}

func (c *KustoCode) addPlugInSymbolsInScope(collector *kustoCodeScopeCollector) {
	if c == nil || c.Globals == nil || collector == nil {
		return
	}
	for _, function := range c.Globals.PlugIns {
		collector.add(function, kustoCodeFunctionSourceBuiltIn)
	}
}

func (c *KustoCode) addOptionSymbolsInScope(collector *kustoCodeScopeCollector) {
	if c == nil || c.Globals == nil || collector == nil {
		return
	}
	for _, option := range c.Globals.Options {
		collector.add(option, kustoCodeFunctionSourceGlobal)
	}
}

func (c *KustoCode) MaxDepth() int {
	if c == nil || c.Tree == nil {
		return 0
	}
	return c.Tree.Depth()
}

func (c *KustoCode) Dialect() KustoDialect {
	if c == nil {
		return KustoDialectUnknown
	}
	switch c.Kind {
	case CodeKindCommand:
		if c.Globals == nil {
			return KustoDialectEngineCommand
		}
		switch c.Globals.ServerKind {
		case ServerKindClusterManager:
			return KustoDialectClusterManagerCommand
		case ServerKindDataManager:
			return KustoDialectDataManagerCommand
		case ServerKindEngine:
			fallthrough
		default:
			return KustoDialectEngineCommand
		}
	case CodeKindQuery, CodeKindDirective:
		return KustoDialectQuery
	default:
		return KustoDialectUnknown
	}
}

func CodeKindOf(text string) CodeKind {
	return codeKindOfWithCancellation(text, CancellationToken{})
}

func codeKindOfWithCancellation(text string, cancellationToken CancellationToken) CodeKind {
	position := 0
	poller := &lexerCancellationPoller{cancellation: cancellationToken}
	for position < len(text) {
		tokenStart := scanTriviaWithCancellation(text, position, poller)
		if tokenStart >= len(text) {
			break
		}
		switch text[tokenStart] {
		case '.':
			return CodeKindCommand
		case '#':
			nextStart := NextLineStart(text, tokenStart+1)
			if nextStart > tokenStart {
				position = nextStart
				continue
			}
		}
		break
	}
	return CodeKindQuery
}

func GetKind(text string) CodeKind {
	return CodeKindOf(text)
}

func (c *KustoCode) TryGetLineAndOffset(position int) (line int, lineOffset int, ok bool) {
	if c == nil {
		return 0, 0, false
	}
	line, lineOffset, ok = TryGetLineAndOffset(c.lineStartsCache.get(c.Text), position)
	if ok && (position < 0 || position > len(c.Text)) {
		return 0, 0, false
	}
	return line, lineOffset, ok
}

func (c *KustoCode) GetTokenIndex(position int) int {
	if c == nil || len(c.Tokens) == 0 {
		return 0
	}
	lastTokenIndex := len(c.Tokens) - 1
	lastToken := c.Tokens[lastTokenIndex]
	lastTokenStart := c.TokenStarts[lastTokenIndex]
	if position >= lastTokenStart+lastToken.Length() {
		return lastTokenIndex
	}
	index := searchTokenStart(c.TokenStarts, position)
	if index >= 0 && index < len(c.TokenStarts) && c.TokenStarts[index] == position {
		return index
	}
	return index - 1
}

func (c *KustoCode) GetLexicalTokens() []LexicalToken {
	if c == nil {
		return []LexicalToken{}
	}
	return append([]LexicalToken{}, c.Tokens...)
}

func searchTokenStart(starts []int, position int) int {
	lo, hi := 0, len(starts)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if starts[mid] < position {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

func tokenStarts(tokens []LexicalToken) []int {
	return tokenStartsWithCancellation(tokens, CancellationToken{})
}

func tokenStartsWithCancellation(tokens []LexicalToken, cancellationToken CancellationToken) []int {
	starts := make([]int, 0, len(tokens)+1)
	start := 0
	for i, token := range tokens {
		if i&255 == 0 {
			cancellationToken.ThrowIfCancellationRequested()
		}
		starts = append(starts, start)
		start += token.Length()
	}
	starts = append(starts, start)
	cancellationToken.ThrowIfCancellationRequested()
	return starts
}
