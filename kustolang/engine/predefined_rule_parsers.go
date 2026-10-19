package kustolang

// Ported from the command-grammar helper surface in
// src/Kusto.Language/Parser/PredefinedRuleParsers.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type PredefinedRuleParsers struct {
	Value                                 ResultParser[LexicalToken, SyntaxElement]
	StringLiteral                         ResultParser[LexicalToken, SyntaxElement]
	BracketedStringLiteral                ResultParser[LexicalToken, SyntaxElement]
	RawGuidLiteral                        ResultParser[LexicalToken, SyntaxElement]
	GuidLiteral                           ResultParser[LexicalToken, SyntaxElement]
	AnyGuidLiteralOrString                ResultParser[LexicalToken, SyntaxElement]
	Type                                  ResultParser[LexicalToken, SyntaxElement]
	NameDeclaration                       ResultParser[LexicalToken, SyntaxElement]
	QualifiedNameDeclaration              ResultParser[LexicalToken, SyntaxElement]
	WildcardedNameDeclaration             ResultParser[LexicalToken, SyntaxElement]
	QualifiedWildcardedNameDeclaration    ResultParser[LexicalToken, SyntaxElement]
	ColumnNameReference                   ResultParser[LexicalToken, SyntaxElement]
	TableNameReference                    ResultParser[LexicalToken, SyntaxElement]
	ExternalTableNameReference            ResultParser[LexicalToken, SyntaxElement]
	MaterializedViewNameReference         ResultParser[LexicalToken, SyntaxElement]
	FunctionNameReference                 ResultParser[LexicalToken, SyntaxElement]
	EntityGroupNameReference              ResultParser[LexicalToken, SyntaxElement]
	DatabaseNameReference                 ResultParser[LexicalToken, SyntaxElement]
	ClusterNameReference                  ResultParser[LexicalToken, SyntaxElement]
	GraphModelNameReference               ResultParser[LexicalToken, SyntaxElement]
	GraphSnapshotNameReference            ResultParser[LexicalToken, SyntaxElement]
	GraphModelSnapshotNameReference       ResultParser[LexicalToken, SyntaxElement]
	DatabaseOrTableNameReference          ResultParser[LexicalToken, SyntaxElement]
	DatabaseOrTableOrColumnNameReference  ResultParser[LexicalToken, SyntaxElement]
	TableOrColumnNameReference            ResultParser[LexicalToken, SyntaxElement]
	DatabaseTableNameReference            ResultParser[LexicalToken, SyntaxElement]
	DatabaseTableColumnNameReference      ResultParser[LexicalToken, SyntaxElement]
	TableColumnNameReference              ResultParser[LexicalToken, SyntaxElement]
	DatabaseExternalTableNameReference    ResultParser[LexicalToken, SyntaxElement]
	DatabaseMaterializedViewNameReference ResultParser[LexicalToken, SyntaxElement]
	DatabaseFunctionNameReference         ResultParser[LexicalToken, SyntaxElement]
	DatabaseEntityGroupNameReference      ResultParser[LexicalToken, SyntaxElement]
	FunctionDeclaration                   ResultParser[LexicalToken, SyntaxElement]
	FunctionBody                          ResultParser[LexicalToken, SyntaxElement]
	QueryInput                            ResultParser[LexicalToken, SyntaxElement]
	ScriptInput                           ResultParser[LexicalToken, SyntaxElement]
	InputText                             ResultParser[LexicalToken, SyntaxElement]
	BracketedInputText                    ResultParser[LexicalToken, SyntaxElement]

	MissingStringLiteral       func(Source[LexicalToken], int) SyntaxElement
	MissingValue               func(Source[LexicalToken], int) SyntaxElement
	MissingType                func(Source[LexicalToken], int) SyntaxElement
	MissingNameReference       func(Source[LexicalToken], int) SyntaxElement
	MissingNameDeclaration     func(Source[LexicalToken], int) SyntaxElement
	MissingFunctionDeclaration func(Source[LexicalToken], int) SyntaxElement
	MissingFunctionBody        func(Source[LexicalToken], int) SyntaxElement
	MissingExpression          func(Source[LexicalToken], int) SyntaxElement
	MissingStatement           func(Source[LexicalToken], int) SyntaxElement
	MissingInputText           func(Source[LexicalToken], int) SyntaxElement
}

