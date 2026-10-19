package kustolang

// Ported from the public type, schema, literal, JSON and skipped-token parser
// fields in src/Kusto.Language/Parser/QueryGrammar.cs.

func (g *QueryGrammar) initializeTypeAndLiteralParsers(stringOrCompound ResultParser[LexicalToken, Expression]) {
	paramTypeToken := queryGrammarParamTypeExtendedTokenParser()
	g.ParamTypeExtended = Rule1(paramTypeToken, func(token *SyntaxToken) TypeExpression {
		return MustNewPrimitiveTypeExpressionTyped(token)
	})

	paramType := queryGrammarParamTypeTokenParser()
	paramTypeExpression := Rule1(paramType, func(token *SyntaxToken) TypeExpression { return MustNewPrimitiveTypeExpressionTyped(token) })
	invalidParamType := Rule1(queryGrammarInvalidParamTypeTokenParser(), func(token *SyntaxToken) TypeExpression {
		return MustNewPrimitiveTypeExpressionTyped(token, DiagnosticInvalidTypeName(token.RawText()))
	})
	extendedNameDeclaration := queryGrammarExtendedNameDeclarationParser(g)

	var schemaType ResultParser[LexicalToken, *SchemaTypeExpression]
	var nameAndType ResultParser[LexicalToken, *NameAndTypeDeclaration]
	schemaTypeForward := Forward(func() ResultParser[LexicalToken, *SchemaTypeExpression] { return schemaType })
	nameAndTypeForward := Forward(func() ResultParser[LexicalToken, *NameAndTypeDeclaration] { return nameAndType })
	nameAndTypeExpression := Rule1(nameAndTypeForward, func(value *NameAndTypeDeclaration) Expression { return value })
	schemaColumns := SyntaxCommaListParser(nameAndTypeExpression, func(Source[LexicalToken], int) Expression {
		return queryGrammarMissingNameAndTypeDeclaration()
	}, false, true)
	starExpression := Rule1(asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindAsteriskToken).Hide()), func(token *SyntaxToken) Expression {
		return MustNewStarExpressionTyped(token)
	})
	schemaAsterisk := Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		starExpression,
		SyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open *SyntaxToken, star Expression, close *SyntaxToken) *SchemaTypeExpression {
			list := AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>", NewSeparatedElement(star, nil)))
			return MustNewSchemaTypeExpressionTyped(open, list, close)
		},
	)
	schemaMultipart := Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		schemaColumns,
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open *SyntaxToken, columns *SeparatedSyntaxListOf[Expression], close *SyntaxToken) *SchemaTypeExpression {
			return MustNewSchemaTypeExpressionTyped(open, columns, close)
		},
	)
	schemaType = FirstValue(
		IfValue(And(SyntaxTokenKindParser(SyntaxKindOpenParenToken), SyntaxTokenKindParser(SyntaxKindAsteriskToken)), schemaAsterisk),
		schemaMultipart,
	)
	g.SchemaType = schemaType

	validOrInvalidType := FirstValue(paramTypeExpression, invalidParamType)
	missingNameBranch := Rule2(
		SyntaxTokenKindParser(SyntaxKindColonToken),
		Required(validOrInvalidType, func(Source[LexicalToken], int) TypeExpression { return createMissingType() }),
		func(colon *SyntaxToken, typ TypeExpression) *NameAndTypeDeclaration {
			return MustNewNameAndTypeDeclarationTyped(createMissingNameDeclaration(), colon, typ)
		},
	)
	nestedSchemaBranch := IfValue(
		And(extendedNameDeclaration, SyntaxTokenKindParser(SyntaxKindColonToken), SyntaxTokenKindParser(SyntaxKindOpenParenToken)),
		Rule3(extendedNameDeclaration, SyntaxTokenKindParser(SyntaxKindColonToken), schemaTypeForward,
			func(name *NameDeclaration, colon *SyntaxToken, typ *SchemaTypeExpression) *NameAndTypeDeclaration {
				return MustNewNameAndTypeDeclarationTyped(name, colon, typ)
			}),
	)
	ordinaryNameBranch := Rule3(
		extendedNameDeclaration,
		RequiredSyntaxTokenKindParser(SyntaxKindColonToken),
		Required(validOrInvalidType, func(Source[LexicalToken], int) TypeExpression { return createMissingType() }),
		func(name *NameDeclaration, colon *SyntaxToken, typ TypeExpression) *NameAndTypeDeclaration {
			return MustNewNameAndTypeDeclarationTyped(name, colon, typ)
		},
	)
	nameAndType = asResultParser[LexicalToken, *NameAndTypeDeclaration](FirstValue(
		missingNameBranch, nestedSchemaBranch, ordinaryNameBranch,
	).WithTag("<name-and-type>"))
	g.nameAndTypeDeclaration = nameAndType

	literalOf := func(tokenKind, expressionKind SyntaxKind, tag string) ResultParser[LexicalToken, Expression] {
		return asResultParser[LexicalToken, Expression](Rule1(SyntaxTokenKindParser(tokenKind), func(token *SyntaxToken) Expression {
			return MustNewLiteralExpressionTyped(expressionKind, token)
		}).WithTag(tag))
	}
	constructorCompletion := func(parser ResultParser[LexicalToken, Expression], name string) ResultParser[LexicalToken, Expression] {
		return WithSyntaxCompletion(parser, NewCompletionItem(
			CompletionKindScalarPrefix,
			name+"()",
			WithCompletionBeforeText(name+"("),
			WithCompletionAfterText(")"),
			WithCompletionMatchText(name),
			WithCompletionRank(CompletionRankFunction),
		))
	}

	g.booleanLiteral = literalOf(SyntaxKindBooleanLiteralToken, SyntaxKindBooleanLiteralExpression, "<bool-literal>")
	g.booleanLiteralWithCompletion = FirstValue(
		IfValue(
			Or(SyntaxTokenTextParser("true"), SyntaxTokenTextParser("false")),
			g.booleanLiteral,
		),
		g.booleanLiteral,
	)
	g.longLiteral = constructorCompletion(literalOf(SyntaxKindLongLiteralToken, SyntaxKindLongLiteralExpression, "<long-literal>"), "long")
	g.realLiteral = WithSyntaxCompletion(
		literalOf(SyntaxKindRealLiteralToken, SyntaxKindRealLiteralExpression, "<real-literal>"),
		NewCompletionItem(CompletionKindScalarPrefix, "real()", WithCompletionBeforeText("real("), WithCompletionAfterText(")"), WithCompletionMatchText("real"), WithCompletionRank(CompletionRankFunction)),
		NewCompletionItem(CompletionKindScalarPrefix, "double()", WithCompletionBeforeText("double("), WithCompletionAfterText(")"), WithCompletionMatchText("double"), WithCompletionRank(CompletionRankFunction)),
	)
	g.decimalLiteral = constructorCompletion(literalOf(SyntaxKindDecimalLiteralToken, SyntaxKindDecimalLiteralExpression, "<decimal-literal>"), "decimal")
	g.intLiteral = constructorCompletion(literalOf(SyntaxKindIntLiteralToken, SyntaxKindIntLiteralExpression, "<int-literal>"), "int")
	g.guidLiteral = constructorCompletion(literalOf(SyntaxKindGuidLiteralToken, SyntaxKindGuidLiteralExpression, "<guid-literal>"), "guid")
	g.rawGuidLiteral = literalOf(SyntaxKindRawGuidLiteralToken, SyntaxKindGuidLiteralExpression, "<raw-guid-literal>")
	g.dateTimeLiteral = constructorCompletion(literalOf(SyntaxKindDateTimeLiteralToken, SyntaxKindDateTimeLiteralExpression, "<datetime-literal>"), "datetime")
	g.timespanLiteral = constructorCompletion(literalOf(SyntaxKindTimespanLiteralToken, SyntaxKindTimespanLiteralExpression, "<timespan-literal>"), "timespan")
	nullLiteral := Rule1(SyntaxTokenKindParser(SyntaxKindNullKeyword), func(token *SyntaxToken) Expression {
		return MustNewLiteralExpressionTyped(SyntaxKindNullLiteralExpression, token)
	})

	signedNumeric := signedNumericLiteralParser(g.longLiteral, g.realLiteral, g.timespanLiteral)
	clientParameter := Rule1(g.BracedName, func(name Name) Expression {
		return MustNewNameReferenceTyped(name, SymbolMatchNone)
	})

	var jsonValue ResultParser[LexicalToken, Expression]
	jsonValueForward := Forward(func() ResultParser[LexicalToken, Expression] { return jsonValue })
	jsonPair := jsonPairParser(jsonValueForward)
	jsonObject := Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenBraceToken),
		SyntaxCommaListParser(jsonPair, func(Source[LexicalToken], int) *JsonPair { return queryGrammarMissingJsonPair() }, false, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBraceToken),
		func(open *SyntaxToken, pairs *SeparatedSyntaxListOf[*JsonPair], close *SyntaxToken) Expression {
			return MustNewJsonObjectExpressionTyped(open, pairs, close)
		},
	)
	jsonArray := Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenBracketToken),
		SyntaxCommaListParser(jsonValueForward, func(Source[LexicalToken], int) Expression { return queryGrammarMissingJsonValue() }, false, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBracketToken),
		func(open *SyntaxToken, values *SeparatedSyntaxListOf[Expression], close *SyntaxToken) Expression {
			return MustNewJsonArrayExpressionTyped(open, values, close)
		},
	)
	dynamicLiteral := Rule4(
		SyntaxTokenKindParser(SyntaxKindDynamicKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		Required(FirstValue(nullLiteral, jsonValueForward), func(Source[LexicalToken], int) Expression { return queryGrammarMissingJsonValue() }),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(dynamic, open *SyntaxToken, value Expression, close *SyntaxToken) Expression {
			return MustNewDynamicExpressionTyped(dynamic, open, value, close)
		},
	)
	g.dynamicLiteral = dynamicLiteral
	jsonNumber := FirstValue(
		Rule2(
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindMinusToken).Hide()),
			FirstValue(g.longLiteral, g.realLiteral),
			func(operator *SyntaxToken, value Expression) Expression {
				return MustNewPrefixUnaryExpressionTyped(SyntaxKindUnaryMinusExpression, operator, value)
			},
		),
		g.longLiteral, g.realLiteral,
	)
	jsonValue = asResultParser[LexicalToken, Expression](FirstValue(
		stringOrCompound,
		jsonNumber,
		g.timespanLiteral,
		g.booleanLiteral,
		g.dateTimeLiteral,
		g.guidLiteral,
		g.decimalLiteral,
		dynamicLiteral,
		nullLiteral,
		jsonObject,
		jsonArray,
		clientParameter,
	).WithTag("<json-value>"))
	g.JsonValue = jsonValue

	typeOfLiteral := queryGrammarTypeOfLiteralParser(paramTypeExpression, g.ParamTypeExtended, nameAndTypeForward, starExpression)
	g.typeOfLiteral = typeOfLiteral
	rawGuidLiteralWithDiagnostic := Rule1(g.rawGuidLiteral, func(expression Expression) Expression {
		literal := expression.(*LiteralExpression)
		return MustNewLiteralExpressionTyped(SyntaxKindGuidLiteralExpression, literal.Token().Clone(true).(*SyntaxToken), DiagnosticRawGuidLiteralNotAllowed())
	})
	g.literalCore = asResultParser[LexicalToken, Expression](FirstValue(
		stringOrCompound,
		g.booleanLiteral,
		g.longLiteral,
		g.realLiteral,
		g.decimalLiteral,
		g.intLiteral,
		g.guidLiteral,
		rawGuidLiteralWithDiagnostic,
		g.dateTimeLiteral,
		g.timespanLiteral,
		signedNumeric,
		dynamicLiteral,
		typeOfLiteral,
		clientParameter,
	).WithTag("<literal>"))
	g.Literal = asResultParser[LexicalToken, Expression](Forward(func() ResultParser[LexicalToken, Expression] {
		return g.literalCore
	}).WithTag("<literal>"))
	g.LiteralList = SyntaxCommaListParser(g.Literal, func(source Source[LexicalToken], start int) Expression {
		return queryGrammarMissingExpression(source, start)
	}, false, true)

	g.SkippedTokens = asResultParser[LexicalToken, *SkippedTokens](OneOrMoreProduce(AnySyntaxTokenButEndParser(), func(tokens []*SyntaxToken) *SkippedTokens {
		elements := make([]SyntaxElement, len(tokens))
		for i, token := range tokens {
			if i == 0 {
				clone := *token
				clone.diagnostics = append(append([]Diagnostic{}, token.diagnostics...), DiagnosticIncompleteFragment())
				token = &clone
			}
			elements[i] = token
		}
		return MustNewSkippedTokensTyped(AsSyntaxListOf[*SyntaxToken](NewTypedSyntaxList("SyntaxToken", elements...)))
	}).WithTag("<skipped-tokens>"))
}

