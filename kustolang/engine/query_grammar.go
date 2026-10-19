package kustolang

import (
	"strings"
	"sync"
)

// QueryGrammar exposes the stable public parser objects used by query and
// command grammars. This first native slice covers the upstream name and
// wildcard rules; the remaining query rules continue to use QueryParser.
type QueryGrammar struct {
	Options                         ParseOptions
	IdentifierName                  ResultParser[LexicalToken, Name]
	BracketedName                   ResultParser[LexicalToken, Name]
	BracedName                      ResultParser[LexicalToken, Name]
	SimpleNameDeclaration           ResultParser[LexicalToken, *NameDeclaration]
	SimpleNameDeclarationExpression ResultParser[LexicalToken, Expression]
	BracketedNameDeclaration        ResultParser[LexicalToken, *NameDeclaration]
	SimpleNameReference             ResultParser[LexicalToken, Expression]
	WildcardedIdentifier            ResultParser[LexicalToken, *SyntaxToken]
	WildcardedNameReference         ResultParser[LexicalToken, Expression]
	StringLiteral                   ResultParser[LexicalToken, Expression]
	ParamTypeExtended               ResultParser[LexicalToken, TypeExpression]
	SchemaType                      ResultParser[LexicalToken, *SchemaTypeExpression]
	Literal                         ResultParser[LexicalToken, Expression]
	JsonValue                       ResultParser[LexicalToken, Expression]
	LiteralList                     ResultParser[LexicalToken, *SeparatedSyntaxListOf[Expression]]
	FunctionParameters              ResultParser[LexicalToken, *FunctionParameters]
	FunctionBody                    ResultParser[LexicalToken, *FunctionBody]
	SkippedTokens                   ResultParser[LexicalToken, *SkippedTokens]
	Directive                       ResultParser[LexicalToken, *Directive]
	Statement                       ResultParser[LexicalToken, Statement]
	StatementList                   ResultParser[LexicalToken, *SeparatedSyntaxListOf[Statement]]
	MacroExpandSubQuery             ResultParser[LexicalToken, *SeparatedSyntaxListOf[Statement]]
	QueryBlock                      ResultParser[LexicalToken, *QueryBlock]
	UnnamedExpression               ResultParser[LexicalToken, Expression]
	NamedExpression                 ResultParser[LexicalToken, Expression]
	Expression                      ResultParser[LexicalToken, Expression]
	PipeExpression                  ResultParser[LexicalToken, Expression]
	PipeSubExpression               ResultParser[LexicalToken, Expression]
	FollowingPipeElementExpression  ResultParser[LexicalToken, QueryOperator]
	QueryOperator                   ResultParser[LexicalToken, QueryOperator]
	nameAndTypeDeclaration          ResultParser[LexicalToken, *NameAndTypeDeclaration]
	booleanLiteral                  ResultParser[LexicalToken, Expression]
	booleanLiteralWithCompletion    ResultParser[LexicalToken, Expression]
	longLiteral                     ResultParser[LexicalToken, Expression]
	realLiteral                     ResultParser[LexicalToken, Expression]
	decimalLiteral                  ResultParser[LexicalToken, Expression]
	intLiteral                      ResultParser[LexicalToken, Expression]
	guidLiteral                     ResultParser[LexicalToken, Expression]
	rawGuidLiteral                  ResultParser[LexicalToken, Expression]
	dateTimeLiteral                 ResultParser[LexicalToken, Expression]
	timespanLiteral                 ResultParser[LexicalToken, Expression]
	literalCore                     ResultParser[LexicalToken, Expression]
	tabularPrimaryExpression        ResultParser[LexicalToken, Expression]
	typeOfLiteral                   ResultParser[LexicalToken, Expression]
	summarizeOperator               ResultParser[LexicalToken, *SummarizeOperator]
	dotCompositeFunctionCall        ResultParser[LexicalToken, Expression]
	functionCallOrPath              ResultParser[LexicalToken, Expression]
	entityPathExpression            ResultParser[LexicalToken, Expression]
	requiredFunctionCall            ResultParser[LexicalToken, *FunctionCallExpression]
	unionExpression                 ResultParser[LexicalToken, Expression]
	wildcardedEntityReference       ResultParser[LexicalToken, Expression]
	bracketedEntityNameSelector     ResultParser[LexicalToken, Expression]
	barePathElementSelector         ResultParser[LexicalToken, Expression]
	nameEqualsClause                ResultParser[LexicalToken, *NameEqualsClause]
	forkExpression                  ResultParser[LexicalToken, *ForkExpression]
	forkOperator                    ResultParser[LexicalToken, QueryOperator]
	forkPipeOperator                ResultParser[LexicalToken, QueryOperator]
	forkPipeExpression              ResultParser[LexicalToken, Expression]
	facetOperator                   ResultParser[LexicalToken, QueryOperator]
	macroExpandOperator             ResultParser[LexicalToken, QueryOperator]
	makeGraphOperator               ResultParser[LexicalToken, QueryOperator]
	graphMarkComponentsOperator     ResultParser[LexicalToken, QueryOperator]
	graphWhereNodesOperator         ResultParser[LexicalToken, QueryOperator]
	graphWhereEdgesOperator         ResultParser[LexicalToken, QueryOperator]
	graphToTableAsClause            ResultParser[LexicalToken, *GraphToTableAsClause]
	graphToTableNodesClause         ResultParser[LexicalToken, *GraphToTableOutputClause]
	graphToTableEdgesClause         ResultParser[LexicalToken, *GraphToTableOutputClause]
	graphToTableOutputClause        ResultParser[LexicalToken, *GraphToTableOutputClause]
	graphToTableOperator            ResultParser[LexicalToken, QueryOperator]
	whereClause                     ResultParser[LexicalToken, *WhereClause]
	projectClause                   ResultParser[LexicalToken, *ProjectClause]
	graphMatchPatternEdgeRange      ResultParser[LexicalToken, *GraphMatchPatternEdgeRange]
	graphMatchPatternEdge           ResultParser[LexicalToken, GraphMatchPatternNotation]
	graphMatchPatternNode           ResultParser[LexicalToken, GraphMatchPatternNotation]
	graphMatchPattern               ResultParser[LexicalToken, *GraphMatchPattern]
	graphMatchPatternClause         ResultParser[LexicalToken, *SeparatedSyntaxListOf[*GraphMatchPattern]]
	graphMatchOperator              ResultParser[LexicalToken, QueryOperator]
	graphShortestPathsOperator      ResultParser[LexicalToken, QueryOperator]
	makeSeriesOperator              ResultParser[LexicalToken, QueryOperator]
	contextualSubExpression         ResultParser[LexicalToken, Expression]
	mvApplyOperator                 ResultParser[LexicalToken, QueryOperator]
	parseOperator                   ResultParser[LexicalToken, QueryOperator]
	parseWhereOperator              ResultParser[LexicalToken, QueryOperator]
	parseKvOperator                 ResultParser[LexicalToken, QueryOperator]
	stringOrCompoundLiteral         ResultParser[LexicalToken, Expression]
	dynamicLiteral                  ResultParser[LexicalToken, Expression]
	functionCallExpression          ResultParser[LexicalToken, Expression]
	simplePathExpression            ResultParser[LexicalToken, Expression]
	partitionOperator               ResultParser[LexicalToken, QueryOperator]
	partitionByOperator             ResultParser[LexicalToken, QueryOperator]
	reduceByOperator                ResultParser[LexicalToken, QueryOperator]
	renderOperator                  ResultParser[LexicalToken, QueryOperator]
	scanOperator                    ResultParser[LexicalToken, QueryOperator]
	functionParameter               ResultParser[LexicalToken, *FunctionParameter]
	functionDeclaration             ResultParser[LexicalToken, *FunctionDeclaration]
	letStatement                    ResultParser[LexicalToken, Statement]
	declareQueryParametersStatement ResultParser[LexicalToken, Statement]
	aliasStatement                  ResultParser[LexicalToken, Statement]
	setOptionStatement              ResultParser[LexicalToken, Statement]
	declarePatternStatement         ResultParser[LexicalToken, Statement]
	restrictStatement               ResultParser[LexicalToken, Statement]
	queryStatement                  ResultParser[LexicalToken, Statement]
	materializeExpression           ResultParser[LexicalToken, Expression]
	entityGroup                     ResultParser[LexicalToken, Expression]
	scanRestrictionEntity           Parser[LexicalToken]
	scanSimpleName                  Parser[LexicalToken]
}

