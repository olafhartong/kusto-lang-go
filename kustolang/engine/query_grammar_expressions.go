package kustolang

import "strings"

// Ported from the scalar expression and named-expression regions in
// src/Kusto.Language/Parser/QueryGrammar.cs. Parenthesized scalar expressions
// intentionally recurse to UnnamedExpression until the pipe grammar is native.

func (g *QueryGrammar) initializeScalarExpressionParsers() {
	var unnamed ResultParser[LexicalToken, Expression]
	var named ResultParser[LexicalToken, Expression]
	unnamedForward := Forward(func() ResultParser[LexicalToken, Expression] { return unnamed })
	namedForward := Forward(func() ResultParser[LexicalToken, Expression] { return named })
	fullExpression := Forward(func() ResultParser[LexicalToken, Expression] { return g.Expression })

	missingExpression := func(source Source[LexicalToken], start int) Expression {
		return queryGrammarMissingExpression(source, start)
	}
	parenthesized := Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		Required(fullExpression, missingExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open *SyntaxToken, expression Expression, close *SyntaxToken) Expression {
			return MustNewParenthesizedExpressionTyped(open, expression, close)
		},
	)
	bracketed := Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenBracketToken),
		Required(unnamedForward, func(Source[LexicalToken], int) Expression { return queryGrammarMissingNameReference() }),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBracketToken),
		func(open *SyntaxToken, expression Expression, close *SyntaxToken) Expression {
			return MustNewBracketedExpressionTyped(open, expression, close)
		},
	)
	starExpression := Rule1(asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindAsteriskToken).Hide()), func(token *SyntaxToken) Expression {
		return MustNewStarExpressionTyped(token)
	})
	atExpression := Rule1(SyntaxTokenKindParser(SyntaxKindAtToken), func(token *SyntaxToken) Expression {
		return MustNewAtExpressionTyped(token)
	})
	specialKeywordTokens := asResultParser[LexicalToken, *SyntaxToken](
		SyntaxTokenKindsParser([]SyntaxKind{SyntaxKindKindKeyword, SyntaxKindWithSourceKeyword, SyntaxKindWith_SourceKeyword}).Hide(),
	)
	specialKeywordReference := asResultParser[LexicalToken, Expression](Rule1(
		specialKeywordTokens,
		func(token *SyntaxToken) Expression {
			return MustNewNameReferenceTyped(MustNewTokenNameTyped(token), SymbolMatchNone)
		},
	).Hide())
	bareNameStart := Match(func(token LexicalToken) bool {
		return token.Kind == SyntaxKindIdentifierToken || token.Kind == SyntaxKindOpenBraceToken ||
			(IsKeywordKind(token.Kind) && SyntaxKindCanBeIdentifier(token.Kind))
	})
	barePathElementSelector := FirstValue(
		atExpression,
		specialKeywordReference,
		IfValue(bareNameStart, g.SimpleNameReference),
	)
	bracketedNameScan := MatchConsumer[LexicalToken](func(source Source[LexicalToken], start int) int {
		if source.IsEnd(start) || source.Peek(start).Kind != SyntaxKindOpenBracketToken {
			return -1
		}
		pos := start + 1
		for !source.IsEnd(pos) && source.Peek(pos).Kind == SyntaxKindStringLiteralToken {
			pos++
		}
		if pos == start+1 || source.IsEnd(pos) || source.Peek(pos).Kind != SyntaxKindCloseBracketToken {
			return -1
		}
		return pos + 1 - start
	})
	bracketedNameReference := IfValue(
		bracketedNameScan,
		g.SimpleNameReference,
	)
	bracketedWildcardReference := IfValue(
		And(SyntaxTokenKindParser(SyntaxKindOpenBracketToken), g.WildcardedIdentifier),
		Rule3(
			SyntaxTokenKindParser(SyntaxKindOpenBracketToken),
			g.WildcardedIdentifier,
			RequiredSyntaxTokenKindParser(SyntaxKindCloseBracketToken),
			func(open *SyntaxToken, wildcard *SyntaxToken, close *SyntaxToken) Expression {
				return MustNewNameReferenceTyped(MustNewBracketedWildcardedNameTyped(open, wildcard, close), SymbolMatchNone)
			},
		),
	)
	bracketedPathElementSelector := FirstValue(bracketedNameReference, bracketedWildcardReference, bracketed)
	pathElementSelector := FirstValue(barePathElementSelector, bracketedPathElementSelector)
	g.simplePathExpression = ApplyZeroOrMore(pathElementSelector, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
		return FirstRight(
			RuleRight2(
				SyntaxTokenKindParser(SyntaxKindDotToken),
				Required(pathElementSelector, func(Source[LexicalToken], int) Expression { return queryGrammarMissingNameReference() }),
				func(left Expression, dot *SyntaxToken, selector Expression) Expression {
					return MustNewPathExpressionTyped(left, dot, selector)
				},
			),
			RuleRight1(bracketedPathElementSelector, func(left Expression, selector Expression) Expression {
				return MustNewElementExpressionTyped(left, selector)
			}),
		)
	})
	dataScopeClause := func(kind CompletionKind) ResultParser[LexicalToken, *DataScopeClause] {
		return WithSyntaxCompletion(
			Rule3(
				HiddenSyntaxTokenKindParser(SyntaxKindDataScopeKeyword),
				RequiredSyntaxTokenKindParser(SyntaxKindEqualToken),
				RequiredSyntaxTokenTextsParser(KustoFacts.DataScopeValues),
				func(keyword, equal, value *SyntaxToken) *DataScopeClause {
					return MustNewDataScopeClauseTyped(keyword, equal, value)
				},
			),
			NewCompletionItem(kind, "datascope"),
		)
	}
	primaryPathSelector := ApplyOptional(pathElementSelector, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
		return RuleRight1(dataScopeClause(CompletionKindTabularSuffix), func(left Expression, clause *DataScopeClause) Expression {
			return MustNewDataScopeExpressionTyped(left, clause)
		})
	})
	invalidKeywordAsNameReference := MatchConsumerValue(
		func(source Source[LexicalToken], start int) int {
			if !source.IsEnd(start) && isKeywordInNamePositionKinds(source.Peek(start).Kind, source.Peek(start+1).Kind) {
				return 1
			}
			return -1
		},
		func(source Source[LexicalToken], start, _ int) Expression {
			token := SyntaxTokenFromLexical(source.Peek(start))
			return MustNewNameReferenceTyped(
				MustNewTokenNameTyped(token),
				SymbolMatchNone,
				DiagnosticNameRequiresBrackets(token.RawText()),
			)
		},
	)

	knownFunctionNames := map[string]struct{}{}
	for _, function := range Functions.All {
		knownFunctionNames[strings.ToLower(function.SymbolName())] = struct{}{}
	}
	for _, aggregate := range Aggregates.All {
		knownFunctionNames[strings.ToLower(aggregate.SymbolName())] = struct{}{}
	}
	knownFunctionToken := asResultParser[LexicalToken, *SyntaxToken](MatchValue(
		func(token LexicalToken) bool {
			_, ok := knownFunctionNames[strings.ToLower(token.Text)]
			return ok
		},
		func(token LexicalToken) *SyntaxToken { return SyntaxTokenFromLexical(token) },
	).Hide())
	knownFunctionReference := asResultParser[LexicalToken, Expression](Rule1(knownFunctionToken, func(token *SyntaxToken) Expression {
		return MustNewNameReferenceTyped(MustNewTokenNameTyped(token), SymbolMatchNone)
	}).Hide())
	inlineExternalFunctionReference := asResultParser[LexicalToken, Expression](Rule1(
		SyntaxTokenKindParser(SyntaxKindInlineExternalTableKeyword),
		func(token *SyntaxToken) Expression {
			return MustNewNameReferenceTyped(MustNewTokenNameTyped(token), SymbolMatchNone)
		},
	).Hide())
	functionCallNames := FirstValue(g.SimpleNameReference, knownFunctionReference, inlineExternalFunctionReference)
	functionName := Rule1(functionCallNames, func(expression Expression) *NameReference {
		return expression.(*NameReference)
	})
	argument := FirstValue(
		IfValue(
			And(SyntaxTokenKindParser(SyntaxKindAsteriskToken), Or(SyntaxTokenKindParser(SyntaxKindCloseParenToken), SyntaxTokenKindParser(SyntaxKindCommaToken))),
			starExpression,
		),
		namedForward,
	)
	argumentList := Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		SyntaxCommaListParser(argument, missingExpression, false, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open *SyntaxToken, arguments *SeparatedSyntaxListOf[Expression], close *SyntaxToken) *ExpressionList {
			return MustNewExpressionListTyped(open, arguments, close)
		},
	)
	functionCallConcrete := asResultParser[LexicalToken, *FunctionCallExpression](Rule2(functionName, argumentList, func(name *NameReference, arguments *ExpressionList) *FunctionCallExpression {
		return MustNewFunctionCallExpressionTyped(name, arguments)
	}).WithTag("<FunctionCall>"))
	functionCall := Rule1(functionCallConcrete, func(call *FunctionCallExpression) Expression { return call })
	g.functionCallExpression = functionCall
	functionCallWithoutArguments := Rule1(functionName, func(name *NameReference) *FunctionCallExpression {
		return MustNewFunctionCallExpressionTyped(name, createMissingArgumentList())
	})
	g.requiredFunctionCall = Required(
		FirstValue(functionCallConcrete, functionCallWithoutArguments),
		func(Source[LexicalToken], int) *FunctionCallExpression { return createMissingFunctionCallExpression() },
	)
	scanBracketedFunctionName := MatchConsumer[LexicalToken](func(source Source[LexicalToken], start int) int {
		if source.IsEnd(start) || source.Peek(start).Kind != SyntaxKindOpenBracketToken {
			return -1
		}
		pos := start + 1
		for !source.IsEnd(pos) && source.Peek(pos).Kind == SyntaxKindStringLiteralToken {
			pos++
		}
		if pos == start+1 {
			return -1
		}
		if !source.IsEnd(pos) && source.Peek(pos).Kind == SyntaxKindCloseBracketToken {
			pos++
		}
		return pos - start
	})
	scanFunctionCallName := First(
		SyntaxTokenKindParser(SyntaxKindIdentifierToken),
		scanBracketedFunctionName,
		knownFunctionToken,
		HiddenSyntaxTokenKindParser(SyntaxKindInlineExternalTableKeyword),
	)
	scanFunctionCall := And(scanFunctionCallName, SyntaxTokenKindParser(SyntaxKindOpenParenToken))
	g.dotCompositeFunctionCall = ApplyZeroOrMore(functionCall, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
		return RuleRight2(
			SyntaxTokenKindParser(SyntaxKindDotToken),
			IfValue(scanFunctionCall, functionCall),
			func(left Expression, dot *SyntaxToken, right Expression) Expression {
				return MustNewPathExpressionTyped(left, dot, right)
			},
		)
	})
	toScalarExpression := Rule5(
		SyntaxTokenKindParser(
			SyntaxKindToScalarKeyword,
			NewCompletionItem(CompletionKindScalarPrefix, "toscalar"),
		),
		OptionalValue(queryGrammarParameter(g, QueryOperatorParameters.ToScalarKindParameter, false)),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		Required(fullExpression, missingExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(keyword *SyntaxToken, kind *NamedParameter, open *SyntaxToken, expression Expression, close *SyntaxToken) Expression {
			return MustNewToScalarExpressionTyped(keyword, kind, open, expression, close)
		},
	)
	toTableExpression := Rule5(
		asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(
			SyntaxKindToTableKeyword,
			NewCompletionItem(CompletionKindTabularPrefix, "totable"),
		).Hide()),
		OptionalValue(queryGrammarParameter(g, QueryOperatorParameters.ToTableKindParameter, false)),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		Required(fullExpression, missingExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(keyword *SyntaxToken, kind *NamedParameter, open *SyntaxToken, expression Expression, close *SyntaxToken) Expression {
			return MustNewToTableExpressionTyped(keyword, kind, open, expression, close)
		},
	)
	tabularPrimary := IfValue(
		FirstValue(
			SyntaxTokenKindParser(SyntaxKindDataTableKeyword),
			SyntaxTokenKindParser(SyntaxKindContextualDataTableKeyword),
			SyntaxTokenKindParser(SyntaxKindExternalDataKeyword),
			SyntaxTokenKindParser(SyntaxKindExternal_DataKeyword),
			SyntaxTokenKindParser(SyntaxKindInlineExternalTableKeyword),
			HiddenSyntaxTokenKindParser(SyntaxKindMaterializedViewCombineKeyword),
		),
		Forward(func() ResultParser[LexicalToken, Expression] { return g.tabularPrimaryExpression }),
	)
	inlineExternalPrimary := IfValue(
		Or(
			And(
				HiddenSyntaxTokenKindParser(SyntaxKindInlineExternalTableKeyword),
				HiddenSyntaxTokenKindParser(SyntaxKindOpenParenToken),
				Not(HiddenSyntaxTokenKindParser(SyntaxKindCloseParenToken)),
			),
			And(
				HiddenSyntaxTokenKindParser(SyntaxKindInlineExternalTableKeyword),
				HiddenSyntaxTokenKindParser(SyntaxKindKindKeyword),
			),
		),
		Forward(func() ResultParser[LexicalToken, Expression] { return g.tabularPrimaryExpression }),
	)

	primary := FirstValue(
		g.Literal,
		parenthesized,
		tabularPrimary,
		primaryPathSelector,
		invalidKeywordAsNameReference,
	)
	selector := FirstValue(IfValue(scanFunctionCall, functionCall), pathElementSelector)
	bracketedEntityNamePathElementSelector := FirstValue(bracketedWildcardReference, bracketedNameReference)
	g.bracketedEntityNameSelector = bracketedEntityNamePathElementSelector
	g.barePathElementSelector = barePathElementSelector
	wildcardedEntityPathRoot := ApplyZeroOrMore(selector, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
		return FirstRight(
			IfRight(
				And(HiddenSyntaxTokenKindParser(SyntaxKindDotToken), Not(g.WildcardedIdentifier)),
				RuleRight2(
					SyntaxTokenKindParser(SyntaxKindDotToken),
					Required(selector, func(Source[LexicalToken], int) Expression { return queryGrammarMissingNameReference() }),
					func(left Expression, dot *SyntaxToken, right Expression) Expression {
						return MustNewPathExpressionTyped(left, dot, right)
					},
				),
			),
			RuleRight1(bracketedPathElementSelector, func(left Expression, right Expression) Expression {
				return MustNewElementExpressionTyped(left, right)
			}),
		)
	})
	wildcardedEntityReference := asResultParser[LexicalToken, Expression](FirstValue(
		g.WildcardedNameReference,
		ApplyOptional(wildcardedEntityPathRoot, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
			return RuleRight2(
				SyntaxTokenKindParser(SyntaxKindDotToken),
				Required(g.WildcardedNameReference, func(Source[LexicalToken], int) Expression { return queryGrammarMissingNameReference() }),
				func(left Expression, dot *SyntaxToken, right Expression) Expression {
					return MustNewPathExpressionTyped(left, dot, right)
				},
			)
		}),
	).WithTag("<wildcarded-entity-reference>"))
	g.wildcardedEntityReference = wildcardedEntityReference
	g.unionExpression = FirstValue(parenthesized, wildcardedEntityReference, bracketedEntityNamePathElementSelector, barePathElementSelector)
	functionCallOrPathWithSuffix := ApplyZeroOrMore(
		FirstValue(toScalarExpression, inlineExternalPrimary, IfValue(scanFunctionCall, functionCall), primary),
		func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
			return FirstRight(
				RuleRight2(
					SyntaxTokenKindParser(SyntaxKindDotToken),
					Required(selector, func(Source[LexicalToken], int) Expression { return queryGrammarMissingNameReference() }),
					func(left Expression, dot *SyntaxToken, right Expression) Expression {
						return MustNewPathExpressionTyped(left, dot, right)
					},
				),
				RuleRight1(bracketed, func(left Expression, right Expression) Expression {
					return MustNewElementExpressionTyped(left, right)
				}),
			)
		},
	)
	g.entityPathExpression = functionCallOrPathWithSuffix
	functionCallOrPath := asResultParser[LexicalToken, Expression](FirstValue(toTableExpression, functionCallOrPathWithSuffix).WithTag("<function-call-or-path>"))
	g.functionCallOrPath = functionCallOrPath

	requiredFunctionCallOrPath := Required(functionCallOrPath, missingExpression)
	unary := FirstValue(
		Rule2(
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindPlusToken).Hide()),
			requiredFunctionCallOrPath,
			func(operator *SyntaxToken, expression Expression) Expression {
				return MustNewPrefixUnaryExpressionTyped(SyntaxKindUnaryPlusExpression, operator, expression)
			},
		),
		Rule2(
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindMinusToken).Hide()),
			requiredFunctionCallOrPath,
			func(operator *SyntaxToken, expression Expression) Expression {
				return MustNewPrefixUnaryExpressionTyped(SyntaxKindUnaryMinusExpression, operator, expression)
			},
		),
		functionCallOrPath,
	)
	requiredUnary := Required(unary, missingExpression)

	stringOperator := queryGrammarStringOperatorTokenParser()
	stringOperation := FirstValue(
		Rule3(
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindAsteriskToken).Hide()),
			stringOperator,
			requiredUnary,
			func(star, operator *SyntaxToken, right Expression) Expression {
				return MustNewBinaryExpressionTyped(stringOperationKind(operator.Kind()), MustNewStarExpressionTyped(star), operator, right)
			},
		),
		ApplyOptional(unary, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
			return RuleRight2(stringOperator, requiredUnary, func(left Expression, operator *SyntaxToken, right Expression) Expression {
				return MustNewBinaryExpressionTyped(stringOperationKind(operator.Kind()), left, operator, right)
			})
		}),
	)

	multiplicative := queryGrammarApplyBinaryZeroOrMore(stringOperation, []SyntaxKind{
		SyntaxKindAsteriskToken, SyntaxKindSlashToken, SyntaxKindPercentToken,
	}, multiplicativeExpressionKind, missingExpression)
	additive := queryGrammarApplyBinaryZeroOrMore(multiplicative, []SyntaxKind{
		SyntaxKindPlusToken, SyntaxKindMinusToken,
	}, additiveExpressionKind, missingExpression)
	relational := queryGrammarApplyBinaryOptional(additive, []SyntaxKind{
		SyntaxKindLessThanToken, SyntaxKindLessThanOrEqualToken, SyntaxKindGreaterThanToken, SyntaxKindGreaterThanOrEqualToken,
	}, relationalExpressionKind, missingExpression)

	inOperatorExpressions := BestValue(
		SyntaxCommaListParser(unnamedForward, missingExpression, true, false),
		Rule1(fullExpression, func(expression Expression) *SeparatedSyntaxListOf[Expression] {
			return AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList(
				"SeparatedElement<Expression>",
				NewSeparatedElement(expression, nil),
			))
		}),
	)
	expressionList := Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		inOperatorExpressions,
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], close *SyntaxToken) *ExpressionList {
			return MustNewExpressionListTyped(open, expressions, close)
		},
	)
	expressionCouple := Rule5(
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		Required(unnamedForward, missingExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindDotDotToken),
		Required(unnamedForward, missingExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open *SyntaxToken, first Expression, dotDot *SyntaxToken, second Expression, close *SyntaxToken) *ExpressionCouple {
			return MustNewExpressionCoupleTyped(open, first, dotDot, second, close)
		},
	)
	requiredRelational := Required(relational, missingExpression)
	equality := FirstValue(
		Rule3(
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindAsteriskToken).Hide()),
			SyntaxTokenKindParser(SyntaxKindEqualEqualToken),
			requiredRelational,
			func(star, operator *SyntaxToken, right Expression) Expression {
				return MustNewBinaryExpressionTyped(SyntaxKindEqualExpression, MustNewStarExpressionTyped(star), operator, right)
			},
		),
		ApplyOptional(relational, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
			return FirstRight(
				queryGrammarBinaryRight(SyntaxKindEqualEqualToken, SyntaxKindEqualExpression, requiredRelational),
				queryGrammarBinaryRight(SyntaxKindBangEqualToken, SyntaxKindNotEqualExpression, requiredRelational),
				queryGrammarBinaryRight(SyntaxKindLessThanGreaterThanToken, SyntaxKindNotEqualExpression, requiredRelational),
				queryGrammarInRight(SyntaxKindInKeyword, SyntaxKindInExpression, expressionList),
				queryGrammarInRight(SyntaxKindInCsKeyword, SyntaxKindInCsExpression, expressionList),
				queryGrammarInRight(SyntaxKindNotInKeyword, SyntaxKindNotInExpression, expressionList),
				queryGrammarInRight(SyntaxKindNotInCsKeyword, SyntaxKindNotInCsExpression, expressionList),
				queryGrammarHasAnyRight(SyntaxKindHasAnyKeyword, false, expressionList),
				queryGrammarHasAnyRight(SyntaxKindHasAllKeyword, true, expressionList),
				queryGrammarBetweenRight(SyntaxKindBetweenKeyword, SyntaxKindBetweenExpression, expressionCouple),
				queryGrammarBetweenRight(SyntaxKindNotBetweenKeyword, SyntaxKindNotBetweenExpression, expressionCouple),
			)
		}),
	)
	logicalAnd := queryGrammarApplyBinaryZeroOrMore(equality, []SyntaxKind{SyntaxKindAndKeyword}, func(SyntaxKind) SyntaxKind {
		return SyntaxKindAndExpression
	}, missingExpression, CompletionKindScalarInfix)
	logicalOr := queryGrammarApplyBinaryZeroOrMore(logicalAnd, []SyntaxKind{SyntaxKindOrKeyword}, func(SyntaxKind) SyntaxKind {
		return SyntaxKindOrExpression
	}, missingExpression, CompletionKindScalarInfix)
	unnamed = asResultParser[LexicalToken, Expression](logicalOr.WithTag("<expression>"))

	renameName := queryGrammarExtendedNameDeclarationParser(g)
	dashedName := queryGrammarDashedNameParser()
	simpleNamed := IfValue(
		And(renameName, SyntaxTokenKindParser(SyntaxKindEqualToken)),
		Rule3(renameName, SyntaxTokenKindParser(SyntaxKindEqualToken), Required(unnamedForward, missingExpression),
			func(name *NameDeclaration, equal *SyntaxToken, expression Expression) Expression {
				return MustNewSimpleNamedExpressionTyped(name, equal, expression)
			}),
	)
	dashedNamed := IfValue(
		And(dashedName, SyntaxTokenKindParser(SyntaxKindEqualToken)),
		Rule3(dashedName, SyntaxTokenKindParser(SyntaxKindEqualToken), Required(unnamedForward, missingExpression),
			func(name *TokenName, equal *SyntaxToken, expression Expression) Expression {
				declaration := MustNewNameDeclarationTyped(name, DiagnosticNameRequiresBrackets(name.Name().RawText()))
				return MustNewSimpleNamedExpressionTyped(declaration, equal, expression)
			}),
	)
	renameList := Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		SyntaxCommaListParser(renameName, func(Source[LexicalToken], int) *NameDeclaration { return createMissingNameDeclaration() }, true, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open *SyntaxToken, names *SeparatedSyntaxListOf[*NameDeclaration], close *SyntaxToken) *RenameList {
			return MustNewRenameListTyped(open, names, close)
		},
	)
	compoundNamed := IfValue(
		And(renameList, SyntaxTokenKindParser(SyntaxKindEqualToken)),
		Rule3(renameList, SyntaxTokenKindParser(SyntaxKindEqualToken), Required(unnamedForward, missingExpression),
			func(names *RenameList, equal *SyntaxToken, expression Expression) Expression {
				return MustNewCompoundNamedExpressionTyped(names, equal, expression)
			}),
	)
	named = asResultParser[LexicalToken, Expression](FirstValue(simpleNamed, dashedNamed, compoundNamed, unnamedForward).WithTag("<expression>"))
	g.UnnamedExpression = unnamed
	g.NamedExpression = named
}