func NewPredefinedRuleParsers(inputs ...ResultParser[LexicalToken, SyntaxElement]) *PredefinedRuleParsers {
	return NewPredefinedRuleParsersWithGrammar(QueryGrammarFrom(DefaultGlobalState), inputs...)
}

func NewPredefinedRuleParsersWithGrammar(queryGrammar *QueryGrammar, inputs ...ResultParser[LexicalToken, SyntaxElement]) *PredefinedRuleParsers {
	if queryGrammar == nil {
		queryGrammar = QueryGrammarFrom(DefaultGlobalState)
	}
	var queryInput ResultParser[LexicalToken, SyntaxElement]
	var scriptInput ResultParser[LexicalToken, SyntaxElement]
	if len(inputs) > 0 {
		queryInput = inputs[0]
	}
	if len(inputs) > 1 {
		scriptInput = inputs[1]
	}
	if queryInput == nil {
		queryInput = neverSyntaxElementParser("<query-input>")
	}
	if scriptInput == nil {
		scriptInput = neverSyntaxElementParser("<script-input>")
	}

	rules := &PredefinedRuleParsers{}
	rules.MissingStringLiteral = func(Source[LexicalToken], int) SyntaxElement { return createMissingStringLiteral() }
	rules.MissingValue = func(Source[LexicalToken], int) SyntaxElement { return createMissingValueExpression() }
	rules.MissingType = func(Source[LexicalToken], int) SyntaxElement { return createMissingType() }
	rules.MissingNameReference = func(Source[LexicalToken], int) SyntaxElement { return createMissingNameReferenceNode() }
	rules.MissingNameDeclaration = func(Source[LexicalToken], int) SyntaxElement { return createMissingNameDeclaration() }
	rules.MissingExpression = func(Source[LexicalToken], int) SyntaxElement { return createMissingExpressionNode() }
	rules.MissingStatement = func(Source[LexicalToken], int) SyntaxElement {
		return MustNewExpressionStatementTyped(createMissingExpressionNode())
	}
	rules.MissingFunctionBody = func(Source[LexicalToken], int) SyntaxElement { return createMissingFunctionBody() }
	rules.MissingFunctionDeclaration = func(Source[LexicalToken], int) SyntaxElement {
		return MustNewFunctionDeclarationTyped(nil, createMissingFunctionParameters(), createMissingFunctionBody())
	}
	rules.MissingInputText = func(Source[LexicalToken], int) SyntaxElement {
		return SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindInputTextToken, "", ""))
	}

	rules.RawGuidLiteral = syntaxElementParser(Rule1(SyntaxTokenKindParser(SyntaxKindRawGuidLiteralToken), func(token *SyntaxToken) *LiteralExpression {
		return MustNewLiteralExpressionTyped(SyntaxKindGuidLiteralExpression, token)
	}))
	rules.GuidLiteral = syntaxElementParser(Rule1(SyntaxTokenKindParser(SyntaxKindGuidLiteralToken), func(token *SyntaxToken) *LiteralExpression {
		return MustNewLiteralExpressionTyped(SyntaxKindGuidLiteralExpression, token)
	}))
	rules.StringLiteral = syntaxElementParser(queryGrammar.StringLiteral)
	rules.BracketedStringLiteral = syntaxElementParser(Rule1(predefinedBracketedStringLiteralTokenParser(), func(token *SyntaxToken) *LiteralExpression {
		return MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, token)
	}))
	rules.AnyGuidLiteralOrString = asResultParser[LexicalToken, SyntaxElement](
		FirstValue(rules.GuidLiteral, rules.StringLiteral, rules.RawGuidLiteral),
	)
	rules.Value = asResultParser[LexicalToken, SyntaxElement](
		FirstValue(rules.GuidLiteral, rules.RawGuidLiteral, syntaxElementParser(queryGrammar.Literal)),
	)
	rules.Type = syntaxElementParser(queryGrammar.ParamTypeExtended)

	name := asResultParser[LexicalToken, Name](FirstValue(
		queryGrammar.IdentifierName,
		queryGrammar.BracketedName,
		queryGrammar.BracedName,
		Rule1(SyntaxTokenKindParser(SyntaxKindStringLiteralToken), func(token *SyntaxToken) Name { return MustNewTokenNameTyped(token) }),
		Rule1(predefinedBracketedStringLiteralTokenParser(), func(token *SyntaxToken) Name { return MustNewTokenNameTyped(token) }),
	).WithTag("<name>"))
	rules.NameDeclaration = syntaxElementParser(
		asResultParser[LexicalToken, *NameDeclaration](
			Rule1(name, func(name Name) *NameDeclaration {
				return MustNewNameDeclarationTyped(name)
			}).WithTag("<name>"),
		),
	)
	rules.QualifiedNameDeclaration = predefinedQualifiedNameParser(rules.NameDeclaration, rules.MissingNameReference, "<qualified_name>")

	rules.ColumnNameReference = predefinedNameReferenceParser(name, SymbolMatchColumn, CompletionHintColumn, "<column>")
	rules.TableNameReference = predefinedNameReferenceParser(name, SymbolMatchTable, CompletionHintTable, "<table>")
	rules.ExternalTableNameReference = predefinedNameReferenceParser(name, SymbolMatchExternalTable, CompletionHintExternalTable, "<externaltable>")
	rules.MaterializedViewNameReference = predefinedNameReferenceParser(name, SymbolMatchMaterializedView, CompletionHintMaterializedView, "<materializedview>")
	rules.EntityGroupNameReference = predefinedNameReferenceParser(name, SymbolMatchEntityGroup, CompletionHintEntityGroup, "<entitygroup>")
	rules.GraphModelNameReference = predefinedNameReferenceParser(name, SymbolMatchGraphModel, CompletionHintGraphModel, "<graphmodel>")
	rules.GraphSnapshotNameReference = predefinedNameReferenceParser(name, SymbolMatchGraphSnapshot, CompletionHintGraphSnapshot, "<graphsnapshot>")
	rules.GraphModelSnapshotNameReference = predefinedOptionalQualifiedNameParser(rules.GraphModelNameReference, rules.GraphSnapshotNameReference, rules.MissingNameReference, "<graphmodelsnapshot>")
	rules.DatabaseNameReference = predefinedNameReferenceParser(name, SymbolMatchDatabase, CompletionHintDatabase, "<database>")
	rules.ClusterNameReference = predefinedNameReferenceParser(name, SymbolMatchCluster, CompletionHintCluster, "<cluster>")
	rules.FunctionNameReference = predefinedNameReferenceParser(name, SymbolMatchFunction, CompletionHintDatabaseFunction, "<function>")
	rules.DatabaseOrTableNameReference = predefinedNameReferenceParser(name, SymbolMatchDatabase|SymbolMatchTable, CompletionHintDatabase|CompletionHintTable, "<database_or_table>")
	rules.DatabaseTableNameReference = predefinedQualifiedNameParser(rules.DatabaseOrTableNameReference, func(source Source[LexicalToken], start int) SyntaxElement {
		return rules.MissingNameReference(source, start)
	}, "<database_table>", rules.TableNameReference)
	rules.DatabaseOrTableOrColumnNameReference = predefinedNameReferenceParser(name, SymbolMatchDatabase|SymbolMatchTable|SymbolMatchColumn, CompletionHintDatabase|CompletionHintTable|CompletionHintColumn, "<database_or_table_or_column>")
	rules.DatabaseTableColumnNameReference = predefinedQualifiedNameParser(rules.DatabaseOrTableOrColumnNameReference, rules.MissingNameReference, "<database_table_column>", rules.DatabaseOrTableOrColumnNameReference)
	rules.TableOrColumnNameReference = predefinedNameReferenceParser(name, SymbolMatchTable|SymbolMatchColumn, CompletionHintTable|CompletionHintColumn, "<table_or_column>")
	rules.TableColumnNameReference = predefinedQualifiedNameParser(rules.TableOrColumnNameReference, rules.MissingNameReference, "<table_column>", rules.TableOrColumnNameReference)
	databaseOrExternalTableNameReference := predefinedNameReferenceParser(name, SymbolMatchDatabase|SymbolMatchExternalTable, CompletionHintDatabase|CompletionHintExternalTable, "<database_or_externaltable>")
	rules.DatabaseExternalTableNameReference = predefinedQualifiedNameParser(databaseOrExternalTableNameReference, rules.MissingNameReference, "<database_externaltable>", rules.ExternalTableNameReference)
	databaseOrMaterializedViewNameReference := predefinedNameReferenceParser(name, SymbolMatchDatabase|SymbolMatchMaterializedView, CompletionHintDatabase|CompletionHintMaterializedView, "<database_or_materializedview>")
	rules.DatabaseMaterializedViewNameReference = predefinedQualifiedNameParser(databaseOrMaterializedViewNameReference, rules.MissingNameReference, "<database_materializedview>", rules.MaterializedViewNameReference)
	databaseOrFunctionNameReference := predefinedNameReferenceParser(name, SymbolMatchDatabase|SymbolMatchFunction, CompletionHintDatabase|CompletionHintDatabaseFunction, "<database_or_function>")
	rules.DatabaseFunctionNameReference = predefinedQualifiedNameParser(databaseOrFunctionNameReference, rules.MissingNameReference, "<database_function>", rules.FunctionNameReference)
	databaseOrEntityGroupNameReference := predefinedNameReferenceParser(name, SymbolMatchDatabase|SymbolMatchEntityGroup, CompletionHintDatabase|CompletionHintEntityGroup, "<database_or_entitygroup>")
	rules.DatabaseEntityGroupNameReference = predefinedQualifiedNameParser(databaseOrEntityGroupNameReference, rules.MissingNameReference, "<database_entitygroup>", rules.EntityGroupNameReference)

	rules.WildcardedNameDeclaration = syntaxElementParser(Rule1(queryGrammar.WildcardedIdentifier, func(id *SyntaxToken) *NameDeclaration {
		return MustNewNameDeclarationTyped(MustNewWildcardedNameTyped(id))
	}))
	wildcardedOrNameDeclaration := asResultParser[LexicalToken, SyntaxElement](
		FirstValue(rules.WildcardedNameDeclaration, rules.NameDeclaration),
	)
	rules.QualifiedWildcardedNameDeclaration = asResultParser[LexicalToken, SyntaxElement](
		FirstValue(
			IfValue(
				And(rules.NameDeclaration, SyntaxTokenKindParser(SyntaxKindDotToken), wildcardedOrNameDeclaration),
				Rule3(
					rules.NameDeclaration,
					SyntaxTokenKindParser(SyntaxKindDotToken),
					wildcardedOrNameDeclaration,
					func(qual SyntaxElement, dot *SyntaxToken, name SyntaxElement) SyntaxElement {
						return MustNewPathExpressionTyped(qual.(Expression), dot, name.(Expression))
					},
				),
			),
			wildcardedOrNameDeclaration,
		),
	)

	functionBody := ForwardValue(queryGrammar.FunctionBody)
	rules.FunctionBody = syntaxElementParser(functionBody)
	rules.FunctionDeclaration = syntaxElementParser(Rule2(
		queryGrammar.FunctionParameters,
		functionBody,
		func(parameters *FunctionParameters, body *FunctionBody) *FunctionDeclaration {
			return MustNewFunctionDeclarationTyped(nil, parameters, body)
		},
	))
	rules.QueryInput = queryInput
	rules.ScriptInput = scriptInput
	rules.InputText = syntaxElementParser(predefinedInputTextParser())
	rules.BracketedInputText = syntaxElementParser(predefinedBracketedInputTextParser())

	return rules
}

