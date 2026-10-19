package kustolang

// Ported from the Query Block and let-statement paths of
// src/Kusto.Language/Parser/QueryParser.cs.

func (p *queryParserState) parseQueryBlockStatement() Statement {
	if p.scanPossibleStatement() {
		switch p.peekToken().Kind {
		case SyntaxKindAliasKeyword:
			return p.parseAliasStatement()
		case SyntaxKindDeclareKeyword:
			if p.peekToken(1).Kind == SyntaxKindPatternKeyword {
				return p.parsePatternStatement()
			}
			return p.parseQueryParametersStatement()
		case SyntaxKindLetKeyword:
			return p.parseLetStatement()
		case SyntaxKindRestrictKeyword:
			return p.parseRestrictStatement()
		case SyntaxKindSetKeyword:
			return p.parseSetOptionStatement()
		}
	}

	if expr := p.parseExpression(); expr != nil {
		return MustNewExpressionStatementTyped(expr)
	}
	return nil
}

func (p *queryParserState) scanPossibleStatement() bool {
	switch p.peekToken().Kind {
	case SyntaxKindAliasKeyword:
		return p.peekToken(1).Kind == SyntaxKindDatabaseKeyword
	case SyntaxKindDeclareKeyword:
		return p.peekToken(1).Kind == SyntaxKindQueryParametersKeyword || p.peekToken(1).Kind == SyntaxKindPatternKeyword
	case SyntaxKindLetKeyword:
		return p.scanName(1) > 0
	case SyntaxKindRestrictKeyword:
		return p.peekToken(1).Kind == SyntaxKindAccessKeyword
	case SyntaxKindSetKeyword:
		return true
	default:
		return false
	}
}

func (p *queryParserState) parseQueryBlockStatementList() *SyntaxList {
	var elements []SyntaxElement
	statement := p.parseQueryBlockStatement()
	for statement != nil {
		semicolon := p.parseToken(SyntaxKindSemicolonToken)
		if semicolon != nil {
			elements = append(elements, NewSeparatedElement(statement, semicolon))
			statement = p.parseQueryBlockStatement()
			continue
		}

		next := p.parseQueryBlockStatement()
		if next != nil {
			elements = append(elements, NewSeparatedElement(statement, createMissingToken(SyntaxKindSemicolonToken)))
		} else {
			elements = append(elements, NewSeparatedElement(statement, nil))
		}
		statement = next
	}
	return NewTypedSyntaxList("SeparatedElement<Statement>", elements...)
}

func (p *queryParserState) parseLetStatement() *LetStatement {
	keyword := p.parseToken(SyntaxKindLetKeyword)
	if keyword == nil {
		return nil
	}
	name := p.parseNameDeclaration()
	if name == nil {
		name = p.createMissingNameDeclaration()
	}
	equal := p.parseRequiredToken(SyntaxKindEqualToken)

	var expr Expression
	if p.peekToken().Kind == SyntaxKindMaterializeKeyword {
		expr = p.parseMaterializeExpression()
	} else if p.peekToken().Kind == SyntaxKindEntityGroupKeyword {
		expr = p.parseEntityGroup()
	} else if p.scanFunctionDeclarationStart() {
		expr = p.parseFunctionDeclaration()
	} else {
		expr = p.parsePipeOnlyRecoveryRootExpression()
	}
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewLetStatementTyped(keyword, name, equal, expr)
}

func (p *queryParserState) parseFunctionParameters() *FunctionParameters {
	open := p.parseToken(SyntaxKindOpenParenToken)
	if open == nil {
		return nil
	}
	parameters := p.parseFunctionParameterCommaList(func() bool { return p.scanCommonListEnd() })
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewFunctionParametersTyped(open, parameters, close)
}

func ParseFunctionParameters(text string) *FunctionParameters {
	return ParseFunctionParametersWithOptions(text, DefaultParseOptions)
}

func ParseFunctionParametersWithOptions(text string, options ParseOptions) *FunctionParameters {
	return newQueryParserStateWithOwnedTokens(LexWithOptions(text, options), options, CancellationToken{}).parseFunctionParameters()
}

func ParseFunctionParametersTokens(tokens []LexicalToken) *FunctionParameters {
	return ParseFunctionParametersTokensFromWithOptions(tokens, 0, DefaultParseOptions)
}

func ParseFunctionParametersTokensWithOptions(tokens []LexicalToken, options ParseOptions) *FunctionParameters {
	return ParseFunctionParametersTokensFromWithOptions(tokens, 0, options)
}

func ParseFunctionParametersTokensFrom(tokens []LexicalToken, start int) *FunctionParameters {
	return ParseFunctionParametersTokensFromWithOptions(tokens, start, DefaultParseOptions)
}

func ParseFunctionParametersTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) *FunctionParameters {
	return newQueryParserStateWithOptions(lexicalTokensFrom(tokens, start), options).parseFunctionParameters()
}

func ParseFunctionParameter(text string) *FunctionParameter {
	return ParseFunctionParameterWithOptions(text, DefaultParseOptions)
}

func ParseFunctionParameterWithOptions(text string, options ParseOptions) *FunctionParameter {
	return newQueryParserStateWithOwnedTokens(LexWithOptions(text, options), options, CancellationToken{}).parseFunctionParameter()
}

func ParseFunctionParameterTokens(tokens []LexicalToken) *FunctionParameter {
	return ParseFunctionParameterTokensFromWithOptions(tokens, 0, DefaultParseOptions)
}

func ParseFunctionParameterTokensWithOptions(tokens []LexicalToken, options ParseOptions) *FunctionParameter {
	return ParseFunctionParameterTokensFromWithOptions(tokens, 0, options)
}

func ParseFunctionParameterTokensFrom(tokens []LexicalToken, start int) *FunctionParameter {
	return ParseFunctionParameterTokensFromWithOptions(tokens, start, DefaultParseOptions)
}

func ParseFunctionParameterTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) *FunctionParameter {
	return newQueryParserStateWithOptions(lexicalTokensFrom(tokens, start), options).parseFunctionParameter()
}

func createMissingFunctionParameters() *FunctionParameters {
	return MustNewFunctionParametersTyped(
		createMissingToken(SyntaxKindOpenParenToken),
		AsSeparatedSyntaxListOf[*FunctionParameter](NewTypedSyntaxList("SeparatedElement<FunctionParameter>")),
		createMissingToken(SyntaxKindCloseParenToken),
	)
}

func (p *queryParserState) parseFunctionBodyStatementElement() *SeparatedElement {
	var statement Statement
	if p.scanPossibleStatement() {
		switch p.peekToken().Kind {
		case SyntaxKindDeclareKeyword:
			statement = p.parseQueryParametersStatement()
		case SyntaxKindLetKeyword:
			statement = p.parseLetStatement()
		}
	}
	if statement == nil {
		return nil
	}
	return NewSeparatedElement(statement, p.parseRequiredToken(SyntaxKindSemicolonToken))
}

func (p *queryParserState) parseFunctionBodyStatementList() *SeparatedSyntaxListOf[Statement] {
	var elements []SyntaxElement
	for !p.scanCommonListEnd() {
		element := p.parseFunctionBodyStatementElement()
		if element == nil {
			break
		}
		elements = append(elements, element)
	}
	return AsSeparatedSyntaxListOf[Statement](NewTypedSyntaxList("SeparatedElement<Statement>", elements...))
}

func (p *queryParserState) parseFunctionBody() *FunctionBody {
	open := p.parseToken(SyntaxKindOpenBraceToken)
	if open == nil {
		return nil
	}
	statements := p.parseFunctionBodyStatementList()
	expr := p.parsePipeOnlyRecoveryRootExpression()
	semicolon := p.parseToken(SyntaxKindSemicolonToken)
	close := p.parseRequiredToken(SyntaxKindCloseBraceToken)
	return MustNewFunctionBodyTyped(open, statements, expr, semicolon, close)
}

func ParseFunctionBody(text string) *FunctionBody {
	return ParseFunctionBodyWithOptions(text, DefaultParseOptions)
}

func ParseFunctionBodyWithOptions(text string, options ParseOptions) *FunctionBody {
	return ParseFunctionBodyTokensWithOptions(LexWithOptions(text, options), options)
}

func ParseFunctionBodyTokens(tokens []LexicalToken) *FunctionBody {
	return ParseFunctionBodyTokensFromWithOptions(tokens, 0, DefaultParseOptions)
}

func ParseFunctionBodyTokensWithOptions(tokens []LexicalToken, options ParseOptions) *FunctionBody {
	return ParseFunctionBodyTokensFromWithOptions(tokens, 0, options)
}

func ParseFunctionBodyTokensFrom(tokens []LexicalToken, start int) *FunctionBody {
	return ParseFunctionBodyTokensFromWithOptions(tokens, start, DefaultParseOptions)
}

func ParseFunctionBodyTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) *FunctionBody {
	return newQueryParserStateWithOptions(lexicalTokensFrom(tokens, start), options).parseFunctionBody()
}

func createMissingFunctionBody() *FunctionBody {
	return MustNewFunctionBodyTyped(
		createMissingToken(SyntaxKindOpenBraceToken),
		AsSeparatedSyntaxListOf[Statement](NewTypedSyntaxList("SeparatedElement<Statement>")),
		nil,
		nil,
		createMissingToken(SyntaxKindCloseBraceToken),
	)
}