func queryGrammarApplyBinaryZeroOrMore(
	left ResultParser[LexicalToken, Expression],
	operatorKinds []SyntaxKind,
	expressionKind func(SyntaxKind) SyntaxKind,
	missing func(Source[LexicalToken], int) Expression,
	completionKind ...CompletionKind,
) ResultParser[LexicalToken, Expression] {
	operator := SyntaxTokenKindsParser(operatorKinds)
	if len(completionKind) > 0 && len(operatorKinds) == 1 {
		operatorText := SyntaxText(operatorKinds[0])
		operator = SyntaxTokenKindParser(
			operatorKinds[0],
			NewCompletionItem(completionKind[0], operatorText),
		)
	}
	return ApplyZeroOrMore(left, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
		return RuleRight2(operator, Required(left, missing), func(left Expression, operator *SyntaxToken, right Expression) Expression {
			return MustNewBinaryExpressionTyped(expressionKind(operator.Kind()), left, operator, right)
		})
	})
}

func queryGrammarApplyBinaryOptional(
	left ResultParser[LexicalToken, Expression],
	operatorKinds []SyntaxKind,
	expressionKind func(SyntaxKind) SyntaxKind,
	missing func(Source[LexicalToken], int) Expression,
) ResultParser[LexicalToken, Expression] {
	operator := SyntaxTokenKindsParser(operatorKinds)
	return ApplyOptional(left, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
		return RuleRight2(operator, Required(left, missing), func(left Expression, operator *SyntaxToken, right Expression) Expression {
			return MustNewBinaryExpressionTyped(expressionKind(operator.Kind()), left, operator, right)
		})
	})
}