func predefinedStringLiteralParser() ResultParser[LexicalToken, *LiteralExpression] {
	return Rule1(SyntaxTokenKindParser(SyntaxKindStringLiteralToken), func(token *SyntaxToken) *LiteralExpression {
		return MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, token)
	})
}

func predefinedLiteralElementParser() ResultParser[LexicalToken, SyntaxElement] {
	return MatchConsumerValue(
		func(source Source[LexicalToken], start int) int {
			parser := newQueryParserState(lexicalTokensFromSource(source, start))
			if expr := parser.parseLiteral(); expr != nil {
				return parser.pos
			}
			return -1
		},
		func(source Source[LexicalToken], start int, _ int) SyntaxElement {
			return newQueryParserState(lexicalTokensFromSource(source, start)).parseLiteral()
		},
	)
}

func predefinedNameParser() ResultParser[LexicalToken, Name] {
	identifierName := Rule1(predefinedIdentifierTokenParser(), func(token *SyntaxToken) Name {
		return MustNewTokenNameTyped(token)
	})
	stringName := Rule1(SyntaxTokenKindParser(SyntaxKindStringLiteralToken), func(token *SyntaxToken) Name {
		return MustNewTokenNameTyped(token)
	})
	bracketedStringLiteralName := Rule1(predefinedBracketedStringLiteralTokenParser(), func(token *SyntaxToken) Name {
		return MustNewTokenNameTyped(token)
	})
	return asResultParser[LexicalToken, Name](
		FirstValue(
			identifierName,
			predefinedBracketedNameParser(),
			predefinedBracedNameParser(),
			stringName,
			bracketedStringLiteralName,
		).WithTag("<name>"),
	)
}