func (p *queryParserState) scanFunctionDeclarationStart(offset ...int) bool {
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	if p.peekToken(delta).Kind == SyntaxKindViewKeyword {
		delta++
	}
	if p.peekToken(delta).Kind != SyntaxKindOpenParenToken {
		return false
	}
	next := p.peekToken(delta + 1).Kind
	if next == SyntaxKindCloseParenToken || next == SyntaxKindAsteriskToken {
		return true
	}
	nameLen := p.scanExtendedName(delta + 1)
	return nameLen > 0 && p.peekToken(delta+1+nameLen).Kind == SyntaxKindColonToken
}

func (p *queryParserState) parseFunctionDeclaration() *FunctionDeclaration {
	switch p.peekToken().Kind {
	case SyntaxKindViewKeyword:
		view := p.parseToken()
		parameters := p.parseFunctionParameters()
		if parameters == nil {
			parameters = createMissingFunctionParameters()
		}
		body := p.parseFunctionBody()
		if body == nil {
			body = createMissingFunctionBody()
		}
		return MustNewFunctionDeclarationTyped(view, parameters, body)
	case SyntaxKindOpenParenToken:
		parameters := p.parseFunctionParameters()
		if parameters == nil {
			parameters = createMissingFunctionParameters()
		}
		body := p.parseFunctionBody()
		if body == nil {
			body = createMissingFunctionBody()
		}
		return MustNewFunctionDeclarationTyped(nil, parameters, body)
	default:
		return nil
	}
}

func (p *queryParserState) parsePatternPathValue() *PatternPathValue {
	dot := p.parseToken(SyntaxKindDotToken)
	if dot == nil {
		return nil
	}
	open := p.parseRequiredToken(SyntaxKindOpenBracketToken)
	value := p.parseLiteral()
	if value == nil {
		value = createMissingStringLiteral()
	}
	close := p.parseRequiredToken(SyntaxKindCloseBracketToken)
	return MustNewPatternPathValueTyped(dot, open, value, close)
}

func (p *queryParserState) parsePatternMatchStatementElement() *SeparatedElement {
	statement := p.parseLetStatement()
	if statement == nil {
		return nil
	}
	return NewSeparatedElement(statement, p.parseRequiredToken(SyntaxKindSemicolonToken))
}

func (p *queryParserState) parsePatternMatchStatementList() *SeparatedSyntaxListOf[Statement] {
	var elements []SyntaxElement
	for !p.scanCommonListEnd() {
		element := p.parsePatternMatchStatementElement()
		if element == nil {
			break
		}
		elements = append(elements, element)
	}
	return AsSeparatedSyntaxListOf[Statement](NewTypedSyntaxList("SeparatedElement<Statement>", elements...))
}

func (p *queryParserState) parseRequiredPatternMatchFunctionBody() *FunctionBody {
	open := p.parseRequiredToken(SyntaxKindOpenBraceToken)
	statements := p.parsePatternMatchStatementList()
	expr := p.parseExpression()
	semicolon := p.parseToken(SyntaxKindSemicolonToken)
	close := p.parseRequiredToken(SyntaxKindCloseBraceToken)
	return MustNewFunctionBodyTyped(open, statements, expr, semicolon, close)
}

func (p *queryParserState) parsePatternMatch() *PatternMatch {
	open := p.parseToken(SyntaxKindOpenParenToken)
	if open == nil {
		return nil
	}
	expressions := parseCommaList(p, "Expression", p.parseUnnamedExpression, func() Expression {
		return createMissingStringLiteral()
	}, p.atCommonListEnd, true, false)
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	pathValue := p.parsePatternPathValue()
	equal := p.parseRequiredToken(SyntaxKindEqualToken)
	body := p.parseRequiredPatternMatchFunctionBody()
	semicolon := p.parseRequiredToken(SyntaxKindSemicolonToken)
	return MustNewPatternMatchTyped(MustNewExpressionListTyped(open, expressions, close), pathValue, equal, body, semicolon)
}

