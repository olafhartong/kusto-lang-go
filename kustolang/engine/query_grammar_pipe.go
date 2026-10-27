package kustolang

// Ported from the common query-operator and pipe-expression productions in
// src/Kusto.Language/Parser/QueryGrammar.cs. The operator alternatives grow
// incrementally as each native combinator production is ported.

func (g *QueryGrammar) initializePipeExpressionParsers() {
	missingExpression := func(source Source[LexicalToken], start int) Expression {
		return queryGrammarMissingExpression(source, start)
	}
	var mvApplyOperatorCore ResultParser[LexicalToken, QueryOperator]
	mvApplyOperator := asResultParser[LexicalToken, QueryOperator](Forward(func() ResultParser[LexicalToken, QueryOperator] {
		return mvApplyOperatorCore
	}).WithTag("<mvapply>"))
	g.mvApplyOperator = mvApplyOperator
	var parseOperatorCore ResultParser[LexicalToken, QueryOperator]
	parseOperator := asResultParser[LexicalToken, QueryOperator](Forward(func() ResultParser[LexicalToken, QueryOperator] {
		return parseOperatorCore
	}).WithTag("<parse>"))
	g.parseOperator = parseOperator
	var parseWhereOperatorCore ResultParser[LexicalToken, QueryOperator]
	parseWhereOperator := asResultParser[LexicalToken, QueryOperator](Forward(func() ResultParser[LexicalToken, QueryOperator] {
		return parseWhereOperatorCore
	}).WithTag("<parse-where>"))
	g.parseWhereOperator = parseWhereOperator
	var parseKvOperatorCore ResultParser[LexicalToken, QueryOperator]
	parseKvOperator := asResultParser[LexicalToken, QueryOperator](Forward(func() ResultParser[LexicalToken, QueryOperator] {
		return parseKvOperatorCore
	}).WithTag("<parse-kv>"))
	g.parseKvOperator = parseKvOperator
	var partitionOperatorCore ResultParser[LexicalToken, QueryOperator]
	partitionOperator := asResultParser[LexicalToken, QueryOperator](Forward(func() ResultParser[LexicalToken, QueryOperator] {
		return partitionOperatorCore
	}).WithTag("<partition>"))
	g.partitionOperator = partitionOperator
	var partitionByOperatorCore ResultParser[LexicalToken, QueryOperator]
	partitionByOperator := Forward(func() ResultParser[LexicalToken, QueryOperator] {
		return partitionByOperatorCore
	})
	g.partitionByOperator = partitionByOperator
	var reduceByOperatorCore ResultParser[LexicalToken, QueryOperator]
	reduceByOperator := asResultParser[LexicalToken, QueryOperator](Forward(func() ResultParser[LexicalToken, QueryOperator] {
		return reduceByOperatorCore
	}).WithTag("<reduce-by>"))
	g.reduceByOperator = reduceByOperator
	var renderOperatorCore ResultParser[LexicalToken, QueryOperator]
	renderOperator := asResultParser[LexicalToken, QueryOperator](Forward(func() ResultParser[LexicalToken, QueryOperator] {
		return renderOperatorCore
	}).WithTag("<render>"))
	g.renderOperator = renderOperator
	var scanOperatorCore ResultParser[LexicalToken, QueryOperator]
	scanOperator := Forward(func() ResultParser[LexicalToken, QueryOperator] {
		return scanOperatorCore
	})
	g.scanOperator = scanOperator
	summarizeBinScan := And(
		HiddenSyntaxTokenKindParser(SyntaxKindBinKeyword),
		HiddenSyntaxTokenKindParser(SyntaxKindEqualToken),
	)
	summarizeBinClause := IfValue(summarizeBinScan, Rule3(
		SyntaxTokenKindParser(SyntaxKindBinKeyword),
		SyntaxTokenKindParser(SyntaxKindEqualToken),
		Required(g.UnnamedExpression, missingExpression),
		func(bin, equal *SyntaxToken, value Expression) *SimpleNamedExpression {
			return MustNewSimpleNamedExpressionTyped(MustNewNameDeclarationTyped(MustNewTokenNameTyped(bin)), equal, value)
		},
	))
	summarizeByExpression := IfValue(
		Not(And(
			HiddenSyntaxTokenKindParser(SyntaxKindBinKeyword),
			HiddenSyntaxTokenKindParser(SyntaxKindEqualToken),
			Match(func(token LexicalToken) bool { return SyntaxCategoryOf(token.Kind) == SyntaxCategoryLiteral }),
		)),
		g.NamedExpression,
	)
	summarizeByClause := asResultParser[LexicalToken, *SummarizeByClause](Rule3(
		SyntaxTokenKindParser(SyntaxKindByKeyword),
		SyntaxCommaListParser(summarizeByExpression, missingExpression, true, false),
		OptionalValue(asResultParser[LexicalToken, *SimpleNamedExpression](summarizeBinClause.Hide())),
		func(by *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], binClause *SimpleNamedExpression) *SummarizeByClause {
			return MustNewSummarizeByClauseTyped(by, expressions, binClause)
		},
	).WithTag("<summarize-by>"))
	summarizeExpression := IfValue(
		Not(HiddenSyntaxTokenKindParser(SyntaxKindByKeyword)),
		g.NamedExpression,
	)
	summarizeOperator := asResultParser[LexicalToken, *SummarizeOperator](Rule4(
		SyntaxTokenKindParser(SyntaxKindSummarizeKeyword, NewCompletionItem(CompletionKindQueryPrefix, "summarize", WithCompletionPriority(CompletionPriorityHigh))),
		queryGrammarParameterList(g, QueryOperatorParameters.SummarizeParameters, true),
		SyntaxCommaListParser(summarizeExpression, missingExpression, false, false, SyntaxKindByKeyword),
		OptionalValue(summarizeByClause),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], aggregates *SeparatedSyntaxListOf[Expression], byClause *SummarizeByClause) *SummarizeOperator {
			return MustNewSummarizeOperatorTyped(keyword, parameters, aggregates, byClause)
		},
	).WithTag("<summarize>"))
	g.summarizeOperator = summarizeOperator
	typeOfLiteral := Rule1(g.typeOfLiteral, func(expression Expression) *TypeOfLiteralExpression {
		return expression.(*TypeOfLiteralExpression)
	})
	toTypeOfClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindToKeyword),
		Required(typeOfLiteral, func(Source[LexicalToken], int) *TypeOfLiteralExpression { return createMissingTypeOfLiteral() }),
		func(to *SyntaxToken, typeOf *TypeOfLiteralExpression) *ToTypeOfClause {
			return MustNewToTypeOfClauseTyped(to, typeOf)
		},
	)
	mvExpandExpression := FirstValue(
		IfValue(
			HiddenSyntaxTokenKindParser(SyntaxKindToKeyword),
			Rule2(
				Required(g.NamedExpression, missingExpression),
				toTypeOfClause,
				func(expression Expression, typeOf *ToTypeOfClause) *MvExpandExpression {
					return MustNewMvExpandExpressionTyped(expression, typeOf)
				},
			),
		),
		Rule2(
			g.NamedExpression,
			OptionalValue(toTypeOfClause),
			func(expression Expression, typeOf *ToTypeOfClause) *MvExpandExpression {
				return MustNewMvExpandExpressionTyped(expression, typeOf)
			},
		),
	)
	missingMvExpandExpression := func(source Source[LexicalToken], start int) *MvExpandExpression {
		return MustNewMvExpandExpressionTyped(queryGrammarMissingExpression(source, start), nil)
	}
	mvExpandListExpression := IfValue(
		Not(HiddenSyntaxTokenKindParser(SyntaxKindLimitKeyword)),
		mvExpandExpression,
	)
	mvExpandExpressions := SyntaxCommaListParser(
		mvExpandListExpression,
		missingMvExpandExpression,
		true,
		false,
		SyntaxKindLimitKeyword,
	)
	mvExpandRowLimitClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindLimitKeyword),
		Required(g.UnnamedExpression, missingExpression),
		func(limit *SyntaxToken, expression Expression) *MvExpandRowLimitClause {
			return MustNewMvExpandRowLimitClauseTyped(limit, expression)
		},
	)
	mvExpandRowLimitAfterCompleteList := IfValue(
		Not(MatchConsumer[LexicalToken](func(source Source[LexicalToken], start int) int {
			if start > 0 && source.Peek(start-1).Kind == SyntaxKindCommaToken {
				return 0
			}
			return -1
		})),
		mvExpandRowLimitClause,
	)
	mvExpandOperator := asResultParser[LexicalToken, QueryOperator](Rule4(
		FirstValue(
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(
				SyntaxKindMvExpandKeyword,
				NewCompletionItem(CompletionKindQueryPrefix, "mvexpand"),
			).Hide()),
			SyntaxTokenKindParser(
				SyntaxKindMvDashExpandKeyword,
				NewCompletionItem(CompletionKindQueryPrefix, "mv-expand"),
			),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.MvExpandParameters, true),
		mvExpandExpressions,
		OptionalValue(mvExpandRowLimitAfterCompleteList),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expressions *SeparatedSyntaxListOf[*MvExpandExpression], limit *MvExpandRowLimitClause) QueryOperator {
			return MustNewMvExpandOperatorTyped(keyword, parameters, expressions, limit)
		},
	).WithTag("<mvexpand>"))
	unionOperator := asResultParser[LexicalToken, QueryOperator](Rule3(
		SyntaxTokenKindParser(
			SyntaxKindUnionKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "union"),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.UnionParameters, true),
		SyntaxCommaListParser(g.unionExpression, missingExpression, true, false),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expressions *SeparatedSyntaxListOf[Expression]) QueryOperator {
			return MustNewUnionOperatorTyped(keyword, parameters, expressions)
		},
	).WithTag("<union>"))
	invokeOperator := asResultParser[LexicalToken, QueryOperator](Rule2(
		SyntaxTokenKindParser(
			SyntaxKindInvokeKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "invoke", WithCompletionPriority(CompletionPriorityLow)),
		),
		Required(g.dotCompositeFunctionCall, missingExpression),
		func(keyword *SyntaxToken, function Expression) QueryOperator {
			return MustNewInvokeOperatorTyped(keyword, function)
		},
	).WithTag("<invoke>"))
	evaluateRowSchema := Rule6(
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		OptionalValue(SyntaxTokenKindParser(SyntaxKindCommaToken)),
		OptionalValue(SyntaxTokenKindParser(SyntaxKindAsteriskToken)),
		OptionalValue(SyntaxTokenKindParser(SyntaxKindCommaToken)),
		SyntaxCommaListParser(g.nameAndTypeDeclaration, func(Source[LexicalToken], int) *NameAndTypeDeclaration {
			return queryGrammarMissingNameAndTypeDeclaration()
		}, false, true),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open, leadingComma, asterisk, asteriskComma *SyntaxToken, columns *SeparatedSyntaxListOf[*NameAndTypeDeclaration], close *SyntaxToken) *EvaluateRowSchema {
			return MustNewEvaluateRowSchemaTyped(open, leadingComma, asterisk, asteriskComma, columns, close)
		},
	)
	missingEvaluateRowSchema := func(Source[LexicalToken], int) *EvaluateRowSchema {
		return MustNewEvaluateRowSchemaTyped(
			createOmittedToken(SyntaxKindOpenParenToken),
			nil,
			nil,
			nil,
			AsSeparatedSyntaxListOf[*NameAndTypeDeclaration](NewTypedSyntaxList("SeparatedElement<NameAndTypeDeclaration>")),
			createOmittedToken(SyntaxKindCloseParenToken),
			DiagnosticMissingSchemaDeclaration(),
		)
	}
	evaluateSchemaClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindColonToken),
		Required(evaluateRowSchema, missingEvaluateRowSchema),
		func(colon *SyntaxToken, schema *EvaluateRowSchema) *EvaluateSchemaClause {
			return MustNewEvaluateSchemaClauseTyped(colon, schema)
		},
	)
	evaluateOperator := asResultParser[LexicalToken, QueryOperator](Rule4(
		SyntaxTokenKindParser(
			SyntaxKindEvaluateKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "evaluate", WithCompletionPriority(CompletionPriorityLow)),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.EvaluateParameters, false),
		g.requiredFunctionCall,
		OptionalValue(evaluateSchemaClause),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], function *FunctionCallExpression, schema *EvaluateSchemaClause) QueryOperator {
			return MustNewEvaluateOperatorTyped(keyword, parameters, function, schema)
		},
	).WithTag("<evaluate>"))
	joinEqualityExpression := ApplyOptional(g.functionCallOrPath, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
		return RuleRight2(
			SyntaxTokenKindParser(
				SyntaxKindEqualEqualToken,
				NewCompletionItem(CompletionKindScalarInfix, "=="),
			),
			Required(g.functionCallOrPath, missingExpression),
			func(left Expression, equal *SyntaxToken, right Expression) Expression {
				return MustNewBinaryExpressionTyped(SyntaxKindEqualExpression, left, equal, right)
			},
		)
	})
	joinAndExpression := ApplyZeroOrMore(joinEqualityExpression, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
		return RuleRight2(
			SyntaxTokenKindParser(
				SyntaxKindAndKeyword,
				NewCompletionItem(CompletionKindScalarInfix, "and"),
			),
			Required(joinEqualityExpression, missingExpression),
			func(left Expression, and *SyntaxToken, right Expression) Expression {
				return MustNewBinaryExpressionTyped(SyntaxKindAndExpression, left, and, right)
			},
		)
	})
	joinOnExpression := BestValue(
		joinAndExpression,
		asResultParser[LexicalToken, Expression](g.UnnamedExpression.Hide()),
	)
	joinOnClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindOnKeyword),
		SyntaxCommaListParser(joinOnExpression, missingExpression, true, false),
		func(on *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression]) *JoinOnClause {
			return MustNewJoinOnClauseTyped(on, expressions)
		},
	)
	joinWhereClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindWhereKeyword),
		Required(g.UnnamedExpression, missingExpression),
		func(where *SyntaxToken, expression Expression) *JoinWhereClause {
			return MustNewJoinWhereClauseTyped(where, expression)
		},
	)
	joinOnCondition := Rule1(joinOnClause, func(clause *JoinOnClause) JoinConditionClause { return clause })
	joinWhereCondition := asResultParser[LexicalToken, JoinConditionClause](Rule1(
		joinWhereClause,
		func(clause *JoinWhereClause) JoinConditionClause { return clause },
	).Hide())
	joinCondition := FirstValue(joinOnCondition, joinWhereCondition)
	joinOperator := asResultParser[LexicalToken, QueryOperator](Rule4(
		SyntaxTokenKindParser(
			SyntaxKindJoinKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "join", WithCompletionPriority(CompletionPriorityHigh)),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.JoinParameters, true),
		Required(g.UnnamedExpression, missingExpression),
		OptionalValue(joinCondition),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, condition JoinConditionClause) QueryOperator {
			return MustNewJoinOperatorTyped(keyword, parameters, expression, condition)
		},
	).WithTag("<join>"))
	lookupOperator := asResultParser[LexicalToken, QueryOperator](Rule4(
		SyntaxTokenKindParser(
			SyntaxKindLookupKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "lookup", WithCompletionPriority(CompletionPriorityHigh)),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.LookupParameters, true),
		Required(g.UnnamedExpression, missingExpression),
		Required(joinOnCondition, func(Source[LexicalToken], int) JoinConditionClause { return createMissingJoinOnClause() }),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, condition JoinConditionClause) QueryOperator {
			return MustNewLookupOperatorTyped(keyword, parameters, expression, condition)
		},
	).WithTag("<lookup>"))
	makeGraphInvocationExpression := FirstValue(
		Rule2(
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindMinusToken).Hide()),
			g.functionCallOrPath,
			func(operator *SyntaxToken, expression Expression) Expression {
				return MustNewPrefixUnaryExpressionTyped(SyntaxKindUnaryMinusExpression, operator, expression)
			},
		),
		Rule2(
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindPlusToken).Hide()),
			g.functionCallOrPath,
			func(operator *SyntaxToken, expression Expression) Expression {
				return MustNewPrefixUnaryExpressionTyped(SyntaxKindUnaryPlusExpression, operator, expression)
			},
		),
		g.functionCallOrPath,
	)
	makeGraphTableAndKeyClause := asResultParser[LexicalToken, *MakeGraphTableAndKeyClause](Rule3(
		makeGraphInvocationExpression,
		RequiredSyntaxTokenKindParser(SyntaxKindOnKeyword),
		Required(g.SimpleNameReference, func(Source[LexicalToken], int) Expression { return createMissingNameReferenceNode() }),
		func(table Expression, on *SyntaxToken, column Expression) *MakeGraphTableAndKeyClause {
			return MustNewMakeGraphTableAndKeyClauseTyped(table, on, column.(*NameReference))
		},
	).WithTag("<table-and-key-clause>"))
	missingMakeGraphTableAndKeyClause := func(Source[LexicalToken], int) *MakeGraphTableAndKeyClause {
		missingReference := func() *NameReference {
			return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone)
		}
		return MustNewMakeGraphTableAndKeyClauseTyped(
			missingReference(),
			createOmittedToken(SyntaxKindOnKeyword),
			missingReference(),
			DiagnosticMissingExpression(),
		)
	}
	makeGraphWithTablesAndKeysClause := asResultParser[LexicalToken, MakeGraphWithClause](Rule2(
		SyntaxTokenKindParser(SyntaxKindWithKeyword),
		SyntaxCommaListParser(
			makeGraphTableAndKeyClause,
			missingMakeGraphTableAndKeyClause,
			true,
			false,
			SyntaxKindPartitionedByKeyword,
		),
		func(with *SyntaxToken, tablesAndKeys *SeparatedSyntaxListOf[*MakeGraphTableAndKeyClause]) MakeGraphWithClause {
			return MustNewMakeGraphWithTablesAndKeysClauseTyped(with, tablesAndKeys)
		},
	).WithTag("<make-graph-with-tables-and-keys-clause>"))
	makeGraphWithImplicitIDClause := asResultParser[LexicalToken, MakeGraphWithClause](Rule3(
		SyntaxTokenKindParser(
			SyntaxKindWithNodeIdKeyword,
			NewCompletionItem(CompletionKindKeyword, "with_node_id", WithCompletionBeforeText("with_node_id=")),
		),
		RequiredSyntaxTokenKindParser(SyntaxKindEqualToken),
		Required(g.SimpleNameDeclaration, func(Source[LexicalToken], int) *NameDeclaration { return createMissingNameDeclaration() }),
		func(withNodeID, equal *SyntaxToken, name *NameDeclaration) MakeGraphWithClause {
			return MustNewMakeGraphWithImplicitIdClauseTyped(withNodeID, equal, name)
		},
	).WithTag("<make-graph-with-implicit-node-id-clause>"))
	makeGraphPartitionedByClause := asResultParser[LexicalToken, *MakeGraphPartitionedByClause](Rule5(
		SyntaxTokenKindParser(SyntaxKindPartitionedByKeyword),
		Required(g.SimpleNameReference, func(Source[LexicalToken], int) Expression { return createMissingNameReferenceNode() }),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		Required(Forward(func() ResultParser[LexicalToken, Expression] { return g.contextualSubExpression }), missingExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(keyword *SyntaxToken, entityExpression Expression, open *SyntaxToken, subquery Expression, close *SyntaxToken) *MakeGraphPartitionedByClause {
			entity := entityExpression.(*NameReference)
			return MustNewMakeGraphPartitionedByClauseTyped(keyword, entity, open, subquery, close)
		},
	).WithTag("<make-graph-partitioned-by-clause>"))
	makeGraphOperator := asResultParser[LexicalToken, QueryOperator](Rule7(
		SyntaxTokenKindParser(
			SyntaxKindMakeGraphKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "make-graph"),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.GraphMakeParameters, true),
		Required(g.SimpleNameReference, func(Source[LexicalToken], int) Expression { return createMissingNameReferenceNode() }),
		Required(
			FirstValue(
				SyntaxTokenTextParser("-->", SyntaxKindDashDashGreaterThanToken),
				asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenTextParser("--", SyntaxKindDashDashToken).Hide()),
			),
			func(Source[LexicalToken], int) *SyntaxToken {
				return CreateMissingSyntaxTokenForKinds([]SyntaxKind{SyntaxKindDashDashGreaterThanToken, SyntaxKindDashDashToken})
			},
		),
		Required(g.SimpleNameReference, func(Source[LexicalToken], int) Expression { return createMissingNameReferenceNode() }),
		OptionalValue(
			FirstValue(makeGraphWithImplicitIDClause, makeGraphWithTablesAndKeysClause),
			func() MakeGraphWithClause { return (*MakeGraphWithImplicitIdClause)(nil) },
		),
		OptionalValue(makeGraphPartitionedByClause),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], source Expression, direction *SyntaxToken, target Expression, with MakeGraphWithClause, partitionedBy *MakeGraphPartitionedByClause) QueryOperator {
			if isNilSyntaxElement(with) {
				with = nil
			}
			return MustNewMakeGraphOperatorTyped(keyword, parameters, source.(*NameReference), direction, target.(*NameReference), with, partitionedBy)
		},
	).WithTag("<make-graph>"))
	g.makeGraphOperator = makeGraphOperator
	graphMarkComponentsOperator := asResultParser[LexicalToken, QueryOperator](Rule2(
		SyntaxTokenKindParser(
			SyntaxKindGraphMarkComponentsKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "graph-mark-components"),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.GraphMarkComponentsParameters, true),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter]) QueryOperator {
			return MustNewGraphMarkComponentsOperatorTyped(keyword, parameters)
		},
	).WithTag("<graph-mark-components>"))
	graphWhereNodesOperator := asResultParser[LexicalToken, QueryOperator](Rule2(
		SyntaxTokenKindParser(
			SyntaxKindGraphWhereNodesKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "graph-where-nodes"),
		),
		Required(Forward(func() ResultParser[LexicalToken, Expression] { return g.Expression }), missingExpression),
		func(keyword *SyntaxToken, expression Expression) QueryOperator {
			return MustNewGraphWhereNodesOperatorTyped(keyword, expression)
		},
	).WithTag("<graph-where-nodes>"))
	graphWhereEdgesOperator := asResultParser[LexicalToken, QueryOperator](Rule2(
		SyntaxTokenKindParser(
			SyntaxKindGraphWhereEdgesKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "graph-where-edges"),
		),
		Required(Forward(func() ResultParser[LexicalToken, Expression] { return g.Expression }), missingExpression),
		func(keyword *SyntaxToken, expression Expression) QueryOperator {
			return MustNewGraphWhereEdgesOperatorTyped(keyword, expression)
		},
	).WithTag("<graph-where-edges>"))
	graphToTableAsClause := asResultParser[LexicalToken, *GraphToTableAsClause](Rule2(
		SyntaxTokenKindParser(
			SyntaxKindAsKeyword,
			NewCompletionItem(CompletionKindKeyword, "as", WithCompletionPriority(CompletionPriorityLow)),
		),
		Required(g.SimpleNameDeclaration, func(Source[LexicalToken], int) *NameDeclaration { return createMissingNameDeclaration() }),
		func(keyword *SyntaxToken, name *NameDeclaration) *GraphToTableAsClause {
			return MustNewGraphToTableAsClauseTyped(keyword, name)
		},
	).WithTag("<graph-to-table-as-clause>"))
	graphToTableNodesClause := asResultParser[LexicalToken, *GraphToTableOutputClause](Rule3(
		SyntaxTokenKindParser(
			SyntaxKindNodesKeyword,
			NewCompletionItem(CompletionKindKeyword, "nodes"),
		),
		OptionalValue(graphToTableAsClause),
		queryGrammarParameterList(g, QueryOperatorParameters.GraphToTableNodesParameters, true),
		func(keyword *SyntaxToken, asClause *GraphToTableAsClause, parameters *SyntaxListOf[*NamedParameter]) *GraphToTableOutputClause {
			return MustNewGraphToTableOutputClauseTyped(keyword, asClause, parameters)
		},
	).WithTag("graph-to-table-output-nodes-clause"))
	graphToTableEdgesClause := asResultParser[LexicalToken, *GraphToTableOutputClause](Rule3(
		SyntaxTokenKindParser(
			SyntaxKindGraphEdgesKeyword,
			NewCompletionItem(CompletionKindKeyword, "edges"),
		),
		OptionalValue(graphToTableAsClause),
		queryGrammarParameterList(g, QueryOperatorParameters.GraphToTableEdgesParameters, true),
		func(keyword *SyntaxToken, asClause *GraphToTableAsClause, parameters *SyntaxListOf[*NamedParameter]) *GraphToTableOutputClause {
			return MustNewGraphToTableOutputClauseTyped(keyword, asClause, parameters)
		},
	).WithTag("graph-to-table-output-edges-clause"))
	graphToTableOutputClause := asResultParser[LexicalToken, *GraphToTableOutputClause](FirstValue(
		graphToTableNodesClause,
		graphToTableEdgesClause,
	).WithTag("<graph-to-table-output-clause>"))
	missingGraphToTableOutputClause := func(Source[LexicalToken], int) *GraphToTableOutputClause {
		return MustNewGraphToTableOutputClauseTyped(
			createOmittedToken(SyntaxKindOpenParenToken),
			MustNewGraphToTableAsClauseTyped(
				createOmittedToken(SyntaxKindOpenParenToken),
				createMissingNameDeclaration(),
			),
			nil,
		)
	}
	graphToTableOperator := asResultParser[LexicalToken, QueryOperator](Rule2(
		SyntaxTokenKindParser(
			SyntaxKindGraphToTableKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "graph-to-table"),
		),
		SyntaxCommaListParser(graphToTableOutputClause, missingGraphToTableOutputClause, true, false),
		func(keyword *SyntaxToken, outputClause *SeparatedSyntaxListOf[*GraphToTableOutputClause]) QueryOperator {
			return MustNewGraphToTableOperatorTyped(keyword, outputClause)
		},
	).WithTag("<graph-to-table>"))
	whereClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindWhereKeyword),
		Required(Forward(func() ResultParser[LexicalToken, Expression] { return g.Expression }), missingExpression),
		func(keyword *SyntaxToken, expression Expression) *WhereClause {
			return MustNewWhereClauseTyped(keyword, expression)
		},
	)
	projectClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindProjectKeyword),
		SyntaxCommaListParser(g.NamedExpression, missingExpression, true, false),
		func(keyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression]) *ProjectClause {
			return MustNewProjectClauseTyped(keyword, expressions)
		},
	)
	invocationExpression := FirstValue(
		Rule2(
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindMinusToken).Hide()),
			g.functionCallOrPath,
			func(operator *SyntaxToken, expression Expression) Expression {
				return MustNewPrefixUnaryExpressionTyped(SyntaxKindUnaryMinusExpression, operator, expression)
			},
		),
		Rule2(
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindPlusToken).Hide()),
			g.functionCallOrPath,
			func(operator *SyntaxToken, expression Expression) Expression {
				return MustNewPrefixUnaryExpressionTyped(SyntaxKindUnaryPlusExpression, operator, expression)
			},
		),
		g.functionCallOrPath,
	)
	graphMatchPatternEdgeRange := Rule4(
		SyntaxTokenKindParser(SyntaxKindAsteriskToken),
		Required(invocationExpression, missingExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindDotDotToken),
		Required(invocationExpression, missingExpression),
		func(asterisk *SyntaxToken, rangeStart Expression, dotDot *SyntaxToken, rangeEnd Expression) *GraphMatchPatternEdgeRange {
			return MustNewGraphMatchPatternEdgeRangeTyped(asterisk, rangeStart, dotDot, rangeEnd)
		},
	)
	graphMatchPatternEdge := FirstValue(
		Rule1(
			FirstValue(
				SyntaxTokenTextParser("-->", SyntaxKindDashDashGreaterThanToken),
				SyntaxTokenTextParser("<--", SyntaxKindLessThanDashDashToken),
				SyntaxTokenTextParser("--", SyntaxKindDashDashToken),
			),
			func(token *SyntaxToken) GraphMatchPatternNotation {
				return MustNewGraphMatchPatternEdgeTyped(token, nil, nil, nil)
			},
		),
		Rule4(
			FirstValue(
				SyntaxTokenTextParser("<-[", SyntaxKindLessThanDashBracketToken),
				SyntaxTokenTextParser("-[", SyntaxKindDashBracketToken),
			),
			OptionalValue(g.SimpleNameDeclaration),
			OptionalValue(graphMatchPatternEdgeRange),
			Required(
				FirstValue(
					SyntaxTokenTextParser("]->", SyntaxKindBracketDashGreaterThanToken),
					SyntaxTokenTextParser("]-", SyntaxKindBracketDashToken),
				),
				func(Source[LexicalToken], int) *SyntaxToken {
					return CreateMissingSyntaxTokenForKinds([]SyntaxKind{SyntaxKindBracketDashGreaterThanToken, SyntaxKindBracketDashToken})
				},
			),
			func(first *SyntaxToken, name *NameDeclaration, edgeRange *GraphMatchPatternEdgeRange, last *SyntaxToken) GraphMatchPatternNotation {
				return MustNewGraphMatchPatternEdgeTyped(first, name, edgeRange, last)
			},
		),
	)
	graphMatchPatternNode := Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		OptionalValue(g.SimpleNameDeclaration),
		SyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open *SyntaxToken, name *NameDeclaration, close *SyntaxToken) GraphMatchPatternNotation {
			return MustNewGraphMatchPatternNodeTyped(open, name, close)
		},
	)
	missingGraphMatchPatternNotation := func(Source[LexicalToken], int) GraphMatchPatternNotation {
		return MustNewGraphMatchPatternNodeTyped(
			createOmittedToken(SyntaxKindOpenParenToken),
			MustNewNameDeclarationTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken))),
			createOmittedToken(SyntaxKindCloseParenToken),
			DiagnosticMissingGraphMatchPatternElement(),
		)
	}
	graphMatchPatternNotation := FirstValue(graphMatchPatternNode, graphMatchPatternEdge)
	graphMatchPatternStart := First(
		HiddenSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		asResultParser[LexicalToken, *SyntaxToken](FirstValue(
			SyntaxTokenTextParser("-->", SyntaxKindDashDashGreaterThanToken),
			SyntaxTokenTextParser("<--", SyntaxKindLessThanDashDashToken),
			SyntaxTokenTextParser("--", SyntaxKindDashDashToken),
			SyntaxTokenTextParser("<-[", SyntaxKindLessThanDashBracketToken),
			SyntaxTokenTextParser("-[", SyntaxKindDashBracketToken),
		).Hide()),
	)
	graphMatchPattern := WithSyntaxCompletion(
		IfValue(
			graphMatchPatternStart,
			Rule1(
				SyntaxListParser(graphMatchPatternNotation, missingGraphMatchPatternNotation, true),
				func(elements *SyntaxListOf[GraphMatchPatternNotation]) *GraphMatchPattern {
					return MustNewGraphMatchPatternTyped(elements)
				},
			),
		),
		NewCompletionItem(CompletionKindSyntax, "(n1)-[e]->(n2)"),
		NewCompletionItem(CompletionKindSyntax, "(n1)-[e1]->(n2)-[e2]->(n3)"),
		NewCompletionItem(CompletionKindSyntax, "(n1)-[e*1..3]->(n2)"),
	)
	missingGraphMatchPattern := func(source Source[LexicalToken], start int) *GraphMatchPattern {
		elements := NewTypedSyntaxListWithDiagnostics(
			"GraphMatchPatternNotation",
			[]SyntaxElement{missingGraphMatchPatternNotation(source, start)},
			DiagnosticMissingGraphMatchPattern(),
		)
		return MustNewGraphMatchPatternTyped(AsSyntaxListOf[GraphMatchPatternNotation](elements))
	}
	graphMatchPatternClause := asResultParser[LexicalToken, *SeparatedSyntaxListOf[*GraphMatchPattern]](SyntaxCommaListParser(
		graphMatchPattern,
		missingGraphMatchPattern,
		true,
		false,
		SyntaxKindWhereKeyword,
		SyntaxKindProjectKeyword,
	).WithTag("<graph-match-pattern-clause>"))
	graphMatchOperator := asResultParser[LexicalToken, QueryOperator](Rule5(
		SyntaxTokenKindParser(
			SyntaxKindGraphMatchKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "graph-match"),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.GraphMatchParameters, true),
		graphMatchPatternClause,
		OptionalValue(whereClause),
		OptionalValue(projectClause),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], patterns *SeparatedSyntaxListOf[*GraphMatchPattern], where *WhereClause, project *ProjectClause) QueryOperator {
			return MustNewGraphMatchOperatorTyped(keyword, parameters, patterns, where, project)
		},
	).WithTag("<graph-match-operator>"))
	graphShortestPathsOperator := asResultParser[LexicalToken, QueryOperator](Rule5(
		SyntaxTokenKindParser(
			SyntaxKindGraphShortestPathsKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "graph-shortest-paths"),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.GraphShortestPathsParameters, true),
		graphMatchPatternClause,
		OptionalValue(whereClause),
		OptionalValue(projectClause),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], patterns *SeparatedSyntaxListOf[*GraphMatchPattern], where *WhereClause, project *ProjectClause) QueryOperator {
			return MustNewGraphShortestPathsOperatorTyped(keyword, parameters, patterns, where, project)
		},
	).WithTag("<graph-shortest-paths-operator>"))
	g.graphMarkComponentsOperator = graphMarkComponentsOperator
	g.graphWhereNodesOperator = graphWhereNodesOperator
	g.graphWhereEdgesOperator = graphWhereEdgesOperator
	g.graphToTableAsClause = graphToTableAsClause
	g.graphToTableNodesClause = graphToTableNodesClause
	g.graphToTableEdgesClause = graphToTableEdgesClause
	g.graphToTableOutputClause = graphToTableOutputClause
	g.graphToTableOperator = graphToTableOperator
	g.whereClause = whereClause
	g.projectClause = projectClause
	g.graphMatchPatternEdgeRange = graphMatchPatternEdgeRange
	g.graphMatchPatternEdge = graphMatchPatternEdge
	g.graphMatchPatternNode = graphMatchPatternNode
	g.graphMatchPattern = graphMatchPattern
	g.graphMatchPatternClause = graphMatchPatternClause
	g.graphMatchOperator = graphMatchOperator
	g.graphShortestPathsOperator = graphShortestPathsOperator
	makeSeriesOnClause := Rule2(
		RequiredSyntaxTokenKindParser(SyntaxKindOnKeyword),
		Required(
			IfValue(
				Not(First(
					SyntaxTokenKindParser(SyntaxKindFromKeyword),
					SyntaxTokenKindParser(SyntaxKindToKeyword),
					SyntaxTokenKindParser(SyntaxKindStepKeyword),
					SyntaxTokenKindParser(SyntaxKindInKeyword),
				)),
				g.NamedExpression,
			),
			missingExpression,
		),
		func(keyword *SyntaxToken, expression Expression) *MakeSeriesOnClause {
			return MustNewMakeSeriesOnClauseTyped(keyword, expression)
		},
	)
	makeSeriesInRangeClause := Rule5(
		asResultParser[LexicalToken, *SyntaxToken](RequiredSyntaxTokenKindParser(SyntaxKindInKeyword).Hide()),
		asResultParser[LexicalToken, *SyntaxToken](RequiredSyntaxTokenKindParser(SyntaxKindRangeKeyword).Hide()),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		SyntaxCommaListParser(g.NamedExpression, missingExpression, false, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(inKeyword, rangeKeyword, open *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], close *SyntaxToken) MakeSeriesRangeClause {
			return MustNewMakeSeriesInRangeClauseTyped(inKeyword, rangeKeyword, MustNewExpressionListTyped(open, expressions, close))
		},
	)
	makeSeriesFromClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindFromKeyword),
		Required(g.UnnamedExpression, missingExpression),
		func(keyword *SyntaxToken, expression Expression) *MakeSeriesFromClause {
			return MustNewMakeSeriesFromClauseTyped(keyword, expression)
		},
	)
	makeSeriesToClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindToKeyword),
		Required(g.UnnamedExpression, missingExpression),
		func(keyword *SyntaxToken, expression Expression) *MakeSeriesToClause {
			return MustNewMakeSeriesToClauseTyped(keyword, expression)
		},
	)
	makeSeriesStepClause := Rule2(
		RequiredSyntaxTokenKindParser(SyntaxKindStepKeyword),
		Required(g.UnnamedExpression, missingExpression),
		func(keyword *SyntaxToken, expression Expression) *MakeSeriesStepClause {
			return MustNewMakeSeriesStepClauseTyped(keyword, expression)
		},
	)
	makeSeriesFromToStepClause := IfValue(
		FirstValue(
			SyntaxTokenKindParser(SyntaxKindFromKeyword),
			SyntaxTokenKindParser(SyntaxKindToKeyword),
			SyntaxTokenKindParser(SyntaxKindStepKeyword),
		),
		Rule3(
			OptionalValue(makeSeriesFromClause),
			OptionalValue(makeSeriesToClause),
			makeSeriesStepClause,
			func(from *MakeSeriesFromClause, to *MakeSeriesToClause, step *MakeSeriesStepClause) MakeSeriesRangeClause {
				return MustNewMakeSeriesFromToStepClauseTyped(from, to, step)
			},
		),
	)
	makeSeriesByClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindByKeyword),
		SyntaxCommaListParser(g.NamedExpression, missingExpression, true, false),
		func(keyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression]) *MakeSeriesByClause {
			return MustNewMakeSeriesByClauseTyped(keyword, expressions)
		},
	)
	defaultExpressionClause := Rule3(
		SyntaxTokenKindParser(SyntaxKindDefaultKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindEqualToken),
		Required(
			IfValue(
				Not(First(
					SyntaxTokenKindParser(SyntaxKindOnKeyword),
					SyntaxTokenKindParser(SyntaxKindFromKeyword),
					SyntaxTokenKindParser(SyntaxKindInKeyword),
					SyntaxTokenKindParser(SyntaxKindByKeyword),
				)),
				g.NamedExpression,
			),
			missingExpression,
		),
		func(keyword, equal *SyntaxToken, expression Expression) *DefaultExpressionClause {
			return MustNewDefaultExpressionClauseTyped(keyword, equal, expression)
		},
	)
	makeSeriesExpression := Rule2(
		g.NamedExpression,
		OptionalValue(defaultExpressionClause),
		func(expression Expression, defaultExpression *DefaultExpressionClause) *MakeSeriesExpression {
			return MustNewMakeSeriesExpressionTyped(expression, defaultExpression)
		},
	)
	makeSeriesExpressionBeforeClause := IfValue(
		Not(First(
			SyntaxTokenKindParser(SyntaxKindOnKeyword),
			SyntaxTokenKindParser(SyntaxKindFromKeyword),
			SyntaxTokenKindParser(SyntaxKindInKeyword),
		)),
		makeSeriesExpression,
	)
	missingMakeSeriesExpression := func(source Source[LexicalToken], start int) *MakeSeriesExpression {
		return MustNewMakeSeriesExpressionTyped(missingExpression(source, start), nil)
	}
	makeSeriesOperator := asResultParser[LexicalToken, QueryOperator](Rule6(
		SyntaxTokenKindParser(
			SyntaxKindMakeSeriesKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "make-series"),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.MakeSeriesParameters, true),
		SyntaxCommaListParser(
			makeSeriesExpressionBeforeClause,
			missingMakeSeriesExpression,
			true,
			false,
			SyntaxKindOnKeyword,
			SyntaxKindFromKeyword,
			SyntaxKindInKeyword,
		),
		makeSeriesOnClause,
		FirstValue(makeSeriesFromToStepClause, makeSeriesInRangeClause),
		OptionalValue(makeSeriesByClause),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], aggregates *SeparatedSyntaxListOf[*MakeSeriesExpression], on *MakeSeriesOnClause, rangeClause MakeSeriesRangeClause, by *MakeSeriesByClause) QueryOperator {
			return MustNewMakeSeriesOperatorTyped(keyword, parameters, aggregates, on, rangeClause, by)
		},
	).WithTag("<make-series>"))
	g.makeSeriesOperator = makeSeriesOperator
	printOperator := Rule2(
		SyntaxTokenKindParser(SyntaxKindPrintKeyword),
		SyntaxCommaListParser(g.NamedExpression, missingExpression, true, false),
		func(keyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression]) QueryOperator {
			return MustNewPrintOperatorTyped(keyword, expressions)
		},
	)
	rangeOperator := IfValue(
		Sequence(SyntaxTokenKindParser(SyntaxKindRangeKeyword), Fails(SyntaxTokenKindParser(SyntaxKindOpenParenToken))),
		Rule8(
			SyntaxTokenKindParser(SyntaxKindRangeKeyword),
			Required(g.SimpleNameDeclaration, func(Source[LexicalToken], int) *NameDeclaration { return createMissingNameDeclaration() }),
			RequiredSyntaxTokenKindParser(SyntaxKindFromKeyword),
			Required(g.UnnamedExpression, missingExpression),
			RequiredSyntaxTokenKindParser(SyntaxKindToKeyword),
			Required(g.UnnamedExpression, missingExpression),
			RequiredSyntaxTokenKindParser(SyntaxKindStepKeyword),
			Required(g.UnnamedExpression, missingExpression),
			func(keyword *SyntaxToken, name *NameDeclaration, fromKeyword *SyntaxToken, from Expression, toKeyword *SyntaxToken, to Expression, stepKeyword *SyntaxToken, step Expression) QueryOperator {
				return MustNewRangeOperatorTyped(keyword, name, fromKeyword, from, toKeyword, to, stepKeyword, step)
			},
		),
	)

	countAs := Rule2(
		asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindAsKeyword).Hide()),
		RequiredSyntaxTokenKindParser(SyntaxKindIdentifierToken),
		func(as, identifier *SyntaxToken) *CountAsIdentifierClause {
			return MustNewCountAsIdentifierClauseTyped(as, identifier)
		},
	)
	countOperator := Rule2(
		SyntaxTokenKindParser(SyntaxKindCountKeyword, NewCompletionItem(CompletionKindQueryPrefix, "count")),
		OptionalValue(countAs),
		func(keyword *SyntaxToken, as *CountAsIdentifierClause) QueryOperator {
			return MustNewCountOperatorTyped(keyword, as)
		},
	)
	extendOperator := Rule2(
		SyntaxTokenKindParser(SyntaxKindExtendKeyword, NewCompletionItem(CompletionKindQueryPrefix, "extend")),
		SyntaxCommaListParser(g.NamedExpression, missingExpression, true, false),
		func(keyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression]) QueryOperator {
			return MustNewExtendOperatorTyped(keyword, expressions)
		},
	)
	filterOperator := Rule3(
		FirstValue(
			SyntaxTokenKindParser(SyntaxKindWhereKeyword, NewCompletionItem(CompletionKindQueryPrefix, "where")),
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindFilterKeyword).Hide()),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.FilterParameters, true),
		Required(g.NamedExpression, missingExpression),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], condition Expression) QueryOperator {
			return MustNewFilterOperatorTyped(keyword, parameters, condition)
		},
	)
	projectOperator := Rule2(
		SyntaxTokenKindParser(SyntaxKindProjectKeyword, NewCompletionItem(CompletionKindQueryPrefix, "project")),
		SyntaxCommaListParser(g.NamedExpression, missingExpression, false, false),
		func(keyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression]) QueryOperator {
			return MustNewProjectOperatorTyped(keyword, expressions)
		},
	)
	distinctElement := FirstValue(
		Rule1(asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindAsteriskToken).Hide()), func(token *SyntaxToken) Expression {
			return MustNewStarExpressionTyped(token)
		}),
		g.NamedExpression,
	)
	distinctParameters := queryGrammarParameterList(g, QueryOperatorParameters.DistinctParameters, true)
	distinctOperator := Rule3(
		SyntaxTokenKindParser(SyntaxKindDistinctKeyword),
		distinctParameters,
		SyntaxCommaListParser(distinctElement, missingExpression, true, false),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expressions *SeparatedSyntaxListOf[Expression]) QueryOperator {
			return MustNewDistinctOperatorTyped(keyword, parameters, expressions)
		},
	)
	takeOperator := Rule3(
		FirstValue(SyntaxTokenKindParser(SyntaxKindLimitKeyword), SyntaxTokenKindParser(SyntaxKindTakeKeyword)),
		queryGrammarParameterList(g, QueryOperatorParameters.TakeParameters, true),
		Required(g.NamedExpression, missingExpression),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression) QueryOperator {
			return MustNewTakeOperatorTyped(keyword, parameters, expression)
		},
	)
	executeAndCacheOperator := Rule1(SyntaxTokenKindParser(SyntaxKindExecuteAndCacheKeyword), func(keyword *SyntaxToken) QueryOperator {
		return MustNewExecuteAndCacheOperatorTyped(keyword)
	})
	getSchemaKind := queryGrammarParameter(g, QueryOperatorParameters.GetSchemaKind, false)
	getSchemaOperator := Rule2(
		SyntaxTokenKindParser(SyntaxKindGetSchemaKeyword),
		OptionalValue(getSchemaKind),
		func(keyword *SyntaxToken, kind *NamedParameter) QueryOperator {
			return MustNewGetSchemaOperatorTyped(keyword, kind)
		},
	)
	serializeOperator := Rule3(
		SyntaxTokenKindParser(SyntaxKindSerializeKeyword),
		queryGrammarParameterList(g, QueryOperatorParameters.SerializedParameters, true),
		SyntaxCommaListParser(g.NamedExpression, missingExpression, false, false),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expressions *SeparatedSyntaxListOf[Expression]) QueryOperator {
			return MustNewSerializeOperatorTyped(keyword, parameters, expressions)
		},
	)
	sampleOperator := Rule3(
		SyntaxTokenKindParser(SyntaxKindSampleKeyword),
		queryGrammarParameterList(g, QueryOperatorParameters.SampleParameters, true),
		Required(g.NamedExpression, missingExpression),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression) QueryOperator {
			return MustNewSampleOperatorTyped(keyword, parameters, expression)
		},
	)
	sampleDistinctOperator := Rule5(
		SyntaxTokenKindParser(SyntaxKindSampleDistinctKeyword),
		queryGrammarParameterList(g, QueryOperatorParameters.SampleDistinctParameters, true),
		Required(g.NamedExpression, missingExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindOfKeyword),
		Required(g.NamedExpression, missingExpression),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, of *SyntaxToken, ofExpression Expression) QueryOperator {
			return MustNewSampleDistinctOperatorTyped(keyword, parameters, expression, of, ofExpression)
		},
	)
	simpleOrWildcarded := FirstValue(g.WildcardedNameReference, g.SimpleNameReference)
	projectAwayOperator := Rule2(
		SyntaxTokenKindParser(SyntaxKindProjectAwayKeyword),
		SyntaxCommaListParser(simpleOrWildcarded, missingExpression, false, false),
		func(keyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression]) QueryOperator {
			return MustNewProjectAwayOperatorTyped(keyword, expressions)
		},
	)
	projectByNamesOperator := Rule2(
		asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindProjectByNamesKeyword).Hide()),
		SyntaxCommaListParser(g.UnnamedExpression, missingExpression, false, false),
		func(keyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression]) QueryOperator {
			return MustNewProjectByNamesOperatorTyped(keyword, expressions)
		},
	)
	projectKeepOperator := Rule2(
		SyntaxTokenKindParser(SyntaxKindProjectKeepKeyword),
		SyntaxCommaListParser(simpleOrWildcarded, missingExpression, true, false),
		func(keyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression]) QueryOperator {
			return MustNewProjectKeepOperatorTyped(keyword, expressions)
		},
	)
	projectRenameOperator := Rule2(
		SyntaxTokenKindParser(SyntaxKindProjectRenameKeyword),
		SyntaxCommaListParser(g.NamedExpression, missingExpression, false, false),
		func(keyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression]) QueryOperator {
			return MustNewProjectRenameOperatorTyped(keyword, expressions)
		},
	)
	reorderOrdering := Rule1(SyntaxTokenKindsParser([]SyntaxKind{
		SyntaxKindAscKeyword, SyntaxKindDescKeyword, SyntaxKindGrannyAscKeyword, SyntaxKindGrannyDescKeyword,
	}), func(token *SyntaxToken) *OrderingClause {
		return MustNewOrderingClauseTyped(token, nil)
	})
	reorderExpression := ApplyOptional(simpleOrWildcarded, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
		return RuleRight1(reorderOrdering, func(left Expression, ordering *OrderingClause) Expression {
			return MustNewOrderedExpressionTyped(left, ordering)
		})
	})
	projectReorderOperator := Rule2(
		SyntaxTokenKindParser(SyntaxKindProjectReorderKeyword),
		SyntaxCommaListParser(reorderExpression, missingExpression, false, false),
		func(keyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression]) QueryOperator {
			return MustNewProjectReorderOperatorTyped(keyword, expressions)
		},
	)

	missingFirstOrLast := func(Source[LexicalToken], int) *SyntaxToken {
		return NewMissingToken(SyntaxKindFirstKeyword, DiagnosticMissingFirstOrLast())
	}
	orderingNullsClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindNullsKeyword),
		Required(
			SyntaxTokenKindsParser([]SyntaxKind{SyntaxKindFirstKeyword, SyntaxKindLastKeyword}),
			missingFirstOrLast,
		),
		func(nulls, firstOrLast *SyntaxToken) *OrderingNullsClause {
			return MustNewOrderingNullsClauseTyped(nulls, firstOrLast)
		},
	)
	orderingClause := Rule2(
		OptionalValue(SyntaxTokenKindsParser([]SyntaxKind{SyntaxKindAscKeyword, SyntaxKindDescKeyword})),
		OptionalValue(orderingNullsClause),
		func(ascOrDesc *SyntaxToken, nulls *OrderingNullsClause) *OrderingClause {
			return MustNewOrderingClauseTyped(ascOrDesc, nulls)
		},
	)
	orderedExpression := ApplyOptional(g.NamedExpression, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
		return RuleRight1(
			IfValue(
				FirstValue(
					SyntaxTokenKindParser(SyntaxKindAscKeyword),
					SyntaxTokenKindParser(SyntaxKindDescKeyword),
					SyntaxTokenKindParser(SyntaxKindNullsKeyword),
				),
				orderingClause,
			),
			func(left Expression, ordering *OrderingClause) Expression {
				return MustNewOrderedExpressionTyped(left, ordering)
			},
		)
	})
	sortParameters := queryGrammarParameterList(g, QueryOperatorParameters.SortParameters, false)
	sortOperator := Rule4(
		FirstValue(SyntaxTokenKindParser(SyntaxKindOrderKeyword), SyntaxTokenKindParser(SyntaxKindSortKeyword)),
		sortParameters,
		RequiredSyntaxTokenKindParser(SyntaxKindByKeyword),
		SyntaxCommaListParser(orderedExpression, missingExpression, true, false),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], by *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression]) QueryOperator {
			return MustNewSortOperatorTyped(keyword, parameters, by, expressions)
		},
	)
	topParameters := queryGrammarParameterList(g, QueryOperatorParameters.TopParameters, true)
	topOperator := Rule5(
		SyntaxTokenKindParser(SyntaxKindTopKeyword),
		topParameters,
		Required(g.NamedExpression, missingExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindByKeyword),
		Required(orderedExpression, missingExpression),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, by *SyntaxToken, byExpression Expression) QueryOperator {
			return MustNewTopOperatorTyped(keyword, parameters, expression, by, byExpression)
		},
	)
	topHittersByClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindByKeyword),
		Required(g.NamedExpression, missingExpression),
		func(by *SyntaxToken, expression Expression) *TopHittersByClause {
			return MustNewTopHittersByClauseTyped(by, expression)
		},
	)
	topHittersOperator := Rule5(
		SyntaxTokenKindParser(SyntaxKindTopHittersKeyword),
		Required(g.NamedExpression, missingExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindOfKeyword),
		Required(g.NamedExpression, missingExpression),
		OptionalValue(topHittersByClause),
		func(keyword *SyntaxToken, expression Expression, of *SyntaxToken, ofExpression Expression, by *TopHittersByClause) QueryOperator {
			return MustNewTopHittersOperatorTyped(keyword, expression, of, ofExpression, by)
		},
	)
	topNestedWithOthersClause := Rule4(
		SyntaxTokenKindParser(SyntaxKindWithKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindOthersKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindEqualToken),
		g.Literal,
		func(with, others, equal *SyntaxToken, expression Expression) *TopNestedWithOthersClause {
			return MustNewTopNestedWithOthersClauseTyped(with, others, equal, expression)
		},
	)
	type optionalTopNestedExpression struct {
		expression Expression
	}
	optionalTopNestedCount := OptionalValue(
		Rule1(g.NamedExpression, func(expression Expression) optionalTopNestedExpression {
			return optionalTopNestedExpression{expression: expression}
		}),
		func() optionalTopNestedExpression { return optionalTopNestedExpression{} },
	)
	topNestedClause := Rule7(
		SyntaxTokenKindParser(SyntaxKindTopNestedKeyword),
		optionalTopNestedCount,
		RequiredSyntaxTokenKindParser(SyntaxKindOfKeyword),
		Required(g.NamedExpression, missingExpression),
		OptionalValue(topNestedWithOthersClause),
		RequiredSyntaxTokenKindParser(SyntaxKindByKeyword),
		Required(orderedExpression, missingExpression),
		func(keyword *SyntaxToken, expression optionalTopNestedExpression, of *SyntaxToken, ofExpression Expression, withOthers *TopNestedWithOthersClause, by *SyntaxToken, byExpression Expression) *TopNestedClause {
			return MustNewTopNestedClauseTyped(keyword, expression.expression, of, ofExpression, withOthers, by, byExpression)
		},
	)
	missingTopNestedClause := func(Source[LexicalToken], int) *TopNestedClause {
		missingName := func() Expression {
			return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone)
		}
		return MustNewTopNestedClauseTyped(
			createOmittedToken(SyntaxKindTopNestedKeyword),
			nil,
			createOmittedToken(SyntaxKindOfKeyword),
			missingName(),
			nil,
			createOmittedToken(SyntaxKindByKeyword),
			missingName(),
			DiagnosticMissingClause(),
		)
	}
	topNestedOperator := asResultParser[LexicalToken, QueryOperator](IfValue(
		SyntaxTokenKindParser(SyntaxKindTopNestedKeyword),
		Rule1(
			SyntaxCommaListParser(topNestedClause, missingTopNestedClause, true, false),
			func(clauses *SeparatedSyntaxListOf[*TopNestedClause]) QueryOperator {
				return MustNewTopNestedOperatorTyped(clauses)
			},
		),
	).WithTag("<top-nested>"))
	consumeOperator := asResultParser[LexicalToken, QueryOperator](Rule2(
		HiddenSyntaxTokenKindParser(SyntaxKindConsumeKeyword),
		queryGrammarParameterList(g, QueryOperatorParameters.ConsumeParameters, false),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter]) QueryOperator {
			return MustNewConsumeOperatorTyped(keyword, parameters)
		},
	).WithTag("<consume>"))
	asOperator := asResultParser[LexicalToken, QueryOperator](Rule3(
		SyntaxTokenKindParser(SyntaxKindAsKeyword),
		queryGrammarParameterList(g, QueryOperatorParameters.AsParameters, true),
		Required(g.SimpleNameDeclaration, queryGrammarMissingNameDeclaration),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], name *NameDeclaration) QueryOperator {
			return MustNewAsOperatorTyped(keyword, parameters, name)
		},
	).WithTag("<as>"))
	dataScopeClause := func(kind CompletionKind) ResultParser[LexicalToken, *DataScopeClause] {
		return Rule3(
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(
				SyntaxKindDataScopeKeyword,
				NewCompletionItem(kind, "datascope"),
			).Hide()),
			RequiredSyntaxTokenKindParser(SyntaxKindEqualToken),
			RequiredSyntaxTokenTextsParser(KustoFacts.DataScopeValues),
			func(keyword, equal, value *SyntaxToken) *DataScopeClause {
				return MustNewDataScopeClauseTyped(keyword, equal, value)
			},
		)
	}
	findOperandName := ApplyOptional(
		FirstValue(g.bracketedEntityNameSelector, g.barePathElementSelector),
		func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
			return RuleRight2(
				SyntaxTokenKindParser(SyntaxKindBarToken),
				Required(asOperator, func(Source[LexicalToken], int) QueryOperator {
					return MustNewBadQueryOperatorTyped(createOmittedToken(SyntaxKindIdentifierToken), DiagnosticQueryOperatorExpected())
				}),
				func(left Expression, bar *SyntaxToken, operator QueryOperator) Expression {
					return MustNewPipeExpressionTyped(left, bar, operator)
				},
			)
		},
	)
	findOperand := BestValue(g.wildcardedEntityReference, findOperandName)
	findInClause := Rule4(
		SyntaxTokenKindParser(SyntaxKindInKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		SyntaxCommaListParser(findOperand, missingExpression, true, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(keyword, open *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], close *SyntaxToken) *FindInClause {
			return MustNewFindInClauseTyped(keyword, open, expressions, close)
		},
	)
	columnNameReference := asResultParser[LexicalToken, Expression](g.SimpleNameReference.WithTag("<column>"))
	typedColumnNameReference := ApplyOptional(columnNameReference, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
		return RuleRight2(
			SyntaxTokenKindParser(SyntaxKindColonToken),
			Required(
				FirstValue(
					g.ParamTypeExtended,
					Rule1(queryGrammarInvalidParamTypeTokenParser(), func(token *SyntaxToken) TypeExpression {
						return MustNewPrimitiveTypeExpressionTyped(token, DiagnosticInvalidTypeName(token.RawText()))
					}),
				),
				func(Source[LexicalToken], int) TypeExpression { return createMissingType() },
			),
			func(left Expression, colon *SyntaxToken, typ TypeExpression) Expression {
				return MustNewTypedColumnReferenceTyped(left.(*NameReference), colon, typ)
			},
		)
	})
	findProjectColumn := FirstValue(
		Rule4(
			SyntaxTokenKindParser(SyntaxKindPackKeyword),
			RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
			RequiredSyntaxTokenKindParser(SyntaxKindAsteriskToken, NewCompletionItem(CompletionKindSyntax, "*")),
			RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
			func(keyword, open, star, close *SyntaxToken) Expression {
				return MustNewPackExpressionTyped(keyword, open, star, close)
			},
		),
		Rule1(SyntaxTokenKindParser(SyntaxKindAsteriskToken), func(token *SyntaxToken) Expression {
			return MustNewStarExpressionTyped(token)
		}),
		typedColumnNameReference,
	)
	emptyFindProjectColumns := func() *SeparatedSyntaxListOf[Expression] {
		return AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>"))
	}
	findProjectClause := FirstValue(
		Rule2(
			SyntaxTokenKindParser(SyntaxKindProjectKeyword),
			SyntaxCommaListParser(findProjectColumn, missingExpression, true, false),
			func(keyword *SyntaxToken, columns *SeparatedSyntaxListOf[Expression]) *FindProjectClause {
				return MustNewFindProjectClauseTyped(keyword, columns)
			},
		),
		Rule1(SyntaxTokenKindParser(SyntaxKindProjectSmartKeyword), func(keyword *SyntaxToken) *FindProjectClause {
			return MustNewFindProjectClauseTyped(keyword, emptyFindProjectColumns())
		}),
	)
	findProjectAwayClause := Rule2(
		SyntaxTokenKindParser(SyntaxKind_ProjectAwayKeyword),
		SyntaxCommaListParser(findProjectColumn, missingExpression, true, false),
		func(keyword *SyntaxToken, columns *SeparatedSyntaxListOf[Expression]) *FindProjectClause {
			return MustNewFindProjectClauseTyped(keyword, columns)
		},
	)
	findOperator := asResultParser[LexicalToken, QueryOperator](Rule8(
		SyntaxTokenKindParser(SyntaxKindFindKeyword, NewCompletionItem(CompletionKindQueryPrefix, "find")),
		OptionalValue(dataScopeClause(CompletionKindSyntax)),
		queryGrammarParameterList(g, QueryOperatorParameters.FindParameters, true),
		OptionalValue(findInClause),
		OptionalValue(SyntaxTokenKindParser(SyntaxKindWhereKeyword)),
		Required(g.UnnamedExpression, missingExpression),
		OptionalValue(findProjectClause),
		asResultParser[LexicalToken, *FindProjectClause](OptionalValue(findProjectAwayClause).Hide()),
		func(keyword *SyntaxToken, scope *DataScopeClause, parameters *SyntaxListOf[*NamedParameter], inClause *FindInClause, where *SyntaxToken, condition Expression, project, projectAway *FindProjectClause) QueryOperator {
			if where == nil && (parameters.Count() > 0 || inClause != nil) {
				where = CreateMissingSyntaxToken(SyntaxKindWhereKeyword)
			}
			return MustNewFindOperatorTyped(keyword, scope, parameters, inClause, where, condition, project, projectAway)
		},
	).WithTag("<find>"))
	starSearchPredicate := ApplyOptional(
		Rule1(SyntaxTokenKindParser(SyntaxKindAsteriskToken), func(token *SyntaxToken) Expression {
			return MustNewStarExpressionTyped(token)
		}),
		func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
			return RuleRight2(
				SyntaxTokenKindParser(SyntaxKindAndKeyword),
				Required(g.UnnamedExpression, missingExpression),
				func(left Expression, operator *SyntaxToken, right Expression) Expression {
					return MustNewBinaryExpressionTyped(SyntaxKindAndExpression, left, operator, right)
				},
			)
		},
	)
	searchOperator := asResultParser[LexicalToken, QueryOperator](Rule5(
		SyntaxTokenKindParser(SyntaxKindSearchKeyword, NewCompletionItem(CompletionKindQueryPrefix, "search")),
		queryGrammarParameterList(g, QueryOperatorParameters.SearchParameters, true),
		OptionalValue(dataScopeClause(CompletionKindSyntax)),
		OptionalValue(findInClause),
		Required(FirstValue(g.UnnamedExpression, starSearchPredicate), missingExpression),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], scope *DataScopeClause, inClause *FindInClause, condition Expression) QueryOperator {
			return MustNewSearchOperatorTyped(keyword, parameters, scope, inClause, condition)
		},
	).WithTag("<search>"))
	var forkPipeOperatorCore ResultParser[LexicalToken, QueryOperator]
	var forkPipeExpressionCore ResultParser[LexicalToken, Expression]
	forkPipeOperator := asResultParser[LexicalToken, QueryOperator](Forward(func() ResultParser[LexicalToken, QueryOperator] {
		return forkPipeOperatorCore
	}).WithTag("<fork-pipe-operator>"))
	forkPipeExpression := asResultParser[LexicalToken, Expression](Forward(func() ResultParser[LexicalToken, Expression] {
		return forkPipeExpressionCore
	}).WithTag("<fork-pipe-expression>"))
	nameEqualsClause := IfValue(
		Or(SyntaxTokenKindParser(SyntaxKindIdentifierToken), SyntaxTokenKindParser(SyntaxKindOpenBracketToken)),
		Rule2(
			g.SimpleNameDeclaration,
			RequiredSyntaxTokenKindParser(SyntaxKindEqualToken),
			func(name *NameDeclaration, equal *SyntaxToken) *NameEqualsClause {
				return MustNewNameEqualsClauseTyped(name, equal)
			},
		),
	)
	g.nameEqualsClause = nameEqualsClause
	forkExpression := Rule4(
		OptionalValue(nameEqualsClause),
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		Required(
			FirstValue(
				forkPipeExpression,
				asResultParser[LexicalToken, Expression](Forward(func() ResultParser[LexicalToken, Expression] {
					return g.Expression
				}).Hide()),
			),
			missingExpression,
		),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(name *NameEqualsClause, open *SyntaxToken, expression Expression, close *SyntaxToken) *ForkExpression {
			return MustNewForkExpressionTyped(name, open, expression, close)
		},
	)
	g.forkExpression = forkExpression
	missingForkExpression := func(source Source[LexicalToken], start int) *ForkExpression {
		return MustNewForkExpressionTyped(
			nil,
			CreateMissingSyntaxToken(SyntaxKindOpenParenToken),
			missingExpression(source, start),
			CreateMissingSyntaxToken(SyntaxKindCloseParenToken),
		)
	}
	forkOperator := asResultParser[LexicalToken, QueryOperator](Rule2(
		asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(
			SyntaxKindForkKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "fork"),
		).Hide()),
		SyntaxListParser(forkExpression, missingForkExpression, true),
		func(keyword *SyntaxToken, expressions *SyntaxListOf[*ForkExpression]) QueryOperator {
			return MustNewForkOperatorTyped(keyword, expressions)
		},
	).WithTag("<fork>"))
	g.forkOperator = forkOperator
	g.forkPipeOperator = forkPipeOperator
	g.forkPipeExpression = forkPipeExpression
	forkPipeOperatorCore = FirstValue(
		countOperator,
		extendOperator,
		filterOperator,
		parseOperator,
		parseWhereOperator,
		parseKvOperator,
		takeOperator,
		topNestedOperator,
		projectOperator,
		projectAwayOperator,
		projectByNamesOperator,
		projectKeepOperator,
		projectRenameOperator,
		projectReorderOperator,
		Rule1(summarizeOperator, func(operator *SummarizeOperator) QueryOperator { return operator }),
		distinctOperator,
		topHittersOperator,
		topOperator,
		sortOperator,
		mvExpandOperator,
		mvApplyOperator,
		reduceByOperator,
		sampleOperator,
		sampleDistinctOperator,
		asOperator,
		invokeOperator,
		executeAndCacheOperator,
		scanOperator,
		asResultParser[LexicalToken, QueryOperator](Forward(func() ResultParser[LexicalToken, QueryOperator] {
			return g.QueryOperator
		}).Hide()),
	)
	forkPipeExpressionCore = ApplyZeroOrMore(
		Rule1(forkPipeOperator, func(operator QueryOperator) Expression { return operator.(Expression) }),
		func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
			return RuleRight2(
				SyntaxTokenKindParser(SyntaxKindBarToken),
				Required(forkPipeOperator, func(Source[LexicalToken], int) QueryOperator {
					return MustNewBadQueryOperatorTyped(createOmittedToken(SyntaxKindIdentifierToken), DiagnosticQueryOperatorExpected())
				}),
				func(left Expression, bar *SyntaxToken, right QueryOperator) Expression {
					return MustNewPipeExpressionTyped(left, bar, right)
				},
			)
		},
	)
	missingFacetQueryOperator := func(Source[LexicalToken], int) QueryOperator {
		return MustNewBadQueryOperatorTyped(createOmittedToken(SyntaxKindIdentifierToken), DiagnosticQueryOperatorExpected())
	}
	missingFacetQueryOperatorExpression := func(Source[LexicalToken], int) Expression {
		return MustNewBadQueryOperatorTyped(createOmittedToken(SyntaxKindIdentifierToken), DiagnosticQueryOperatorExpected())
	}
	facetWithExpressionClause := Rule4(
		SyntaxTokenKindParser(SyntaxKindWithKeyword),
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		Required(forkPipeExpression, missingFacetQueryOperatorExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(with, open *SyntaxToken, expression Expression, close *SyntaxToken) FacetWithClause {
			return MustNewFacetWithExpressionClauseTyped(with, open, expression, close)
		},
	)
	facetWithOperatorClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindWithKeyword),
		Required(forkPipeOperator, missingFacetQueryOperator),
		func(with *SyntaxToken, operator QueryOperator) FacetWithClause {
			return MustNewFacetWithOperatorClauseTyped(with, operator)
		},
	)
	facetOperator := asResultParser[LexicalToken, QueryOperator](Rule4(
		asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(
			SyntaxKindFacetKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "facet"),
		).Hide()),
		RequiredSyntaxTokenKindParser(SyntaxKindByKeyword),
		SyntaxSeparatedListParser(
			g.simplePathExpression,
			SyntaxKindCommaToken,
			func(Source[LexicalToken], int) Expression { return createMissingNameReferenceNode() },
			nil,
			true,
			false,
		),
		OptionalValue(FirstValue(facetWithExpressionClause, facetWithOperatorClause)),
		func(facet, by *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], with FacetWithClause) QueryOperator {
			return MustNewFacetOperatorTyped(facet, by, expressions, with)
		},
	).WithTag("<facet>"))
	g.facetOperator = facetOperator
	entityGroup := Rule4(
		SyntaxTokenKindParser(SyntaxKindEntityGroupKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenBracketToken),
		SyntaxCommaListParser(g.UnnamedExpression, missingExpression, true, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBracketToken),
		func(keyword, open *SyntaxToken, entities *SeparatedSyntaxListOf[Expression], close *SyntaxToken) Expression {
			return MustNewEntityGroupTyped(keyword, open, entities, close)
		},
	)
	g.entityGroup = entityGroup
	macroExpandScopeReferenceName := Rule2(
		RequiredSyntaxTokenKindParser(SyntaxKindAsKeyword),
		asResultParser[LexicalToken, *NameDeclaration](Rule1(SyntaxTokenKindParser(SyntaxKindIdentifierToken), func(token *SyntaxToken) *NameDeclaration {
			return MustNewNameDeclarationTyped(MustNewTokenNameTyped(token))
		}).WithTag("<identifer>")),
		func(as *SyntaxToken, name *NameDeclaration) *MacroExpandScopeReferenceName {
			return MustNewMacroExpandScopeReferenceNameTyped(as, name)
		},
	)
	scanQualifiedEntityStart := And(
		Or(HiddenSyntaxTokenTextParser("database"), HiddenSyntaxTokenTextParser("cluster")),
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
	)
	macroExpandTarget := FirstValue(
		entityGroup,
		IfValue(scanQualifiedEntityStart, g.entityPathExpression),
		g.functionCallExpression,
		g.SimpleNameReference,
	)
	macroExpandSubQuery := Forward(func() ResultParser[LexicalToken, *SeparatedSyntaxListOf[Statement]] {
		return g.MacroExpandSubQuery
	})
	macroExpandOperator := asResultParser[LexicalToken, QueryOperator](Rule7(
		SyntaxTokenKindParser(
			SyntaxKindMacroExpandKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "macro-expand", WithCompletionPriority(CompletionPriorityHigh)),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.MacroExpandParameters, true),
		macroExpandTarget,
		OptionalValue(macroExpandScopeReferenceName),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		macroExpandSubQuery,
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], target Expression, scope *MacroExpandScopeReferenceName, open *SyntaxToken, statements *SeparatedSyntaxListOf[Statement], close *SyntaxToken) QueryOperator {
			return MustNewMacroExpandOperatorTyped(keyword, parameters, target, scope, open, statements, close)
		},
	).WithTag("<macro-expand>"))
	g.macroExpandOperator = macroExpandOperator
	rowSchema := Rule4(
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		OptionalValue(SyntaxTokenKindParser(SyntaxKindCommaToken)),
		SyntaxCommaListParser(
			g.nameAndTypeDeclaration,
			func(Source[LexicalToken], int) *NameAndTypeDeclaration {
				return queryGrammarMissingNameAndTypeDeclaration()
			},
			false,
			true,
		),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open, leadingComma *SyntaxToken, columns *SeparatedSyntaxListOf[*NameAndTypeDeclaration], close *SyntaxToken) *RowSchema {
			return MustNewRowSchemaTyped(open, leadingComma, columns, close)
		},
	)
	mandatoryRowSchema := Rule4(
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		OptionalValue(SyntaxTokenKindParser(SyntaxKindCommaToken)),
		SyntaxCommaListParser(
			g.nameAndTypeDeclaration,
			func(Source[LexicalToken], int) *NameAndTypeDeclaration {
				return queryGrammarMissingNameAndTypeDeclaration()
			},
			true,
			true,
		),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open, leadingComma *SyntaxToken, columns *SeparatedSyntaxListOf[*NameAndTypeDeclaration], close *SyntaxToken) *RowSchema {
			return MustNewRowSchemaTyped(open, leadingComma, columns, close)
		},
	)
	dataTableExpression := Rule7(
		SyntaxTokenKindParser(SyntaxKindDataTableKeyword),
		queryGrammarParameterList(g, QueryOperatorParameters.DataTableParameters, false),
		Required(rowSchema, func(Source[LexicalToken], int) *RowSchema { return createMissingRowSchema() }),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenBracketToken),
		OptionalValue(SyntaxTokenKindParser(SyntaxKindCommaToken)),
		g.LiteralList,
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBracketToken),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], schema *RowSchema, openBracket, leadingComma *SyntaxToken, values *SeparatedSyntaxListOf[Expression], closeBracket *SyntaxToken) Expression {
			return MustNewDataTableExpressionTyped(keyword, parameters, schema, openBracket, leadingComma, values, closeBracket)
		},
	)
	contextualDataTableExpression := Rule3(
		HiddenSyntaxTokenKindParser(SyntaxKindContextualDataTableKeyword),
		Required(g.UnnamedExpression, queryGrammarMissingExpression),
		Required(rowSchema, func(Source[LexicalToken], int) *RowSchema { return createMissingRowSchema() }),
		func(keyword *SyntaxToken, id Expression, schema *RowSchema) Expression {
			return MustNewContextualDataTableExpressionTyped(keyword, id, schema)
		},
	)
	requiredContextualFollowingOperator := Required(
		asResultParser[LexicalToken, QueryOperator](Forward(func() ResultParser[LexicalToken, QueryOperator] {
			return g.FollowingPipeElementExpression
		})),
		func(Source[LexicalToken], int) QueryOperator {
			return MustNewBadQueryOperatorTyped(createOmittedToken(SyntaxKindIdentifierToken), DiagnosticQueryOperatorExpected())
		},
	)
	contextualDataTablePipeline := ApplyZeroOrMore(
		contextualDataTableExpression,
		func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
			return RuleRight2(
				SyntaxTokenKindParser(SyntaxKindBarToken),
				requiredContextualFollowingOperator,
				func(left Expression, bar *SyntaxToken, right QueryOperator) Expression {
					return MustNewPipeExpressionTyped(left, bar, right)
				},
			)
		},
	)
	contextualSubExpression := asResultParser[LexicalToken, Expression](FirstValue(
		contextualDataTablePipeline,
		Forward(func() ResultParser[LexicalToken, Expression] { return g.PipeSubExpression }),
	).WithTag("<contextual-sub-expression>"))
	g.contextualSubExpression = contextualSubExpression
	mvApplyExpression := FirstValue(
		IfValue(
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindToKeyword).Hide()),
			Rule1(toTypeOfClause, func(clause *ToTypeOfClause) *MvApplyExpression {
				return MustNewMvApplyExpressionTyped(queryGrammarMissingExpression(nil, 0), clause)
			}),
		),
		Rule2(
			g.NamedExpression,
			OptionalValue(toTypeOfClause),
			func(expression Expression, clause *ToTypeOfClause) *MvApplyExpression {
				return MustNewMvApplyExpressionTyped(expression, clause)
			},
		),
	)
	singleMvApplyTypeExpression := IfValue(
		And(
			SyntaxTokenKindParser(SyntaxKindToKeyword),
			g.typeOfLiteral,
			Fails(SyntaxTokenKindParser(SyntaxKindCommaToken)),
		),
		asResultParser[LexicalToken, *SeparatedSyntaxListOf[*MvApplyExpression]](Rule1(toTypeOfClause, func(clause *ToTypeOfClause) *SeparatedSyntaxListOf[*MvApplyExpression] {
			expression := MustNewMvApplyExpressionTyped(nil, clause)
			return AsSeparatedSyntaxListOf[*MvApplyExpression](NewTypedSyntaxList(
				"SeparatedElement<MvApplyExpression>",
				NewSeparatedElement(expression, nil),
			))
		}).Hide()),
	)
	mvApplyExpressionBeforeClause := IfValue(
		Not(First(
			SyntaxTokenKindParser(SyntaxKindLimitKeyword),
			SyntaxTokenKindParser(SyntaxKindIdKeyword),
			SyntaxTokenKindParser(SyntaxKindOnKeyword),
		)),
		mvApplyExpression,
	)
	mvApplyExpressionList := FirstValue(
		singleMvApplyTypeExpression,
		SyntaxCommaListParser(
			mvApplyExpressionBeforeClause,
			func(source Source[LexicalToken], start int) *MvApplyExpression {
				return MustNewMvApplyExpressionTyped(missingExpression(source, start), nil)
			},
			true,
			false,
			SyntaxKindLimitKeyword,
			SyntaxKindIdKeyword,
			SyntaxKindOnKeyword,
		),
	)
	mvApplyRowLimitClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindLimitKeyword),
		Required(g.UnnamedExpression, missingExpression),
		func(keyword *SyntaxToken, expression Expression) *MvApplyRowLimitClause {
			return MustNewMvApplyRowLimitClauseTyped(keyword, expression)
		},
	)
	mvApplyContextIDClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindIdKeyword),
		Required(g.UnnamedExpression, missingExpression),
		func(keyword *SyntaxToken, expression Expression) *MvApplyContextIdClause {
			return MustNewMvApplyContextIdClauseTyped(keyword, expression)
		},
	)
	mvApplySubqueryExpression := Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		Required(contextualSubExpression, missingExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open *SyntaxToken, expression Expression, close *SyntaxToken) *MvApplySubqueryExpression {
			return MustNewMvApplySubqueryExpressionTyped(open, expression, close)
		},
	)
	missingMvApplySubqueryExpression := func(source Source[LexicalToken], start int) *MvApplySubqueryExpression {
		return MustNewMvApplySubqueryExpressionTyped(
			CreateMissingSyntaxToken(SyntaxKindOpenParenToken),
			missingExpression(source, start),
			CreateMissingSyntaxToken(SyntaxKindCloseParenToken),
		)
	}
	mvApplyOperatorCore = asResultParser[LexicalToken, QueryOperator](Rule7(
		FirstValue(
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(
				SyntaxKindMvApplyKeyword,
				NewCompletionItem(CompletionKindQueryPrefix, "mvapply"),
			).Hide()),
			SyntaxTokenKindParser(
				SyntaxKindMvDashApplyKeyword,
				NewCompletionItem(CompletionKindQueryPrefix, "mv-apply"),
			),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.MvApplyParameters, true),
		mvApplyExpressionList,
		OptionalValue(mvApplyRowLimitClause),
		asResultParser[LexicalToken, *MvApplyContextIdClause](OptionalValue(mvApplyContextIDClause).Hide()),
		RequiredSyntaxTokenKindParser(SyntaxKindOnKeyword),
		Required(mvApplySubqueryExpression, missingMvApplySubqueryExpression),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expressions *SeparatedSyntaxListOf[*MvApplyExpression], rowLimit *MvApplyRowLimitClause, contextID *MvApplyContextIdClause, on *SyntaxToken, subquery *MvApplySubqueryExpression) QueryOperator {
			return MustNewMvApplyOperatorTyped(keyword, parameters, expressions, rowLimit, contextID, on, subquery)
		},
	))
	parseParameterType := Rule1(queryGrammarParamTypeTokenParser(), func(token *SyntaxToken) TypeExpression {
		return MustNewPrimitiveTypeExpressionTyped(token)
	})
	parseInvalidParameterType := Rule1(queryGrammarInvalidParamTypeTokenParser(), func(token *SyntaxToken) TypeExpression {
		return MustNewPrimitiveTypeExpressionTyped(token, DiagnosticInvalidTypeName(token.RawText()))
	})
	parseDeclarationName := Rule1(queryGrammarExtendedNameDeclarationParser(g), func(name *NameDeclaration) Expression {
		return name
	})
	nameAndOptionalTypeDeclaration := FirstValue(
		ApplyOptional(parseDeclarationName, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
			return RuleRight2(
				SyntaxTokenKindParser(SyntaxKindColonToken),
				Required(FirstValue(parseParameterType, parseInvalidParameterType), func(Source[LexicalToken], int) TypeExpression {
					return createMissingType()
				}),
				func(left Expression, colon *SyntaxToken, typ TypeExpression) Expression {
					return MustNewNameAndTypeDeclarationTyped(left.(*NameDeclaration), colon, typ)
				},
			)
		}),
		asResultParser[LexicalToken, Expression](Rule2(
			asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindColonToken).Hide()),
			Required(FirstValue(parseParameterType, parseInvalidParameterType), func(Source[LexicalToken], int) TypeExpression {
				return createMissingType()
			}),
			func(colon *SyntaxToken, typ TypeExpression) Expression {
				return MustNewNameAndTypeDeclarationTyped(createMissingNameDeclaration(), colon, typ)
			},
		)),
	)
	parseWithExpression := FirstValue(
		Rule1(SyntaxTokenKindParser(SyntaxKindAsteriskToken), func(token *SyntaxToken) Expression {
			return MustNewStarExpressionTyped(token)
		}),
		g.stringOrCompoundLiteral,
		nameAndOptionalTypeDeclaration,
	)
	parsePatterns := SyntaxListParser(
		Rule1(parseWithExpression, func(expression Expression) SyntaxNodeElement { return expression }),
		nil,
		false,
	)
	parseSourceExpression := Required(
		IfValue(Not(SyntaxTokenKindParser(SyntaxKindWithKeyword)), g.UnnamedExpression),
		missingExpression,
	)
	parseOperatorCore = asResultParser[LexicalToken, QueryOperator](Rule5(
		SyntaxTokenKindParser(
			SyntaxKindParseKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "parse", WithCompletionPriority(CompletionPriorityLow)),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.ParseParameters, true),
		parseSourceExpression,
		RequiredSyntaxTokenKindParser(SyntaxKindWithKeyword),
		parsePatterns,
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, with *SyntaxToken, patterns *SyntaxListOf[SyntaxNodeElement]) QueryOperator {
			return MustNewParseOperatorTyped(keyword, parameters, expression, with, patterns)
		},
	))
	parseWhereOperatorCore = asResultParser[LexicalToken, QueryOperator](Rule5(
		SyntaxTokenKindParser(
			SyntaxKindParseWhereKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "parse-where", WithCompletionPriority(CompletionPriorityLow)),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.ParseParameters, true),
		parseSourceExpression,
		RequiredSyntaxTokenKindParser(SyntaxKindWithKeyword),
		parsePatterns,
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, with *SyntaxToken, patterns *SyntaxListOf[SyntaxNodeElement]) QueryOperator {
			return MustNewParseWhereOperatorTyped(keyword, parameters, expression, with, patterns)
		},
	))
	parseKvWithClause := Rule4(
		asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindWithKeyword).Hide()),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		queryGrammarParameterCommaList(g, QueryOperatorParameters.ParseKvWithProperties, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(keyword, open *SyntaxToken, properties *SeparatedSyntaxListOf[*NamedParameter], close *SyntaxToken) *ParseKvWithClause {
			return MustNewParseKvWithClauseTyped(keyword, open, properties, close)
		},
	)
	parseKvSourceExpression := Required(
		IfValue(Not(SyntaxTokenKindParser(SyntaxKindAsKeyword)), g.UnnamedExpression),
		missingExpression,
	)
	parseKvOperatorCore = asResultParser[LexicalToken, QueryOperator](Rule5(
		SyntaxTokenKindParser(
			SyntaxKindParseKvKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "parse-kv", WithCompletionPriority(CompletionPriorityLow)),
		),
		parseKvSourceExpression,
		RequiredSyntaxTokenKindParser(SyntaxKindAsKeyword),
		Required(rowSchema, func(Source[LexicalToken], int) *RowSchema { return createMissingRowSchema() }),
		OptionalValue(parseKvWithClause),
		func(keyword *SyntaxToken, expression Expression, as *SyntaxToken, keys *RowSchema, with *ParseKvWithClause) QueryOperator {
			return MustNewParseKvOperatorTyped(keyword, expression, as, keys, with)
		},
	))
	partitionScopeClause := Rule2(
		asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindInKeyword).Hide()),
		Required(
			FirstValue(g.functionCallExpression, g.dynamicLiteral),
			missingExpression,
		),
		func(keyword *SyntaxToken, expression Expression) *PartitionScope {
			return MustNewPartitionScopeTyped(keyword, expression)
		},
	)
	partitionQueryExpression := Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenBraceToken),
		Required(Forward(func() ResultParser[LexicalToken, Expression] { return g.Expression }), missingExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBraceToken),
		func(open *SyntaxToken, expression Expression, close *SyntaxToken) PartitionOperand {
			return MustNewPartitionQueryTyped(open, expression, close)
		},
	)
	partitionSubqueryBody := FirstValue(
		Forward(func() ResultParser[LexicalToken, Expression] { return g.PipeSubExpression }),
		asResultParser[LexicalToken, Expression](Forward(func() ResultParser[LexicalToken, Expression] {
			return g.Expression
		}).Hide()),
	)
	partitionSubqueryExpression := Rule3(
		SyntaxTokenKindParser(SyntaxKindOpenParenToken),
		Required(partitionSubqueryBody, missingExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open *SyntaxToken, expression Expression, close *SyntaxToken) PartitionOperand {
			return MustNewPartitionSubqueryTyped(open, expression, close)
		},
	)
	partitionOperand := FirstValue(partitionSubqueryExpression, partitionQueryExpression)
	missingPartitionOperand := func(source Source[LexicalToken], start int) PartitionOperand {
		return MustNewPartitionSubqueryTyped(
			CreateMissingSyntaxToken(SyntaxKindOpenParenToken),
			missingExpression(source, start),
			CreateMissingSyntaxToken(SyntaxKindCloseParenToken),
		)
	}
	partitionOperatorCore = asResultParser[LexicalToken, QueryOperator](Rule6(
		SyntaxTokenKindParser(
			SyntaxKindPartitionKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "partition", WithCompletionPriority(CompletionPriorityLow)),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.PartitionParameters, false),
		RequiredSyntaxTokenKindParser(SyntaxKindByKeyword),
		Required(g.simplePathExpression, func(Source[LexicalToken], int) Expression { return createMissingNameReferenceNode() }),
		OptionalValue(partitionScopeClause),
		Required(partitionOperand, missingPartitionOperand),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], by *SyntaxToken, expression Expression, scope *PartitionScope, operand PartitionOperand) QueryOperator {
			return MustNewPartitionOperatorTyped(keyword, parameters, by, expression, scope, operand)
		},
	))
	partitionByIDClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindIdKeyword),
		g.Literal,
		func(keyword *SyntaxToken, value Expression) *PartitionByIdClause {
			return MustNewPartitionByIdClauseTyped(keyword, value)
		},
	)
	partitionByOperatorCore = asResultParser[LexicalToken, QueryOperator](Rule7(
		HiddenSyntaxTokenKindParser(SyntaxKindPartitionByKeyword),
		queryGrammarParameterList(g, QueryOperatorParameters.PartitionByParameters, false),
		Required(g.simplePathExpression, func(Source[LexicalToken], int) Expression { return createMissingNameReferenceNode() }),
		OptionalValue(partitionByIDClause),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		Required(g.contextualSubExpression, missingExpression),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], entity Expression, id *PartitionByIdClause, open *SyntaxToken, subquery Expression, close *SyntaxToken) QueryOperator {
			return MustNewPartitionByOperatorTyped(keyword, parameters, entity, id, open, subquery, close)
		},
	))
	reduceByWithClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindWithKeyword),
		queryGrammarParameterCommaList(g, QueryOperatorParameters.ReduceWithParameters, false),
		func(keyword *SyntaxToken, parameters *SeparatedSyntaxListOf[*NamedParameter]) *ReduceByWithClause {
			return MustNewReduceByWithClauseTyped(keyword, parameters)
		},
	)
	reduceByExpression := Required(
		IfValue(Not(SyntaxTokenKindParser(SyntaxKindWithKeyword)), g.NamedExpression),
		missingExpression,
	)
	reduceByOperatorCore = asResultParser[LexicalToken, QueryOperator](Rule5(
		SyntaxTokenKindParser(
			SyntaxKindReduceKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "reduce", WithCompletionPriority(CompletionPriorityLow)),
		),
		queryGrammarParameterList(g, QueryOperatorParameters.ReduceParameters, false),
		RequiredSyntaxTokenKindParser(SyntaxKindByKeyword),
		reduceByExpression,
		OptionalValue(reduceByWithClause),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], by *SyntaxToken, expression Expression, with *ReduceByWithClause) QueryOperator {
			return MustNewReduceByOperatorTyped(keyword, parameters, by, expression, with)
		},
	))
	renderWithClause := Rule5(
		SyntaxTokenKindParser(SyntaxKindWithKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		OptionalValue(SyntaxTokenKindParser(SyntaxKindCommaToken)),
		queryGrammarParameterCommaList(g, QueryOperatorParameters.RenderWithProperties, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(with, open, leadingComma *SyntaxToken, properties *SeparatedSyntaxListOf[*NamedParameter], close *SyntaxToken) *RenderWithClause {
			return MustNewRenderWithClauseTyped(with, open, leadingComma, properties, close)
		},
	)
	hiddenChartTypes := make(map[string]struct{}, len(KustoFacts.HiddenChartTypes))
	for _, chartType := range KustoFacts.HiddenChartTypes {
		hiddenChartTypes[chartType] = struct{}{}
	}
	chartTypeParsers := make([]ResultParser[LexicalToken, *SyntaxToken], 0, len(KustoFacts.ChartTypes)+1)
	for _, chartType := range KustoFacts.ChartTypes {
		if _, hidden := hiddenChartTypes[chartType]; hidden {
			chartTypeParsers = append(chartTypeParsers, HiddenSyntaxTokenTextParser(chartType))
			continue
		}
		base := asResultParser[LexicalToken, *SyntaxToken](SyntaxMatchTextParser(chartType).WithTag(defaultSyntaxParserTagForText(chartType)))
		chartTypeParsers = append(chartTypeParsers, WithSyntaxCompletion(base, NewCompletionItem(CompletionKindRenderChart, chartType)))
	}
	chartTypeParsers = append(chartTypeParsers, SyntaxTokenKindParser(SyntaxKindIdentifierToken))
	renderChartType := FirstValue(chartTypeParsers...)
	deprecatedRenderByName := IfValue(
		Not(Or(
			SyntaxTokenTextParser("kind"),
			SyntaxTokenTextParser("title"),
			SyntaxTokenTextParser("accumulate"),
			SyntaxTokenKindParser(SyntaxKindWithKeyword),
		)),
		Rule1(g.SimpleNameReference, func(expression Expression) *NameReference { return expression.(*NameReference) }),
	)
	deprecatedRenderBy := Rule2(
		HiddenSyntaxTokenKindParser(SyntaxKindByKeyword),
		queryGrammarNameReferenceList(deprecatedRenderByName, nil),
		func(keyword *SyntaxToken, names Expression) *NamedParameter {
			return MustNewNamedParameterTyped(
				MustNewNameDeclarationTyped(MustNewTokenNameTyped(keyword)),
				createOmittedToken(SyntaxKindEqualToken),
				names,
				CompletionHintNone,
			)
		},
	)
	deprecatedRenderWith := IfValue(
		And(HiddenSyntaxTokenKindParser(SyntaxKindWithKeyword), Not(SyntaxTokenKindParser(SyntaxKindOpenParenToken))),
		Rule2(
			HiddenSyntaxTokenKindParser(SyntaxKindWithKeyword),
			Required(g.Literal, func(Source[LexicalToken], int) Expression { return createMissingValueExpression() }),
			func(keyword *SyntaxToken, value Expression) *NamedParameter {
				return MustNewNamedParameterTyped(
					MustNewNameDeclarationTyped(MustNewTokenNameTyped(keyword)),
					createOmittedToken(SyntaxKindEqualToken),
					value,
					CompletionHintNone,
				)
			},
		),
	)
	deprecatedRenderProperty := FirstValue(
		queryGrammarParameter(g, QueryOperatorParameters.RenderParameters[0], false),
		queryGrammarParameter(g, QueryOperatorParameters.RenderParameters[1], false),
		queryGrammarParameter(g, QueryOperatorParameters.RenderParameters[2], false),
		deprecatedRenderWith,
		deprecatedRenderBy,
	)
	deprecatedRenderProperties := ZeroOrMoreProduce(deprecatedRenderProperty, func(parameters []*NamedParameter) *SyntaxListOf[*NamedParameter] {
		elements := make([]SyntaxElement, len(parameters))
		for i, parameter := range parameters {
			elements[i] = parameter
		}
		return AsSyntaxListOf[*NamedParameter](NewTypedSyntaxList("NamedParameter", elements...))
	})
	renderOperatorCore = asResultParser[LexicalToken, QueryOperator](Rule4(
		SyntaxTokenKindParser(
			SyntaxKindRenderKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "render", WithCompletionPriority(CompletionPriorityHigh)),
		),
		Required(renderChartType, func(Source[LexicalToken], int) *SyntaxToken {
			return CreateMissingSyntaxTokenForTexts(KustoFacts.ChartTypes)
		}),
		deprecatedRenderProperties,
		OptionalValue(renderWithClause),
		func(keyword, chart *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], with *RenderWithClause) QueryOperator {
			return MustNewRenderOperatorTyped(keyword, chart, parameters, with)
		},
	))
	scanAssignment := asResultParser[LexicalToken, *ScanAssignment](Rule3(
		queryGrammarExtendedNameReferenceParser(g),
		SyntaxTokenKindParser(SyntaxKindEqualToken),
		Required(g.UnnamedExpression, missingExpression),
		func(name *NameReference, equal *SyntaxToken, expression Expression) *ScanAssignment {
			return MustNewScanAssignmentTyped(name, equal, expression)
		},
	).WithTag("<assignment>"))
	scanComputationClause := Rule2(
		SyntaxTokenKindParser(SyntaxKindFatArrowToken),
		SyntaxCommaListParser(scanAssignment, func(Source[LexicalToken], int) *ScanAssignment {
			return createMissingScanAssignment()
		}, true, false),
		func(arrow *SyntaxToken, assignments *SeparatedSyntaxListOf[*ScanAssignment]) *ScanComputationClause {
			return MustNewScanComputationClauseTyped(arrow, assignments)
		},
	)
	scanStepOutput := Rule3(
		SyntaxTokenKindParser(SyntaxKindOutputKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindEqualToken),
		RequiredSyntaxTokenTextsParser(KustoFacts.ScanStepOutputValues),
		func(output, equal, kind *SyntaxToken) *ScanStepOutput {
			return MustNewScanStepOutputTyped(output, equal, kind)
		},
	)
	renameName := queryGrammarExtendedNameDeclarationParser(g)
	scanStep := Rule8(
		SyntaxTokenKindParser(SyntaxKindStepKeyword),
		Required(renameName, func(Source[LexicalToken], int) *NameDeclaration { return createMissingNameDeclaration() }),
		OptionalValue(HiddenSyntaxTokenKindParser(SyntaxKindOptionalKeyword)),
		OptionalValue(asResultParser[LexicalToken, *ScanStepOutput](scanStepOutput.Hide())),
		RequiredSyntaxTokenKindParser(SyntaxKindColonToken),
		Required(g.UnnamedExpression, missingExpression),
		OptionalValue(scanComputationClause),
		RequiredSyntaxTokenKindParser(SyntaxKindSemicolonToken),
		func(step *SyntaxToken, name *NameDeclaration, optional *SyntaxToken, output *ScanStepOutput, colon *SyntaxToken, condition Expression, computation *ScanComputationClause, semicolon *SyntaxToken) *ScanStep {
			return MustNewScanStepTyped(step, name, optional, output, colon, condition, computation, semicolon)
		},
	)
	scanOrderByClause := Rule3(
		HiddenSyntaxTokenKindParser(SyntaxKindOrderKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindByKeyword),
		SyntaxCommaListParser(
			orderedExpression,
			missingExpression,
			true,
			false,
			SyntaxKindPartitionKeyword,
			SyntaxKindDeclareKeyword,
			SyntaxKindWithKeyword,
		),
		func(order, by *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression]) *ScanOrderByClause {
			return MustNewScanOrderByClauseTyped(order, by, expressions)
		},
	)
	scanPartitionByClause := Rule3(
		HiddenSyntaxTokenKindParser(SyntaxKindPartitionKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindByKeyword),
		SyntaxCommaListParser(
			g.UnnamedExpression,
			missingExpression,
			true,
			false,
			SyntaxKindDeclareKeyword,
			SyntaxKindWithKeyword,
		),
		func(partition, by *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression]) *ScanPartitionByClause {
			return MustNewScanPartitionByClauseTyped(partition, by, expressions)
		},
	)
	scanDeclareClause := Rule4(
		SyntaxTokenKindParser(SyntaxKindDeclareKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		SyntaxCommaListParser(g.functionParameter, func(Source[LexicalToken], int) *FunctionParameter {
			return createMissingFunctionParameter()
		}, false, false, SyntaxKindWithKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(declare, open *SyntaxToken, declarations *SeparatedSyntaxListOf[*FunctionParameter], close *SyntaxToken) *ScanDeclareClause {
			return MustNewScanDeclareClauseTyped(declare, open, declarations, close)
		},
	)
	scanSteps := SyntaxListParser(scanStep, nil, false)
	scanOperatorCore = asResultParser[LexicalToken, QueryOperator](Rule9(
		SyntaxTokenKindParser(SyntaxKindScanKeyword, NewCompletionItem(CompletionKindQueryPrefix, "scan")),
		queryGrammarParameterList(g, QueryOperatorParameters.ScanParameters, false),
		OptionalValue(scanOrderByClause),
		OptionalValue(scanPartitionByClause),
		OptionalValue(scanDeclareClause),
		RequiredSyntaxTokenKindParser(
			SyntaxKindWithKeyword,
			NewCompletionItem(CompletionKindSyntax, "with", WithCompletionBeforeText("with ("), WithCompletionAfterText(")")),
		),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		scanSteps,
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(scan *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], order *ScanOrderByClause, partition *ScanPartitionByClause, declare *ScanDeclareClause, with, open *SyntaxToken, steps *SyntaxListOf[*ScanStep], close *SyntaxToken) QueryOperator {
			return MustNewScanOperatorTyped(scan, parameters, order, partition, declare, with, open, steps, close)
		},
	))
	inlineKindClause := Rule3(
		SyntaxTokenKindParser(SyntaxKindKindKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindEqualToken),
		RequiredSyntaxTokenTextsParser(KustoFacts.InlineExternalTableKinds),
		func(keyword, equal, value *SyntaxToken) *InlineExternalTableKindClause {
			return MustNewInlineExternalTableKindClauseTyped(keyword, equal, value)
		},
	)
	missingInlineKindClause := func(Source[LexicalToken], int) *InlineExternalTableKindClause {
		return MustNewInlineExternalTableKindClauseTyped(
			createOmittedToken(SyntaxKindKindKeyword),
			createOmittedToken(SyntaxKindEqualToken),
			createOmittedToken(SyntaxKindStringLiteralToken),
			DiagnosticMissingExternalTableKind(),
		)
	}
	inlineDataFormatClause := Rule3(
		SyntaxTokenKindParser(SyntaxKindDataFormatKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindEqualToken),
		RequiredSyntaxTokenKindParser(SyntaxKindIdentifierToken),
		func(keyword, equal, value *SyntaxToken) *InlineExternalTableDataFormatClause {
			return MustNewInlineExternalTableDataFormatClauseTyped(keyword, equal, value)
		},
	)
	missingInlineDataFormatClause := func() *InlineExternalTableDataFormatClause {
		return MustNewInlineExternalTableDataFormatClauseTyped(
			createOmittedToken(SyntaxKindDataFormatKeyword),
			createOmittedToken(SyntaxKindEqualToken),
			createOmittedToken(SyntaxKindStringLiteralToken),
			DiagnosticMissingDataFormat(),
		)
	}
	omittedInlineDataFormatClause := func() *InlineExternalTableDataFormatClause {
		return MustNewInlineExternalTableDataFormatClauseTyped(
			createOmittedToken(SyntaxKindDataFormatKeyword),
			createOmittedToken(SyntaxKindEqualToken),
			createOmittedToken(SyntaxKindStringLiteralToken),
		)
	}
	omittedInlineRowSchema := func() *RowSchema {
		return MustNewRowSchemaTyped(
			createOmittedToken(SyntaxKindOpenParenToken),
			nil,
			AsSeparatedSyntaxListOf[*NameAndTypeDeclaration](NewTypedSyntaxList("SeparatedElement<NameAndTypeDeclaration>")),
			createOmittedToken(SyntaxKindCloseParenToken),
		)
	}
	isInlineDelta := func(kind *InlineExternalTableKindClause) bool {
		return kind != nil && kind.Value() != nil && ordinalEqualFold(kind.Value().ValueText(), "delta")
	}
	partitionTypeToken := FirstValue(
		SyntaxTokenKindParser(SyntaxKindDateTimeKeyword, NewCompletionItem(CompletionKindScalarType, "datetime")),
		SyntaxTokenKindParser(SyntaxKindLongKeyword, NewCompletionItem(CompletionKindScalarType, "long")),
		SyntaxTokenKindParser(SyntaxKindStringKeyword, NewCompletionItem(CompletionKindScalarType, "string")),
	)
	partitionType := Rule1(partitionTypeToken, func(token *SyntaxToken) TypeExpression {
		return MustNewPrimitiveTypeExpressionTyped(token)
	})
	invalidPartitionType := Rule1(queryGrammarInvalidParamTypeTokenParser(), func(token *SyntaxToken) TypeExpression {
		return MustNewPrimitiveTypeExpressionTyped(token, DiagnosticInvalidTypeName(token.RawText()))
	})
	type optionalInlineExpression struct {
		expression Expression
	}
	optionalPartitionExpression := OptionalValue(
		Rule1(g.UnnamedExpression, func(expression Expression) optionalInlineExpression {
			return optionalInlineExpression{expression: expression}
		}),
		func() optionalInlineExpression { return optionalInlineExpression{} },
	)
	partitionColumn := Rule5(
		queryGrammarExtendedNameDeclarationParser(g),
		RequiredSyntaxTokenKindParser(SyntaxKindColonToken),
		Required(FirstValue(partitionType, invalidPartitionType), func(Source[LexicalToken], int) TypeExpression { return createMissingType() }),
		OptionalValue(SyntaxTokenKindParser(SyntaxKindEqualToken)),
		optionalPartitionExpression,
		func(name *NameDeclaration, colon *SyntaxToken, typ TypeExpression, equal *SyntaxToken, expression optionalInlineExpression) *PartitionColumnDeclaration {
			return MustNewPartitionColumnDeclarationTyped(name, colon, typ, equal, expression.expression)
		},
	)
	missingPartitionColumn := func(source Source[LexicalToken], start int) *PartitionColumnDeclaration {
		return MustNewPartitionColumnDeclarationTyped(
			MustNewNameDeclarationTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken))),
			createOmittedToken(SyntaxKindColonToken),
			MustNewPrimitiveTypeExpressionTyped(createOmittedToken(SyntaxKindStringKeyword)),
			createOmittedToken(SyntaxKindEqualToken),
			queryGrammarMissingExpression(source, start),
			DiagnosticMissingPartitionColumnDeclaration(),
		)
	}
	inlinePartitionClause := Rule6(
		SyntaxTokenKindParser(SyntaxKindPartitionKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindByKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		OptionalValue(SyntaxTokenKindParser(SyntaxKindCommaToken)),
		SyntaxCommaListParser(partitionColumn, missingPartitionColumn, false, true),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(keyword, by, open, leadingComma *SyntaxToken, columns *SeparatedSyntaxListOf[*PartitionColumnDeclaration], close *SyntaxToken) *InlineExternalTablePartitionClause {
			return MustNewInlineExternalTablePartitionClauseTyped(keyword, by, open, leadingComma, columns, close)
		},
	)
	stringLiteral := Rule1(g.StringLiteral, func(expression Expression) *LiteralExpression {
		return expression.(*LiteralExpression)
	})
	identifierReference := Rule1(SyntaxTokenKindParser(SyntaxKindIdentifierToken), func(token *SyntaxToken) Expression {
		return MustNewNameReferenceTyped(MustNewTokenNameTyped(token), SymbolMatchNone)
	})
	dateTimePattern := Rule6(
		SyntaxTokenKindParser(SyntaxKindDateTimePatternKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		Required(stringLiteral, func(Source[LexicalToken], int) *LiteralExpression { return createMissingStringLiteral() }),
		RequiredSyntaxTokenKindParser(SyntaxKindCommaToken),
		Required(Rule1(SyntaxTokenKindParser(SyntaxKindIdentifierToken), func(token *SyntaxToken) *NameReference {
			return MustNewNameReferenceTyped(MustNewTokenNameTyped(token), SymbolMatchNone)
		}), func(Source[LexicalToken], int) *NameReference {
			return queryGrammarMissingNameReference().(*NameReference)
		}),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(keyword, open *SyntaxToken, pattern *LiteralExpression, comma *SyntaxToken, column *NameReference, close *SyntaxToken) Expression {
			return MustNewDateTimePatternTyped(keyword, open, pattern, comma, column, close)
		},
	)
	inlinePathElement := Rule2(
		FirstValue(identifierReference, dateTimePattern),
		OptionalValue(stringLiteral),
		func(column Expression, separator *LiteralExpression) *InlineExternalTablePathFormatPartitionColumnReference {
			return MustNewInlineExternalTablePathFormatPartitionColumnReferenceTyped(column, separator)
		},
	)
	missingInlinePathElement := func(source Source[LexicalToken], start int) *InlineExternalTablePathFormatPartitionColumnReference {
		return MustNewInlineExternalTablePathFormatPartitionColumnReferenceTyped(
			queryGrammarMissingExpression(source, start),
			MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, createOmittedToken(SyntaxKindStringLiteralToken)),
			DiagnosticMissingPathFormatTokens(),
		)
	}
	inlinePathFormatClause := Rule6(
		SyntaxTokenKindParser(SyntaxKindPathFormatKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindEqualToken),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		OptionalValue(stringLiteral),
		SyntaxListParser(inlinePathElement, missingInlinePathElement, true),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(keyword, equal, open *SyntaxToken, separator *LiteralExpression, path *SyntaxListOf[*InlineExternalTablePathFormatPartitionColumnReference], close *SyntaxToken) *InlineExternalTablePathFormatClause {
			return MustNewInlineExternalTablePathFormatClauseTyped(keyword, equal, open, separator, path, close)
		},
	)
	inlineConnectionStringsClause := Rule3(
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		SyntaxCommaListParser(g.UnnamedExpression, missingExpression, true, true),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open *SyntaxToken, values *SeparatedSyntaxListOf[Expression], close *SyntaxToken) *InlineExternalTableConnectionStringsClause {
			return MustNewInlineExternalTableConnectionStringsClauseTyped(open, values, close)
		},
	)
	missingInlineConnectionStrings := func(Source[LexicalToken], int) *InlineExternalTableConnectionStringsClause {
		return MustNewInlineExternalTableConnectionStringsClauseTyped(
			createOmittedToken(SyntaxKindOpenParenToken),
			AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>")),
			createOmittedToken(SyntaxKindCloseParenToken),
			DiagnosticMissingConnectionStrings(),
		)
	}
	externalDataLiteral := func(tokenKind, expressionKind SyntaxKind) ResultParser[LexicalToken, Expression] {
		return Rule1(SyntaxTokenKindParser(tokenKind), func(token *SyntaxToken) Expression {
			return MustNewLiteralExpressionTyped(expressionKind, token)
		})
	}
	externalDataPropertyLiteral := FirstValue(
		g.StringLiteral,
		externalDataLiteral(SyntaxKindLongLiteralToken, SyntaxKindLongLiteralExpression),
		externalDataLiteral(SyntaxKindRealLiteralToken, SyntaxKindRealLiteralExpression),
		externalDataLiteral(SyntaxKindBooleanLiteralToken, SyntaxKindBooleanLiteralExpression),
		externalDataLiteral(SyntaxKindDateTimeLiteralToken, SyntaxKindDateTimeLiteralExpression),
		g.typeOfLiteral,
		externalDataLiteral(SyntaxKindRawGuidLiteralToken, SyntaxKindGuidLiteralExpression),
	)
	externalDataPropertyName := queryGrammarExtendedNameDeclarationParser(g)
	externalDataPropertyValue := FirstValue(
		externalDataPropertyLiteral,
		Rule1(externalDataPropertyName, func(name *NameDeclaration) Expression { return name }),
	)
	externalDataProperty := Rule3(
		externalDataPropertyName,
		SyntaxTokenKindParser(SyntaxKindEqualToken),
		Required(externalDataPropertyValue, func(Source[LexicalToken], int) Expression { return createMissingValueExpression() }),
		func(name *NameDeclaration, equals *SyntaxToken, value Expression) *NamedParameter {
			return MustNewNamedParameterTyped(name, equals, value, CompletionHintNone)
		},
	)
	externalDataWithClause := Rule4(
		SyntaxTokenKindParser(SyntaxKindWithKeyword),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		SyntaxCommaListParser(externalDataProperty, func(Source[LexicalToken], int) *NamedParameter {
			return createMissingNamedParameter()
		}, false, false),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(keyword, open *SyntaxToken, properties *SeparatedSyntaxListOf[*NamedParameter], close *SyntaxToken) *ExternalDataWithClause {
			return MustNewExternalDataWithClauseTyped(keyword, open, properties, close)
		},
	)
	inlineKeyword := func() ResultParser[LexicalToken, *SyntaxToken] {
		return SyntaxTokenKindParser(SyntaxKindInlineExternalTableKeyword,
			NewCompletionItem(CompletionKindQueryPrefix, "inline_external_table"))
	}
	inlineParameters := func() ResultParser[LexicalToken, *SyntaxListOf[*NamedParameter]] {
		return queryGrammarParameterList(g, QueryOperatorParameters.InlineExternalTableProperties, false, true)
	}
	inlineDataFormatForKind := func(kind *InlineExternalTableKindClause, clause *InlineExternalTableDataFormatClause) *InlineExternalTableDataFormatClause {
		if clause != nil {
			return clause
		}
		if isInlineDelta(kind) {
			return omittedInlineDataFormatClause()
		}
		return missingInlineDataFormatClause()
	}
	inlineWithSchema := IfValue(
		And(
			HiddenSyntaxTokenKindParser(SyntaxKindInlineExternalTableKeyword),
			HiddenSyntaxTokenKindParser(SyntaxKindOpenParenToken),
			Not(HiddenSyntaxTokenKindParser(SyntaxKindCloseParenToken)),
		),
		Rule9(
			inlineKeyword(),
			inlineParameters(),
			Required(mandatoryRowSchema, func(Source[LexicalToken], int) *RowSchema { return createMissingRowSchema() }),
			Required(inlineKindClause, missingInlineKindClause),
			OptionalValue(inlinePartitionClause),
			OptionalValue(inlinePathFormatClause),
			OptionalValue(inlineDataFormatClause),
			Required(inlineConnectionStringsClause, missingInlineConnectionStrings),
			OptionalValue(externalDataWithClause),
			func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], schema *RowSchema, kind *InlineExternalTableKindClause, partition *InlineExternalTablePartitionClause, path *InlineExternalTablePathFormatClause, dataFormat *InlineExternalTableDataFormatClause, connections *InlineExternalTableConnectionStringsClause, withClause *ExternalDataWithClause) Expression {
				return MustNewInlineExternalTableExpressionTyped(keyword, parameters, schema, kind, partition, path, inlineDataFormatForKind(kind, dataFormat), connections, withClause)
			},
		),
	)
	inlineWithoutSchema := IfValue(
		And(HiddenSyntaxTokenKindParser(SyntaxKindInlineExternalTableKeyword), HiddenSyntaxTokenKindParser(SyntaxKindKindKeyword)),
		Rule8(
			inlineKeyword(),
			inlineParameters(),
			Required(inlineKindClause, missingInlineKindClause),
			OptionalValue(inlinePartitionClause),
			OptionalValue(inlinePathFormatClause),
			OptionalValue(inlineDataFormatClause),
			Required(inlineConnectionStringsClause, missingInlineConnectionStrings),
			OptionalValue(externalDataWithClause),
			func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], kind *InlineExternalTableKindClause, partition *InlineExternalTablePartitionClause, path *InlineExternalTablePathFormatClause, dataFormat *InlineExternalTableDataFormatClause, connections *InlineExternalTableConnectionStringsClause, withClause *ExternalDataWithClause) Expression {
				schema := createMissingRowSchema()
				if isInlineDelta(kind) {
					schema = omittedInlineRowSchema()
				}
				return MustNewInlineExternalTableExpressionTyped(keyword, parameters, schema, kind, partition, path, inlineDataFormatForKind(kind, dataFormat), connections, withClause)
			},
		),
	)
	inlineExternalTableExpression := FirstValue(inlineWithSchema, inlineWithoutSchema)
	externalDataKeyword := FirstValue(
		SyntaxTokenKindParser(SyntaxKindExternalDataKeyword, NewCompletionItem(CompletionKindQueryPrefix, "externaldata")),
		HiddenSyntaxTokenKindParser(SyntaxKindExternal_DataKeyword),
	)
	externalDataExpression := Rule7(
		externalDataKeyword,
		queryGrammarParameterList(g, QueryOperatorParameters.ExternalDataWithClauseProperties, false),
		Required(rowSchema, func(Source[LexicalToken], int) *RowSchema { return createMissingRowSchema() }),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenBracketToken),
		SyntaxCommaListParser(g.UnnamedExpression, queryGrammarMissingExpression, true, true),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseBracketToken),
		OptionalValue(externalDataWithClause),
		func(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], schema *RowSchema, open *SyntaxToken, uris *SeparatedSyntaxListOf[Expression], close *SyntaxToken, withClause *ExternalDataWithClause) Expression {
			return MustNewExternalDataExpressionTyped(keyword, parameters, schema, open, uris, close, withClause)
		},
	)
	materializedExpression := Forward(func() ResultParser[LexicalToken, Expression] { return g.Expression })
	materializedNameClause := Rule3(
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		WithSyntaxCompletionHint(Required(materializedExpression, missingExpression), CompletionHintLiteral),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(open *SyntaxToken, value Expression, close *SyntaxToken) *MaterializedViewCombineNameClause {
			return MustNewMaterializedViewCombineNameClauseTyped(open, value, close)
		},
	)
	materializedClause := func(name string, hint CompletionHint, beforeText string) ResultParser[LexicalToken, *MaterializedViewCombineClause] {
		keyword := SyntaxTokenTextParser(name)
		if beforeText != "" {
			keyword = SyntaxTokenTextParser(name, SyntaxKindIdentifierToken)
			keyword = WithSyntaxCompletion(keyword, NewCompletionItem(CompletionKindSyntax, name,
				WithCompletionBeforeText(beforeText), WithCompletionAfterText(")")))
		}
		return Rule4(
			keyword,
			RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
			WithSyntaxCompletionHint(Required(materializedExpression, missingExpression), hint),
			RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
			func(keyword, open *SyntaxToken, expression Expression, close *SyntaxToken) *MaterializedViewCombineClause {
				return MustNewMaterializedViewCombineClauseTyped(keyword, open, expression, close)
			},
		)
	}
	materializedBaseClause := materializedClause("base", CompletionHintTable, "base (")
	materializedDeltaClause := materializedClause("delta", CompletionHintNonScalar, "delta (")
	summarizeAsQueryOperator := Rule1(summarizeOperator, func(operator *SummarizeOperator) QueryOperator { return operator })
	materializedAggregationsClause := Rule4(
		WithSyntaxCompletion(SyntaxTokenTextParser("aggregations", SyntaxKindIdentifierToken),
			NewCompletionItem(CompletionKindSyntax, "aggregations", WithCompletionBeforeText("aggregations (summarize "), WithCompletionAfterText(")"))),
		RequiredSyntaxTokenKindParser(SyntaxKindOpenParenToken),
		WithSyntaxCompletionHint(Required(summarizeAsQueryOperator, func(Source[LexicalToken], int) QueryOperator {
			return MustNewBadQueryOperatorTyped(createOmittedToken(SyntaxKindIdentifierToken), DiagnosticQueryOperatorExpected())
		}), CompletionHintQuery),
		RequiredSyntaxTokenKindParser(SyntaxKindCloseParenToken),
		func(keyword, open *SyntaxToken, operator QueryOperator, close *SyntaxToken) *MaterializedViewCombineClause {
			return MustNewMaterializedViewCombineClauseTyped(keyword, open, operator.(Expression), close)
		},
	)
	missingMaterializedNameClause := func(source Source[LexicalToken], start int) *MaterializedViewCombineNameClause {
		return MustNewMaterializedViewCombineNameClauseTyped(
			createOmittedToken(SyntaxKindOpenParenToken),
			queryGrammarMissingExpression(source, start),
			createOmittedToken(SyntaxKindCloseParenToken),
		)
	}
	missingMaterializedClause := func(name string) func(Source[LexicalToken], int) *MaterializedViewCombineClause {
		return func(source Source[LexicalToken], start int) *MaterializedViewCombineClause {
			return MustNewMaterializedViewCombineClauseTyped(
				createOmittedToken(SyntaxKindMaterializedViewCombineClause),
				createOmittedToken(SyntaxKindOpenParenToken),
				queryGrammarMissingExpression(source, start),
				createOmittedToken(SyntaxKindCloseParenToken),
				DiagnosticMissingMaterializedViewCombineClause(name),
			)
		}
	}
	materializedViewCombineExpression := Rule5(
		asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(
			SyntaxKindMaterializedViewCombineKeyword,
			NewCompletionItem(CompletionKindTabularPrefix, "materialized-view-combine"),
		).Hide()),
		Required(materializedNameClause, missingMaterializedNameClause),
		Required(materializedBaseClause, missingMaterializedClause("base")),
		Required(materializedDeltaClause, missingMaterializedClause("delta")),
		Required(materializedAggregationsClause, missingMaterializedClause("aggregates")),
		func(keyword *SyntaxToken, name *MaterializedViewCombineNameClause, base, delta, aggregations *MaterializedViewCombineClause) Expression {
			return MustNewMaterializedViewCombineExpressionTyped(keyword, name, base, delta, aggregations)
		},
	)
	g.tabularPrimaryExpression = FirstValue(dataTableExpression, contextualDataTableExpression, externalDataExpression, inlineExternalTableExpression, materializedViewCombineExpression)
	assertSchemaOperator := asResultParser[LexicalToken, QueryOperator](Rule2(
		HiddenSyntaxTokenKindParser(SyntaxKindAssertSchemaKeyword),
		Required(rowSchema, func(Source[LexicalToken], int) *RowSchema { return createMissingRowSchema() }),
		func(keyword *SyntaxToken, schema *RowSchema) QueryOperator {
			return MustNewAssertSchemaOperatorTyped(keyword, schema)
		},
	).Hide())

	prePipe := FirstValue(evaluateOperator, findOperator, searchOperator, unionOperator, macroExpandOperator, rangeOperator, printOperator)
	postPipe := FirstValue(
		asOperator,
		assertSchemaOperator,
		consumeOperator,
		countOperator,
		distinctOperator,
		evaluateOperator,
		executeAndCacheOperator,
		extendOperator,
		facetOperator,
		filterOperator,
		forkOperator,
		getSchemaOperator,
		graphMatchOperator,
		graphShortestPathsOperator,
		graphMarkComponentsOperator,
		asResultParser[LexicalToken, QueryOperator](graphWhereNodesOperator.Hide()),
		asResultParser[LexicalToken, QueryOperator](graphWhereEdgesOperator.Hide()),
		graphToTableOperator,
		invokeOperator,
		joinOperator,
		lookupOperator,
		makeGraphOperator,
		makeSeriesOperator,
		mvApplyOperator,
		mvExpandOperator,
		parseOperator,
		parseWhereOperator,
		parseKvOperator,
		partitionByOperator,
		partitionOperator,
		projectOperator,
		projectAwayOperator,
		projectByNamesOperator,
		projectKeepOperator,
		projectRenameOperator,
		projectReorderOperator,
		reduceByOperator,
		renderOperator,
		sampleOperator,
		sampleDistinctOperator,
		scanOperator,
		searchOperator,
		serializeOperator,
		sortOperator,
		Rule1(summarizeOperator, func(operator *SummarizeOperator) QueryOperator { return operator }),
		takeOperator,
		topHittersOperator,
		topOperator,
		topNestedOperator,
		unionOperator,
	)
	g.QueryOperator = FirstValue(prePipe, postPipe)
	badOperator := Rule1(SyntaxTokenKindParser(SyntaxKindIdentifierToken), func(keyword *SyntaxToken) QueryOperator {
		return MustNewBadQueryOperatorTyped(keyword, DiagnosticQueryOperatorExpected())
	})
	g.FollowingPipeElementExpression = FirstValue(
		postPipe,
		asResultParser[LexicalToken, QueryOperator](prePipe.Hide()),
		asResultParser[LexicalToken, QueryOperator](badOperator.Hide()),
	)

	toExpression := func(operator QueryOperator) Expression { return operator.(Expression) }
	initial := FirstValue(
		Rule1(prePipe, toExpression),
		IfValue(
			Not(g.scanSimpleName),
			asResultParser[LexicalToken, Expression](Rule1(postPipe, toExpression).Hide()),
		),
		g.UnnamedExpression,
	)
	requiredFollowing := Required(g.FollowingPipeElementExpression, func(Source[LexicalToken], int) QueryOperator {
		return MustNewBadQueryOperatorTyped(createOmittedToken(SyntaxKindIdentifierToken), DiagnosticQueryOperatorExpected())
	})
	pipe := ApplyZeroOrMore(initial, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
		return RuleRight2(
			SyntaxTokenKindParser(SyntaxKindBarToken),
			requiredFollowing,
			func(left Expression, bar *SyntaxToken, right QueryOperator) Expression {
				return MustNewPipeExpressionTyped(left, bar, right)
			},
		)
	})
	pipeSubInitial := FirstValue(
		Rule1(postPipe, toExpression),
		asResultParser[LexicalToken, Expression](initial.Hide()),
	)
	pipeSub := ApplyZeroOrMore(pipeSubInitial, func(LeftValue[Expression]) RightParser[LexicalToken, Expression] {
		return RuleRight2(
			SyntaxTokenKindParser(SyntaxKindBarToken),
			requiredFollowing,
			func(left Expression, bar *SyntaxToken, right QueryOperator) Expression {
				return MustNewPipeExpressionTyped(left, bar, right)
			},
		)
	})
	g.PipeExpression = asResultParser[LexicalToken, Expression](pipe.WithTag("<pipe-expression>"))
	g.PipeSubExpression = asResultParser[LexicalToken, Expression](pipeSub.WithTag("<pipe-sub-expression>"))
	g.Expression = asResultParser[LexicalToken, Expression](Forward(func() ResultParser[LexicalToken, Expression] {
		return g.PipeExpression
	}).WithTag("<expression>"))
}

