package kustolang

import "strings"

// Ported from the expression/literal/function-call portions of
// src/Kusto.Language/Parser/QueryParser.cs.
// This is still an incremental subset: many query operators and several
// specialized expression forms remain tracked as pending parser work.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func ParseExpression(text string) Expression {
	return ParseExpressionWithOptions(text, DefaultParseOptions)
}

func ParseExpressionWithOptions(text string, options ParseOptions) Expression {
	return newQueryParserStateWithOwnedTokens(LexWithOptions(text, options), options, CancellationToken{}).parseExpression()
}

func ParseExpressionTokens(tokens []LexicalToken) Expression {
	return ParseExpressionTokensFromWithOptions(tokens, 0, DefaultParseOptions)
}

func ParseExpressionTokensWithOptions(tokens []LexicalToken, options ParseOptions) Expression {
	return ParseExpressionTokensFromWithOptions(tokens, 0, options)
}

func ParseExpressionTokensFrom(tokens []LexicalToken, start int) Expression {
	return ParseExpressionTokensFromWithOptions(tokens, start, DefaultParseOptions)
}

func ParseExpressionTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) Expression {
	return newQueryParserStateWithOptions(lexicalTokensFrom(tokens, start), options).parseExpression()
}

func ParseLiteral(text string) Expression {
	return ParseLiteralWithOptions(text, DefaultParseOptions)
}

func ParseLiteralWithOptions(text string, options ParseOptions) Expression {
	return newQueryParserStateWithOwnedTokens(LexWithOptions(text, options), options, CancellationToken{}).parseLiteral()
}

func ParseLiteralTokens(tokens []LexicalToken) Expression {
	return ParseLiteralTokensFromWithOptions(tokens, 0, DefaultParseOptions)
}

func ParseLiteralTokensWithOptions(tokens []LexicalToken, options ParseOptions) Expression {
	return ParseLiteralTokensFromWithOptions(tokens, 0, options)
}

func ParseLiteralTokensFrom(tokens []LexicalToken, start int) Expression {
	return ParseLiteralTokensFromWithOptions(tokens, start, DefaultParseOptions)
}

func ParseLiteralTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) Expression {
	return newQueryParserStateWithOptions(lexicalTokensFrom(tokens, start), options).parseLiteral()
}

func ParseEntityGroup(text string) Expression {
	return ParseEntityGroupWithOptions(text, DefaultParseOptions)
}

func ParseEntityGroupWithOptions(text string, options ParseOptions) Expression {
	return newQueryParserStateWithOwnedTokens(LexWithOptions(text, options), options, CancellationToken{}).parseEntityGroup()
}

func ParseEntityGroupTokens(tokens []LexicalToken) Expression {
	return ParseEntityGroupTokensFromWithOptions(tokens, 0, DefaultParseOptions)
}

func ParseEntityGroupTokensWithOptions(tokens []LexicalToken, options ParseOptions) Expression {
	return ParseEntityGroupTokensFromWithOptions(tokens, 0, options)
}

func ParseEntityGroupTokensFrom(tokens []LexicalToken, start int) Expression {
	return ParseEntityGroupTokensFromWithOptions(tokens, start, DefaultParseOptions)
}

func ParseEntityGroupTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) Expression {
	return newQueryParserStateWithOptions(lexicalTokensFrom(tokens, start), options).parseEntityGroup()
}

func ParseEntityPath(text string) Expression {
	return ParseEntityPathWithOptions(text, DefaultParseOptions)
}

func ParseEntityPathWithOptions(text string, options ParseOptions) Expression {
	return newQueryParserStateWithOwnedTokens(LexWithOptions(text, options), options, CancellationToken{}).parseEntityPathExpression()
}

func ParseEntityPathTokens(tokens []LexicalToken) Expression {
	return ParseEntityPathTokensFromWithOptions(tokens, 0, DefaultParseOptions)
}

func ParseEntityPathTokensWithOptions(tokens []LexicalToken, options ParseOptions) Expression {
	return ParseEntityPathTokensFromWithOptions(tokens, 0, options)
}

func ParseEntityPathTokensFrom(tokens []LexicalToken, start int) Expression {
	return ParseEntityPathTokensFromWithOptions(tokens, start, DefaultParseOptions)
}

func ParseEntityPathTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) Expression {
	return newQueryParserStateWithOptions(lexicalTokensFrom(tokens, start), options).parseEntityPathExpression()
}

func (p *queryParserState) parseExpression() Expression {
	return p.parsePipeExpression()
}

func (p *queryParserState) parseUnnamedExpression() Expression {
	return p.parseLogicalOrExpression()
}

func (p *queryParserState) parseLiteral() Expression {
	switch p.peekToken().Kind {
	case SyntaxKindStringLiteralToken:
		return p.parseStringOrCompoundStringLiteral()
	case SyntaxKindBooleanLiteralToken:
		return MustNewLiteralExpressionTyped(SyntaxKindBooleanLiteralExpression, p.parseToken())
	case SyntaxKindLongLiteralToken:
		return MustNewLiteralExpressionTyped(SyntaxKindLongLiteralExpression, p.parseToken())
	case SyntaxKindRealLiteralToken:
		return MustNewLiteralExpressionTyped(SyntaxKindRealLiteralExpression, p.parseToken())
	case SyntaxKindDecimalLiteralToken:
		return MustNewLiteralExpressionTyped(SyntaxKindDecimalLiteralExpression, p.parseToken())
	case SyntaxKindIntLiteralToken:
		return MustNewLiteralExpressionTyped(SyntaxKindIntLiteralExpression, p.parseToken())
	case SyntaxKindGuidLiteralToken:
		return MustNewLiteralExpressionTyped(SyntaxKindGuidLiteralExpression, p.parseToken())
	case SyntaxKindRawGuidLiteralToken:
		return MustNewLiteralExpressionTyped(SyntaxKindGuidLiteralExpression, p.parseToken(), DiagnosticRawGuidLiteralNotAllowed())
	case SyntaxKindDateTimeLiteralToken:
		return MustNewLiteralExpressionTyped(SyntaxKindDateTimeLiteralExpression, p.parseToken())
	case SyntaxKindTimespanLiteralToken:
		return MustNewLiteralExpressionTyped(SyntaxKindTimespanLiteralExpression, p.parseToken())
	case SyntaxKindDynamicKeyword:
		return p.parseDynamicLiteral()
	case SyntaxKindPlusToken, SyntaxKindMinusToken:
		return p.parseSignedNumericLiteral()
	case SyntaxKindTypeOfKeyword:
		if literal := p.parseTypeOfLiteral(); literal != nil {
			return literal
		}
		return nil
	case SyntaxKindOpenBraceToken:
		reference := p.parseBracedNameReference()
		if reference == nil {
			return nil
		}
		return reference
	default:
		return nil
	}
}

func (p *queryParserState) parseStringOrCompoundStringLiteral() Expression {
	if p.peekToken().Kind != SyntaxKindStringLiteralToken {
		return nil
	}
	if p.peekToken(1).Kind != SyntaxKindStringLiteralToken {
		return MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, p.parseToken())
	}
	var tokens []SyntaxElement
	for p.peekToken().Kind == SyntaxKindStringLiteralToken {
		tokens = append(tokens, p.parseToken())
	}
	return MustNewCompoundStringLiteralExpressionTyped(AsSyntaxListOf[*SyntaxToken](NewTypedSyntaxList("SyntaxToken", tokens...)))
}

func (p *queryParserState) scanStringOrCompoundStringLiteral(offset ...int) int {
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	start := delta
	for p.peekToken(delta).Kind == SyntaxKindStringLiteralToken {
		delta++
	}
	if delta == start {
		return -1
	}
	return delta - start
}

func (p *queryParserState) parseSignedNumericLiteral() Expression {
	sign := p.peekToken()
	number := p.peekToken(1)
	if (sign.Kind != SyntaxKindPlusToken && sign.Kind != SyntaxKindMinusToken) || number.Trivia != "" {
		return nil
	}
	if number.Text == "" || number.Text[0] < '0' || number.Text[0] > '9' {
		return nil
	}
	switch number.Kind {
	case SyntaxKindLongLiteralToken, SyntaxKindRealLiteralToken, SyntaxKindTimespanLiteralToken:
	default:
		return nil
	}
	signToken := p.parseToken()
	numberToken := p.parseToken()
	combined := NewLexicalToken(number.Kind, signToken.Trivia(), signToken.RawText()+numberToken.RawText())
	token := SyntaxTokenFromLexical(combined)
	switch number.Kind {
	case SyntaxKindLongLiteralToken:
		return MustNewLiteralExpressionTyped(SyntaxKindLongLiteralExpression, token)
	case SyntaxKindRealLiteralToken:
		return MustNewLiteralExpressionTyped(SyntaxKindRealLiteralExpression, token)
	case SyntaxKindTimespanLiteralToken:
		return MustNewLiteralExpressionTyped(SyntaxKindTimespanLiteralExpression, token)
	default:
		return nil
	}
}

func (p *queryParserState) parseForcedRealLiteral() Expression {
	token := p.peekToken()
	switch token.Kind {
	case SyntaxKindLongLiteralToken, SyntaxKindRealLiteralToken:
		if token.Text != "" && token.Text[0] >= '0' && token.Text[0] <= '9' {
			return MustNewLiteralExpressionTyped(SyntaxKindRealLiteralExpression, p.parseToken())
		}
	case SyntaxKindPlusToken, SyntaxKindMinusToken:
		return p.parseSignedForcedRealLiteral()
	case SyntaxKindOpenBraceToken:
		return p.parseBracedNameReference()
	}
	return nil
}