func queryGrammarBinaryRight(operatorKind, expressionKind SyntaxKind, right ResultParser[LexicalToken, Expression]) RightParser[LexicalToken, Expression] {
	return RuleRight2(SyntaxTokenKindParser(operatorKind), right, func(left Expression, operator *SyntaxToken, right Expression) Expression {
		return MustNewBinaryExpressionTyped(expressionKind, left, operator, right)
	})
}

func queryGrammarInRight(operatorKind, expressionKind SyntaxKind, list ResultParser[LexicalToken, *ExpressionList]) RightParser[LexicalToken, Expression] {
	operatorText := SyntaxText(operatorKind)
	operator := SyntaxTokenKindParser(
		operatorKind,
		NewCompletionItem(
			CompletionKindScalarInfix,
			operatorText,
			WithCompletionApplyTextMarkers(operatorText+" (|)"),
			WithCompletionMatchText(operatorText),
		),
	)
	rule := RuleRight2(operator, list, func(left Expression, operator *SyntaxToken, right *ExpressionList) Expression {
		return MustNewInExpressionTyped(expressionKind, left, operator, right)
	})
	if operatorKind == SyntaxKindInKeyword {
		return IfRight(
			Sequence(HiddenSyntaxTokenKindParser(SyntaxKindInKeyword), Fails(SyntaxTokenKindParser(SyntaxKindRangeKeyword))),
			rule,
		)
	}
	return rule
}