func queryGrammarParameterList(g *QueryGrammar, declared []*QueryOperatorParameter, equalsNeeded bool, declaredOnly ...bool) ResultParser[LexicalToken, *SyntaxListOf[*NamedParameter]] {
	parameters := make([]*QueryOperatorParameter, 0, len(declared)+len(QueryOperatorParameters.AllParameters))
	seen := map[string]struct{}{}
	groups := [][]*QueryOperatorParameter{declared, QueryOperatorParameters.AllParameters}
	if len(declaredOnly) > 0 && declaredOnly[0] {
		groups = groups[:1]
	}
	for _, group := range groups {
		for _, parameter := range group {
			if parameter == nil {
				continue
			}
			name := parameter.SymbolName()
			if queryGrammarExcludesPracticalCompatibilityParameter(declared, name) {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			parameters = append(parameters, parameter)
		}
	}

	parsers := make([]ResultParser[LexicalToken, *NamedParameter], 0, len(parameters))
	for _, parameter := range parameters {
		parsers = append(parsers, queryGrammarParameter(g, parameter, equalsNeeded, queryGrammarParameterUniverse(parameter, declared)))
	}
	if len(parsers) == 0 {
		return ZeroOrMoreProduce(
			MatchValue(func(LexicalToken) bool { return false }, func(LexicalToken) *NamedParameter { return nil }),
			func([]*NamedParameter) *SyntaxListOf[*NamedParameter] {
				return AsSyntaxListOf[*NamedParameter](NewTypedSyntaxList("NamedParameter"))
			},
		)
	}
	return ZeroOrMoreProduce(FirstValue(parsers...), func(parameters []*NamedParameter) *SyntaxListOf[*NamedParameter] {
		elements := make([]SyntaxElement, len(parameters))
		for i, parameter := range parameters {
			elements[i] = parameter
		}
		return AsSyntaxListOf[*NamedParameter](NewTypedSyntaxList("NamedParameter", elements...))
	})
}

func queryGrammarParameterCommaList(g *QueryGrammar, declared []*QueryOperatorParameter, equalsNeeded bool, declaredOnly ...bool) ResultParser[LexicalToken, *SeparatedSyntaxListOf[*NamedParameter]] {
	parameters := make([]*QueryOperatorParameter, 0, len(declared)+len(QueryOperatorParameters.AllParameters))
	seen := map[string]struct{}{}
	groups := [][]*QueryOperatorParameter{declared, QueryOperatorParameters.AllParameters}
	if len(declaredOnly) > 0 && declaredOnly[0] {
		groups = groups[:1]
	}
	for _, group := range groups {
		for _, parameter := range group {
			if parameter == nil || queryGrammarExcludesPracticalCompatibilityParameter(declared, parameter.SymbolName()) {
				continue
			}
			if _, ok := seen[parameter.SymbolName()]; ok {
				continue
			}
			seen[parameter.SymbolName()] = struct{}{}
			parameters = append(parameters, parameter)
		}
	}
	parsers := make([]ResultParser[LexicalToken, *NamedParameter], 0, len(parameters))
	for _, parameter := range parameters {
		parsers = append(parsers, queryGrammarParameter(g, parameter, equalsNeeded, queryGrammarParameterUniverse(parameter, declared)))
	}
	return SyntaxCommaListParser(
		FirstValue(parsers...),
		func(Source[LexicalToken], int) *NamedParameter { return createMissingNamedParameter() },
		false,
		false,
	)
}

func queryGrammarExcludesPracticalCompatibilityParameter(declared []*QueryOperatorParameter, name string) bool {
	for _, parameter := range declared {
		if parameter != nil && parameter.SymbolName() == name {
			return false
		}
	}
	switch name {
	case "with_source_id", "with_target_id", "with_node_id":
		return true
	default:
		return false
	}
}

func queryGrammarParameterUniverse(parameter *QueryOperatorParameter, declared []*QueryOperatorParameter) []*QueryOperatorParameter {
	for _, candidate := range declared {
		if candidate != nil && candidate.SymbolName() == parameter.SymbolName() {
			return declared
		}
	}
	return QueryOperatorParameters.AllParameters
}

func queryGrammarParameter(g *QueryGrammar, parameter *QueryOperatorParameter, equalsNeeded bool, allParameters ...[]*QueryOperatorParameter) ResultParser[LexicalToken, *NamedParameter] {
	nameParser := queryGrammarParameterNameParser(parameter)
	valueParser, missingValue := queryGrammarParameterValueParser(g, parameter, allParameters...)
	hint := CompletionHintNone
	if parameter.ValueKind() == QueryOperatorParameterValueColumn || parameter.ValueKind() == QueryOperatorParameterValueColumnList {
		hint = CompletionHintColumn
	}

	var parser ResultParser[LexicalToken, *NamedParameter]
	if parameter.HasNoEquals() {
		parser = Rule2(
			nameParser,
			Required(valueParser, missingValue),
			func(name *SyntaxToken, value Expression) *NamedParameter {
				declaration := MustNewNameDeclarationTyped(MustNewTokenNameTyped(name))
				return MustNewNamedParameterTyped(declaration, createOmittedToken(SyntaxKindEqualToken), value, hint)
			},
		)
	} else {
		parser = Rule3(
			nameParser,
			RequiredSyntaxTokenKindParser(SyntaxKindEqualToken),
			Required(valueParser, missingValue),
			func(name, equal *SyntaxToken, value Expression) *NamedParameter {
				declaration := MustNewNameDeclarationTyped(MustNewTokenNameTyped(name))
				return MustNewNamedParameterTyped(declaration, equal, value, hint)
			},
		)
		if equalsNeeded {
			parser = IfValue(And(nameParser, SyntaxTokenKindParser(SyntaxKindEqualToken)), parser)
		}
	}
	return parser
}

func queryGrammarParameterNameParser(parameter *QueryOperatorParameter) ResultParser[LexicalToken, *SyntaxToken] {
	name := parameter.SymbolName()
	var primary ResultParser[LexicalToken, *SyntaxToken]
	if parameter.IsHidden() {
		primary = HiddenSyntaxTokenTextParser(name)
	} else if parameter.HasNoEquals() {
		primary = SyntaxTokenTextParser(name)
	} else {
		applyText := name + "=|"
		if parameter.ValueKind() == QueryOperatorParameterValueStringLiteral {
			applyText = name + "=\"|\""
		}
		base := asResultParser[LexicalToken, *SyntaxToken](SyntaxMatchTextParser(name).WithTag(defaultSyntaxParserTagForText(name)))
		primary = WithSyntaxCompletion(base, NewCompletionItem(
			syntaxCompletionKindForText(name),
			name,
			WithCompletionApplyTextMarkers(applyText),
		))
	}
	parsers := []ResultParser[LexicalToken, *SyntaxToken]{primary}
	for _, alias := range parameter.Aliases() {
		parsers = append(parsers, HiddenSyntaxTokenTextParser(alias))
	}
	if len(parsers) == 1 {
		return parsers[0]
	}
	return FirstValue(parsers...)
}

func queryGrammarParameterValueParser(g *QueryGrammar, parameter *QueryOperatorParameter, allParameters ...[]*QueryOperatorParameter) (ResultParser[LexicalToken, Expression], func(Source[LexicalToken], int) Expression) {
	tokenLiteral := Rule1(MatchValue(func(token LexicalToken) bool {
		category := SyntaxCategoryOf(token.Kind)
		return category == SyntaxCategoryIdentifier || category == SyntaxCategoryKeyword
	}, func(token LexicalToken) *SyntaxToken { return SyntaxTokenFromLexical(token) }), func(token *SyntaxToken) Expression {
		return MustNewLiteralExpressionTyped(SyntaxKindTokenLiteralExpression, token)
	})
	anyValue := FirstValue(g.Literal, tokenLiteral, g.SimpleNameReference)
	missing := func(Source[LexicalToken], int) Expression { return createMissingValueExpression() }

	switch parameter.ValueKind() {
	case QueryOperatorParameterValueBoolLiteral:
		boolean := Rule1(SyntaxTokenKindParser(SyntaxKindBooleanLiteralToken), func(token *SyntaxToken) Expression {
			return MustNewLiteralExpressionTyped(SyntaxKindBooleanLiteralExpression, token)
		})
		return FirstValue(boolean, anyValue), func(Source[LexicalToken], int) Expression { return createMissingBoolLiteral() }
	case QueryOperatorParameterValueStringLiteral:
		return anyValue, func(Source[LexicalToken], int) Expression { return createMissingStringLiteral() }
	case QueryOperatorParameterValueIntegerLiteral, QueryOperatorParameterValueNumericLiteral, QueryOperatorParameterValueSummableLiteral:
		return anyValue, func(Source[LexicalToken], int) Expression { return createMissingLongLiteral() }
	case QueryOperatorParameterValueForcedRealLiteral:
		unsigned := Rule1(MatchValue(func(token LexicalToken) bool {
			switch token.Kind {
			case SyntaxKindLongLiteralToken, SyntaxKindIntLiteralToken, SyntaxKindRealLiteralToken, SyntaxKindDecimalLiteralToken:
				return true
			default:
				return false
			}
		}, func(token LexicalToken) *SyntaxToken { return SyntaxTokenFromLexical(token) }), func(token *SyntaxToken) Expression {
			return MustNewLiteralExpressionTyped(SyntaxKindRealLiteralExpression, token)
		})
		signed := IfValue(
			And(
				Or(SyntaxTokenKindParser(SyntaxKindMinusToken), SyntaxTokenKindParser(SyntaxKindPlusToken)),
				Match(func(token LexicalToken) bool {
					return token.Trivia == "" && (token.Kind == SyntaxKindLongLiteralToken || token.Kind == SyntaxKindRealLiteralToken)
				}),
			),
			Rule2(
				FirstValue(SyntaxTokenKindParser(SyntaxKindMinusToken), SyntaxTokenKindParser(SyntaxKindPlusToken)),
				FirstValue(SyntaxTokenKindParser(SyntaxKindLongLiteralToken), SyntaxTokenKindParser(SyntaxKindRealLiteralToken)),
				func(sign, number *SyntaxToken) Expression {
					combined := NewLexicalToken(number.Kind(), sign.Trivia(), sign.RawText()+number.RawText())
					return MustNewLiteralExpressionTyped(SyntaxKindRealLiteralExpression, SyntaxTokenFromLexical(combined))
				},
			),
		)
		return FirstValue(unsigned, signed, anyValue), func(Source[LexicalToken], int) Expression { return createMissingRealLiteral() }
	case QueryOperatorParameterValueString:
		return g.functionCallOrPath, missing
	case QueryOperatorParameterValueWord, QueryOperatorParameterValueWordOrNumber:
		values := parameter.Values()
		if len(values) > 0 {
			allowedTokens := MatchValue(func(token LexicalToken) bool {
				for _, value := range values {
					if (parameter.IsCaseSensitive() && token.Text == value) || (!parameter.IsCaseSensitive() && ordinalEqualFold(token.Text, value)) {
						return true
					}
				}
				return false
			}, func(token LexicalToken) *SyntaxToken { return SyntaxTokenFromLexical(token) })
			items := make([]CompletionItem, 0, len(values))
			for _, value := range values {
				items = append(items, NewCompletionItem(CompletionKindSyntax, value))
			}
			allowed := Rule1(WithSyntaxCompletion(allowedTokens, items...), func(token *SyntaxToken) Expression {
				return MustNewLiteralExpressionTyped(SyntaxKindTokenLiteralExpression, token)
			})
			return FirstValue(allowed, anyValue), func(Source[LexicalToken], int) Expression { return createMissingTokenLiteral(values...) }
		}
		return anyValue, func(Source[LexicalToken], int) Expression { return createMissingTokenLiteral("token") }
	case QueryOperatorParameterValueNameDeclaration:
		return FirstValue(g.SimpleNameDeclarationExpression, anyValue), func(Source[LexicalToken], int) Expression {
			return createMissingNameDeclaration()
		}
	case QueryOperatorParameterValueColumn:
		return FirstValue(g.SimpleNameReference, anyValue), func(Source[LexicalToken], int) Expression { return createMissingNameReferenceNode() }
	case QueryOperatorParameterValueColumnList:
		boundaryNames := []string{}
		if len(allParameters) > 0 {
			for _, candidate := range allParameters[0] {
				if candidate != nil {
					boundaryNames = append(boundaryNames, candidate.SymbolName())
				}
			}
		}
		nameList := queryGrammarNameReferenceList(queryGrammarExtendedNameReferenceParser(g), boundaryNames)
		return FirstValue(nameList, anyValue), func(Source[LexicalToken], int) Expression {
			return MustNewNameReferenceListTyped(MakeSyntaxSeparatedList[*NameReference]([]any{
				createMissingNameReferenceNode(),
			}))
		}
	default:
		return anyValue, missing
	}
}

func queryGrammarExtendedNameReferenceParser(g *QueryGrammar) ResultParser[LexicalToken, *NameReference] {
	simple := Rule1(g.SimpleNameReference, func(expression Expression) *NameReference { return expression.(*NameReference) })
	extendedKeyword := Rule1(MatchValue(func(token LexicalToken) bool {
		return isExtendedKeywordAsIdentifier(token.Kind)
	}, func(token LexicalToken) *SyntaxToken { return SyntaxTokenFromLexical(token) }), func(token *SyntaxToken) *NameReference {
		return MustNewNameReferenceTyped(MustNewTokenNameTyped(token), SymbolMatchNone)
	})
	return asResultParser[LexicalToken, *NameReference](FirstValue(simple, extendedKeyword).WithTag("<name>"))
}

func queryGrammarNameReferenceList(nameParser ResultParser[LexicalToken, *NameReference], boundaryNames []string) ResultParser[LexicalToken, Expression] {
	comma := asResultParser[LexicalToken, *SyntaxToken](SyntaxTokenKindParser(SyntaxKindCommaToken))
	if len(boundaryNames) > 0 {
		comma = IfValue(
			Not(And(
				SyntaxTokenKindParser(SyntaxKindCommaToken),
				SyntaxTokenTextsParser(boundaryNames),
				SyntaxTokenKindParser(SyntaxKindEqualToken),
			)),
			comma,
		)
	}
	list := OList(
		nameParser,
		comma,
		func(Source[LexicalToken], int) *NameReference {
			return createMissingNameReferenceNode()
		},
		nil,
		EndOfCommaListParser(),
		true,
		false,
		func(elements []any) *SeparatedSyntaxListOf[*NameReference] {
			return MakeSyntaxSeparatedList[*NameReference](elements)
		},
	)
	return Rule1(list, func(names *SeparatedSyntaxListOf[*NameReference]) Expression {
		return MustNewNameReferenceListTyped(names)
	})
}

func queryGrammarMissingNameDeclaration(source Source[LexicalToken], start int) *NameDeclaration {
	diagnostic := DiagnosticMissingName()
	if source != nil && !source.IsEnd(start) && IsKeywordKind(source.Peek(start).Kind) {
		diagnostic = DiagnosticMissingNameWithKeyword(source.Peek(start).Text)
	}
	return MustNewNameDeclarationTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), diagnostic)
}