func signedNumericLiteralParser(
	longLiteral ResultParser[LexicalToken, Expression],
	realLiteral ResultParser[LexicalToken, Expression],
	timespanLiteral ResultParser[LexicalToken, Expression],
) ResultParser[LexicalToken, Expression] {
	test := MatchConsumer[LexicalToken](func(source Source[LexicalToken], start int) int {
		if source.IsEnd(start + 1) {
			return -1
		}
		sign, number := source.Peek(start), source.Peek(start+1)
		if sign.Kind != SyntaxKindPlusToken && sign.Kind != SyntaxKindMinusToken {
			return -1
		}
		if number.Trivia != "" || len(number.Text) == 0 || number.Text[0] < '0' || number.Text[0] > '9' {
			return -1
		}
		if number.Kind != SyntaxKindLongLiteralToken && number.Kind != SyntaxKindRealLiteralToken && number.Kind != SyntaxKindTimespanLiteralToken {
			return -1
		}
		return 2
	})
	return IfValue(test, Rule2(
		asResultParser[LexicalToken, *SyntaxToken](FirstValue(
			SyntaxTokenKindParser(SyntaxKindMinusToken), SyntaxTokenKindParser(SyntaxKindPlusToken),
		).Hide()),
		FirstValue(longLiteral, realLiteral, timespanLiteral),
		func(sign *SyntaxToken, expression Expression) Expression {
			literal := expression.(*LiteralExpression)
			combined := SyntaxTokenFromLexical(NewLexicalToken(literal.Token().Kind(), sign.Trivia(), sign.RawText()+literal.Token().RawText()))
			return MustNewLiteralExpressionTyped(literal.Kind(), combined)
		},
	))
}