func queryGrammarHasAnyRight(operatorKind SyntaxKind, all bool, list ResultParser[LexicalToken, *ExpressionList]) RightParser[LexicalToken, Expression] {
	operatorText := SyntaxText(operatorKind)
	operator := SyntaxTokenKindParser(
		operatorKind,
		NewCompletionItem(
			CompletionKindScalarInfix,
			operatorText,
			WithCompletionApplyTextMarkers(operatorText+" (|)"),
			WithCompletionMatchText(operatorText),
		),
	)
	return RuleRight2(operator, list, func(left Expression, operator *SyntaxToken, right *ExpressionList) Expression {
		if all {
			return MustNewHasAllExpressionTyped(SyntaxKindHasAllExpression, left, operator, right)
		}
		return MustNewHasAnyExpressionTyped(SyntaxKindHasAnyExpression, left, operator, right)
	})
}

func queryGrammarBetweenRight(operatorKind, expressionKind SyntaxKind, couple ResultParser[LexicalToken, *ExpressionCouple]) RightParser[LexicalToken, Expression] {
	operatorText := SyntaxText(operatorKind)
	operator := SyntaxTokenKindParser(
		operatorKind,
		NewCompletionItem(
			CompletionKindScalarInfix,
			operatorText,
			WithCompletionApplyTextMarkers(operatorText+" (| .. )"),
		),
	)
	return RuleRight2(operator, couple, func(left Expression, operator *SyntaxToken, right *ExpressionCouple) Expression {
		return MustNewBetweenExpressionTyped(expressionKind, left, operator, right)
	})
}