func (p *queryParserState) parseSignedForcedRealLiteral() Expression {
	sign := p.peekToken()
	number := p.peekToken(1)
	if (sign.Kind != SyntaxKindPlusToken && sign.Kind != SyntaxKindMinusToken) || number.Trivia != "" {
		return nil
	}
	switch number.Kind {
	case SyntaxKindLongLiteralToken, SyntaxKindRealLiteralToken:
	default:
		return nil
	}
	signToken := p.parseToken()
	numberToken := p.parseToken()
	combined := NewLexicalToken(number.Kind, signToken.Trivia(), signToken.RawText()+numberToken.RawText())
	return MustNewLiteralExpressionTyped(SyntaxKindRealLiteralExpression, SyntaxTokenFromLexical(combined))
}

func (p *queryParserState) parsePrimaryExpression() Expression {
	if p.peekToken().Kind == SyntaxKindTypeOfKeyword {
		if literal := p.parseTypeOfLiteral(); literal != nil {
			return literal
		}
	}
	switch p.peekToken().Kind {
	case SyntaxKindStringLiteralToken,
		SyntaxKindBooleanLiteralToken,
		SyntaxKindLongLiteralToken,
		SyntaxKindRealLiteralToken,
		SyntaxKindDecimalLiteralToken,
		SyntaxKindIntLiteralToken,
		SyntaxKindGuidLiteralToken,
		SyntaxKindRawGuidLiteralToken,
		SyntaxKindDateTimeLiteralToken,
		SyntaxKindTimespanLiteralToken,
		SyntaxKindDynamicKeyword:
		return p.parseLiteral()
	case SyntaxKindDataTableKeyword:
		return p.parseDataTableExpression()
	case SyntaxKindContextualDataTableKeyword:
		return p.parseContextualDataTableExpression()
	case SyntaxKindExternalDataKeyword, SyntaxKindExternal_DataKeyword:
		return p.parseExternalDataExpression()
	case SyntaxKindInlineExternalTableKeyword:
		if expr := p.parseInlineExternalTableExpression(); expr != nil {
			return expr
		}
		if p.scanFunctionCallStart() {
			return p.parseFunctionCallExpression()
		}
		return p.parsePrimaryPathSelector()
	case SyntaxKindEntityGroupKeyword:
		if p.peekToken(1).Kind == SyntaxKindOpenBracketToken {
			return p.parseEntityGroup()
		}
		if p.scanFunctionCallStart() {
			return p.parseFunctionCallExpression()
		}
		return p.parsePrimaryPathSelector()
	case SyntaxKindMaterializedViewCombineKeyword:
		return p.parseMaterializedViewCombineExpression()
	case SyntaxKindOpenParenToken:
		return p.parseParenthesizedExpression()
	case SyntaxKindToScalarKeyword:
		return p.parseToScalarExpression()
	default:
		if p.scanFunctionCallStart() {
			return p.parseFunctionCallExpression()
		}
		if expr := p.parsePrimaryPathSelector(); expr != nil {
			return expr
		}
		if expr := p.parseInvalidKeywordAsNameReference(); expr != nil {
			return expr
		}
		return nil
	}
}

func (p *queryParserState) parsePrimaryPathSelector() Expression {
	selector := p.parsePathElementSelector()
	if selector == nil {
		return nil
	}
	dataScope := p.parseDataScopeClause()
	if dataScope != nil {
		return MustNewDataScopeExpressionTyped(selector, dataScope)
	}
	return selector
}

func (p *queryParserState) parseParenthesizedExpression() Expression {
	if p.peekToken().Kind != SyntaxKindOpenParenToken {
		return nil
	}
	open := p.parseToken()
	expr := p.parseExpression()
	if isNilSyntaxElement(expr) {
		expr = p.createMissingExpression()
	}
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewParenthesizedExpressionTyped(open, expr, close)
}

func (p *queryParserState) parseDynamicLiteral() Expression {
	if p.peekToken().Kind != SyntaxKindDynamicKeyword || p.peekToken(1).Kind != SyntaxKindOpenParenToken {
		return nil
	}
	keyword := p.parseToken()
	open := p.parseToken()
	value := p.parseJsonValue()
	if value == nil {
		value = createMissingJsonValue()
	}
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewDynamicExpressionTyped(keyword, open, value, close)
}

func (p *queryParserState) parseJsonValue() Expression {
	switch p.peekToken().Kind {
	case SyntaxKindLongLiteralToken, SyntaxKindRealLiteralToken, SyntaxKindMinusToken:
		return p.parseJsonNumber()
	case SyntaxKindBooleanLiteralToken,
		SyntaxKindDateTimeLiteralToken,
		SyntaxKindTimespanLiteralToken,
		SyntaxKindGuidLiteralToken,
		SyntaxKindRawGuidLiteralToken,
		SyntaxKindDecimalLiteralToken,
		SyntaxKindStringLiteralToken,
		SyntaxKindDynamicKeyword:
		return p.parseLiteral()
	case SyntaxKindNullKeyword:
		return MustNewLiteralExpressionTyped(SyntaxKindNullLiteralExpression, p.parseToken())
	case SyntaxKindOpenBracketToken:
		return p.parseJsonArray()
	case SyntaxKindOpenBraceToken:
		if p.scanBracedName() > 0 {
			return p.parseBracedNameReference()
		}
		return p.parseJsonObject()
	default:
		return nil
	}
}

func (p *queryParserState) parseJsonNumber() Expression {
	switch p.peekToken().Kind {
	case SyntaxKindLongLiteralToken, SyntaxKindRealLiteralToken:
		return p.parseLiteral()
	case SyntaxKindMinusToken:
		number := p.peekToken(1)
		if number.Kind == SyntaxKindLongLiteralToken || number.Kind == SyntaxKindRealLiteralToken {
			operator := p.parseToken()
			literal := p.parseLiteral()
			return MustNewPrefixUnaryExpressionTyped(SyntaxKindUnaryMinusExpression, operator, literal)
		}
	}
	return nil
}

func (p *queryParserState) parseJsonArray() Expression {
	if p.peekToken().Kind != SyntaxKindOpenBracketToken {
		return nil
	}
	open := p.parseToken()
	values := p.parseJsonValueCommaList()
	close := p.parseRequiredToken(SyntaxKindCloseBracketToken)
	return MustNewJsonArrayExpressionTyped(open, values, close)
}

func (p *queryParserState) parseJsonObject() Expression {
	if p.peekToken().Kind != SyntaxKindOpenBraceToken {
		return nil
	}
	open := p.parseToken()
	pairs := p.parseJsonPairCommaList()
	close := p.parseRequiredToken(SyntaxKindCloseBraceToken)
	return MustNewJsonObjectExpressionTyped(open, pairs, close)
}

func (p *queryParserState) parseJsonPair() *JsonPair {
	switch p.peekToken().Kind {
	case SyntaxKindStringLiteralToken:
		name := p.parseToken()
		colon := p.parseRequiredToken(SyntaxKindColonToken)
		value := p.parseJsonValue()
		if value == nil {
			value = createMissingJsonValue()
		}
		return MustNewJsonPairTyped(name, colon, value)
	case SyntaxKindColonToken:
		name := NewMissingToken(SyntaxKindStringLiteralToken, DiagnosticMissingString())
		colon := p.parseToken()
		value := p.parseJsonValue()
		if value == nil {
			value = createMissingJsonValue()
		}
		return MustNewJsonPairTyped(name, colon, value)
	default:
		value := p.parseJsonValue()
		if value == nil {
			return nil
		}
		return MustNewJsonPairTyped(
			NewMissingToken(SyntaxKindStringLiteralToken, DiagnosticMissingString()),
			createMissingToken(SyntaxKindColonToken),
			value,
		)
	}
}

func (p *queryParserState) parseJsonValueCommaList() *SeparatedSyntaxListOf[Expression] {
	return parseCommaList(p, "Expression", p.parseJsonValue, createMissingJsonValue, p.atCommonListEnd, false, false)
}

func (p *queryParserState) parseJsonPairCommaList() *SeparatedSyntaxListOf[*JsonPair] {
	return parseCommaList(p, "JsonPair", p.parseJsonPair, createMissingJsonPair, p.atCommonListEnd, false, false)
}

func createMissingJsonValue() Expression {
	return MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, createOmittedToken(SyntaxKindStringLiteralToken), DiagnosticMissingJsonValue())
}

func createMissingJsonPair() *JsonPair {
	return MustNewJsonPairTyped(
		createOmittedToken(SyntaxKindStringLiteralToken),
		createOmittedToken(SyntaxKindColonToken),
		MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, createOmittedToken(SyntaxKindStringLiteralToken)),
		DiagnosticMissingJsonPair(),
	)
}