func predefinedIdentifierTokenParser() ResultParser[LexicalToken, *SyntaxToken] {
	return asResultParser[LexicalToken, *SyntaxToken](
		MatchValue(func(token LexicalToken) bool {
			return token.Kind == SyntaxKindIdentifierToken || (IsKeywordKind(token.Kind) && SyntaxKindCanBeIdentifier(token.Kind))
		}, func(token LexicalToken) *SyntaxToken {
			return SyntaxTokenFromLexical(token)
		}).WithTag("<identifier>"),
	)
}

func predefinedBracketedNameParser() ResultParser[LexicalToken, Name] {
	return Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenBracketToken),
		predefinedStringLiteralParser(),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBracketToken),
		func(open *SyntaxToken, name *LiteralExpression, close *SyntaxToken) Name {
			return MustNewBracketedNameTyped(open, name, close)
		},
	)
}

func predefinedBracedNameParser() ResultParser[LexicalToken, Name] {
	return Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenBraceToken),
		predefinedIdentifierTokenParser(),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBraceToken),
		func(open *SyntaxToken, name *SyntaxToken, close *SyntaxToken) Name {
			return MustNewBracedNameTyped(open, name, close)
		},
	)
}

func predefinedNameReferenceParser(nameParser ResultParser[LexicalToken, Name], match SymbolMatch, hint CompletionHint, tag string) ResultParser[LexicalToken, SyntaxElement] {
	parser := Rule1(nameParser, func(name Name) *NameReference {
		return MustNewNameReferenceTyped(name, match)
	})
	parser = asResultParser[LexicalToken, *NameReference](WithSyntaxCompletionHint(parser, hint))
	return syntaxElementParser(asResultParser[LexicalToken, *NameReference](parser.WithTag(tag)))
}