func jsonPairParser(jsonValue ResultParser[LexicalToken, Expression]) ResultParser[LexicalToken, *JsonPair] {
	missingValue := func(Source[LexicalToken], int) Expression { return queryGrammarMissingJsonValue() }
	return asResultParser[LexicalToken, *JsonPair](FirstValue(
		Rule3(
			SyntaxTokenKindParser(SyntaxKindStringLiteralToken),
			RequiredSyntaxTokenKindParser(SyntaxKindColonToken),
			Required(jsonValue, missingValue),
			func(name, colon *SyntaxToken, value Expression) *JsonPair {
				return MustNewJsonPairTyped(name, colon, value)
			},
		),
		Rule2(
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindColonToken).Hide()),
			Required(jsonValue, missingValue),
			func(colon *SyntaxToken, value Expression) *JsonPair {
				return MustNewJsonPairTyped(NewMissingToken(SyntaxKindStringLiteralToken, DiagnosticMissingString()), colon, value)
			},
		),
		Rule1(asResultParser[LexicalToken, Expression](jsonValue.Hide()), func(value Expression) *JsonPair {
			return MustNewJsonPairTyped(NewMissingToken(SyntaxKindStringLiteralToken, DiagnosticMissingString()), CreateMissingSyntaxToken(SyntaxKindColonToken), value)
		}),
	).WithTag("<json-pair>"))
}