func createMissingPatternMatch() *PatternMatch {
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

func (p *queryParserState) parsePatternMatchList(oneOrMore bool) *SyntaxListOf[*PatternMatch] {
	var elements []SyntaxElement
	for !p.scanCommonListEnd() {
		match := p.parsePatternMatch()
		if match == nil {
			if len(elements) == 0 && oneOrMore {
				elements = append(elements, createMissingPatternMatch())
			}
			break
		}
		elements = append(elements, match)
	}
	if len(elements) == 0 && oneOrMore {
		elements = append(elements, createMissingPatternMatch())
	}
	return AsSyntaxListOf[*PatternMatch](NewTypedSyntaxList("PatternMatch", elements...))
}

func (p *queryParserState) parsePatternPathParameter() *PatternPathParameter {
	open := p.parseToken(SyntaxKindOpenBracketToken)
	if open == nil {
		return nil
	}
	parameter := p.parseNameAndTypeDeclaration()
	if parameter == nil {
		parameter = createMissingNameAndTypeDeclaration()
	}
	close := p.parseRequiredToken(SyntaxKindCloseBracketToken)
	return MustNewPatternPathParameterTyped(open, parameter, close)
}

func (p *queryParserState) parsePatternDeclaration() *PatternDeclaration {
	equal := p.parseToken(SyntaxKindEqualToken)
	if equal == nil {
		return nil
	}
	openParen := p.parseRequiredToken(SyntaxKindOpenParenToken)
	parameters := p.parseNameAndTypeDeclarationCommaList(false)
	closeParen := p.parseRequiredToken(SyntaxKindCloseParenToken)
	pathParameter := p.parsePatternPathParameter()
	openBrace := p.parseRequiredToken(SyntaxKindOpenBraceToken)
	patterns := p.parsePatternMatchList(true)
	closeBrace := p.parseRequiredToken(SyntaxKindCloseBraceToken)
	return MustNewPatternDeclarationTyped(equal, openParen, parameters, closeParen, pathParameter, openBrace, patterns, closeBrace)
}

func (p *queryParserState) parsePatternStatement() *PatternStatement {
	if p.peekToken().Kind != SyntaxKindDeclareKeyword || p.peekToken(1).Kind != SyntaxKindPatternKeyword {
		return nil
	}
	declareKeyword := p.parseToken()
	patternKeyword := p.parseToken()
	name := p.parseNameDeclaration()
	if name == nil {
		name = p.createMissingNameDeclaration()
	}
	return MustNewPatternStatementTyped(declareKeyword, patternKeyword, name, p.parsePatternDeclaration())
}

func (p *queryParserState) parseAliasStatement() *AliasStatement {
	keyword := p.parseToken(SyntaxKindAliasKeyword)
	if keyword == nil {
		return nil
	}
	databaseKeyword := p.parseRequiredToken(SyntaxKindDatabaseKeyword)
	name := p.parseNameDeclaration()
	if name == nil {
		name = p.createMissingNameDeclaration()
	}
	equal := p.parseRequiredToken(SyntaxKindEqualToken)
	expr := p.parseUnnamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewAliasStatementTyped(keyword, databaseKeyword, name, equal, expr)
}

func (p *queryParserState) parseOptionValueClause() *OptionValueClause {
	equal := p.parseToken(SyntaxKindEqualToken)
	if equal == nil {
		return nil
	}
	expr := p.parseUnnamedExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewOptionValueClauseTyped(equal, expr)
}

func (p *queryParserState) parseSetOptionStatement() *SetOptionStatement {
	keyword := p.parseToken(SyntaxKindSetKeyword)
	if keyword == nil {
		return nil
	}
	name := p.parseNameDeclaration()
	if name == nil {
		name = p.createMissingNameDeclaration()
	}
	return MustNewSetOptionStatementTyped(keyword, name, p.parseOptionValueClause())
}

func (p *queryParserState) parseQueryParametersStatement() *QueryParametersStatement {
	keyword := p.parseToken(SyntaxKindDeclareKeyword)
	if keyword == nil {
		return nil
	}
	parametersKeyword := p.parseRequiredToken(SyntaxKindQueryParametersKeyword)
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	parameters := p.parseRequiredFunctionParameterCommaList(func() bool { return p.scanCommonListEnd() })
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewQueryParametersStatementTyped(keyword, parametersKeyword, open, parameters, close)
}

func (p *queryParserState) parseRestrictStatement() *RestrictStatement {
	keyword := p.parseToken(SyntaxKindRestrictKeyword)
	if keyword == nil {
		return nil
	}
	accessKeyword := p.parseRequiredToken(SyntaxKindAccessKeyword)
	toKeyword := p.parseRequiredToken(SyntaxKindToKeyword)
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	expressions := p.parseRestrictExpressionCommaList()
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	withClause := p.parseRestrictStatementWithClause()
	return MustNewRestrictStatementTyped(keyword, accessKeyword, toKeyword, open, expressions, close, withClause)
}

func (p *queryParserState) parseRestrictExpressionCommaList() *SeparatedSyntaxListOf[Expression] {
	return p.parseExpressionCommaListWithEnd(true, p.atCommonListEnd, p.parseRestrictExpression)
}

func (p *queryParserState) parseRestrictExpression() Expression {
	if expr := p.parseWildcardedEntityExpression(); expr != nil {
		return expr
	}
	return p.parseNameReference()
}

func (p *queryParserState) parseRestrictStatementWithClause() *RestrictStatementWithClause {
	keyword := p.parseToken(SyntaxKindWithKeyword)
	if keyword == nil {
		return nil
	}
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	properties := p.parseExternalDataPropertyCommaList()
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewRestrictStatementWithClauseTyped(keyword, open, properties, close)
}