func predefinedQualifiedNameParser(
	leftParser ResultParser[LexicalToken, SyntaxElement],
	missingSelector func(Source[LexicalToken], int) SyntaxElement,
	tag string,
	selectorParser ...ResultParser[LexicalToken, SyntaxElement],
) ResultParser[LexicalToken, SyntaxElement] {
	right := leftParser
	if len(selectorParser) > 0 && selectorParser[0] != nil {
		right = selectorParser[0]
	}
	parser := ApplyZeroOrMore(leftParser, func(LeftValue[SyntaxElement]) RightParser[LexicalToken, SyntaxElement] {
		return RuleRight2(
			asResultParser[LexicalToken, *SyntaxToken](HiddenSyntaxTokenKindParser(SyntaxKindDotToken)),
			Required(right, missingSelector),
			func(left SyntaxElement, dot *SyntaxToken, selector SyntaxElement) SyntaxElement {
				return MustNewPathExpressionTyped(left.(Expression), dot, selector.(Expression))
			},
		)
	})
	return asResultParser[LexicalToken, SyntaxElement](parser.WithTag(tag))
}

func predefinedOptionalQualifiedNameParser(
	leftParser ResultParser[LexicalToken, SyntaxElement],
	rightParser ResultParser[LexicalToken, SyntaxElement],
	missingSelector func(Source[LexicalToken], int) SyntaxElement,
	tag string,
) ResultParser[LexicalToken, SyntaxElement] {
	parser := ApplyOptional(leftParser, func(LeftValue[SyntaxElement]) RightParser[LexicalToken, SyntaxElement] {
		return RuleRight2(
			asResultParser[LexicalToken, *SyntaxToken](HiddenSyntaxTokenKindParser(SyntaxKindDotToken)),
			Required(rightParser, missingSelector),
			func(left SyntaxElement, dot *SyntaxToken, selector SyntaxElement) SyntaxElement {
				return MustNewPathExpressionTyped(left.(Expression), dot, selector.(Expression))
			},
		)
	})
	return asResultParser[LexicalToken, SyntaxElement](parser.WithTag(tag))
}

