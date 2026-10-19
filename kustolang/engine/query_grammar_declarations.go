package kustolang

// Ported from the function-parameter declaration region in
// src/Kusto.Language/Parser/QueryGrammar.cs.
func (g *QueryGrammar) initializeDeclarationParsers() {
	missingExpression := func(source Source[LexicalToken], start int) Expression {
		return queryGrammarMissingExpression(source, start)
	}
	missingNameDeclaration := func(Source[LexicalToken], int) *NameDeclaration {
		return createMissingNameDeclaration()
	}

	var letStatementCore ResultParser[LexicalToken, Statement]
	var declareQueryParametersStatementCore ResultParser[LexicalToken, Statement]
	g.letStatement = asResultParser[LexicalToken, Statement](Forward(func() ResultParser[LexicalToken, Statement] {
		return letStatementCore
	}).WithTag("<let>"))
	g.declareQueryParametersStatement = asResultParser[LexicalToken, Statement](Forward(func() ResultParser[LexicalToken, Statement] {
		return declareQueryParametersStatementCore
	}).WithTag("<query-parameters>"))

	queryPrefixToken := func(kind SyntaxKind) ResultParser[LexicalToken, *SyntaxToken] {
		return SyntaxTokenKindParser(kind, NewCompletionItem(CompletionKindQueryPrefix, SyntaxText(kind)))
	}

	g.aliasStatement = asResultParser[LexicalToken, Statement](Rule5(
		asResultParser[LexicalToken, *SyntaxToken](queryPrefixToken(SyntaxKindAliasKeyword).Hide()),
		RequiredSyntaxTokenKindParser(SyntaxKindDatabaseKeyword),
		Required(g.SimpleNameDeclaration, missingNameDeclaration),
		RequiredSyntaxTokenKindParser(SyntaxKindEqualToken),
		Required(Forward(func() ResultParser[LexicalToken, Expression] { return g.UnnamedExpression }), missingExpression),
		func(alias, database *SyntaxToken, name *NameDeclaration, equal *SyntaxToken, expression Expression) Statement {
			return MustNewAliasStatementTyped(alias, database, name, equal, expression)
		},
	).WithTag("<alias>"))

	g.materializeExpression = Rule4(
		SyntaxTokenKindParser(SyntaxKindMaterializeKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		Required(Forward(func() ResultParser[LexicalToken, Expression] { return g.PipeExpression }), missingExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(keyword, open *SyntaxToken, expression Expression, close *SyntaxToken) Expression {
			return MustNewMaterializeExpressionTyped(keyword, open, expression, close)
		},
	)

	nameTokenLiteral := Rule1(
		MatchValue(func(token LexicalToken) bool {
			return token.Kind == SyntaxKindIdentifierToken || (IsKeywordKind(token.Kind) && SyntaxKindCanBeIdentifier(token.Kind))
		}, func(token LexicalToken) *SyntaxToken {
			return SyntaxTokenFromLexical(token)
		}),
		func(token *SyntaxToken) Expression {
			return MustNewLiteralExpressionTyped(SyntaxKindTokenLiteralExpression, token)
		},
	)
	defaultValueDeclaration := Rule2(
		SyntaxTokenKindParser(SyntaxKindEqualToken),
		Required(FirstValue(g.Literal, nameTokenLiteral), missingExpression),
		func(equal *SyntaxToken, value Expression) *DefaultValueDeclaration {
			return MustNewDefaultValueDeclarationTyped(equal, value)
		},
	)
	functionParameter := asResultParser[LexicalToken, *FunctionParameter](Rule2(
		g.nameAndTypeDeclaration,
		OptionalValue(defaultValueDeclaration),
		func(nameAndType *NameAndTypeDeclaration, defaultValue *DefaultValueDeclaration) *FunctionParameter {
			return MustNewFunctionParameterTyped(nameAndType, defaultValue)
		},
	).WithTag("<function-parameter>"))
	g.functionParameter = functionParameter
	g.FunctionParameters = Rule3(
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		SyntaxCommaListParser(functionParameter, func(Source[LexicalToken], int) *FunctionParameter {
			return createMissingFunctionParameter()
		}, false, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open *SyntaxToken, parameters *SeparatedSyntaxListOf[*FunctionParameter], close *SyntaxToken) *FunctionParameters {
			return MustNewFunctionParametersTyped(open, parameters, close)
		},
	)
	functionBodyLeadingStatement := IfValue(
		First(
			SyntaxTokenKindParser(SyntaxKindLetKeyword),
			And(
				SyntaxTokenKindParser(SyntaxKindDeclareKeyword),
				SyntaxTokenKindParser(SyntaxKindQueryParametersKeyword),
			),
		),
		Rule2(FirstValue(
			g.letStatement,
			g.declareQueryParametersStatement,
		),
			RequiredSyntaxTokenKindParser(SyntaxKindSemicolonToken),
			func(statement Statement, semicolon *SyntaxToken) *SeparatedElement {
				return NewSeparatedElement(statement, semicolon)
			},
		),
	)
	missingFunctionBodyStatement := func(Source[LexicalToken], int) *SeparatedElement {
		return queryGrammarMissingStatementElement()
	}
	functionBodyStatementList := asResultParser[LexicalToken, *SyntaxListOf[*SeparatedElement]](
		SyntaxListParser(functionBodyLeadingStatement, missingFunctionBodyStatement, false).WithTag("<statement-list>"),
	)
	g.FunctionBody = Rule5(
		RequiredSyntaxTokenKindParser(SyntaxKindOpenBraceToken),
		functionBodyStatementList,
		OptionalValue(Forward(func() ResultParser[LexicalToken, Expression] { return g.Expression })),
		OptionalValue(SyntaxTokenKindParser(SyntaxKindSemicolonToken)),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBraceToken),
		func(open *SyntaxToken, statementElements *SyntaxListOf[*SeparatedElement], expression Expression, semicolon, close *SyntaxToken) *FunctionBody {
			statements := AsSeparatedSyntaxListOf[Statement](statementElements.SyntaxList)
			return MustNewFunctionBodyTyped(open, statements, expression, semicolon, close)
		},
	)
	g.functionDeclaration = asResultParser[LexicalToken, *FunctionDeclaration](Rule3(
		OptionalValue(SyntaxTokenKindParser(SyntaxKindViewKeyword)),
		g.FunctionParameters,
		g.FunctionBody,
		func(view *SyntaxToken, parameters *FunctionParameters, body *FunctionBody) *FunctionDeclaration {
			return MustNewFunctionDeclarationTyped(view, parameters, body)
		},
	).WithTag("<function-declaration>"))

	letStart := queryPrefixToken(SyntaxKindLetKeyword)
	functionDeclarationGuard := And(
		letStart,
		g.scanSimpleName,
		SyntaxTokenKindParser(SyntaxKindEqualToken),
		OptionalValue(SyntaxTokenKindParser(SyntaxKindViewKeyword)),
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		Or(
			SyntaxTokenKindParser(SyntaxKindCloseParenToken),
			SyntaxTokenKindParser(SyntaxKindAsteriskToken),
			And(g.scanSimpleName, SyntaxTokenKindParser(SyntaxKindColonToken)),
		),
	)
	functionLet := IfValue(functionDeclarationGuard, Rule4(
		SyntaxTokenKindParser(SyntaxKindLetKeyword),
		g.SimpleNameDeclaration,
		SyntaxTokenKindParser(SyntaxKindEqualToken),
		g.functionDeclaration,
		func(keyword *SyntaxToken, name *NameDeclaration, equal *SyntaxToken, declaration *FunctionDeclaration) Statement {
			return MustNewLetStatementTyped(keyword, name, equal, declaration)
		},
	))
	materializeLet := IfValue(
		And(letStart, g.scanSimpleName, SyntaxTokenKindParser(SyntaxKindEqualToken), SyntaxTokenKindParser(SyntaxKindMaterializeKeyword)),
		Rule4(
			SyntaxTokenKindParser(SyntaxKindLetKeyword),
			Required(g.SimpleNameDeclaration, missingNameDeclaration),
			SyntaxTokenKindParser(SyntaxKindEqualToken),
			g.materializeExpression,
			func(keyword *SyntaxToken, name *NameDeclaration, equal *SyntaxToken, expression Expression) Statement {
				return MustNewLetStatementTyped(keyword, name, equal, expression)
			},
		),
	)
	entityGroupLet := IfValue(
		And(letStart, g.scanSimpleName, SyntaxTokenKindParser(SyntaxKindEqualToken), SyntaxTokenKindParser(SyntaxKindEntityGroupKeyword)),
		Rule4(
			SyntaxTokenKindParser(SyntaxKindLetKeyword),
			Required(g.SimpleNameDeclaration, missingNameDeclaration),
			SyntaxTokenKindParser(SyntaxKindEqualToken),
			Forward(func() ResultParser[LexicalToken, Expression] { return g.entityGroup }),
			func(keyword *SyntaxToken, name *NameDeclaration, equal *SyntaxToken, expression Expression) Statement {
				return MustNewLetStatementTyped(keyword, name, equal, expression)
			},
		),
	)
	regularLet := Rule4(
		queryPrefixToken(SyntaxKindLetKeyword),
		Required(g.SimpleNameDeclaration, missingNameDeclaration),
		RequiredSyntaxTokenKindParser(SyntaxKindEqualToken),
		Required(Forward(func() ResultParser[LexicalToken, Expression] { return g.Expression }), missingExpression),
		func(keyword *SyntaxToken, name *NameDeclaration, equal *SyntaxToken, expression Expression) Statement {
			return MustNewLetStatementTyped(keyword, name, equal, expression)
		},
	)
	letStatementCore = FirstValue(functionLet, materializeLet, entityGroupLet, regularLet)

	optionValueClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindEqualToken),
		Required(Forward(func() ResultParser[LexicalToken, Expression] { return g.UnnamedExpression }), missingExpression),
		func(equal *SyntaxToken, expression Expression) *OptionValueClause {
			return MustNewOptionValueClauseTyped(equal, expression)
		},
	)
	g.setOptionStatement = asResultParser[LexicalToken, Statement](Rule3(
		queryPrefixToken(SyntaxKindSetKeyword),
		Required(g.SimpleNameDeclaration, missingNameDeclaration),
		OptionalValue(optionValueClause),
		func(keyword *SyntaxToken, name *NameDeclaration, value *OptionValueClause) Statement {
			return MustNewSetOptionStatementTyped(keyword, name, value)
		},
	).WithTag("<set-option>"))

	declareQueryParametersStatementCore = Rule5(
		asResultParser[LexicalToken, *SyntaxToken](queryPrefixToken(SyntaxKindDeclareKeyword).Hide()),
		RequiredSyntaxTokenKindParser(SyntaxKindQueryParametersKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		SyntaxCommaListParser(g.functionParameter, func(Source[LexicalToken], int) *FunctionParameter {
			return createMissingFunctionParameter()
		}, true, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(declare, queryParameters, open *SyntaxToken, parameters *SeparatedSyntaxListOf[*FunctionParameter], close *SyntaxToken) Statement {
			return MustNewQueryParametersStatementTyped(declare, queryParameters, open, parameters, close)
		},
	)

	restriction := WithSyntaxCompletionHint(
		FirstValue(
			Forward(func() ResultParser[LexicalToken, Expression] {
				return IfValue(g.scanRestrictionEntity, g.wildcardedEntityReference)
			}),
			g.SimpleNameReference,
		),
		CompletionHintTable|CompletionHintMaterializedView|CompletionHintExternalTable|CompletionHintGraphModel,
	)
	restrictWithClause := Rule4(
		SyntaxTokenKindParser(SyntaxKindWithKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		queryGrammarParameterCommaList(g, QueryOperatorParameters.RestrictStatementParameters, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(with, open *SyntaxToken, properties *SeparatedSyntaxListOf[*NamedParameter], close *SyntaxToken) *RestrictStatementWithClause {
			return MustNewRestrictStatementWithClauseTyped(with, open, properties, close)
		},
	)
	g.restrictStatement = asResultParser[LexicalToken, Statement](Rule7(
		asResultParser[LexicalToken, *SyntaxToken](queryPrefixToken(SyntaxKindRestrictKeyword).Hide()),
		RequiredSyntaxTokenKindParser(SyntaxKindAccessKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindToKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		SyntaxCommaListParser(restriction, missingExpression, true, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		OptionalValue(restrictWithClause),
		func(restrict, access, to, open *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], close *SyntaxToken, with *RestrictStatementWithClause) Statement {
			return MustNewRestrictStatementTyped(restrict, access, to, open, expressions, close, with)
		},
	).WithTag("<restrict>"))

	patternPathValue := Rule4(
		SyntaxTokenKindParser(SyntaxKindDotToken),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenBracketToken),
		Required(g.StringLiteral, func(Source[LexicalToken], int) Expression { return createMissingStringLiteral() }),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBracketToken),
		func(dot, open *SyntaxToken, value Expression, close *SyntaxToken) *PatternPathValue {
			return MustNewPatternPathValueTyped(dot, open, value, close)
		},
	)
	patternMatchStatementElement := Rule2(
		g.letStatement,
		RequiredSyntaxTokenKindParser(SyntaxKindSemicolonToken),
		func(statement Statement, semicolon *SyntaxToken) *SeparatedElement {
			return NewSeparatedElement(statement, semicolon)
		},
	)
	patternMatchBody := Rule5(
		RequiredSyntaxTokenKindParser(SyntaxKindOpenBraceToken),
		SyntaxListParser(patternMatchStatementElement, func(Source[LexicalToken], int) *SeparatedElement {
			return queryGrammarMissingStatementElement()
		}, false),
		OptionalValue(Forward(func() ResultParser[LexicalToken, Expression] { return g.Expression })),
		OptionalValue(SyntaxTokenKindParser(SyntaxKindSemicolonToken)),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBraceToken),
		func(open *SyntaxToken, statementElements *SyntaxListOf[*SeparatedElement], expression Expression, semicolon, close *SyntaxToken) *FunctionBody {
			return MustNewFunctionBodyTyped(open, AsSeparatedSyntaxListOf[Statement](statementElements.SyntaxList), expression, semicolon, close)
		},
	)
	patternMatchValue := FirstValue(g.StringLiteral, asResultParser[LexicalToken, Expression](Forward(func() ResultParser[LexicalToken, Expression] {
		return g.UnnamedExpression
	}).Hide()))
	patternMatch := Rule7(
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		SyntaxCommaListParser(patternMatchValue, func(Source[LexicalToken], int) Expression { return createMissingStringLiteral() }, true, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		OptionalValue(patternPathValue),
		RequiredSyntaxTokenKindParser(SyntaxKindEqualToken),
		patternMatchBody,
		RequiredSyntaxTokenKindParser(SyntaxKindSemicolonToken),
		func(open *SyntaxToken, values *SeparatedSyntaxListOf[Expression], close *SyntaxToken, path *PatternPathValue, equal *SyntaxToken, body *FunctionBody, semicolon *SyntaxToken) *PatternMatch {
			return MustNewPatternMatchTyped(MustNewExpressionListTyped(open, values, close), path, equal, body, semicolon)
		},
	)
	patternPathParameter := Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenBracketToken),
		Required(g.nameAndTypeDeclaration, func(Source[LexicalToken], int) *NameAndTypeDeclaration {
			return queryGrammarMissingNameAndTypeDeclaration()
		}),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBracketToken),
		func(open *SyntaxToken, parameter *NameAndTypeDeclaration, close *SyntaxToken) *PatternPathParameter {
			return MustNewPatternPathParameterTyped(open, parameter, close)
		},
	)
	patternDeclaration := Rule8(
		SyntaxTokenKindParser(SyntaxKindEqualToken),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		SyntaxCommaListParser(g.nameAndTypeDeclaration, func(Source[LexicalToken], int) *NameAndTypeDeclaration {
			return queryGrammarMissingNameAndTypeDeclaration()
		}, false, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		OptionalValue(patternPathParameter),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenBraceToken),
		SyntaxListParser(patternMatch, func(Source[LexicalToken], int) *PatternMatch { return queryGrammarMissingPatternMatch() }, true),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBraceToken),
		func(equal, openParen *SyntaxToken, parameters *SeparatedSyntaxListOf[*NameAndTypeDeclaration], closeParen *SyntaxToken, path *PatternPathParameter, openBrace *SyntaxToken, patterns *SyntaxListOf[*PatternMatch], closeBrace *SyntaxToken) *PatternDeclaration {
			return MustNewPatternDeclarationTyped(equal, openParen, parameters, closeParen, path, openBrace, patterns, closeBrace)
		},
	)
	g.declarePatternStatement = asResultParser[LexicalToken, Statement](Rule4(
		asResultParser[LexicalToken, *SyntaxToken](queryPrefixToken(SyntaxKindDeclareKeyword).Hide()),
		SyntaxTokenKindParser(SyntaxKindPatternKeyword),
		Required(g.SimpleNameDeclaration, missingNameDeclaration),
		OptionalValue(patternDeclaration),
		func(declare, pattern *SyntaxToken, name *NameDeclaration, declaration *PatternDeclaration) Statement {
			return MustNewPatternStatementTyped(declare, pattern, name, declaration)
		},
	).WithTag("<pattern-statement>"))

	g.queryStatement = Rule1(
		Forward(func() ResultParser[LexicalToken, Expression] { return g.Expression }),
		func(expression Expression) Statement { return MustNewExpressionStatementTyped(expression) },
	)
}

func queryGrammarMissingStatementElement() *SeparatedElement {
	return NewSeparatedElement(queryGrammarMissingStatement(), nil)
}

func queryGrammarMissingPatternMatch() *PatternMatch {
	return MustNewPatternMatchTyped(
		MustNewExpressionListTyped(
			createOmittedToken(SyntaxKindOpenParenToken),
			AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>")),
			createOmittedToken(SyntaxKindCloseParenToken),
		),
		nil,
		createOmittedToken(SyntaxKindEqualToken),
		MustNewFunctionBodyTyped(
			createOmittedToken(SyntaxKindOpenBraceToken),
			AsSeparatedSyntaxListOf[Statement](NewTypedSyntaxList("SeparatedElement<Statement>")),
			nil,
			nil,
			createOmittedToken(SyntaxKindCloseBraceToken),
		),
		createOmittedToken(SyntaxKindSemicolonToken),
		DiagnosticMissingPatternMatch(),
	)
}