func (p *queryParserState) parseFunctionCallOrPath() Expression {
	if p.peekToken().Kind == SyntaxKindToTableKeyword {
		return p.parseToTableExpression()
	}
	expr := p.parsePrimaryExpression()
	for expr != nil {
		switch p.peekToken().Kind {
		case SyntaxKindDotToken:
			dot := p.parseToken()
			selector := p.parsePathElementSelectorOrFunctionCall()
			if selector == nil {
				selector = p.createMissingNameReference()
			}
			expr = MustNewPathExpressionTyped(expr, dot, selector)
		case SyntaxKindOpenBracketToken:
			selector := p.parseBracketedExpression()
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

func (p *queryParserState) parseToScalarExpression() Expression {
	keyword := p.parseToken(SyntaxKindToScalarKeyword)
	if keyword == nil {
		return nil
	}
	kind := p.parseQueryOperatorParameter(QueryOperatorParameters.ToScalarKindParameter)
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	expr := p.parseExpression()
	if isNilSyntaxElement(expr) {
		expr = p.createMissingExpression()
	}
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewToScalarExpressionTyped(keyword, kind, open, expr, close)
}

func (p *queryParserState) parseToTableExpression() Expression {
	keyword := p.parseToken(SyntaxKindToTableKeyword)
	if keyword == nil {
		return nil
	}
	kind := p.parseQueryOperatorParameter(QueryOperatorParameters.ToTableKindParameter)
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	expr := p.parseExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewToTableExpressionTyped(keyword, kind, open, expr, close)
}

func (p *queryParserState) parseMaterializeExpression() *MaterializeExpression {
	keyword := p.parseToken(SyntaxKindMaterializeKeyword)
	if keyword == nil {
		return nil
	}
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	expr := p.parsePipeExpression()
	if expr == nil {
		expr = p.createMissingExpression()
	}
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewMaterializeExpressionTyped(keyword, open, expr, close)
}

func (p *queryParserState) parseUnaryPlusOrMinusExpression() Expression {
	switch p.peekToken().Kind {
	case SyntaxKindMinusToken:
		return MustNewPrefixUnaryExpressionTyped(SyntaxKindUnaryMinusExpression, p.parseToken(), p.parseFunctionCallOrPathOrMissing())
	case SyntaxKindPlusToken:
		return MustNewPrefixUnaryExpressionTyped(SyntaxKindUnaryPlusExpression, p.parseToken(), p.parseFunctionCallOrPathOrMissing())
	default:
		return p.parseFunctionCallOrPath()
	}
}

func (p *queryParserState) parseInvocationExpression() Expression {
	return p.parseUnaryPlusOrMinusExpression()
}

func (p *queryParserState) parseStringOperation() Expression {
	if p.peekToken().Kind == SyntaxKindAsteriskToken {
		if kind := stringOperationKind(p.peekToken(1).Kind); kind != SyntaxKindNone {
			left := MustNewStarExpressionTyped(p.parseToken())
			operator := p.parseToken()
			right := p.parseUnaryPlusOrMinusExpression()
			if right == nil {
				right = p.createMissingExpression()
			}
			return MustNewBinaryExpressionTyped(kind, left, operator, right)
		}
	}

	expr := p.parseUnaryPlusOrMinusExpression()
	if expr != nil {
		if kind := stringOperationKind(p.peekToken().Kind); kind != SyntaxKindNone {
			operator := p.parseToken()
			right := p.parseUnaryPlusOrMinusExpression()
			if right == nil {
				right = p.createMissingExpression()
			}
			expr = MustNewBinaryExpressionTyped(kind, expr, operator, right)
		}
	}
	return expr
}

func (p *queryParserState) parseMultiplicativeExpression() Expression {
	expr := p.parseStringOperation()
	for expr != nil {
		kind := multiplicativeExpressionKind(p.peekToken().Kind)
		if kind == SyntaxKindNone {
			return expr
		}
		operator := p.parseToken()
		right := p.parseStringOperation()
		if right == nil {
			right = p.createMissingExpression()
		}
		expr = MustNewBinaryExpressionTyped(kind, expr, operator, right)
	}
	return expr
}

func (p *queryParserState) parseAdditiveExpression() Expression {
	expr := p.parseMultiplicativeExpression()
	for expr != nil {
		kind := additiveExpressionKind(p.peekToken().Kind)
		if kind == SyntaxKindNone {
			return expr
		}
		operator := p.parseToken()
		right := p.parseMultiplicativeExpression()
		if right == nil {
			right = p.createMissingExpression()
		}
		expr = MustNewBinaryExpressionTyped(kind, expr, operator, right)
	}
	return expr
}

func (p *queryParserState) parseRelationalExpression() Expression {
	expr := p.parseAdditiveExpression()
	if expr != nil {
		if kind := relationalExpressionKind(p.peekToken().Kind); kind != SyntaxKindNone {
			operator := p.parseToken()
			right := p.parseAdditiveExpression()
			if right == nil {
				right = p.createMissingExpression()
			}
			expr = MustNewBinaryExpressionTyped(kind, expr, operator, right)
		}
	}
	return expr
}

func (p *queryParserState) parseEqualityExpression() Expression {
	if p.peekToken().Kind == SyntaxKindAsteriskToken && p.peekToken(1).Kind == SyntaxKindEqualEqualToken {
		left := MustNewStarExpressionTyped(p.parseToken())
		operator := p.parseToken()
		right := p.parseRelationalExpression()
		if right == nil {
			right = p.createMissingExpression()
		}
		return MustNewBinaryExpressionTyped(SyntaxKindEqualExpression, left, operator, right)
	}

	expr := p.parseRelationalExpression()
	if expr == nil {
		return nil
	}

	switch p.peekToken().Kind {
	case SyntaxKindEqualEqualToken:
		operator := p.parseToken()
		right := p.parseRelationalExpression()
		if right == nil {
			right = p.createMissingExpression()
		}
		return MustNewBinaryExpressionTyped(SyntaxKindEqualExpression, expr, operator, right)
	case SyntaxKindBangEqualToken, SyntaxKindLessThanGreaterThanToken:
		operator := p.parseToken()
		right := p.parseRelationalExpression()
		if right == nil {
			right = createMissingTokenLiteral()
		}
		return MustNewBinaryExpressionTyped(SyntaxKindNotEqualExpression, expr, operator, right)
	case SyntaxKindInKeyword:
		if p.peekToken(1).Kind != SyntaxKindRangeKeyword {
			return MustNewInExpressionTyped(SyntaxKindInExpression, expr, p.parseToken(), p.parseRequiredInOperatorExpressionList())
		}
	case SyntaxKindInCsKeyword:
		return MustNewInExpressionTyped(SyntaxKindInCsExpression, expr, p.parseToken(), p.parseRequiredInOperatorExpressionList())
	case SyntaxKindNotInKeyword:
		return MustNewInExpressionTyped(SyntaxKindNotInExpression, expr, p.parseToken(), p.parseRequiredInOperatorExpressionList())
	case SyntaxKindNotInCsKeyword:
		return MustNewInExpressionTyped(SyntaxKindNotInCsExpression, expr, p.parseToken(), p.parseRequiredInOperatorExpressionList())
	case SyntaxKindHasAnyKeyword:
		return MustNewHasAnyExpressionTyped(SyntaxKindHasAnyExpression, expr, p.parseToken(), p.parseRequiredInOperatorExpressionList())
	case SyntaxKindHasAllKeyword:
		return MustNewHasAllExpressionTyped(SyntaxKindHasAllExpression, expr, p.parseToken(), p.parseRequiredInOperatorExpressionList())
	case SyntaxKindBetweenKeyword:
		return MustNewBetweenExpressionTyped(SyntaxKindBetweenExpression, expr, p.parseToken(), p.parseRequiredExpressionCouple())
	case SyntaxKindNotBetweenKeyword:
		return MustNewBetweenExpressionTyped(SyntaxKindNotBetweenExpression, expr, p.parseToken(), p.parseRequiredExpressionCouple())
	}
	return expr
}

func (p *queryParserState) parseLogicalAndExpression() Expression {
	expr := p.parseEqualityExpression()
	for expr != nil && p.peekToken().Kind == SyntaxKindAndKeyword {
		operator := p.parseToken()
		right := p.parseEqualityExpression()
		if right == nil {
			right = p.createMissingExpression()
		}
		expr = MustNewBinaryExpressionTyped(SyntaxKindAndExpression, expr, operator, right)
	}
	return expr
}

func (p *queryParserState) parseLogicalOrExpression() Expression {
	expr := p.parseLogicalAndExpression()
	for expr != nil && p.peekToken().Kind == SyntaxKindOrKeyword {
		operator := p.parseToken()
		right := p.parseLogicalAndExpression()
		if right == nil {
			right = p.createMissingExpression()
		}
		expr = MustNewBinaryExpressionTyped(SyntaxKindOrExpression, expr, operator, right)
	}
	return expr
}

func (p *queryParserState) parseRequiredInOperatorExpressionList() *ExpressionList {
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	var expressions *SeparatedSyntaxListOf[Expression]
	if p.scanIsQueryExpression() {
		query := p.parseExpression()
		if query == nil {
			query = p.createMissingExpression()
		}
		expressions = AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>", NewSeparatedElement(query, nil)))
	} else {
		expressions = p.parseUnnamedExpressionCommaList(true)
	}
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewExpressionListTyped(open, expressions, close)
}

func (p *queryParserState) scanIsQueryExpression() bool {
	if p.scanPossibleQueryOperator() {
		return true
	}
	return p.scanIsPipeBeforeEndOfExpression()
}

func (p *queryParserState) scanIsPipeBeforeEndOfExpression() bool {
	parenDepth := 0
	for offset := 0; ; offset++ {
		switch p.peekToken(offset).Kind {
		case SyntaxKindNone, SyntaxKindEndOfTextToken, SyntaxKindSemicolonToken:
			return false
		case SyntaxKindCommaToken:
			if parenDepth == 0 {
				return false
			}
		case SyntaxKindCloseParenToken:
			if parenDepth == 0 {
				return false
			}
			parenDepth--
		case SyntaxKindOpenParenToken:
			parenDepth++
		case SyntaxKindBarToken:
			if parenDepth == 0 {
				return true
			}
		}
	}
}

func (p *queryParserState) parseRequiredExpressionCouple() *ExpressionCouple {
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	first := p.parseUnnamedExpression()
	if first == nil {
		first = p.createMissingExpression()
	}
	dotDot := p.parseRequiredToken(SyntaxKindDotDotToken)
	second := p.parseUnnamedExpression()
	if second == nil {
		second = p.createMissingExpression()
	}
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewExpressionCoupleTyped(open, first, dotDot, second, close)
}

func (p *queryParserState) parseFunctionCallOrPathOrMissing() Expression {
	if expr := p.parseFunctionCallOrPath(); expr != nil {
		return expr
	}
	return p.createMissingExpression()
}

func (p *queryParserState) parseEntityPathExpression() Expression {
	expr := p.parsePathElementSelectorOrFunctionCall()
	for expr != nil {
		switch p.peekToken().Kind {
		case SyntaxKindDotToken:
			dot := p.parseToken()
			selector := p.parsePathElementSelectorOrFunctionCall()
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

func (p *queryParserState) parsePathElementSelectorOrFunctionCall() Expression {
	if p.scanFunctionCallStart() {
		return p.parseFunctionCallExpression()
	}
	return p.parsePathElementSelector()
}

func (p *queryParserState) parsePathElementSelector() Expression {
	if p.peekToken().Kind == SyntaxKindOpenBracketToken {
		if reference := p.parseBracketedEntityNamePathElementSelector(); reference != nil {
			return reference
		}
		return p.parseBracketedExpression()
	}
	return p.parseBarePathElementSelector()
}

func (p *queryParserState) parseBracketedEntityNamePathElementSelector() Expression {
	if p.peekToken().Kind != SyntaxKindOpenBracketToken {
		return nil
	}
	if reference := p.parseBracketedWildcardedNameReference(); reference != nil {
		return reference
	}
	if reference := p.parseBracketedNameReference(); reference != nil {
		return reference
	}
	return nil
}

func (p *queryParserState) parseBarePathElementSelector() Expression {
	if p.peekToken().Kind == SyntaxKindAtToken {
		return MustNewAtExpressionTyped(p.parseToken())
	}
	if isSpecialKeywordAsIdentifier(p.peekToken().Kind) {
		return MustNewNameReferenceTyped(MustNewTokenNameTyped(p.parseToken()), SymbolMatchNone)
	}
	reference := p.parseNameReference()
	if reference == nil {
		return nil
	}
	return reference
}

func isSpecialKeywordAsIdentifier(kind SyntaxKind) bool {
	switch kind {
	case SyntaxKindKindKeyword, SyntaxKindWithSourceKeyword, SyntaxKindWith_SourceKeyword:
		return true
	default:
		return false
	}
}

func (p *queryParserState) parseBracketedExpression() Expression {
	if p.peekToken().Kind != SyntaxKindOpenBracketToken {
		return nil
	}
	open := p.parseToken()
	expr := p.parseExpression()
	if expr == nil {
		expr = p.createMissingNameReference()
	}
	close := p.parseRequiredToken(SyntaxKindCloseBracketToken)
	return MustNewBracketedExpressionTyped(open, expr, close)
}

func (p *queryParserState) parseFunctionCallExpression() *FunctionCallExpression {
	if !p.scanFunctionCallStart() {
		return nil
	}
	name := p.parseFunctionCallName()
	arguments := p.parseArgumentList()
	return MustNewFunctionCallExpressionTyped(name, arguments)
}

func (p *queryParserState) parseDotCompositeFunctionCall() Expression {
	var expr Expression = p.parseFunctionCallExpression()
	for expr != nil && p.peekToken().Kind == SyntaxKindDotToken && p.scanFunctionCallStart(1) {
		dot := p.parseToken()
		call := p.parseFunctionCallExpression()
		expr = MustNewPathExpressionTyped(expr, dot, call)
	}
	return expr
}

func (p *queryParserState) scanTypeOfScalar(offset ...int) bool {
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	if p.peekToken(delta).Kind != SyntaxKindTypeOfKeyword || p.peekToken(delta+1).Kind != SyntaxKindOpenParenToken {
		return false
	}
	if isExtendedParamTypeKind(p.peekToken(delta + 2).Kind) {
		return true
	}
	return p.peekToken(delta+2).Kind == SyntaxKindIdentifierToken && p.peekToken(delta+3).Kind == SyntaxKindCloseParenToken
}

func (p *queryParserState) parseTypeOfLiteral() *TypeOfLiteralExpression {
	if p.peekToken().Kind != SyntaxKindTypeOfKeyword {
		return nil
	}
	if p.scanTypeOfScalar() {
		keyword := p.parseToken()
		open := p.parseRequiredToken(SyntaxKindOpenParenToken)
		typ := p.parseParamTypeExtended()
		if typ == nil {
			typ = p.parseIdentifierTypeExpression()
		}
		if typ == nil {
			typ = createMissingType()
		}
		close := p.parseRequiredToken(SyntaxKindCloseParenToken)
		typeExpression, _ := typ.(Expression)
		if typeExpression == nil {
			typeExpression = p.createMissingExpression()
		}
		return MustNewTypeOfLiteralExpressionTyped(keyword, open, singleExpressionList(typeExpression), close)
	}
	if p.peekToken(1).Kind == SyntaxKindOpenParenToken {
		keyword := p.parseToken()
		open := p.parseToken()
		types := p.parseTypeOfElementCommaList()
		close := p.parseRequiredToken(SyntaxKindCloseParenToken)
		return MustNewTypeOfLiteralExpressionTyped(keyword, open, types, close)
	}
	return nil
}

func (p *queryParserState) parseTypeOfElement() Expression {
	if p.peekToken().Kind == SyntaxKindAsteriskToken {
		return MustNewStarExpressionTyped(p.parseToken())
	}
	if declaration := p.parseNameAndTypeDeclaration(); declaration != nil {
		return declaration
	}
	return nil
}

func (p *queryParserState) parseTypeOfElementCommaList() *SeparatedSyntaxListOf[Expression] {
	return parseCommaList(
		p,
		"Expression",
		p.parseTypeOfElement,
		func() Expression {
			expression, _ := createMissingType().(Expression)
			return expression
		},
		p.atCommonListEnd,
		false,
		false,
	)
}

func (p *queryParserState) parseTypeOfLiteralOrMissing() *TypeOfLiteralExpression {
	if literal := p.parseTypeOfLiteral(); literal != nil {
		return literal
	}
	return createMissingTypeOfLiteral()
}

func createMissingTypeOfLiteral() *TypeOfLiteralExpression {
	return MustNewTypeOfLiteralExpressionTyped(
		createOmittedToken(SyntaxKindTypeOfKeyword),
		createOmittedToken(SyntaxKindOpenParenToken),
		AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>")),
		createOmittedToken(SyntaxKindCloseParenToken),
		DiagnosticMissingTypeOfLiteral(),
	)
}

func singleExpressionList(expr Expression) *SeparatedSyntaxListOf[Expression] {
	return AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>", NewSeparatedElement(expr, nil)))
}

func (p *queryParserState) parseDataTableExpression() *DataTableExpression {
	keyword := p.parseToken(SyntaxKindDataTableKeyword)
	if keyword == nil {
		return nil
	}
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.DataTableParameters, false)
	schema := p.parseRowSchema(false)
	if schema == nil {
		schema = createMissingRowSchema()
	}
	open := p.parseRequiredToken(SyntaxKindOpenBracketToken)
	leadingComma := p.parseToken(SyntaxKindCommaToken)
	values := p.parseDataTableValueCommaList()
	close := p.parseRequiredToken(SyntaxKindCloseBracketToken)
	return MustNewDataTableExpressionTyped(keyword, parameters, schema, open, leadingComma, values, close)
}

func (p *queryParserState) parseExternalDataExpression() *ExternalDataExpression {
	switch p.peekToken().Kind {
	case SyntaxKindExternalDataKeyword, SyntaxKindExternal_DataKeyword:
	default:
		return nil
	}
	keyword := p.parseToken()
	parameters := p.parseQueryOperatorParameterList(QueryOperatorParameters.ExternalDataWithClauseProperties, false)
	schema := p.parseRowSchema(false)
	if schema == nil {
		schema = createMissingRowSchema()
	}
	open := p.parseRequiredToken(SyntaxKindOpenBracketToken)
	values := p.parseExternalDataURICommaList()
	close := p.parseRequiredToken(SyntaxKindCloseBracketToken)
	clause := p.parseExternalDataWithClause()
	return MustNewExternalDataExpressionTyped(keyword, parameters, schema, open, values, close, clause)
}

func (p *queryParserState) parseInlineExternalTableExpression() *InlineExternalTableExpression {
	if p.peekToken().Kind != SyntaxKindInlineExternalTableKeyword {
		return nil
	}
	if !p.scanExternalTableSchema(1) && !p.scanInlineExternalTableKindClause(1) {
		return nil
	}
	keyword := p.parseToken()
	parameters := AsSyntaxListOf[*NamedParameter](NewTypedSyntaxList("NamedParameter"))
	schema := (*RowSchema)(nil)
	if p.scanExternalTableSchema() {
		schema = p.parseRowSchema(true)
	}
	kind := p.parseInlineExternalTableKindClause()
	if schema == nil {
		if isInlineExternalTableDeltaKind(kind) {
			schema = createOmittedInlineExternalTableRowSchema()
		} else {
			schema = createMissingRowSchema()
		}
	}
	partition := p.parseInlineExternalTablePartitionClause()
	pathFormat := p.parseInlineExternalTablePathFormatClause()
	dataFormat := p.parseInlineExternalTableDataFormatClause(kind)
	connectionStrings := p.parseInlineExternalTableConnectionStringsClause()
	withClause := p.parseExternalDataWithClause()
	return MustNewInlineExternalTableExpressionTyped(keyword, parameters, schema, kind, partition, pathFormat, dataFormat, connectionStrings, withClause)
}

func (p *queryParserState) scanExternalTableSchema(offset ...int) bool {
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	if p.peekToken(delta).Kind != SyntaxKindOpenParenToken {
		return false
	}
	for index := delta + 1; ; index++ {
		token := p.peekToken(index)
		switch token.Kind {
		case SyntaxKindNone, SyntaxKindColonToken:
			return true
		case SyntaxKindCloseParenToken:
			return false
		}
	}
}

func (p *queryParserState) scanInlineExternalTableKindClause(offset ...int) bool {
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	return p.peekToken(delta).Kind == SyntaxKindKindKeyword
}

func createOmittedInlineExternalTableRowSchema() *RowSchema {
	return MustNewRowSchemaTyped(
		createOmittedToken(SyntaxKindOpenParenToken),
		nil,
		AsSeparatedSyntaxListOf[*NameAndTypeDeclaration](NewTypedSyntaxList("SeparatedElement<NameAndTypeDeclaration>")),
		createOmittedToken(SyntaxKindCloseParenToken),
	)
}

func (p *queryParserState) parseInlineExternalTableKindClause() *InlineExternalTableKindClause {
	keyword := p.parseToken(SyntaxKindKindKeyword)
	if keyword == nil {
		return MustNewInlineExternalTableKindClauseTyped(
			createOmittedToken(SyntaxKindKindKeyword),
			createOmittedToken(SyntaxKindEqualToken),
			createOmittedToken(SyntaxKindStringLiteralToken),
			DiagnosticMissingExternalTableKind(),
		)
	}
	equal := p.parseRequiredToken(SyntaxKindEqualToken)
	value := p.parseInlineExternalTableKindValueToken()
	return MustNewInlineExternalTableKindClauseTyped(keyword, equal, value)
}

func isInlineExternalTableDeltaKind(kindClause *InlineExternalTableKindClause) bool {
	return kindClause != nil && kindClause.Value() != nil && strings.EqualFold(kindClause.Value().ValueText(), "delta")
}

func (p *queryParserState) parseInlineExternalTableDataFormatClause(kindClause *InlineExternalTableKindClause) *InlineExternalTableDataFormatClause {
	keyword := p.parseToken(SyntaxKindDataFormatKeyword)
	if keyword == nil {
		if isInlineExternalTableDeltaKind(kindClause) {
			return createOmittedInlineExternalTableDataFormatClause()
		}
		return createMissingInlineExternalTableDataFormatClause()
	}
	equal := p.parseRequiredToken(SyntaxKindEqualToken)
	value := p.parseInlineExternalTableIdentifierToken()
	return MustNewInlineExternalTableDataFormatClauseTyped(keyword, equal, value)
}

func createMissingInlineExternalTableDataFormatClause() *InlineExternalTableDataFormatClause {
	return MustNewInlineExternalTableDataFormatClauseTyped(
		createOmittedToken(SyntaxKindDataFormatKeyword),
		createOmittedToken(SyntaxKindEqualToken),
		createOmittedToken(SyntaxKindStringLiteralToken),
		DiagnosticMissingDataFormat(),
	)
}

func createOmittedInlineExternalTableDataFormatClause() *InlineExternalTableDataFormatClause {
	return MustNewInlineExternalTableDataFormatClauseTyped(
		createOmittedToken(SyntaxKindDataFormatKeyword),
		createOmittedToken(SyntaxKindEqualToken),
		createOmittedToken(SyntaxKindStringLiteralToken),
	)
}

func (p *queryParserState) parseInlineExternalTableKindValueToken() *SyntaxToken {
	kind := p.peekToken().Kind
	if kind == SyntaxKindIdentifierToken || IsKeywordKind(kind) {
		token := p.peekToken()
		if stringSliceContainsFold(KustoFacts.InlineExternalTableKinds, token.Text) {
			return p.parseToken()
		}
	}
	return CreateMissingSyntaxTokenForTexts(KustoFacts.InlineExternalTableKinds)
}

func (p *queryParserState) parseInlineExternalTableIdentifierToken() *SyntaxToken {
	kind := p.peekToken().Kind
	if kind == SyntaxKindIdentifierToken || IsKeywordKind(kind) {
		return p.parseToken()
	}
	return NewMissingToken(SyntaxKindIdentifierToken, DiagnosticMissingName())
}

func (p *queryParserState) parseInlineExternalTablePartitionClause() *InlineExternalTablePartitionClause {
	keyword := p.parseToken(SyntaxKindPartitionKeyword)
	if keyword == nil {
		return nil
	}
	by := p.parseRequiredToken(SyntaxKindByKeyword)
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	leadingComma := p.parseToken(SyntaxKindCommaToken)
	columns := p.parsePartitionColumnDeclarationCommaList()
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewInlineExternalTablePartitionClauseTyped(keyword, by, open, leadingComma, columns, close)
}

func (p *queryParserState) parsePartitionColumnDeclarationCommaList() *SeparatedSyntaxListOf[*PartitionColumnDeclaration] {
	var elements []SyntaxElement
	element := (*PartitionColumnDeclaration)(nil)
	if !p.scanCommonListEnd() {
		element = p.parsePartitionColumnDeclaration()
	}
	if element == nil && p.peekToken().Kind == SyntaxKindCommaToken {
		element = createMissingPartitionColumnDeclaration()
	}
	for element != nil {
		if p.peekToken().Kind == SyntaxKindEndOfTextToken || p.peekToken().Kind == SyntaxKindNone || p.scanCommonListEnd() {
			elements = append(elements, NewSeparatedElement(element, nil))
			return AsSeparatedSyntaxListOf[*PartitionColumnDeclaration](NewTypedSyntaxList("SeparatedElement<PartitionColumnDeclaration>", elements...))
		}
		if p.peekToken().Kind == SyntaxKindCommaToken {
			comma := p.parseToken()
			elements = append(elements, NewSeparatedElement(element, comma))
			if p.scanCommonListEnd() {
				break
			}
			next := p.parsePartitionColumnDeclaration()
			if next != nil {
				element = next
				continue
			}
			element = createMissingPartitionColumnDeclaration()
			continue
		}
		next := p.parsePartitionColumnDeclaration()
		if next == nil {
			elements = append(elements, NewSeparatedElement(element, createMissingToken(SyntaxKindCommaToken)))
			return AsSeparatedSyntaxListOf[*PartitionColumnDeclaration](NewTypedSyntaxList("SeparatedElement<PartitionColumnDeclaration>", elements...))
		}
		elements = append(elements, NewSeparatedElement(element, createMissingToken(SyntaxKindCommaToken)))
		element = next
	}
	return AsSeparatedSyntaxListOf[*PartitionColumnDeclaration](NewTypedSyntaxList("SeparatedElement<PartitionColumnDeclaration>", elements...))
}

func createMissingPartitionColumnDeclaration() *PartitionColumnDeclaration {
	return MustNewPartitionColumnDeclarationTyped(
		createMissingNameDeclaration(),
		createMissingToken(SyntaxKindColonToken),
		createMissingType(),
		createOmittedToken(SyntaxKindEqualToken),
		nil,
		DiagnosticMissingPartitionColumnDeclaration(),
	)
}

func (p *queryParserState) parsePartitionColumnDeclaration() *PartitionColumnDeclaration {
	name := p.parseExtendedNameDeclaration()
	if name == nil {
		return nil
	}
	colon := p.parseRequiredToken(SyntaxKindColonToken)
	typ := p.parsePartitionColumnType()
	if typ == nil {
		typ = createMissingType()
	}
	equal := p.parseToken(SyntaxKindEqualToken)
	var expr Expression
	if equal != nil {
		expr = p.parseUnnamedExpression()
	}
	return MustNewPartitionColumnDeclarationTyped(name, colon, typ, equal, expr)
}

func (p *queryParserState) parsePartitionColumnType() TypeExpression {
	switch p.peekToken().Kind {
	case SyntaxKindDateTimeKeyword, SyntaxKindLongKeyword, SyntaxKindStringKeyword:
		return MustNewPrimitiveTypeExpressionTyped(p.parseToken())
	case SyntaxKindNone, SyntaxKindEndOfTextToken, SyntaxKindCommaToken, SyntaxKindCloseParenToken, SyntaxKindEqualToken:
		return nil
	default:
		return MustNewPrimitiveTypeExpressionTyped(p.parseToken(), DiagnosticWrongPartitionColumnType())
	}
}

func (p *queryParserState) parseInlineExternalTablePathFormatClause() *InlineExternalTablePathFormatClause {
	keyword := p.parseToken(SyntaxKindPathFormatKeyword)
	if keyword == nil {
		return nil
	}
	equal := p.parseRequiredToken(SyntaxKindEqualToken)
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	separator := p.parseStringLiteral()
	pathExpressions := p.parseInlineExternalTablePathFormatList()
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewInlineExternalTablePathFormatClauseTyped(keyword, equal, open, separator, pathExpressions, close)
}

func (p *queryParserState) parseInlineExternalTablePathFormatList() *SyntaxListOf[*InlineExternalTablePathFormatPartitionColumnReference] {
	var elements []SyntaxElement
	for !p.scanCommonListEnd() {
		element := p.parseExternalTablePathFormatToken()
		if element == nil {
			element = createMissingExternalTablePathFormatToken()
		}
		elements = append(elements, element)
	}
	if len(elements) == 0 {
		elements = append(elements, createMissingExternalTablePathFormatToken())
	}
	return AsSyntaxListOf[*InlineExternalTablePathFormatPartitionColumnReference](NewTypedSyntaxList("InlineExternalTablePathFormatPartitionColumnReference", elements...))
}

func createMissingExternalTablePathFormatToken() *InlineExternalTablePathFormatPartitionColumnReference {
	return MustNewInlineExternalTablePathFormatPartitionColumnReferenceTyped(
		createMissingValueExpression(),
		createMissingStringLiteral(),
		DiagnosticMissingPathFormatTokens(),
	)
}

func (p *queryParserState) parseExternalTablePathFormatToken() *InlineExternalTablePathFormatPartitionColumnReference {
	var partitionColumnReference Expression
	if p.peekToken().Kind == SyntaxKindDateTimePatternKeyword {
		keyword := p.parseRequiredToken(SyntaxKindDateTimePatternKeyword)
		open := p.parseRequiredToken(SyntaxKindOpenParenToken)
		pattern := p.parseStringLiteral()
		if pattern == nil {
			pattern = createMissingStringLiteral()
		}
		comma := p.parseRequiredToken(SyntaxKindCommaToken)
		partitionColumn := p.parseNameReference()
		if partitionColumn == nil {
			partitionColumn = MustNewNameReferenceTyped(MustNewTokenNameTyped(createMissingToken(SyntaxKindIdentifierToken)), SymbolMatchNone, DiagnosticMissingName())
		}
		close := p.parseRequiredToken(SyntaxKindCloseParenToken)
		partitionColumnReference = MustNewDateTimePatternTyped(keyword, open, pattern, comma, partitionColumn, close)
	} else {
		if nameReference := p.parseNameReference(); nameReference != nil {
			partitionColumnReference = nameReference
		}
	}
	if partitionColumnReference == nil {
		if p.peekToken().Kind != SyntaxKindNone && p.peekToken().Kind != SyntaxKindEndOfTextToken {
			p.parseToken()
		}
		return MustNewInlineExternalTablePathFormatPartitionColumnReferenceTyped(
			createMissingValueExpression(),
			createMissingStringLiteral(),
			DiagnosticUnknownTokenInPathFormatDefinition(),
		)
	}
	return MustNewInlineExternalTablePathFormatPartitionColumnReferenceTyped(partitionColumnReference, p.parseStringLiteral())
}

func (p *queryParserState) parseStringLiteral() *LiteralExpression {
	if p.peekToken().Kind != SyntaxKindStringLiteralToken {
		return nil
	}
	return MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, p.parseToken())
}

func createMissingStringLiteral() *LiteralExpression {
	return MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, createOmittedToken(SyntaxKindStringLiteralToken), DiagnosticMissingString())
}

func createMissingBoolLiteral() *LiteralExpression {
	return MustNewLiteralExpressionTyped(SyntaxKindBooleanLiteralExpression, createOmittedToken(SyntaxKindBooleanLiteralToken), DiagnosticMissingBoolean())
}

func createMissingLongLiteral() *LiteralExpression {
	return MustNewLiteralExpressionTyped(SyntaxKindLongLiteralExpression, createOmittedToken(SyntaxKindLongLiteralToken), DiagnosticMissingNumber())
}

func createMissingRealLiteral() *LiteralExpression {
	return MustNewLiteralExpressionTyped(SyntaxKindRealLiteralExpression, createOmittedToken(SyntaxKindRealLiteralToken), DiagnosticMissingNumber())
}

func createMissingTokenLiteral(tokens ...string) *LiteralExpression {
	if len(tokens) > 0 {
		return MustNewLiteralExpressionTyped(SyntaxKindTokenLiteralExpression, createOmittedToken(SyntaxKindIdentifierToken), DiagnosticTermsExpected(tokens...))
	}
	return MustNewLiteralExpressionTyped(SyntaxKindTokenLiteralExpression, createOmittedToken(SyntaxKindIdentifierToken), DiagnosticTermsExpected("token"))
}

func (p *queryParserState) parseInlineExternalTableConnectionStringsClause() *InlineExternalTableConnectionStringsClause {
	open := p.parseToken(SyntaxKindOpenParenToken)
	if open == nil {
		return MustNewInlineExternalTableConnectionStringsClauseTyped(
			createOmittedToken(SyntaxKindOpenParenToken),
			AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>")),
			createOmittedToken(SyntaxKindCloseParenToken),
			DiagnosticMissingConnectionStrings(),
		)
	}
	values := p.parseExternalDataURICommaList()
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewInlineExternalTableConnectionStringsClauseTyped(open, values, close)
}

func stringSliceContainsFold(values []string, needle string) bool {
	for _, value := range values {
		if strings.EqualFold(value, needle) {
			return true
		}
	}
	return false
}

func (p *queryParserState) parseExternalDataURICommaList() *SeparatedSyntaxListOf[Expression] {
	return parseCommaList(p, "Expression", p.parseExpression, p.createMissingValue, p.atCommonListEnd, false, true)
}

func (p *queryParserState) parseExternalDataPropertyValue() Expression {
	switch p.peekToken().Kind {
	case SyntaxKindStringLiteralToken,
		SyntaxKindLongLiteralToken,
		SyntaxKindRealLiteralToken,
		SyntaxKindBooleanLiteralToken,
		SyntaxKindDateTimeLiteralToken,
		SyntaxKindTypeOfLiteralExpression:
		return p.parseLiteral()
	case SyntaxKindRawGuidLiteralToken, SyntaxKindGuidLiteralToken:
		return MustNewLiteralExpressionTyped(SyntaxKindGuidLiteralExpression, p.parseToken())
	default:
		return p.parseRenameNameDeclaration()
	}
}

func (p *queryParserState) parseExternalDataProperty() *NamedParameter {
	name := p.parseExtendedNameDeclaration()
	if name == nil {
		return nil
	}
	equal := p.parseRequiredToken(SyntaxKindEqualToken)
	value := p.parseExternalDataPropertyValue()
	if isNilSyntaxElement(value) {
		value = p.createMissingValue()
	}
	return MustNewNamedParameterTyped(name, equal, value, CompletionHintNone)
}

func (p *queryParserState) parseExternalDataPropertyCommaList() *SeparatedSyntaxListOf[*NamedParameter] {
	return parseCommaList(p, "NamedParameter", p.parseExternalDataProperty, createMissingNamedParameter, p.atCommonListEnd, false, false)
}

func (p *queryParserState) parseExternalDataWithClause() *ExternalDataWithClause {
	keyword := p.parseToken(SyntaxKindWithKeyword)
	if keyword == nil {
		return nil
	}
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	properties := p.parseExternalDataPropertyCommaList()
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewExternalDataWithClauseTyped(keyword, open, properties, close)
}

func (p *queryParserState) parseRequiredMaterializedViewNameClause() *MaterializedViewCombineNameClause {
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	value := p.parseExpression()
	if value == nil {
		value = p.createMissingExpression()
	}
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewMaterializedViewCombineNameClauseTyped(open, value, close)
}

func (p *queryParserState) parseRequiredMaterializedViewCombineClause(keywordName string) *MaterializedViewCombineClause {
	keyword := p.parseTokenText([]string{keywordName})
	if keyword == nil {
		return createMissingMaterializedViewCombineClause(keywordName)
	}
	open := p.parseRequiredToken(SyntaxKindOpenParenToken)
	var expr Expression
	if keywordName == "aggregations" {
		expr = p.parseSummarizeOperator()
	} else {
		expr = p.parseExpression()
	}
	if isNilSyntaxElement(expr) {
		expr = p.createMissingExpression()
	}
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewMaterializedViewCombineClauseTyped(keyword, open, expr, close)
}

func (p *queryParserState) parseMaterializedViewCombineExpression() *MaterializedViewCombineExpression {
	keyword := p.parseToken(SyntaxKindMaterializedViewCombineKeyword)
	if keyword == nil {
		return nil
	}
	nameClause := p.parseRequiredMaterializedViewNameClause()
	baseClause := p.parseRequiredMaterializedViewCombineClause("base")
	deltaClause := p.parseRequiredMaterializedViewCombineClause("delta")
	aggregationsClause := p.parseRequiredMaterializedViewCombineClause("aggregations")
	return MustNewMaterializedViewCombineExpressionTyped(keyword, nameClause, baseClause, deltaClause, aggregationsClause)
}

func createMissingMaterializedViewCombineClause(name string) *MaterializedViewCombineClause {
	return MustNewMaterializedViewCombineClauseTyped(
		createOmittedToken(SyntaxKindMaterializedViewCombineClause),
		createOmittedToken(SyntaxKindOpenParenToken),
		MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone),
		createOmittedToken(SyntaxKindCloseParenToken),
		DiagnosticMissingMaterializedViewCombineClause(name),
	)
}

func (p *queryParserState) parseEntityGroup() Expression {
	if p.peekToken().Kind != SyntaxKindEntityGroupKeyword {
		return nil
	}
	if functionCall := p.parseFunctionCallExpression(); functionCall != nil {
		return functionCall
	}
	keyword := p.parseToken()
	openBracket := p.parseRequiredToken(SyntaxKindOpenBracketToken)
	expressions := p.parseUnnamedExpressionCommaList(true)
	closeBracket := p.parseRequiredToken(SyntaxKindCloseBracketToken)
	return MustNewEntityGroupTyped(keyword, openBracket, expressions, closeBracket)
}

func (p *queryParserState) parseContextualDataTableExpression() *ContextualDataTableExpression {
	keyword := p.parseToken(SyntaxKindContextualDataTableKeyword)
	if keyword == nil {
		return nil
	}
	id := p.parseLiteral()
	if id == nil {
		id = p.parseUnnamedExpression()
	}
	if id == nil {
		id = p.createMissingExpression()
	}
	schema := p.parseRowSchema(false)
	if schema == nil {
		schema = createMissingRowSchema()
	}
	return MustNewContextualDataTableExpressionTyped(keyword, id, schema)
}

func (p *queryParserState) parseDataTableValueCommaList() *SeparatedSyntaxListOf[Expression] {
	return parseCommaList(p, "Expression", p.parseLiteral, p.createMissingValue, p.atCommonListEnd, false, true)
}

func (p *queryParserState) parseRequiredFunctionCallExpression() *FunctionCallExpression {
	if call := p.parseFunctionCallExpression(); call != nil {
		return call
	}
	if p.scanFunctionCallName() > 0 {
		name := p.parseFunctionCallName()
		return MustNewFunctionCallExpressionTyped(
			name,
			createMissingArgumentList(),
		)
	}
	return createMissingFunctionCallExpression()
}

func createMissingArgumentList() *ExpressionList {
	return MustNewExpressionListTyped(
		NewMissingToken(SyntaxKindOpenParenToken, DiagnosticTermsExpected(missingTokenText(SyntaxKindOpenParenToken))),
		AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>")),
		NewMissingToken(SyntaxKindCloseParenToken, DiagnosticTermsExpected(missingTokenText(SyntaxKindCloseParenToken))),
	)
}

func createOmittedArgumentList() *ExpressionList {
	return MustNewExpressionListTyped(
		createOmittedToken(SyntaxKindOpenParenToken),
		AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>")),
		createOmittedToken(SyntaxKindCloseParenToken),
	)
}

func createMissingFunctionCallExpression() *FunctionCallExpression {
	name := MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone)
	return MustNewFunctionCallExpressionTyped(
		name,
		createOmittedArgumentList(),
		DiagnosticMissingFunctionCall(),
	)
}

func (p *queryParserState) scanFunctionCallStart(offset ...int) bool {
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	length := p.scanFunctionCallName(delta)
	return length > 0 && p.peekToken(delta+length).Kind == SyntaxKindOpenParenToken
}

func (p *queryParserState) scanFunctionCallName(offset ...int) int {
	if length := p.scanName(offset...); length > 0 {
		return length
	}
	return -1
}

func (p *queryParserState) parseFunctionCallName() *NameReference {
	if name := p.parseNameReference(); name != nil {
		return name
	}
	return nil
}

func (p *queryParserState) parseArgumentList() *ExpressionList {
	if p.peekToken().Kind != SyntaxKindOpenParenToken {
		return nil
	}
	open := p.parseToken()
	args := p.parseExpressionCommaList(false)
	close := p.parseToken(SyntaxKindCloseParenToken)
	if close == nil {
		close = createExpectedToken(SyntaxKindCloseParenToken)
	}
	return MustNewExpressionListTyped(open, args, close)
}

func (p *queryParserState) parseUnnamedExpressionCommaList(oneOrMore bool) *SeparatedSyntaxListOf[Expression] {
	return parseCommaList(p, "Expression", p.parseUnnamedExpression, p.createMissingExpression, p.atCommonListEnd, oneOrMore, false)
}

func (p *queryParserState) parseExpressionCommaList(oneOrMore bool) *SeparatedSyntaxListOf[Expression] {
	return parseCommaList(p, "Expression", p.parseArgument, p.createMissingExpression, p.atCommonListEnd, oneOrMore, false)
}

func (p *queryParserState) parseArgument() Expression {
	if p.peekToken().Kind == SyntaxKindAsteriskToken && (p.peekToken(1).Kind == SyntaxKindCloseParenToken || p.peekToken(1).Kind == SyntaxKindCommaToken) {
		return MustNewStarExpressionTyped(p.parseToken())
	}
	return p.parseNamedExpression()
}

func (p *queryParserState) parseNamedExpression() Expression {
	if length := p.scanRenameName(); length > 0 && p.peekToken(length).Kind == SyntaxKindEqualToken {
		name := p.parseRenameNameDeclaration()
		equal := p.parseToken(SyntaxKindEqualToken)
		expr := p.parseUnnamedExpression()
		if expr == nil {
			expr = p.createMissingExpression()
		}
		return MustNewSimpleNamedExpressionTyped(name, equal, expr)
	}
	if IsKeywordKind(p.peekToken().Kind) && p.peekToken(1).Kind == SyntaxKindEqualToken {
		token := p.parseToken()
		name := MustNewNameDeclarationTyped(
			MustNewTokenNameTyped(token),
			DiagnosticNameRequiresBrackets(token.RawText()),
		)
		equal := p.parseToken(SyntaxKindEqualToken)
		expr := p.parseUnnamedExpression()
		if expr == nil {
			expr = p.createMissingExpression()
		}
		return MustNewSimpleNamedExpressionTyped(name, equal, expr)
	}
	if length := p.scanDashedName(); length > 0 && p.peekToken(length).Kind == SyntaxKindEqualToken {
		name := p.parseDashedName()
		rawName := ""
		if name != nil && name.Name() != nil {
			rawName = name.Name().RawText()
		}
		nameDecl := MustNewNameDeclarationTyped(name, DiagnosticNameRequiresBrackets(rawName))
		equal := p.parseToken(SyntaxKindEqualToken)
		expr := p.parseUnnamedExpression()
		if expr == nil {
			expr = p.createMissingExpression()
		}
		return MustNewSimpleNamedExpressionTyped(nameDecl, equal, expr)
	}
	if length := p.scanRenameList(); length > 0 && p.peekToken(length).Kind == SyntaxKindEqualToken {
		open := p.parseToken(SyntaxKindOpenParenToken)
		names := p.parseRenameNameDeclarationCommaList(true)
		close := p.parseRequiredToken(SyntaxKindCloseParenToken)
		equal := p.parseRequiredToken(SyntaxKindEqualToken)
		expr := p.parseUnnamedExpression()
		if expr == nil {
			expr = p.createMissingExpression()
		}
		return MustNewCompoundNamedExpressionTyped(MustNewRenameListTyped(open, names, close), equal, expr)
	}
	return p.parseUnnamedExpression()
}

func (p *queryParserState) scanDashedName(offset ...int) int {
	start := 0
	if len(offset) > 0 {
		start = offset[0]
	}
	token := p.peekToken(start)
	if token.Kind != SyntaxKindIdentifierToken && !IsKeywordKind(token.Kind) {
		return -1
	}
	hasDash := strings.Contains(token.Text, "-")
	delta := start + 1
	for {
		token = p.peekToken(delta)
		if token.Kind == SyntaxKindNone || token.Trivia != "" {
			break
		}
		if token.Kind != SyntaxKindIdentifierToken && !IsKeywordKind(token.Kind) && token.Kind != SyntaxKindMinusToken {
			break
		}
		if token.Kind == SyntaxKindMinusToken || strings.Contains(token.Text, "-") {
			hasDash = true
		}
		delta++
	}
	if !hasDash {
		return -1
	}
	return delta - start
}

func (p *queryParserState) parseDashedName() *TokenName {
	length := p.scanDashedName()
	if length <= 0 {
		return nil
	}
	token := p.parseCombinedToken(length)
	if token == nil {
		return nil
	}
	return MustNewTokenNameTyped(token)
}

func (p *queryParserState) parseRenameNameDeclarationCommaList(oneOrMore bool) *SeparatedSyntaxListOf[*NameDeclaration] {
	return parseCommaList(p, "NameDeclaration", p.parseRenameNameDeclaration, p.createMissingNameDeclaration, p.atCommonListEnd, oneOrMore, false)
}

func (p *queryParserState) scanRenameName(offset ...int) int {
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	switch p.peekToken(delta).Kind {
	case SyntaxKindIdentifierToken:
		return 1
	case SyntaxKindOpenBracketToken:
		return p.scanBracketedName(delta)
	case SyntaxKindOpenBraceToken:
		if p.scanIdentifierOrKeywordAsIdentifier(delta+1) && p.peekToken(delta+2).Kind == SyntaxKindCloseBraceToken {
			return 3
		}
		return -1
	default:
		if p.isExtendedKeywordAsIdentifier(p.peekToken(delta).Kind) {
			return 1
		}
		return -1
	}
}

func (p *queryParserState) parseRenameNameDeclaration() *NameDeclaration {
	name := p.parseRenameName()
	if name == nil {
		return nil
	}
	return MustNewNameDeclarationTyped(name)
}

func (p *queryParserState) parseRenameName() Name {
	switch p.peekToken().Kind {
	case SyntaxKindOpenBracketToken:
		return p.parseBracketedName()
	case SyntaxKindOpenBraceToken:
		return p.parseBracedName()
	default:
		if name := p.parseIdentifierName(); name != nil {
			return name
		}
		return p.parseExtendedKeywordAsIdentifierName()
	}
}

func (p *queryParserState) parseName() Name {
	switch p.peekToken().Kind {
	case SyntaxKindOpenBracketToken:
		return p.parseBracketedName()
	case SyntaxKindOpenBraceToken:
		return p.parseBracedName()
	default:
		return p.parseIdentifierName()
	}
}

func (p *queryParserState) parseNameReference() *NameReference {
	name := p.parseName()
	if name == nil {
		return nil
	}
	return MustNewNameReferenceTyped(name, SymbolMatchNone)
}

func (p *queryParserState) parseBracketedNameReference() *NameReference {
	name := p.parseBracketedName()
	if name == nil {
		return nil
	}
	return MustNewNameReferenceTyped(name, SymbolMatchNone)
}

func (p *queryParserState) parseBracedNameReference() *NameReference {
	name := p.parseBracedName()
	if name == nil {
		return nil
	}
	return MustNewNameReferenceTyped(name, SymbolMatchNone)
}

func (p *queryParserState) parseInvalidKeywordAsNameReference() *NameReference {
	if !p.isKeywordInNamePosition() {
		return nil
	}
	token := p.parseToken()
	if token == nil {
		return nil
	}
	return MustNewNameReferenceTyped(MustNewTokenNameTyped(token), SymbolMatchNone, DiagnosticNameRequiresBrackets(token.RawText()))
}

func (p *queryParserState) isKeywordInNamePosition(offset ...int) bool {
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	return isKeywordInNamePositionKinds(p.peekToken(delta).Kind, p.peekToken(delta+1).Kind)
}

func isKeywordInNamePositionKinds(current, next SyntaxKind) bool {
	if !IsKeywordKind(current) {
		return false
	}
	switch next {
	case SyntaxKindAndKeyword,
		SyntaxKindOrKeyword,
		SyntaxKindEqualEqualToken,
		SyntaxKindBangEqualToken,
		SyntaxKindEqualTildeToken,
		SyntaxKindBangTildeToken,
		SyntaxKindGreaterThanToken,
		SyntaxKindGreaterThanOrEqualToken,
		SyntaxKindLessThanToken,
		SyntaxKindLessThanOrEqualToken,
		SyntaxKindAsteriskToken,
		SyntaxKindSlashToken,
		SyntaxKindPercentToken,
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
		SyntaxKindNotBetweenKeyword,
		SyntaxKindCloseParenToken,
		SyntaxKindCloseBracketToken,
		SyntaxKindCloseBraceToken,
		SyntaxKindDotToken,
		SyntaxKindCommaToken,
		SyntaxKindColonToken,
		SyntaxKindBarToken:
		return true
	default:
		return false
	}
}

func (p *queryParserState) scanBracedName(offset ...int) int {
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	if p.peekToken(delta).Kind != SyntaxKindOpenBraceToken {
		return -1
	}
	pos := delta + 1
	token := p.peekToken(pos)
	if (token.Kind != SyntaxKindIdentifierToken && !IsKeywordKind(token.Kind)) || token.Trivia != "" {
		return -1
	}
	pos++
	token = p.peekToken(pos)
	if token.Kind == SyntaxKindOpenBracketToken && token.Trivia == "" {
		pos++
		token = p.peekToken(pos)
		if token.Kind == SyntaxKindMinusToken && token.Trivia == "" {
			pos++
		}
		token = p.peekToken(pos)
		if token.Kind != SyntaxKindLongLiteralToken || token.Trivia != "" {
			return -1
		}
		pos++
		token = p.peekToken(pos)
		if token.Kind != SyntaxKindCloseBracketToken || token.Trivia != "" {
			return -1
		}
		pos++
	}
	token = p.peekToken(pos)
	if token.Kind != SyntaxKindCloseBraceToken || token.Trivia != "" {
		return -1
	}
	return pos + 1 - delta
}

func (p *queryParserState) scanName(offset ...int) int {
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	switch p.peekToken(delta).Kind {
	case SyntaxKindIdentifierToken:
		return 1
	case SyntaxKindOpenBracketToken:
		return p.scanBracketedName(delta)
	case SyntaxKindOpenBraceToken:
		return p.scanBracedName(delta)
	default:
		if p.scanIdentifierOrKeywordAsIdentifier(delta) {
			return 1
		}
		return -1
	}
}

func (p *queryParserState) scanExtendedName(offset ...int) int {
	if length := p.scanName(offset...); length > 0 {
		return length
	}
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	if p.isExtendedKeywordAsIdentifier(p.peekToken(delta).Kind) {
		return 1
	}
	return -1
}

func (p *queryParserState) scanRenameList(offset ...int) int {
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	if p.peekToken(delta).Kind != SyntaxKindOpenParenToken {
		return -1
	}
	delta++
	length := p.scanExtendedName(delta)
	if length <= 0 {
		return -1
	}
	for {
		delta += length
		if p.peekToken(delta).Kind != SyntaxKindCommaToken {
			break
		}
		delta++
		length = p.scanExtendedName(delta)
		if length <= 0 {
			return -1
		}
	}
	if p.peekToken(delta).Kind != SyntaxKindCloseParenToken {
		return -1
	}
	return delta + 1
}

func (p *queryParserState) createMissingExpression() Expression {
	diagnostic := DiagnosticMissingExpression()
	token := p.peekToken()
	if IsKeywordKind(token.Kind) {
		diagnostic = DiagnosticMissingExpressionWithKeyword(token.Text)
	}
	return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone, diagnostic)
}