func predefinedWildcardedIdentifierParser() ResultParser[LexicalToken, *SyntaxToken] {
	return queryStateSyntaxParser("<wildcarded_identifier>", func(p *queryParserState) *SyntaxToken {
		return p.parseWildcardedIdentifier()
	})
}

func predefinedBracketedStringLiteralTokenParser() ResultParser[LexicalToken, *SyntaxToken] {
	return asResultParser[LexicalToken, *SyntaxToken](
		MatchConsumerValue(predefinedBracketedStringLiteralLength, func(source Source[LexicalToken], start int, length int) *SyntaxToken {
			return predefinedCombinedToken(source, start, length, SyntaxKindStringLiteralToken)
		}).WithTag("<bracketed-string>"),
	)
}

func predefinedBracketedStringLiteralLength(source Source[LexicalToken], start int) int {
	if source.IsEnd(start) || source.Peek(start).Text != "[" {
		return -1
	}
	length := 1
	for !source.IsEnd(start + length) {
		token := source.Peek(start + length)
		if token.Kind == SyntaxKindEndOfTextToken {
			return length
		}
		if token.Text == "]" {
			return length + 1
		}
		if token.Text == "[" || HasLineBreaks(token.Trivia) || HasLineBreaks(token.Text) {
			return length
		}
		length++
	}
	return length
}

func predefinedInputTextParser() ResultParser[LexicalToken, *SyntaxToken] {
	return Convert(ZeroOrMore(AnySyntaxTokenButEndParser()), func(source Source[LexicalToken], start int, length int) *SyntaxToken {
		return predefinedCombinedToken(source, start, length, SyntaxKindInputTextToken)
	})
}

func predefinedBracketedInputTextParser() ResultParser[LexicalToken, *SyntaxToken] {
	end := First(HiddenSyntaxTokenTextParser("]"), EndOfTextSyntaxParser())
	return Convert(ZeroOrMore(Not(end)), func(source Source[LexicalToken], start int, length int) *SyntaxToken {
		return predefinedCombinedToken(source, start, length, SyntaxKindInputTextToken)
	})
}

func predefinedCombinedToken(source Source[LexicalToken], start int, length int, kind SyntaxKind) *SyntaxToken {
	if length <= 0 {
		return SyntaxTokenFromLexical(NewLexicalToken(kind, "", ""))
	}
	first := source.Peek(start)
	text := first.Text
	for i := 1; i < length; i++ {
		token := source.Peek(start + i)
		text += token.Trivia + token.Text
	}
	return SyntaxTokenFromLexical(NewLexicalToken(kind, first.Trivia, text))
}

func queryStateSyntaxParser[T SyntaxElement](tag string, parse func(*queryParserState) T) ResultParser[LexicalToken, T] {
	return asResultParser[LexicalToken, T](
		MatchConsumerValue(
			func(source Source[LexicalToken], start int) int {
				parser := newQueryParserState(lexicalTokensFromSource(source, start))
				value := parse(parser)
				if isNilSyntaxElement(value) {
					return -1
				}
				return parser.pos
			},
			func(source Source[LexicalToken], start int, _ int) T {
				return parse(newQueryParserState(lexicalTokensFromSource(source, start)))
			},
		).WithTag(tag),
	)
}

func lexicalTokensFromSource(source Source[LexicalToken], start int) []LexicalToken {
	tokens := []LexicalToken{}
	for offset := 0; !source.IsEnd(start + offset); offset++ {
		token := source.Peek(start + offset)
		tokens = append(tokens, token)
		if token.Kind == SyntaxKindEndOfTextToken {
			break
		}
	}
	return tokens
}

func syntaxElementParser[T SyntaxElement](parser ResultParser[LexicalToken, T]) ResultParser[LexicalToken, SyntaxElement] {
	converted := Rule1(parser, func(value T) SyntaxElement {
		return value
	})
	return asResultParser[LexicalToken, SyntaxElement](converted.WithAnnotations(parser.Annotations()))
}

func neverSyntaxElementParser(tag string) ResultParser[LexicalToken, SyntaxElement] {
	return asResultParser[LexicalToken, SyntaxElement](
		MatchConsumerValue(
			func(Source[LexicalToken], int) int { return -1 },
			func(Source[LexicalToken], int, int) SyntaxElement { return nil },
		).WithTag(tag),
	)
}