type cachedQueryGrammar struct {
	grammar *QueryGrammar
	options ParseOptions
}

var recentQueryGrammar struct {
	sync.Mutex
	value *cachedQueryGrammar
}

func QueryGrammarFrom(globals *GlobalState) *QueryGrammar {
	if globals == nil {
		globals = DefaultGlobalState
	}

	recentQueryGrammar.Lock()
	defer recentQueryGrammar.Unlock()
	if recent := recentQueryGrammar.value; recent != nil && recent.options.EqualExceptForParseKind(globals.ParseOptions) {
		return recent.grammar
	}

	var cached *cachedQueryGrammar
	if globals.Cache != nil {
		cached, _ = KustoCacheTryGetValue[*cachedQueryGrammar](globals.Cache)
	}
	if cached == nil || !cached.options.EqualExceptForParseKind(globals.ParseOptions) {
		cached = &cachedQueryGrammar{grammar: newQueryGrammar(globals.ParseOptions), options: globals.ParseOptions}
		if globals.Cache != nil {
			cached = KustoCacheGetOrCreate(globals.Cache, func() *cachedQueryGrammar { return cached })
		}
	}
	recentQueryGrammar.value = cached
	return cached.grammar
}

func newQueryGrammar(options ParseOptions) *QueryGrammar {
	g := &QueryGrammar{Options: options}

	identifierToken := predefinedIdentifierTokenParser()
	g.IdentifierName = Rule1(identifierToken, func(token *SyntaxToken) Name {
		return MustNewTokenNameTyped(token)
	})

	stringLiteralToken := SyntaxTokenKindParser(SyntaxKindStringLiteralToken)
	stringOrCompound := OneOrMoreProduce(stringLiteralToken, func(tokens []*SyntaxToken) Expression {
		if len(tokens) == 1 {
			return MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, tokens[0])
		}
		elements := make([]SyntaxElement, len(tokens))
		for i, token := range tokens {
			elements[i] = token
		}
		return MustNewCompoundStringLiteralExpressionTyped(AsSyntaxListOf[*SyntaxToken](NewTypedSyntaxList("SyntaxToken", elements...)))
	})
	g.stringOrCompoundLiteral = asResultParser[LexicalToken, Expression](stringOrCompound.WithTag("<string-literal>"))
	g.StringLiteral = asResultParser[LexicalToken, Expression](Rule1(stringLiteralToken, func(token *SyntaxToken) Expression {
		return MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, token)
	}).WithTag("<string-literal>"))

	g.BracketedName = Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenBracketToken),
		stringOrCompound,
		SyntaxTokenKindParser(SyntaxKindCloseBracketToken),
		func(open *SyntaxToken, name Expression, close *SyntaxToken) Name {
			return MustNewBracketedNameTyped(open, name, close)
		},
	)
	g.BracedName = bracedNameParser()

	identifierDeclaration := Rule1(SyntaxTokenKindParser(SyntaxKindIdentifierToken), func(token *SyntaxToken) *NameDeclaration {
		return MustNewNameDeclarationTyped(MustNewTokenNameTyped(token))
	})
	keywordDeclaration := Rule1(predefinedKeywordAsIdentifierTokenParser(), func(token *SyntaxToken) *NameDeclaration {
		return MustNewNameDeclarationTyped(MustNewTokenNameTyped(token))
	})
	g.BracketedNameDeclaration = Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenBracketToken),
		Required(stringOrCompound, func(Source[LexicalToken], int) Expression { return createMissingStringLiteral() }),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBracketToken),
		func(open *SyntaxToken, name Expression, close *SyntaxToken) *NameDeclaration {
			return MustNewNameDeclarationTyped(MustNewBracketedNameTyped(open, name, close))
		},
	)
	g.SimpleNameDeclaration = asResultParser[LexicalToken, *NameDeclaration](FirstValue(
		identifierDeclaration,
		g.BracketedNameDeclaration,
		keywordDeclaration,
	).WithTag("<name>"))
	g.SimpleNameDeclarationExpression = Rule1(g.SimpleNameDeclaration, func(name *NameDeclaration) Expression { return name })

	identifierReference := Rule1(SyntaxTokenKindParser(SyntaxKindIdentifierToken), func(token *SyntaxToken) Expression {
		return MustNewNameReferenceTyped(MustNewTokenNameTyped(token), SymbolMatchNone)
	})
	bracketedReference := Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenBracketToken),
		Required(stringOrCompound, func(Source[LexicalToken], int) Expression { return createMissingStringLiteral() }),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBracketToken),
		func(open *SyntaxToken, name Expression, close *SyntaxToken) Expression {
			return MustNewNameReferenceTyped(MustNewBracketedNameTyped(open, name, close), SymbolMatchNone)
		},
	)
	keywordReference := Rule1(predefinedKeywordAsIdentifierTokenParser(), func(token *SyntaxToken) Expression {
		return MustNewNameReferenceTyped(MustNewTokenNameTyped(token), SymbolMatchNone)
	})
	clientParameter := Rule1(g.BracedName, func(name Name) Expression {
		return MustNewNameReferenceTyped(name, SymbolMatchNone)
	})
	g.SimpleNameReference = asResultParser[LexicalToken, Expression](FirstValue(
		identifierReference, bracketedReference, keywordReference, clientParameter,
	).WithTag("<name>"))
	g.scanSimpleName = Or(
		SyntaxTokenKindParser(SyntaxKindIdentifierToken),
		And(
			SyntaxTokenKindParser(SyntaxKindOpenBracketToken),
			OneOrMore(SyntaxTokenKindParser(SyntaxKindStringLiteralToken)),
			OptionalValue(SyntaxTokenKindParser(SyntaxKindCloseBracketToken)),
		),
		Match(func(token LexicalToken) bool {
			return IsKeywordKind(token.Kind) && SyntaxKindCanBeIdentifier(token.Kind)
		}),
	)

	g.WildcardedIdentifier = wildcardedIdentifierParser(options)
	knownFunctionNames := map[string]struct{}{}
	for _, function := range Functions.All {
		knownFunctionNames[strings.ToLower(function.SymbolName())] = struct{}{}
	}
	for _, aggregate := range Aggregates.All {
		knownFunctionNames[strings.ToLower(aggregate.SymbolName())] = struct{}{}
	}
	scanKnownFunctionName := Match(func(token LexicalToken) bool {
		_, ok := knownFunctionNames[strings.ToLower(token.Text)]
		return ok
	})
	g.scanRestrictionEntity = Or(
		g.WildcardedIdentifier,
		And(
			First(SyntaxTokenKindParser(SyntaxKindIdentifierToken), g.BracketedName, scanKnownFunctionName),
			SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		),
	)
	g.WildcardedNameReference = asResultParser[LexicalToken, Expression](Rule1(g.WildcardedIdentifier, func(token *SyntaxToken) Expression {
		return MustNewNameReferenceTyped(MustNewWildcardedNameTyped(token), SymbolMatchNone)
	}).WithTag("<wildcard>"))
	g.initializeTypeAndLiteralParsers(stringOrCompound)
	g.initializeDeclarationParsers()
	g.initializeScalarExpressionParsers()
	g.initializePipeExpressionParsers()
	g.initializeQueryBlockParsers()
	return g
}