func queryGrammarParamTypeExtendedTokenParser() ResultParser[LexicalToken, *SyntaxToken] {
	canonical := func(kind SyntaxKind) ResultParser[LexicalToken, *SyntaxToken] {
		item, ok := CreateSyntaxCompletionItemForKind(kind, CompletionKindScalarType, CompletionPriorityNormal)
		if ok {
			return SyntaxTokenKindParser(kind, item)
		}
		return SyntaxTokenKindParser(kind)
	}
	return FirstValue(
		canonical(SyntaxKindBoolKeyword),
		HiddenSyntaxTokenKindParser(SyntaxKindBooleanKeyword),
		HiddenSyntaxTokenKindParser(SyntaxKindDateKeyword),
		canonical(SyntaxKindDateTimeKeyword),
		canonical(SyntaxKindDecimalKeyword),
		HiddenSyntaxTokenKindParser(SyntaxKindDoubleKeyword),
		canonical(SyntaxKindDynamicKeyword),
		HiddenSyntaxTokenKindParser(SyntaxKindFloatKeyword),
		canonical(SyntaxKindGuidKeyword),
		canonical(SyntaxKindIntKeyword),
		HiddenSyntaxTokenKindParser(SyntaxKindInt16Keyword),
		HiddenSyntaxTokenKindParser(SyntaxKindInt32Keyword),
		HiddenSyntaxTokenKindParser(SyntaxKindInt64Keyword),
		HiddenSyntaxTokenKindParser(SyntaxKindInt8Keyword),
		canonical(SyntaxKindLongKeyword),
		canonical(SyntaxKindRealKeyword),
		canonical(SyntaxKindStringKeyword),
		HiddenSyntaxTokenKindParser(SyntaxKindTimeKeyword),
		canonical(SyntaxKindTimespanKeyword),
		HiddenSyntaxTokenKindParser(SyntaxKindUIntKeyword),
		HiddenSyntaxTokenKindParser(SyntaxKindUInt16Keyword),
		HiddenSyntaxTokenKindParser(SyntaxKindUInt32Keyword),
		HiddenSyntaxTokenKindParser(SyntaxKindUInt64Keyword),
		HiddenSyntaxTokenKindParser(SyntaxKindUInt8Keyword),
		HiddenSyntaxTokenKindParser(SyntaxKindULongKeyword),
		HiddenSyntaxTokenKindParser(SyntaxKindUniqueIdKeyword),
	)
}