func queryGrammarStringOperatorTokenParser() ResultParser[LexicalToken, *SyntaxToken] {
	return asResultParser[LexicalToken, *SyntaxToken](MatchValue(func(token LexicalToken) bool {
		return stringOperationKind(token.Kind) != SyntaxKindNone
	}, func(token LexicalToken) *SyntaxToken { return SyntaxTokenFromLexical(token) }).WithTag("<string-operator>"))
}

func queryGrammarDashedNameParser() ResultParser[LexicalToken, *TokenName] {
	return MatchConsumerValue(func(source Source[LexicalToken], start int) int {
		if source.IsEnd(start) {
			return -1
		}
		first := source.Peek(start)
		if first.Kind != SyntaxKindIdentifierToken && !IsKeywordKind(first.Kind) {
			return -1
		}
		pos := start + 1
		for !source.IsEnd(pos) {
			token := source.Peek(pos)
			if token.Trivia != "" || (token.Kind != SyntaxKindIdentifierToken && !IsKeywordKind(token.Kind) && token.Kind != SyntaxKindMinusToken) {
				break
			}
			pos++
		}
		return pos - start
	}, func(source Source[LexicalToken], start, length int) *TokenName {
		return MustNewTokenNameTyped(ProduceSyntaxTokenFromText(source, start, length, CombinedLexicalTokenText(source, start, length)))
	})
}

func queryGrammarMissingNameReference() Expression {
	return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone, DiagnosticMissingName())
}