func predefinedKeywordAsIdentifierTokenParser() ResultParser[LexicalToken, *SyntaxToken] {
	return asResultParser[LexicalToken, *SyntaxToken](MatchValue(func(token LexicalToken) bool {
		return IsKeywordKind(token.Kind) && SyntaxKindCanBeIdentifier(token.Kind)
	}, func(token LexicalToken) *SyntaxToken { return SyntaxTokenFromLexical(token) }).WithTag("<keywordAsIdentifier>"))
}

func bracedNameParser() ResultParser[LexicalToken, Name] {
	return asResultParser[LexicalToken, Name](MatchConsumerValue(func(source Source[LexicalToken], start int) int {
		pos := start
		if source.IsEnd(pos) || source.Peek(pos).Kind != SyntaxKindOpenBraceToken {
			return -1
		}
		pos++
		if source.IsEnd(pos) {
			return -1
		}
		part := source.Peek(pos)
		if (part.Kind != SyntaxKindIdentifierToken && !IsKeywordKind(part.Kind)) || part.Trivia != "" {
			return -1
		}
		pos++
		if !source.IsEnd(pos) && source.Peek(pos).Kind == SyntaxKindOpenBracketToken && source.Peek(pos).Trivia == "" {
			pos++
			if !source.IsEnd(pos) && source.Peek(pos).Kind == SyntaxKindMinusToken && source.Peek(pos).Trivia == "" {
				pos++
			}
			if source.IsEnd(pos) || source.Peek(pos).Kind != SyntaxKindLongLiteralToken || source.Peek(pos).Trivia != "" {
				return -1
			}
			pos++
			if source.IsEnd(pos) || source.Peek(pos).Kind != SyntaxKindCloseBracketToken || source.Peek(pos).Trivia != "" {
				return -1
			}
			pos++
		}
		if source.IsEnd(pos) || source.Peek(pos).Kind != SyntaxKindCloseBraceToken || source.Peek(pos).Trivia != "" {
			return -1
		}
		return pos + 1 - start
	}, func(source Source[LexicalToken], start, length int) Name {
		text := ""
		for i := 1; i < length-1; i++ {
			text += source.Peek(start + i).Text
		}
		return MustNewBracedNameTyped(
			SyntaxTokenFromLexical(source.Peek(start)),
			SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", text)),
			SyntaxTokenFromLexical(source.Peek(start+length-1)),
		)
	}).WithTag("<client-parameter>"))
}