func queryGrammarParamTypeTokenParser() ResultParser[LexicalToken, *SyntaxToken] {
	return asResultParser[LexicalToken, *SyntaxToken](MatchValue(func(token LexicalToken) bool {
		return isParamTypeKind(token.Kind)
	}, func(token LexicalToken) *SyntaxToken { return SyntaxTokenFromLexical(token) }).WithTag("<param-type>"))
}

func queryGrammarInvalidParamTypeTokenParser() ResultParser[LexicalToken, *SyntaxToken] {
	return MatchValue(func(token LexicalToken) bool {
		return isExtendedParamTypeKind(token.Kind) || token.Kind == SyntaxKindIdentifierToken || IsKeywordKind(token.Kind)
	}, func(token LexicalToken) *SyntaxToken { return SyntaxTokenFromLexical(token) })
}

func queryGrammarExtendedNameDeclarationParser(g *QueryGrammar) ResultParser[LexicalToken, *NameDeclaration] {
	identifier := Rule1(SyntaxTokenKindParser(SyntaxKindIdentifierToken), func(token *SyntaxToken) *NameDeclaration {
		return MustNewNameDeclarationTyped(MustNewTokenNameTyped(token))
	})
	extendedKeyword := Rule1(MatchValue(func(token LexicalToken) bool {
		return isExtendedKeywordAsIdentifier(token.Kind)
	}, func(token LexicalToken) *SyntaxToken { return SyntaxTokenFromLexical(token) }), func(token *SyntaxToken) *NameDeclaration {
		return MustNewNameDeclarationTyped(MustNewTokenNameTyped(token))
	})
	return asResultParser[LexicalToken, *NameDeclaration](FirstValue(identifier, g.BracketedNameDeclaration, extendedKeyword).WithTag("<name>"))
}