func (p *queryParserState) createMissingNameReference() Expression {
	return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone, p.missingNameDiagnostic())
}

func stringOperationKind(kind SyntaxKind) SyntaxKind {
	switch kind {
	case SyntaxKindEqualTildeToken:
		return SyntaxKindEqualTildeExpression
	case SyntaxKindBangTildeToken:
		return SyntaxKindBangTildeExpression
	case SyntaxKindHasKeyword:
		return SyntaxKindHasExpression
	case SyntaxKindColonToken:
		return SyntaxKindSearchExpression
	case SyntaxKindNotHasKeyword:
		return SyntaxKindNotHasExpression
	case SyntaxKindHasCsKeyword:
		return SyntaxKindHasCsExpression
	case SyntaxKindNotHasCsKeyword:
		return SyntaxKindNotHasCsExpression
	case SyntaxKindHasPrefixKeyword:
		return SyntaxKindHasPrefixExpression
	case SyntaxKindNotHasPrefixKeyword:
		return SyntaxKindNotHasPrefixExpression
	case SyntaxKindHasPrefixCsKeyword:
		return SyntaxKindHasPrefixCsExpression
	case SyntaxKindNotHasPrefixCsKeyword:
		return SyntaxKindNotHasPrefixCsExpression
	case SyntaxKindHasSuffixKeyword:
		return SyntaxKindHasSuffixExpression
	case SyntaxKindNotHasSuffixKeyword:
		return SyntaxKindNotHasSuffixExpression
	case SyntaxKindHasSuffixCsKeyword:
		return SyntaxKindHasSuffixCsExpression
	case SyntaxKindNotHasSuffixCsKeyword:
		return SyntaxKindNotHasSuffixCsExpression
	case SyntaxKindLikeKeyword:
		return SyntaxKindLikeExpression
	case SyntaxKindNotLikeKeyword:
		return SyntaxKindNotLikeExpression
	case SyntaxKindLikeCsKeyword:
		return SyntaxKindLikeCsExpression
	case SyntaxKindNotLikeCsKeyword:
		return SyntaxKindNotLikeCsExpression
	case SyntaxKindContainsKeyword:
		return SyntaxKindContainsExpression
	case SyntaxKindNotContainsKeyword, SyntaxKindNotBangContainsKeyword:
		return SyntaxKindNotContainsExpression
	case SyntaxKindContainsCsKeyword, SyntaxKindContains_CsKeyword:
		return SyntaxKindContainsCsExpression
	case SyntaxKindNotContainsCsKeyword, SyntaxKindNotBangContainsCsKeyword:
		return SyntaxKindNotContainsCsExpression
	case SyntaxKindStartsWithKeyword:
		return SyntaxKindStartsWithExpression
	case SyntaxKindNotStartsWithKeyword:
		return SyntaxKindNotStartsWithExpression
	case SyntaxKindStartsWithCsKeyword:
		return SyntaxKindStartsWithCsExpression
	case SyntaxKindNotStartsWithCsKeyword:
		return SyntaxKindNotStartsWithCsExpression
	case SyntaxKindEndsWithKeyword:
		return SyntaxKindEndsWithExpression
	case SyntaxKindNotEndsWithKeyword:
		return SyntaxKindNotEndsWithExpression
	case SyntaxKindEndsWithCsKeyword:
		return SyntaxKindEndsWithCsExpression
	case SyntaxKindNotEndsWithCsKeyword:
		return SyntaxKindNotEndsWithCsExpression
	case SyntaxKindMatchesRegexKeyword:
		return SyntaxKindMatchesRegexExpression
	default:
		return SyntaxKindNone
	}
}

func multiplicativeExpressionKind(kind SyntaxKind) SyntaxKind {
	switch kind {
	case SyntaxKindAsteriskToken:
		return SyntaxKindMultiplyExpression
	case SyntaxKindSlashToken:
		return SyntaxKindDivideExpression
	case SyntaxKindPercentToken:
		return SyntaxKindModuloExpression
	default:
		return SyntaxKindNone
	}
}

func additiveExpressionKind(kind SyntaxKind) SyntaxKind {
	switch kind {
	case SyntaxKindPlusToken:
		return SyntaxKindAddExpression
	case SyntaxKindMinusToken:
		return SyntaxKindSubtractExpression
	default:
		return SyntaxKindNone
	}
}

func relationalExpressionKind(kind SyntaxKind) SyntaxKind {
	switch kind {
	case SyntaxKindLessThanToken:
		return SyntaxKindLessThanExpression
	case SyntaxKindLessThanOrEqualToken:
		return SyntaxKindLessThanOrEqualExpression
	case SyntaxKindGreaterThanToken:
		return SyntaxKindGreaterThanExpression
	case SyntaxKindGreaterThanOrEqualToken:
		return SyntaxKindGreaterThanOrEqualExpression
	default:
		return SyntaxKindNone
	}
}