func wildcardedIdentifierParser(options ParseOptions) ResultParser[LexicalToken, *SyntaxToken] {
	part := func(token LexicalToken) bool {
		return token.Kind == SyntaxKindIdentifierToken || isExtendedKeywordAsIdentifier(token.Kind)
	}
	adjacent := func(token LexicalToken) bool { return token.Trivia == "" || options.AllowNonAdjacentWildcardParts }
	start := First(
		And(Match[LexicalToken](part), Match(func(token LexicalToken) bool { return token.Kind == SyntaxKindAsteriskToken && adjacent(token) })),
		SyntaxTokenKindParser(SyntaxKindAsteriskToken),
	)
	tail := ZeroOrMore(Match(func(token LexicalToken) bool {
		return adjacent(token) && (part(token) || token.Kind == SyntaxKindLongLiteralToken || token.Kind == SyntaxKindAsteriskToken)
	}))
	scan := And(start, tail).Hide()
	return asResultParser[LexicalToken, *SyntaxToken](Convert(scan, func(source Source[LexicalToken], start, length int) *SyntaxToken {
		raw, value := "", ""
		for i := 0; i < length; i++ {
			token := source.Peek(start + i)
			if i > 0 {
				raw += token.Trivia
			}
			raw += token.Text
			value += token.Text
		}
		return &SyntaxToken{kind: SyntaxKindIdentifierToken, trivia: source.Peek(start).Trivia, text: raw, valueText: value}
	}).WithTag("<wildcard>"))
}