func queryGrammarTypeOfLiteralParser(
	paramType ResultParser[LexicalToken, TypeExpression],
	paramTypeExtended ResultParser[LexicalToken, TypeExpression],
	nameAndType ResultParser[LexicalToken, *NameAndTypeDeclaration],
	starExpression ResultParser[LexicalToken, Expression],
) ResultParser[LexicalToken, Expression] {
	identifierType := Rule1(FirstValue(SyntaxTokenKindParser(SyntaxKindIdentifierToken), predefinedKeywordAsIdentifierTokenParser()), func(token *SyntaxToken) TypeExpression {
		return MustNewPrimitiveTypeExpressionTyped(token)
	})
	scalarScan := And(
		asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindTypeOfKeyword).Hide()),
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		Or(paramType, paramTypeExtended.Hide(), And(SyntaxTokenKindParser(SyntaxKindIdentifierToken).Hide(), SyntaxTokenKindParser(SyntaxKindCloseParenToken))),
	)
	scalar := Rule4(
		asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindTypeOfKeyword).Hide()),
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		FirstValue(paramType, asResultParser[LexicalToken, TypeExpression](paramTypeExtended.Hide()), asResultParser[LexicalToken, TypeExpression](identifierType.Hide())),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(keyword, open *SyntaxToken, typ TypeExpression, close *SyntaxToken) Expression {
			list := AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>", NewSeparatedElement(typ, nil)))
			return MustNewTypeOfLiteralExpressionTyped(keyword, open, list, close)
		},
	)
	typeOfElement := FirstValue(
		Rule1(nameAndType, func(value *NameAndTypeDeclaration) Expression { return value }),
		starExpression,
	)
	tabular := Rule4(
		asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindTypeOfKeyword).Hide()),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		SyntaxCommaListParser(typeOfElement, func(Source[LexicalToken], int) Expression { return queryGrammarMissingTypeExpression() }, true, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(keyword, open *SyntaxToken, types *SeparatedSyntaxListOf[Expression], close *SyntaxToken) Expression {
			return MustNewTypeOfLiteralExpressionTyped(keyword, open, types, close)
		},
	)
	return asResultParser[LexicalToken, Expression](FirstValue(IfValue(scalarScan, scalar), IfValue(
		And(SyntaxTokenKindParser(SyntaxKindTypeOfKeyword), SyntaxTokenKindParser(SyntaxKindOpenParenToken)), tabular,
	)).WithTag("<typeof-literal>"))
}

func queryGrammarMissingExpression(source Source[LexicalToken], start int) Expression {
	diagnostic := DiagnosticMissingExpression()
	if source != nil && !source.IsEnd(start) && IsKeywordKind(source.Peek(start).Kind) {
		diagnostic = DiagnosticMissingExpressionWithKeyword(source.Peek(start).Text)
	}
	return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone, diagnostic)
}

func queryGrammarMissingJsonValue() Expression {
	return MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, createOmittedToken(SyntaxKindStringLiteralToken), DiagnosticMissingJsonValue())
}

func queryGrammarMissingJsonPair() *JsonPair {
	return MustNewJsonPairTyped(
		createOmittedToken(SyntaxKindStringLiteralToken),
		createOmittedToken(SyntaxKindColonToken),
		MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, createOmittedToken(SyntaxKindStringLiteralToken)),
		DiagnosticMissingJsonPair(),
	)
}

func queryGrammarMissingNameAndTypeDeclaration() *NameAndTypeDeclaration {
	return MustNewNameAndTypeDeclarationTyped(
		MustNewNameDeclarationTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken))),
		createOmittedToken(SyntaxKindColonToken),
		MustNewPrimitiveTypeExpressionTyped(createOmittedToken(SyntaxKindStringKeyword)),
		DiagnosticMissingParameter(),
	)
}

func queryGrammarMissingTypeExpression() *PrimitiveTypeExpression {
	return MustNewPrimitiveTypeExpressionTyped(createOmittedToken(SyntaxKindIdentifierToken), DiagnosticMissingTypeName())
}
